package chroot

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2017-03-30/compute"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2016-06-01/subscriptions"
	"github.com/hashicorp/packer/builder/azure/common"
	"github.com/mitchellh/multistep"
	"github.com/stretchr/testify/assert"
)

func Test_StepInstanceInfo_ErrorWhenNoMetadata(t *testing.T) {
	state := new(multistep.BasicStateBag)
	state.Put("azcli", &testAzCli{})
	state.Put("config", &Config{})
	state.Put("ui", &testUi{t})
	sut := StepInstanceInfo{}

	action := sut.Run(state)

	assert.Equal(t, action, multistep.ActionHalt)
	assert.NotNil(t, state.Get("error"))
}

type testAzCli struct{}

func (t *testAzCli) GetComputeMetadata() (common.ComputeMetadata, error) {
	return common.ComputeMetadata{}, nil
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
