# Windows based Geneva-cli

This is a Geneva-based WinDivert tunnel, it takes a Geneva strategy and utilizes WinDivert to capture packets, modify them, and re-inject the packet to avoid censorship. One would use this tool when trying to evade censorship and can be run along side Lantern VPN to further obscure traffic.

[Geneva](https://geneva.cs.umd.edu/) is a genetic algorithm based solution to censorship evasion.

[WinDivert](https://www.reqrypt.org/windivert.html) is a user-mode packet capture-and-divert package for Windows 10, Windows 11, and Windows Server.

## Prerequisites

- Be on a Windows machine
- Have Powershell installed
- Run Powershell as admin mode
- All commands are ran in the project directory
- Go 1.20 is installed

## How to Build

First download WinDivert 2.2 [here](https://www.reqrypt.org/windivert.html) and then extract the following files and place them directly in your geneva-cli folder, the dlls will need to be renamed.
- x86/WinDivert32.sys
- x86/WinDivert.dll
- x64/WinDivert64.sys
- x64/WinDivert.dll

Rename the x86 dll to WinDivert32.dll and the x64 dll to WinDivert64.dll.

`go build`

Additionally you can install the program by adding the folder to your powershell path. 
This will allow you to call with just geneva-cli instead of .\geneva-cli.exe from anywhere. 

Any filepaths you paths should be absolute paths, the working directory will be changed to the executable directory.

## How to run
First you will need a valid Geneva strategy, an straightforward one is included in s.txt. Strategies.txt contains many strategies for you to use.

Then you can run the program using

`.\geneva-cli.exe intercept --interface <interface-name> -strategyFile .\s.txt`

or

`.\geneva-cli.exe intercept --interface <interface-name> -strategy "[TCP:flags:PA]-fragment{tcp:-1:True}-| \/"`

You can find a list of available interfaces using

`.\geneva-cli.exe list-adapters`

### How to find strategies
You can find a list of strategies in strategies.txt or in the [Geneva paper](http://geneva.cs.umd.edu/papers/geneva_ccs19.pdf) on page 9 with tested success rates for China, India, and Kazakhstan.

## How to test

You can test or rather verify that the intercept mode is running by simply running it with a validated strategy. The output should show that packets are being rerouted. You can monitor your adapter in [WireShark](https://www.wireshark.org/) and you should notice a huge uptick in packets being sent from your adapter.

You can predict what the Wireshark output should look like using the [Geneva documentation](https://github.com/getlantern/geneva?tab=readme-ov-file#strategies-forests-and-action-trees), each strategy will look different. If you take the following strategy and use Wireshark or the included pcap function you will notice that outbound [PSH, ACK] packets will be split into two packets. The pcap feature will show exactly which packets split and into how many.

```[TCP:flags:PA]-fragment{tcp:5:True}-| \/```

If you are correlating the include pcap output with Wireshark, be aware that Wireshark numbers packets start at 1 and not 0.

A simple Python 3 script, bulk_pcap.py, has been included to generate output pcap files from strategies.txt. You will need to obtain your own input pcap file using Wireshark's capture feature. The output will be placed in testdata. Tested with Python 3.12

bulk_pcap.py takes a single argument, your input pcap file.

You can run `go test -v` to run unit tests.

## How to install as service

Open PowerShell in administrator mode before running any commands

### Install Command, in manual mode

```.\geneva-cli.exe intercept --service install --strategyFile <path_to_file> --adapter <adapter_name>```

AT this time you must specify the parameters at service installation.

You can specify a strategy, strategy file, a saved command, or nothing to let it run from the default strategy (`s.txt`)
While running the service you can check [Event Viewer](https://learn.microsoft.com/en-us/shows/inside/event-viewer) for output.

You can obtain a list of adapters using the list-adapters command

### Start
```.\geneva-cli.exe intercept --service start```

### Stop

```.\geneva-cli.exe intercept --service stop```

### Uninstall

```.\geneva-cli.exe intercept --service uninstall```

## Notes

- This was tested on a 64-bit Windows 10 machine
- Not tested on WSL
- `strategies.txt` contains several validated strategies
- Currently only IPv4 and TCP are supported, this may change if the core geneva library adds support for more protocols
- strategies can trigger on the reserved flags in a TCP packet but this is not well tested and not recommended, modifying these flags is not supported either

## Help Output
```
NAME:
   geneva - Genetic Evasion for windows

USAGE:
   geneva-cli.exe [global options] command [command options] [arguments...]

COMMANDS:
   dot                  (unavailable on windows) output the strategy graph as an SVG
   intercept            Run a strategy on live network traffic
   list-adapters        Lists the available adapters
   list-saved-commands  Lists the saved-commands
   load-command         Runs commands from config file
   run-pcap             Run a PCAP file through a strategy and output the resulting packets in a new PCAP
   validate             validate that a strategy is well-formed
   help, h              Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)

GLOBAL OPTIONS:
   --help, -h  show help (default: false)```