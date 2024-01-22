package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"github.com/Crosse/godivert"
	"github.com/getlantern/geneva/strategy"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/urfave/cli/v2"
	"golang.org/x/sys/windows"
)

var (
	qpc      uintptr
	iphlpapi *windows.LazyDLL
)

func init() {
	dll := windows.NewLazySystemDLL("kernel32.dll")
	if err := dll.Load(); err != nil {
		panic(fmt.Errorf("error loading kernel32.dll: %v", err))
	}

	_qpc := dll.NewProc("QueryPerformanceCounter")
	if err := _qpc.Find(); err != nil {
		panic(fmt.Errorf("error finding QueryPerformanceCounter: %v", err))
	}

	qpc = _qpc.Addr()

	iphlpapi = windows.NewLazySystemDLL("Iphlpapi.dll")
	if err := iphlpapi.Load(); err != nil {
		panic(fmt.Errorf("error loading Iphlpapi.dll"))
	}
}

func NewInterceptor(iface string) (Interceptor, error) {
	return &interceptor{
		quit:  make(chan struct{}),
		iface: iface,
	}, nil
}

func (p *interceptor) Intercept() error {
	go func() {
		logger.Info(http.ListenAndServe("localhost:6060", nil))
	}()

	filter := "ip and !loopback"
	if p.iface != "" {
		idx, err := getAdapter(p.iface)
		if err != nil {
			return err
		}
		filter = fmt.Sprintf("%s and ifIdx == %d", filter, idx)

		logger.Infof("intercepting traffic on %s (idx %d)\n", p.iface, idx)
	} else {
		logger.Info("intercepting traffic on all interfaces")
	}

	ips := make([]string, 0, len(p.ips))
	for _, ip := range p.ips {
		if ip == nil {
			continue
		}

		ips = append(ips, fmt.Sprintf("remoteAddr == %s", ip))
	}

	if len(ips) > 0 {
		ipFilter := strings.Join(ips, " or ")
		if filter == "" {
			filter = ipFilter
		} else {
			filter = fmt.Sprintf("%s and (%s)", filter, ipFilter)
		}
	}

	logger.Infof("using filter %q\n", filter)

	ex, err := os.Executable()
	if err != nil {
		return fmt.Errorf("Can't locate dll")
	}
	exPath := filepath.Dir(ex)
	dll64 := filepath.Join(exPath, "WinDivert64.dll")
	dll32 := filepath.Join(exPath, "WinDivert32.dll")

	logger.Info("opening handle to WinDivert")
	godivert.LoadDLL(dll64, dll32)

	winDivert, err := godivert.OpenHandle(
		filter,
		godivert.LayerNetwork,
		godivert.PriorityDefault,
		godivert.OpenFlagFragments,
	)
	if err != nil {
		return fmt.Errorf("error initializing WinDivert: %v", err)
	}

	defer func() {
		logger.Info("closing WinDivert handle")

		if err := winDivert.Close(); err != nil {
			logger.Infof("error closing WinDivert handle: %v\n", err)
		}
	}()

	packetChan, err := winDivert.Packets()
	if err != nil {
		return fmt.Errorf("error getting packets: %v\n", err)
	}

	for {
		select {
		case pkt := <-packetChan:
			if err = p.processPacket(winDivert, pkt); err != nil {
				logger.Error(err)
			}
		case <-p.quit:
			return nil
		}
	}
}

func (p *interceptor) processPacket(winDivert *godivert.WinDivertHandle, pkt *godivert.Packet) error {
	pkt.VerifyParsed()

	var dir strategy.Direction
	if pkt.Direction() == godivert.WinDivertDirectionInbound {
		dir = strategy.DirectionInbound
	} else {
		dir = strategy.DirectionOutbound
	}

	p.statistics.Increment(Intercepted, dir)

	var firstLayer gopacket.LayerType

	switch pkt.IpVersion() {
	case 4:
		firstLayer = layers.LayerTypeIPv4
	case 6:
		firstLayer = layers.LayerTypeIPv6
	default:
		logger.Info("bypassing Geneva for non-IP packet")
		p.statistics.Increment(Injected, dir)
		return sendPacket(winDivert, pkt, dir, &p.statistics)
	}

	gopkt := gopacket.NewPacket(pkt.Raw, firstLayer, gopacket.Default)

	results, err := p.strategy.Apply(gopkt, dir)
	if err != nil {
		p.statistics.Increment(Errors, dir)
		p.statistics.Increment(Injected, dir)
		if err2 := sendPacket(winDivert, pkt, dir, &p.statistics); err != nil {
			return fmt.Errorf("failed to send packet after error applying strategy: %w", err2)
		}

		return fmt.Errorf("error applying strategy: %v\n", err)
	}

	for i, packet := range results {
		newPkt := godivert.Packet{
			Raw: packet.Data(),
			Addr: &godivert.WinDivertAddress{
				Timestamp: pkt.Addr.Timestamp + int64(i),
				Flags:     pkt.Addr.Flags,
				Data:      pkt.Addr.Data,
			},
			PacketLen: uint(len(packet.Data())),
		}

		newPkt.VerifyParsed()

		if err = sendPacket(winDivert, &newPkt, dir, &p.statistics); err != nil {
			return err
		}
	}

	return nil
}

func sendPacket(handle *godivert.WinDivertHandle, pkt *godivert.Packet, dir strategy.Direction, stats *Statistics) error {
	if sent, err := handle.Send(pkt); err != nil {
		stats.Increment(Errors, dir)
		return fmt.Errorf("error sending packet: %v", err)
	} else if sent != pkt.PacketLen {
		stats.Increment(Errors, dir)
		return fmt.Errorf("sent %d bytes, but expected %d", sent, pkt.PacketLen)
	}

	stats.Increment(Injected, dir)
	return nil
}

func now() int64 {
	var now uint64

	_, _, _ = syscall.Syscall(qpc, 1, uintptr(unsafe.Pointer(&now)), 0, 0)

	return int64(now)
}

func getAdapter(iface string) (uint32, error) {
	// https://docs.microsoft.com/en-us/windows/win32/api/iphlpapi/nf-iphlpapi-getadaptersaddresses
	// "The recommended method of calling the GetAi daptersAddresses function is to pre-allocate a
	// 15KB working buffer pointed to by the AdapterAddresses parameter. On typical computers,
	// this dramatically reduces the chances that the GetAdaptersAddresses function returns
	// ERROR_BUFFER_OVERFLOW, which would require calling GetAdaptersAddresses function multiple
	// times."
	bufLenBytes := 15 * 1024
	info := make(
		[]windows.IpAdapterAddresses,
		bufLenBytes/int(unsafe.Sizeof(windows.IpAdapterAddresses{})),
	)
	ol := uint32(bufLenBytes)

	err := windows.GetAdaptersAddresses(windows.AF_UNSPEC, 0, 0, &info[0], &ol)
	if err != nil {
		return 0, err
	}

	a := &info[0]
	for a != nil {
		if windows.BytePtrToString(a.AdapterName) == iface ||
			windows.UTF16PtrToString(a.FriendlyName) == iface {
			return a.IfIndex, nil
		}

		a = a.Next
	}

	return 0, fmt.Errorf("no adapter found")
}

func listAdaptersWrapper(c *cli.Context) error {
	return ListAdapters()
}

func ListAdapters() error {
	bufLenBytes := 15 * 1024
	info := make(
		[]windows.IpAdapterAddresses,
		bufLenBytes/int(unsafe.Sizeof(windows.IpAdapterAddresses{})),
	)
	ol := uint32(bufLenBytes)

	err := windows.GetAdaptersAddresses(windows.AF_UNSPEC, 0, 0, &info[0], &ol)
	if err != nil {
		return err
	}

	fmt.Println("Found Adapters")
	a := &info[0]
	for a != nil {
		fmt.Printf("%s, id=%v\n", windows.UTF16PtrToString(a.FriendlyName), a.IfIndex)
		fmt.Printf("\t%s\n", windows.BytePtrToString(a.AdapterName))

		a = a.Next
	}

	return nil

}
