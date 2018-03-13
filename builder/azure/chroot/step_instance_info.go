package chroot

import (
	"errors"
	"fmt"

	"github.com/Azure/go-autorest/autorest"
	azcommon "github.com/hashicorp/packer/builder/azure/common"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

//StepInstanceInfo verifies that this builder is running on an Azure instance.
type StepInstanceInfo struct{}

var GetAuthorizer = func(*Config) (autorest.Authorizer, error) {
	return nil, errors.New("Not implemented")
}

func (s *StepInstanceInfo) Run(state multistep.StateBag) multistep.StepAction {
	azcli := state.Get("azcli").(azcommon.AzureClient)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	md, err := azcli.GetComputeMetadata()
	if err != nil ||

		md.SubscriptionID == "" ||
		md.Name == "" ||
		md.Location == "" ||
		md.ResourceGroupName == "" {

		err := fmt.Errorf(
			"Error retrieving VM resource ID for the instance Packer is running on.\n" +
				"Please verify Packer is running on an Azure VM")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	config.SubscriptionID = md.SubscriptionID
	config.location = md.Location

	authorizer, err := GetAuthorizer(config)
	if err != nil {
		wrappedErr := fmt.Errorf("Error retrieving credentials: %s", err)
		state.Put("error", wrappedErr)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	azcli.SetAuthorizer(authorizer)

	return multistep.ActionContinue
}
