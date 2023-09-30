# packer-plugin-wim

## General overview

A packer plugin to support creation of WIM image from VHD/VHDx files. Main goal is to support building of WIM image on Windows and Linux platform.

This project use cgo and C under the hood to create WIM images for cross platform and better performance.

## Build

To build the plugin run simply `go build -x ./main.go`. For additional important information check below.

## Important notes

- To use this plugin, a require DLL file from wimlib is necessary called libwim-15.dll. It can be found in wimlib archive for Windows runtime: https://wimlib.net/downloads/wimlib-1.14.3-windows-x86_64-bin.zip. In future it will be provided with plugin in correct ZIP archive to work without any additional actions but for now it must be included in one of those paths:
    1) The directory of the executable file being called.
    2) The current working directory from which the executable was called.
    3) The %SystemRoot%\SYSTEM32 directory.
    4) The %SystemRoot% directory.
    5) The directories listed in the Path environment variable.
- Example JSON assume you have OSCDIMG tool installed in PATH env var for creation of secondary ISO with unattended file. It is the eases way to enable it.
- Example JSON assume you have correct unattended file. For reference check: https://developer.hashicorp.com/packer/integrations/hashicorp/hyperv/latest/components/builder/iso#example-for-windows-server-2012-r2-generation-2