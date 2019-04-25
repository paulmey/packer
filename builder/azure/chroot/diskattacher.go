package chroot

import (
	"context"
	"errors"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
)

type DisksAndVMClientAPI interface {
	// TODO-paulmey: replace by computeapi interfaces
	VirtualMachinesClient() VirtualMachinesClientAPI
}

type VirtualMachinesClientAPI interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName string, VMName string, parameters compute.VirtualMachine) (
		result compute.VirtualMachinesCreateOrUpdateFuture, err error)
	Get(ctx context.Context, resourceGroupName string, VMName string, expand compute.InstanceViewTypes) (
		result compute.VirtualMachine, err error)
}

type DiskAttacher interface {
	AttachDisk(vm, disk string) (lun int, err error)
}

func NewDiskAttacher(azureClient DisksAndVMClientAPI) DiskAttacher {
	return diskAttacher{azureClient}
}

type diskAttacher struct {
	azcli DisksAndVMClientAPI
}

func (da diskAttacher) AttachDisk(vm, disk string) (int, error) {

	return 0, errors.New("not implemented")
}
