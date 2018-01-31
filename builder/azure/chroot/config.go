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
	// - Managed disk snapshot resource id
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
			ctx := context.Background()
			res, err := cli.List(ctx,
				c.location,
				parts[0], // publisherName
				parts[1], // offer
				parts[2], // sku
				// order by 'Version name' and take the top 1 hit, i.e. the latest image
				"", to.Int32Ptr(1), "name desc")
			if err != nil {
				if res.StatusCode == 404 {
					return fmt.Errorf("Config: Image URN not found: %s", s)
				}
				return err
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			if res.Value == nil || len(*res.Value) != 1 {
				return fmt.Errorf("Config: expected to find one and only on image in array: %+v", *res.Value)
			}
			imageVersion := ((*res.Value)[0])
			log.Printf("Found candidate image: %+v", imageVersion.ID)

			c.sourceDisk.DiskProperties=&compute.DiskProperties{
				CreationData: &compute.CreationData.CreateOptio


		} else {
			ctx := context.Background()
			res, err := cli.Get(ctx,
				c.location,
				parts[0],
				parts[1],
				parts[2],
				parts[3])
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
