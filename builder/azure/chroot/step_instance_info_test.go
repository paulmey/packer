package chroot

import (
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2017-03-30/compute"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2016-06-01/subscriptions"
	"github.com/Azure/go-autorest/autorest"
	"github.com/hashicorp/packer/builder/azure/common"
	"github.com/mitchellh/multistep"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_StepInstanceInfo_ErrorWhenNoMetadata(t *testing.T) {
	state := new(multistep.BasicStateBag)
	state.Put("azcli", &testAzCli{})
	state.Put("config", &Config{})
	state.Put("ui", &testUi{t})
	sut := StepInstanceInfo{}

	action := sut.Run(state)

	assert.Equal(t, action, multistep.ActionHalt)
	require.NotNil(t, state.Get("error"))
	assert.Contains(t, state.Get("error").(error).Error(),
		"Please verify Packer is running on an Azure VM")
}

func Test_StepInstanceInfo_ErrorWhenError(t *testing.T) {
	state := new(multistep.BasicStateBag)
	azcli := NewTestAzCli()
	azcli.ComputeMetadataError = errors.New("An error here")
	state.Put("azcli", &azcli)
	state.Put("config", &Config{})
	state.Put("ui", &testUi{t})
	sut := StepInstanceInfo{}

	action := sut.Run(state)

	assert.Equal(t, action, multistep.ActionHalt)
	require.NotNil(t, state.Get("error"))
	assert.Contains(t, state.Get("error").(error).Error(),
		"Please verify Packer is running on an Azure VM")
}

func NewTestAzCli() testAzCli {
	return testAzCli{
		ComputeMetadata: common.ComputeMetadata{
			Location:          "westus2",
			Name:              "packer-test-chroot-host",
			ResourceGroupName: "testrg",
			SubscriptionID:    "568974ae-267e-11e8-996e-73f1ea39fde3",
		},
	}
}

type testAzCli struct {
	common.ComputeMetadata
	ComputeMetadataError error
}

func (t *testAzCli) GetComputeMetadata() (common.ComputeMetadata, error) {
	return t.ComputeMetadata, t.ComputeMetadataError
}

func (t *testAzCli) SubscriptionsClient() subscriptions.Client {
	panic("not implemented")
}

func (t *testAzCli) PlatformImagesClient() compute.VirtualMachineImagesClient {
	panic("not implemented")
}

func (t *testAzCli) ManagedDisksClient() compute.DisksClient {
	panic("not implemented")
}

func (t *testAzCli) SetAuthorizer(autorest.Authorizer) {
}

type testUi struct{ *testing.T }

func (t *testUi) Ask(string) (string, error) {
	panic("not implemented")
}

func (t *testUi) Say(s string) {
	t.Logf("Ui.Say: %s", s)
}

func (t *testUi) Message(s string) {
	t.Logf("Ui.Message: %s", s)
}

func (t *testUi) Error(s string) {
	t.Logf("Ui.Error: %s", s)
}

func (t *testUi) Machine(s string, a ...string) {
	t.Logf("Ui.Machine: "+s, a)
}
