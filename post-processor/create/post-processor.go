//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package wimcreate

import (
	"context"
	"fmt"
	"log"
	"os"
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
		pp.config.ImageName = "" // Add deafult name later.
	}

	// Return no errors if everything is good.
	return nil
}

func (pp PostProcessor) PostProcess(context context.Context, ui packer.Ui, baseArtifact packer.Artifact) (newArtifact packer.Artifact, keep, mustKeep bool, err error) {

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

	// Get current working directory.
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	ui.Message(fmt.Sprintf("Current directory: '%s'", currentDir))

	// Create base directory to be used as workspace for artifact creation.
	baseDir := strings.Join([]string{currentDir, "wim"}, "\\")
	err = os.Mkdir(baseDir, 0777)
	if err != nil {
		log.Fatal(err)
	}
	ui.Message(fmt.Sprintf("Base directory for artifact created: '%s'", baseDir))

	// Create temp mount directory. The directory will be created with random number suffix.
	mountDir, err := os.MkdirTemp(baseDir, "mount_")
	if err != nil {
		log.Fatal(err)
	}
	ui.Message(fmt.Sprintf("Mount directory created: '%s'", mountDir))
	//defer os.RemoveAll(mountDir)

	// Final return.
	return newArtifact, keep, mustKeep, err
}
