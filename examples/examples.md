# Examples

This document describe some simple examples on how to create WIM file with some minor changes.

## Hyper-V

This example shows how to make simple modification on OS level and create a WIM image for further usage. This image can be created either from ISO or VHDX. It can be then use id different deployment processes, for example, create custom setup ISO or deploy it via PXE over the network. This example will focus on custom ISO of Windows 2019.

### Get necessary files

You can get necessary base ISO file for [ISO exmaple](#iso) from https://www.microsoft.com/en-us/evalcenter/download-windows-server-2019.

You can get necessary base VHDX file for [VHDX exmaple](#vhdx) from https://www.microsoft.com/en-us/evalcenter/download-windows-server-2019.

In some processes a `oscdimg.exe` binary is used, for example, to create supplemental ISO with additional files, like autounattended.xml or to create custom modified setup ISO. This binary can be collected from [Windows ADK](https://learn.microsoft.com/en-us/windows-hardware/get-started/adk-install#other-adk-downloads). For smooth work add this binary to your `$env:path`. In this example it is also used to create custom setup ISO.

### ISO

Store file path of downloaded ISO image in variable for reference, starting from root of partition:

```powershell
$IsoPath = "<Path to ISO>".Replace("\", "\\")
```

As packer require a file hash for validation purposes, calculate and store it in variable for future reference:

```powershell
$IsoFileHash = (Get-FileHash -Path $IsoPath).Hash
```

Before running packer make sure you install the plugin correctly. For reference check [installation](..\README.md#Installation) in main readme file.

Start the build process with following command:

```powershell
pushd .\examples
packer.exe build -force -var "iso_url=$IsoPath" -var "iso_checksum=$IsoFileHash" .\hyperv_iso_win2019_BIOS_example.json
```

The template use `iso_url` variable as reference to file on disk and `iso_checksum` to do not calculate file hash on each execution.

NOTE: For UEFI based installation, go with `\hyperv_iso_win2019_UEFI_example.json` file as JSON template.

#### Create custom base setup ISO

After creating WIM image, extract a stock Microsoft ISO image to directory and replace `\sources\install.wim` with your custom created WIM file. Then create a custom ISO file with bellow command.

```powershell
oscdimg.exe -u1 -m -h -lWIN_SETUP "<root path to Extracted ISO>" "<root path to new iso>"
```

### VHDX

For VHDX installation a unattended setup file need to be injected into image in order to correctly boot the machine and perform necessary and steps.
You can use bellow snippet to modify the image before using it in packer

```powershell
pushd .\examples
# Get VHDX file
$VhdxPath = "<Path to VHDX>".Replace("\", "\\")

# Mount VHDX file in temp directory
$MountPoint = New-Item -Type Directory -Path $env:TEMP -Name "mount_$(Get-Random)"
Mount-WindowsImage -Path $MountPoint -ImagePath $VhdxPath -Index 1

# Add unattended file perform necessary steps
New-Item -ItemType Directory -Path "$($MountPoint.FullName)\Windows" -Name "Panther" -Force
Copy-Item ".\unattended\win_2019\UEFI\autounattend.xml" "$($MountPoint.FullName)\Windows\Panther\unattend.xml"

# Save the image
Dismount-WindowsImage -Path $MountPoint -Save
Remove-item $MountPoint -Force -Recurse
```

Now when we have correctly modified image we can proceed with creation forward. Unfortunately, packer do not allow directly usage of VHDX disk file as source for new VM. We can either import a VM from a disk (From appropriate directory structure that contains the "Virtual Machines", "Snapshots", and/or "Virtual Hard Disks" subdirectories) or clone existing one. This example shows the second option.

NOTE: To test first option mentioned, you can combine two flows, ISO and VHDX. This way of processing allow to make two-step build where first from ISO template a exported base virtual machine is created and later via VHDX template a WIM image is created. Be sure to include `output_directory` with appropriate path (used later VHDX template) and `skip_export` set to `true` in ISO template.

Before running packer make sure you install the plugin correctly. For reference check [installation](..\README.md#Installation) in main readme file.

Start the build process with following snippet:

```powershell
pushd .\examples
# Create template VM for reference
$TemplateVM = New-VM -Name "Template_win2019" -BootDevice VHD -VHDPath $VhdxPath

# Start packer to create WIM image
packer.exe build -force -var "vm_name=$($TemplateVM.Name)" .\hyperv_vhdx_win2019_UEFI_example.json

# Remove template VM
Remove-VM $TemplateVM -Force
```