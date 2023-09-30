//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package create

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/powa458/packer-plugin-wim/post-processor/utils"
	"github.com/powa458/packer-plugin-wim/post-processor/wim"
)

const (
	qemuBuilderID   = "transcend.qemu"
	hypervBuilderID = "MSOpenTech.hyperv"
)

type Config struct {
	// Main packer paramters from common module.
	common.PackerConfig `mapstructure:",squash"`

	// Paramters specific to this post processor.
	// Those paramters names should start from capital leter.
	// Only such properties are exported to other packages then local scope.
	ImageName        string `mapstructure:"name required:"true""`
	ImageDescription string `mapstructure:"description"`
	ImageCompression uint32 `mapstructure:"compression" required:"true"`

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

	// If error occurred then return it.
	if err != nil {
		return err
	}

	// Set any defaults if needed or validate.
	if pp.config.ImageName == "" {
		pp.config.ImageName = "default"
	}

	if pp.config.ImageCompression > 3 {
		return fmt.Errorf("Unsupported value for property 'compression': %d. Available values: 0 = None, 1 = XPRESS, 2 = LZX, 3 = LZMS.", pp.config.ImageCompression)
	}

	// Return no errors if everything is good.
	return nil
}

func (pp PostProcessor) PostProcess(context context.Context, ui packer.Ui, baseArtifact packer.Artifact) (packer.Artifact, bool, bool, error) {

	// Get BuilderId
	bid := baseArtifact.BuilderId()

	switch bid {
	case qemuBuilderID, hypervBuilderID:
		break
	default:
		err := fmt.Errorf("unsupported artifact type %q: this post-processor only supports "+
			"artifacts from QEMU/Hyper-V builders.", bid)
		return nil, false, false, err
	}

	// Get current working directory.
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, false, false, fmt.Errorf("Unable to get current working directory path")
	}
	ui.Message(fmt.Sprintf("Current directory: '%s'", currentDir))

	// Declare new final artifact
	newArtifact := &wim.WimArtifact{
		Path:        filepath.Join(currentDir, "wim"),
		Name:        pp.config.ImageName,
		Compression: pp.config.ImageCompression,
	}

	// Create base directory to be used as workspace for artifact creation.
	// Should the directory be erased with all content within?
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

	// Defer removal of temp mount directory.
	defer os.RemoveAll(mountDir)

	// Mount the VM image.
	switch bid {
	case hypervBuilderID:

		// Find source image
		source := ""
		for _, i := range baseArtifact.Files() {
			if filepath.Ext(i) == ".vhdx" || filepath.Ext(i) == ".vhd" {
				source = i
				ui.Message(fmt.Sprintf("Found VM image file: '%s'", source))
				break
			} else {
				return nil, false, false, fmt.Errorf("No image file has been found")
			}
		}

		// Mount image to directory
		err = utils.MountImageVHD(context, source, mountDir)
		if err != nil {
			return nil, false, false, err
		}
		ui.Message(fmt.Sprintf("Image %s successfully mounted to: '%s'", source, mountDir))

		// Defer unmounting. We here set err variable to result of this action because this defer must be done always, in successful scenario and on error. With this, we can pass error to packer when unmount fail.
		defer func(string) {
			err = utils.UnmountImageVHD(mountDir)
		}(mountDir)

		// Create WIM image.
		err = wim.CreateWimWindows(context, ui, mountDir, *newArtifact)
		if err != nil {
			return nil, false, false, err
		}

	case qemuBuilderID:
		return nil, false, false, fmt.Errorf("NOT YET IMPLEMENTED")
	}

	/* Final return.
	TODO: Should we also add support for Keep and MustKeep parameter here? Currently both set to false.
	We use err value here and not nil because we got unmounting in defer that return only error value as it is done right before close of program.
	*/
	return newArtifact, false, false, err
}
