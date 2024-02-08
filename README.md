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

Rename the x86 dll to WinDivert32.dll
Rename the x64 dll to WinDivert64.dll

`go build`

Additionally you can install the program by adding the folder to your powershell path. 
This will allow you to call with just geneva-cli instead of .\geneva-cli.exe from anywhere. 

Any filepaths you paths should be absolute paths, the working directory will be changed to the executable directory.

## How to run
First you will need a valid Geneva strategy, one is included in s.txt.

Then you can run the program using

`.\geneva-cli.exe intercept --interface <interface-name> -strategyFile .\s.txt`

You can find a list of available interfaces using

`.\geneva-cli.exe list-adapters`

## How to test

You can test or rather verify that the intercept mode is running by simply running it with a validated strategy. The output should show that packets are being rerouted. You can monitor your adapter in [WireShark](https://www.wireshark.org/) and you should notice a huge uptick in packets being sent from your adapter.

You can predict what the wireshark output should look like using the [Geneva documentation](https://github.com/getlantern/geneva?tab=readme-ov-file#strategies-forests-and-action-trees), each strategy will look different. The default strategy will take all incoming tcp packets and split them in half, on wireshark you should twice as many outgoing packets

You can run `go test -c; .\geneva-cli.test.exe` to run unit tests

## How to install as service

Open PowerShell in administrator mode before running any commands

### Install Command, in manual mode

```.\geneva-cli.exe intercept --service install```

### Start

```.\geneva-cli.exe intercept --service start --strategyFile <path_to_file>```
You can specify a strategy, strategy file, a saved command, or nothing to let it run from the default strategy (`s.txt`)
While running the service you can check [Event Viewer](https://learn.microsoft.com/en-us/shows/inside/event-viewer) for output

### Stop

```.\geneva-cli.exe intercept --service stop```

### Uninstall

```.\geneva-cli.exe intercept --service uninstall```

## Notes

This was tested on a 64-bit Windows 10 machine
Not tested on WSL
`strategies.txt` contains several validated strategies

## Help Output
```
NAME:
   geneva - Genetic Evasion for windows

USAGE:
   geneva-cli.exe [global options] command [command options] [arguments...]

COMMANDS:

   dot            (unavailable on windows) output the strategy graph 
   as an SVG
   intercept      Run a strategy on live network traffic
   list-adapters  Lists the available adapters
   run-pcap       Run a PCAP file through a strategy and output the resulting packets in a new PCAP
   saved-command  Runs commands from config file
   validate       validate that a strategy is well-formed
   help, h        Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)```