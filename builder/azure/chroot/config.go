package chroot

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2017-03-30/compute"
	"github.com/Azure/go-autorest/autorest/to"
	azcommon "github.com/hashicorp/packer/builder/azure/common"
	"log"
	"strings"

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
	location   string

	ctx interpolate.Context
}

func (c *Config) ResolveSource(azcli azcommon.AzureClient) error {
	s := c.Source

	cli := azcli.PlatformImagesClient()
	if parts := strings.Split(s, ":"); len(parts) == 4 {
		log.Printf("Config: source looks like an image URN")
		if strings.ToLower(parts[3]) == "latest" {
			// figure out what the latest version of the image is
			// (until this is implemented in the APIs ???)
			//	cli := c.DisksClient()
			ctx := context.Background()
			res, err := cli.List(ctx,
				c.location,
				parts[0], // publisherName
				parts[1], // offer
				parts[2], // sku
				"", to.Int32Ptr(0), "Version")
			if err != nil {
				if res.StatusCode == 404 {
					return fmt.Errorf("Config: Image URN not found: %s", s)
				}
				return err
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			for _, v := range *res.Value {
				log.Printf("Found candidate image: %+v", *v.ID)
			}
		} else {
			ctx := context.Background()
			res, err := cli.Get(ctx,
				c.location,
				parts[0],
				parts[1],
				parts[2],
				parts[2])
			if err != nil {
				return err
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			_ = res
			// todo: finish
		}

	}
	//	compute.DisksClient
	return nil
}
