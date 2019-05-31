package chroot

import (
	"context"
	"fmt"
	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"log"
)

var _ multistep.Step = &StepInstanceInfo{}

type StepInstanceInfo struct {
}

func (StepInstanceInfo) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	azcli := state.Get("azureclient").(client.AzureClientSet)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Gathering information about this Azure VM...")

	info, err := azcli.MetadataClient().GetComputeInfo()
	if err != nil {
		log.Printf("StepInstanceInfo: error: %+v", err)
		err := fmt.Errorf(
			"Error retrieving information ARM resource ID and location" +
				"of the VM that Packer is running on.\n" +
				"Please verify that Packer is running on a proper Azure VM.")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	log.Printf("StepInstanceInfo: ID: %s", info.ResourceID())
	log.Printf("StepInstanceInfo: location: %s", info.Location)

	state.Put("instance", info)

	return multistep.ActionContinue
}

func (StepInstanceInfo) Cleanup(multistep.StateBag) {}
