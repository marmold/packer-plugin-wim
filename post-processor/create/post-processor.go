//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package wimcreate

import (
	"context"
	"fmt"
	"path/filepath"

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

	if pp.config.ImagePath == "" {
		pp.config.ImagePath = "" // Add default path later.
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

	// Final return.
	return newArtifact, keep, mustKeep, err
}
