package chroot

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testvm   = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testGroup/Microsoft.Compute/virtualMachines/testVM"
	testdisk = "/subscriptions/00000000-0000-0000-0000-000000000001/resourceGroups/testGroup2/Microsoft.Compute/disks/testDisk"
)

func Test_DiskAttacherAttachesDiskToVM(t *testing.T) {
	if os.Getenv("AZURE_DISK_ATTACHER_TESTS") == "" {
		t.Skipf("AZURE_DISK_ATTACHER_TESTS not set")
	}
	azcli := fakeAzureClient()
	da := NewDiskAttacher(azcli)

	_, err := da.AttachDisk(testvm, testdisk)
	assert.NotNil(t, err)
}

func fakeAzureClient() DisksAndVMClientAPI { return nil }
