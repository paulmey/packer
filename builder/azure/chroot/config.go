package chroot

import (
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2017-03-30/compute"
	azcommon "github.com/hashicorp/packer/builder/azure/common"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/template/interpolate"
)

// Config is the configuration that is chained through the steps and
// settable from the template.
type Config struct {
	common.PackerConfig   `mapstructure:",squash"`
	azcommon.ClientConfig `mapstructure:",squash"`
	azcommon.DiskConfig   `mapstructure:",squash"`
	azcommon.ImageConfig  `mapstructure:",squash"`

	ChrootMounts      [][]string `mapstructure:"chroot_mounts"`
	CommandWrapper    string     `mapstructure:"command_wrapper"`
	CopyFiles         []string   `mapstructure:"copy_files"`
	DevicePath        string     `mapstructure:"device_path"`
	FromScratch       bool       `mapstructure:"from_scratch"`
	MountOptions      []string   `mapstructure:"mount_options"`
	MountPartition    int        `mapstructure:"mount_partition"`
	MountPath         string     `mapstructure:"mount_path"`
	PostMountCommands []string   `mapstructure:"post_mount_commands"`
	PreMountCommands  []string   `mapstructure:"pre_mount_commands"`
	RootDeviceName    string     `mapstructure:"root_device_name"`
	RootVolumeSize    int64      `mapstructure:"root_volume_size"`
	//	SourceAmi         string                     `mapstructure:"source_ami"`
	//	SourceAmiFilter   awscommon.AmiFilterOptions `mapstructure:"source_ami_filter"`

	// Disk source can be:
	// - Blob uri to a vhd blob in the same location as the VM
	// - Managed disk resource id
	// - Managed disk image resource id
	// - Platform image urn (publisher:offer:sku:version)
	Source string `mapstructure:"source"`

	sourceDisk compute.Disk

	ctx interpolate.Context
}

func (c *Config) ResolveSource() error {
	//	compute.DisksClient
	return nil
}
