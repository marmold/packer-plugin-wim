# packer-plugin-wim

## General overview

A packer plugin to support creation of WIM image from VHD/VHDx files. Main goal is to support building of WIM image on Windows and Linux platform.

This project use cgo and C under the hood to create WIM images for cross platform and better performance.

## Installation

Currently to use the plugin with packer you need to either download latest release binary for platform you are interested from: https://github.com/marmold/packer-plugin-wim/releases or build it from sources.

To enable plugin in packer make sure the plugin is extracted under `$env:PACKER_PLUGIN_PATH` if configured or in default directory path for plugins equal to `$HOME/.config/packer/plugins` on UNIX, or `%APPDATA%\packer.d\plugins` for Windows. For more information check: https://developer.hashicorp.com/packer/docs/configure.

## Build

### Windows

To build the plugin you need a gcc compiler, go runtime +1.17 and wimlib files.

To be able to build the plugin on Windows, install GCC compiler that suit you the most, for example  https://jmeubank.github.io/tdm-gcc/download/. For Go, follow official instructions. Download wimlib from https://wimlib.net/downloads or compile it from sources. Make sure to place `libwim-15.dll` in `\.out` directory and `libwim.lib`, `wimlib.h` to `.\.lib\devel` directory as this directory is linked in cgo. Run `go build -x -o .\.out\packer-plugin-wim.exe .` from root directory of repository. For additional important information check below.

## Config parameters

| Parameter name |  Description                                                                                 |
| -------------- |  ------------------------------------------------------------------------------------------- |
| **image_name** |  Allow you to set custom name for your file. If not set, a `default.wim` name will be used   |
| **image_path** |  Allow for custom path where result file should be placed. This can be either a root path or relative path where binary runs. If not set a current directory where binary started will be used                                   |
| **compression**|  A compression which should be used when creating WIM file. Supported formats are: `0` = None, `1` = XPRESS, `2` = LZX, `3` = LZMS.                                                                                                |

## Examples

For some simple examples follow [examples](examples/examples.md)

## Important notes

- To use this plugin on Windows, a required DLL file from wimlib is necessary called libwim-15.dll. It can be found in wimlib archive for Windows runtime: https://wimlib.net/downloads. This DLL is part of archive in released artifacts so its ready to go bundle. If you build plugin from sources then you need to make sure to have this library correctly placed in one of those paths:
    - The directory of the executable file being called.
    - The current working directory from which the executable was called.
    - The %SystemRoot%\SYSTEM32 directory.
    - The %SystemRoot% directory.
    - The directories listed in the Path environment variable.
- Example JSON assume you have OSCDIMG tool installed in PATH env var for creation of secondary ISO with unattended file. It is the eases way to enable it.
- Example JSON assume you have correct unattended file. For reference check: https://developer.hashicorp.com/packer/integrations/hashicorp/hyperv/latest/components/builder/iso#example-for-windows-server-2012-r2-generation-2
- LZMS type compression won't be supported in DISM tool in /Apply-Image command.
