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
			},
		},
	}, raws...)
	if err != nil {
		return nil, err
	}

	// Accumulate any errors or warnings
	var errs *packer.MultiError
	var warns []string

	md, err := azcommon.GetComputeMetadata()
	if err != nil {
		return warns, fmt.Errorf("Error retrieving VM metadata (%q), "+
			"is this an Azure VM?", err)
	}
	if md.SubscriptionID == "" ||
		md.Name == "" ||
		md.Location == "" ||
		md.ResourceGroupName == "" {
		return warns, fmt.Errorf("VM metadata is not complete (%+v), "+
			"is this an Azure VM?", md)
	}
	if b.config.SubscriptionID != "" &&
		b.config.SubscriptionID != md.SubscriptionID {
		warns = append(warns, fmt.Sprintf("subscription_id (%s) is overridden "+
			"with VM subscription id (%s)",
			b.config.SubscriptionID,
			md.SubscriptionID))
	}
	b.config.SubscriptionID = md.SubscriptionID

	// Bail early if the creds are no good
	err = b.config.ResolveClient()
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
		if err := b.config.ResolveSource(); err != nil {
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
	panic("not implemented")

}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
