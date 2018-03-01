package chroot

import (
	"errors"
	"fmt"
	"log"

	azcommon "github.com/hashicorp/packer/builder/azure/common"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/mitchellh/multistep"
)

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"tags",
				"command_wrapper",
				"post_mount_commands",
				"pre_mount_commands",
				"mount_path",
				"AzureClient",
			},
		},
	}, raws...)
	if err != nil {
		return nil, err
	}

	// Accumulate any errors or warnings
	var errs *packer.MultiError
	var warns []string

	// Bail early if the creds are no good
	azcli, err := b.config.GetClient()
	if err != nil {
		return warns, err
	}

	if b.config.FromScratch {
		if b.config.Source != "" {
			warns = append(warns, "source is unused when from_scratch is true")
		}
	} else {
		if b.config.Source == "" {
			errs = packer.MultiErrorAppend(
				errs, errors.New("source is required."))
		}
		if err := b.config.ResolveSource(azcli); err != nil {
			errs = packer.MultiErrorAppend(errs, err)
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return warns, errs
	}

	log.Println(common.ScrubConfig(b.config,
		b.config.OAuthToken, b.config.ClientSecret))
	return warns, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	if runtime.GOOS != "linux" {
		return nil, errors.New("The amazon-chroot builder only works on Linux environments.")
	}

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("azcli", azcli)

	// Build the steps
	steps := []multistep.Step{
		&StepInstanceInfo{},
	}

	panic("not implemented")
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
