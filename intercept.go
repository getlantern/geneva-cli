package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/getlantern/common"
	"github.com/getlantern/geneva"
	"github.com/getlantern/geneva/strategy"
	"github.com/getlantern/rot13"
	"github.com/getlantern/yaml"
	"github.com/kardianos/service"
	"github.com/urfave/cli/v2"
)

type myLog struct {
	svcLogger     service.Logger
	genericLogger *log.Logger
	logFile       *os.File
}

func newLogger(svc service.Service) *myLog {
	svcLogger, err := svc.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	logFile, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		log.Fatal(err)
	}
	fileLogger := log.Default()
	fileLogger.SetOutput(logFile)

	return &myLog{svcLogger, fileLogger, logFile}
}

func (l *myLog) Info(v ...interface{}) {
	l.svcLogger.Info(v...)
	l.genericLogger.Print(v...)
}

func (l *myLog) Infof(format string, a ...interface{}) {
	l.svcLogger.Infof(format, a...)
	l.genericLogger.Printf(format, a...)
}

func (l *myLog) Warning(v ...interface{}) {
	l.svcLogger.Warning(v...)
	l.genericLogger.Print(v...)
}

func (l *myLog) Warningf(format string, a ...interface{}) {
	l.svcLogger.Warningf(format, a...)
	l.genericLogger.Printf(format, a...)
}

func (l *myLog) Error(v ...interface{}) {
	l.svcLogger.Error(v...)
	l.genericLogger.Print(v...)
}

func (l *myLog) Errorf(format string, a ...interface{}) {
	l.svcLogger.Errorf(format, a...)
	l.genericLogger.Printf(format, a...)
}

var logger *myLog

const (
	LOG_FILE                = "/geneva-proxy.log"
	STATISTICS_INTERVAL_CLI = 5
	STATISTICS_INTERVAL_SVC = 300
)

type StatisticsType int

const (
	Intercepted = iota
	Injected
	Errors
)

type Statistics struct {
	InterceptedPackets [2]uint64
	InjectedPackets    [2]uint64
	Errors             uint64
}

func (s *Statistics) Increment(st StatisticsType, dir strategy.Direction) {
	switch st {
	case Intercepted:
		s.InterceptedPackets[dir]++
	case Injected:
		s.InjectedPackets[dir]++
	case Errors:
		s.Errors++
	}
}

func (s *Statistics) Print() {
	logger.Infof("intercepted/injected packets: %d/%d inbound, %d/%d outbound; %d errors\n",
		s.InterceptedPackets[strategy.DirectionInbound],
		s.InjectedPackets[strategy.DirectionInbound],
		s.InterceptedPackets[strategy.DirectionOutbound],
		s.InjectedPackets[strategy.DirectionOutbound],
		s.Errors,
	)
}

type Interceptor interface {
	Intercept() error
	SetProxyIPs([]net.IP)
	SetStrategy(strategy string) error
	Strategy() strategy.Strategy

	service.Interface
}

type interceptor struct {
	wg   sync.WaitGroup
	quit chan struct{}

	strategy   *strategy.Strategy
	iface      string
	ips        []net.IP
	statistics Statistics
}

func (i *interceptor) SetProxyIPs(ips []net.IP) {
	i.ips = ips
}

func (i *interceptor) SetStrategy(strat string) error {
	s, err := geneva.NewStrategy(strat)
	if err != nil {
		return err
	}
	i.strategy = s
	return nil
}

func (i *interceptor) Strategy() strategy.Strategy {
	return *i.strategy
}

func (i *interceptor) Start(s service.Service) error {
	logger.Info("starting up")

	i.wg.Add(1)
	go func() {
		logger.Info("starting statistics thread")
		period := STATISTICS_INTERVAL_SVC * time.Second
		if service.Interactive() {
			period = STATISTICS_INTERVAL_CLI * time.Second
		}
		timer := time.NewTicker(period)
		for {
			select {
			case <-i.quit:
				i.wg.Done()
				return
			case <-timer.C:
				i.statistics.Print()
			}
		}
	}()

	i.wg.Add(1)
	go func() {
		if err := i.Intercept(); err != nil {
			logger.Error(err)
		}

		i.wg.Done()

		if !service.Interactive() {
			s.Stop()
		}
	}()

	return nil
}

func (i *interceptor) Stop(s service.Service) error {
	logger.Info("stopping all threads")
	close(i.quit)
	i.wg.Wait()
	i.statistics.Print()

	logger.logFile.Close()

	if service.Interactive() {
		os.Exit(0)
	}

	return nil
}

func init() {
	app.Commands = append(app.Commands,
		&cli.Command{
			Name:  "intercept",
			Usage: "Run a strategy on live network traffic",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "strategy",
					Aliases: []string{"s"},
					Usage:   "A Geneva `STRATEGY` to run",
				},
				&cli.StringFlag{
					Name:  "strategyFile",
					Usage: "Load Geneva strategy from `FILE`",
				},
				&cli.StringFlag{
					Name:  "fromFlashlight",
					Usage: "Load strategy and endpoint information from a given flashlight proxies.yaml file (overrides other options)",
				},
				&cli.StringFlag{
					Name:    "interface",
					Aliases: []string{"i"},
					Usage:   "Network interface on which to intercept traffic",
				},
				&cli.StringFlag{
					Name:  "ips",
					Usage: "comma-separated list of IP:port tuples to proxy",
				},
				&cli.StringFlag{
					Name:  "service",
					Usage: "Control the system service",
				},
				&cli.StringFlag{
					Name:  "saved-command",
					Usage: "Run a named command from saved_command.json, this overwrites all other options",
				},
			},
			Action: intercept,
		})

}

func intercept(c *cli.Context) error {
	interceptor, err := NewInterceptor(c.String("interface"))
	if err != nil {
		return err
	}

	s_file := filepath.Join(getExecPath(), "s.txt")

	svcConfig := service.Config{
		Name:        "geneva-proxy",
		DisplayName: "Geneva proxy",
		Description: "Geneva proxy",
		Arguments:   []string{"intercept", "--strategyFile", "s.txt"},
	}

	serviceArgs := []string{}
	for _, v := range c.FlagNames() {
		fmt.Println(v)
		if v != "service" {
			serviceArgs = append(serviceArgs, v)
			serviceArgs = append(serviceArgs, c.String(v))
		}
	}

	cmd := c.String("service")
	fmt.Println(serviceArgs)
	if len(serviceArgs) > 0 && cmd != "" {
		svcConfig.Arguments = serviceArgs
	}

	if saved := c.String("saved-command"); saved != "" {
		sc, err := getSavedCommands("saved_commands.json")
		if err != nil {
			return cli.Exit(err, 1)
		}

		com, err := sc.getCommand(saved)
		if err != nil {
			return cli.Exit(err, 1)
		}
		if com.CmdStr != "intercept" {
			return cli.Exit(errors.New("only run intercept command as service"), 1)
		}
		svcConfig.Arguments = com.Args
	}

	svc, err := service.New(interceptor, &svcConfig)
	if err != nil {
		return cli.Exit(err, 1)
	}
	logger = newLogger(svc)

	if cmd != "" {
		if err := service.Control(svc, cmd); err != nil {
			logger.Errorf("Valid actions: %s\n", service.ControlAction)
			return cli.Exit(err, 1)
		}
		return nil
	}

	logger.Infof("Uh: %s", s_file)

	var ips []net.IP
	var strat string

	flashPath := c.String("fromFlashlight")

	if flashPath != "" {
		ips, strat, err = parseProxyFile(flashPath)
		if err != nil {
			logger.Errorf("error parsing proxies.yaml: %v\n", err)
			return cli.Exit(err, 1)
		}
		interceptor.SetProxyIPs(ips)

		logger.Info("parsed proxies.yaml\n")

	}

	// command-line options can override some of flashlight's config
	if c.String("strategy") != "" && c.String("strategyFile") != "" {
		return cli.Exit("only one of -strategy and -strategyFile should be used", 1)
	}

	s := c.String("strategy")
	if s == "" {
		strategyFile := c.String("strategyFile")
		if strategyFile == "" {
			errStr := "must provide one of -strategy or -strategyFile"
			logger.Error(errStr)
			return cli.Exit(errStr, 1)
		}

		in, err := os.ReadFile(strategyFile)
		if err != nil {

			errStr := fmt.Sprintf("cannot open %s: %v", strategyFile, err)

			logger.Error(errStr)
			return cli.Exit(errStr, 1)
		}
		s = string(in)
	}

	if s != "" {
		if strat != "" && flashPath != "" {
			logger.Info("overriding strategy from proxies.yaml with command-line argument")
		}
		strat = s
	}

	interceptor.SetStrategy(strat)
	if err != nil {
		logger.Errorf("Can't set strat %s", strat)
		return cli.Exit(err, 1)
	}

	ipsFromArgs := strings.Split(c.String("ips"), ",")
	if len(ipsFromArgs) > 0 {
		if flashPath != "" {
			logger.Info("appending IPs from command line to list from proxies.yaml")
		}

		for _, ip := range ipsFromArgs {
			ips = append(ips, net.ParseIP(strings.Split(ip, ":")[0]))
		}
		interceptor.SetProxyIPs(ips)
	}

	logger.Infof("outbound \\/ inbound %q", interceptor.Strategy())

	err = svc.Run()

	return err
}

func parseProxyFile(proxiesFilepath string) ([]net.IP, string, error) {
	ips := make([]net.IP, 0, 4)

	// get the values from proxies.yaml
	proxiesFile, err := os.Open(proxiesFilepath)

	if err != nil {
		return nil, "", cli.Exit(fmt.Sprintf("cannot open proxies.yaml: %v", err), 1)
	}
	defer proxiesFile.Close()

	wrapped := rot13.NewReader(proxiesFile)
	data, err := io.ReadAll(wrapped)
	if err != nil {
		return nil, "", cli.Exit(fmt.Sprintf("cannot read proxies.yaml: %v", err), 1)
	}

	config := make(map[string]common.ChainedServerInfo, 2)
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, "", cli.Exit(fmt.Sprintf("cannot parse proxies.yaml: %v", err), 1)
	}

	for _, v := range config {
		switch v.Addr {
		case "multiplexed":
			ips = append(ips, net.ParseIP(v.MultiplexedAddr))
		default:
			logger.Warningf("unhandled transport %q\n", v.Addr)
		}
	}

	// TODO: actually get strategy from proxies.yaml
	// s, err := strategy.ParseStrategy(strat)
	// if err != nil {
	// 	return nil, "", err
	// }

	return ips, "", nil
}
