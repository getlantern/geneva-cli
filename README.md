# Windows based Geneva-cli

This is a Geneva-based WinDivert tunnel. Built in Go 1.17

## Prerequistes

- Be on a Wiondows machine
- Have Popwershell installed
- Run Powershell as admin mode
- All commands are ran in the project directory
- Go 1.17 is installed

## How to Build

First download windivert 2.2 [here](https://www.reqrypt.org/windivert.html) and then extract the following files and place them in your geneva-cli folder.
- x86/WinDivert32.sys
- x86/WinDivert.dll
- x64/WinDivert64.sys
- x64/WinDivert.dll

Rename the dll's to WinDivert32.dll and WinDivert64.dll respectively.

`go build`

## How to run

First you will need a valid Geneva strategy, one is included in s.txt.

Then you can run the program using

`.\geneva-cli.exe intercept --interface <interface-name> -strategyFile .\s.txt`

You can find a list of available interfaces using

`.\geneva-cli.exe list-adapters`

## Notes

This was tested on a 64-bit Windows 10 machine, if you run this on Windows 11 or on 32-bit please let us know how it goes.

If you run it under LSW I cannot gaurantee it will run, if it does please let us know.

## Help Output
```
NAME:
   geneva - A new cli application

USAGE:
   geneva-cli.exe [global options] command [command options] [arguments...]

COMMANDS:

   dot            (unavailable on windows) output the strategy graph 
   as an SVG
   intercept      Run a strategy on live network traffic
   list-adapters  Lists the available adapters
   run-pcap       Run a PCAP file through a strategy and output the resulting packets in a new PCAP
   saved-command  Runs commands from from config file
   validate       validate that a strategy is well-formed
   help, h        Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)```