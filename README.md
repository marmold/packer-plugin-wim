# packer-plugin-wim

## General overview

A packer plugin to support creation of WIM image from VHD/VHDx files. Main goal is to support building of WIM image on Windows and Linux platform.

This project use cgo and C under the hood to create WIM images for cross platform and better performance.

## Installation

Currently to use the plugin with packer you need to either download latest release binary from: https://github.com/marmold/packer-plugin-wim/releases or build it from sources.

To enable plugin in packer make sure the plugin is placed under `$env:PACKER_PLUGIN_PATH` if configured or in default directory path for plugins. For more information check: https://developer.hashicorp.com/packer/docs/plugins/install-plugins.

## Build

### Windows

To build the plugin you need a gcc compiler. go runtime +1.17 and wimlib files.

To be able to build the plugin on Windows, install GCC compiler, for example  https://jmeubank.github.io/tdm-gcc/download/. For Go, follow official instructions. Download wimlib from https://wimlib.net/downloads or compile it from sources. Make sure to place libwim.lib and wimlib.h to .\lib\devel directory as this directory is linked in cgo. Run `go build -x -o ./out/packer-plugin-wim.exe .`. For additional important information check below.

## Important notes

- To use this plugin on Windows, a required DLL file from wimlib is necessary called libwim-15.dll. It can be found in wimlib archive for Windows runtime: https://wimlib.net/downloads. This DLL is part of archive in released artifacts so its ready to go bundle. If you build plugin from sources then you need to make sure to have this library correctly placed in one of those paths:
    1. The directory of the executable file being called.
    2. The current working directory from which the executable was called.
    3. The %SystemRoot%\SYSTEM32 directory.
    4. The %SystemRoot% directory.
    5. The directories listed in the Path environment variable.
- Example JSON assume you have OSCDIMG tool installed in PATH env var for creation of secondary ISO with unattended file. It is the eases way to enable it.
- Example JSON assume you have correct unattended file. For reference check: https://developer.hashicorp.com/packer/integrations/hashicorp/hyperv/latest/components/builder/iso#example-for-windows-server-2012-r2-generation-2
- LZMS type compression won't be supported in DISM tool in /Apply-Image command.
