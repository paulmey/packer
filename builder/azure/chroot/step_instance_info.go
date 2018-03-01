package chroot

import (
	"fmt"

	azcommon "github.com/hashicorp/packer/builder/azure/common"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

//StepInstanceInfo verifies that this builder is running on an Azure instance.
type StepInstanceInfo struct{}

func (s *StepInstanceInfo) Run(state multistep.StateBag) multistep.StepAction {
	azcli := state.Get("azcli").(*azcommon.AzureClient)
	ui := state.Get("ui").(packer.Ui)

	md, err := azcli.GetComputeMetadata()
	if err != nil ||

		md.SubscriptionID == "" ||
		md.Name == "" ||
		md.Location == "" ||
		md.ResourceGroupName == "" {

		err := fmt.Errorf(
			"Error retrieving VM resource ID for the instance Packer is running on.\n" +
				"Please verify Pakcer is running on an Azure VM")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	if b.config.SubscriptionID != "" &&
		b.config.SubscriptionID != md.SubscriptionID {
		warns = append(warns, fmt.Sprintf("subscription_id (%s) is overridden "+
			"with VM subscription id (%s)",
			b.config.SubscriptionID,
			md.SubscriptionID))
	}
	b.config.SubscriptionID = md.SubscriptionID
	b.config.location = md.Location

}
