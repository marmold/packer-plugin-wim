//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package wimcreate

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type Config struct {
	// Main packer paramters from common module.
	common.PackerConfig `mapstructure:",squash"`

	// Paramters specific to this post processor.
	// Those paramters names should start from capital leter.
	// Only such properties are exported to other packages then local scope.
	ImageName        string `mapstructure:"image_name"`
	ImagePath        string `mapstructure:"image_path"`
	ImageDescription string `mapstructure:"image_description"`

	ctx interpolate.Context
}

type PostProcessor struct {
	config Config
}

func (pp *PostProcessor) ConfigSpec() hcldec.ObjectSpec {
	return pp.config.FlatMapstructure().HCL2Spec()
}

func (pp *PostProcessor) Configure(raws ...interface{}) error {

	// Decode configuration
	err := config.Decode(&pp.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &pp.config.ctx,
	}, raws...)

	// If error occured then return it.
	if err != nil {
		return err
	}

	// Set any defaults if needed.
	if pp.config.ImageName == "" {
		pp.config.ImageName = "test-wim-name"
	}

	// Return no errors if everything is good.
	return nil
}

func (pp PostProcessor) PostProcess(context context.Context, ui packer.Ui, baseArtifact packer.Artifact) (packer.Artifact, bool, bool, error) {

	// Get current working directory.
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, false, false, fmt.Errorf("Unabel to get current working directory path")
	}
	ui.Message(fmt.Sprintf("Current directory: '%s'", currentDir))

	// Declare new final artifact
	newArtifact := &WimArtifact{
		Path: strings.Join([]string{currentDir, "wim"}, "\\"),
		Name: pp.config.ImageName,
	}

	// Check if the source file is VHDX (VHD also?) format.
	source := ""
	for _, i := range baseArtifact.Files() {
		if filepath.Ext(i) == ".vhdx" {
			source = i
			ui.Message(fmt.Sprintf("Found VHDX file: '%s'", source))
			break
		} else {
			ui.Message(fmt.Sprintf("No VHDX file has been found"))
			return nil, false, false, fmt.Errorf("No VHDX file has been found")
		}
	}

	// Create base directory to be used as workspace for artifact creation.
	// Should the directory be ereased with all content within?
	err = os.MkdirAll(newArtifact.Path, 0777)
	if err != nil {
		return nil, false, false, fmt.Errorf("Unable to create final directory for opeartions: '%s'.", newArtifact.Path)
	}
	ui.Message(fmt.Sprintf("Base directory for artifact created: '%s'", newArtifact.Path))

	// Create temp mount directory. The directory will be created with random number suffix.
	mountDir, err := os.MkdirTemp(newArtifact.Path, "mount_")
	if err != nil {
		log.Fatal(err)
	}
	ui.Message(fmt.Sprintf("Mount directory created: '%s'", mountDir))
	defer os.RemoveAll(mountDir)

	// Mount VHDX image to mount directory.
	err = exec.CommandContext(context, "cmd", "/c", "dism", "/mount-image", strings.Join([]string{"/imagefile", source}, ":"), "/Index:1", strings.Join([]string{"/mountdir", mountDir}, ":")).Run()
	if err != nil {
		//ui.Message(fmt.Sprintf("Unable to mount image %s to mount dir: %s", source, mountDir))
		return nil, false, false, fmt.Errorf("Unable to mount image %s to mount dir: %s", source, mountDir)
	}
	ui.Message(fmt.Sprintf("VHDX Image %s successfully mounted to: '%s'", source, mountDir))

	// Create WIM image from mounted directory.
	wimPath := newArtifact.Path + "\\" + newArtifact.Name + ".wim"
	ui.Message(fmt.Sprintf("Creating new WIM image under %s", wimPath))
	err = exec.CommandContext(context, "cmd", "/c", "dism", "/Capture-Image", strings.Join([]string{"/ImageFile", wimPath}, ":"), strings.Join([]string{"/CaptureDir", mountDir}, ":"), strings.Join([]string{"/Name", "Test"}, ":")).Run()
	if err != nil {
		//ui.Message(fmt.Sprintf("Failed to create WIM image from mount dir: %s. Unmounting ...", mountDir))
		exec.Command("cmd", "/c", "dism", "/Unmount-image", strings.Join([]string{"/mountdir", mountDir}, ":"), "/Discard").Run()
		return nil, false, false, fmt.Errorf("Failed to create WIM image from mount dir: %s. Unmounting ...", mountDir)
	}
	ui.Message(fmt.Sprintf("WIM Image %s successfully created from: '%s'", wimPath, mountDir))

	// Unmount VHDX image from mount directory if eveyrthing went well.
	err = exec.CommandContext(context, "cmd", "/c", "dism", "/Unmount-image", strings.Join([]string{"/mountdir", mountDir}, ":"), "/Discard").Run()
	if err != nil {
		ui.Message(fmt.Sprintf("Failed to unmount image %s from mount dir: %s", source, mountDir))
		// log.Fatal(err)
		return nil, false, false, err
	}
	ui.Message(fmt.Sprintf("VHDX Image %s successfully unmounted from: '%s'", source, mountDir))

	// Final return.
	//TODO: Should we also add support for Keep and MustKeep paramter here? Currnetly both set to false.
	return newArtifact, false, false, err
}
