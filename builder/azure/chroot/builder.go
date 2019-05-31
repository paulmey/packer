package chroot

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/hashicorp/packer/builder/amazon/chroot"
	"github.com/hashicorp/packer/builder/azure/chroot/client"
	azcommon "github.com/hashicorp/packer/builder/azure/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	FromScratch bool `mapstructure:"from_scratch"`

	CommandWrapper    string     `mapstructure:"command_wrapper"`
	PreMountCommands  []string   `mapstructure:"pre_mount_commands"`
	MountOptions      []string   `mapstructure:"mount_options"`
	MountPartition    string     `mapstructure:"mount_partition"`
	MountPath         string     `mapstructure:"mount_path"`
	PostMountCommands []string   `mapstructure:"post_mount_commands"`
	ChrootMounts      [][]string `mapstructure:"chroot_mounts"`
	CopyFiles         []string   `mapstructure:"copy_files"`

	OSDiskSizeGB             int32  `mapstructure:"osdisk_size_gb"`
	OSDiskStorageAccountType string `mapstructure:"osdisk_storageaccounttype"`

	ctx interpolate.Context
}

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	b.config.ctx.Funcs = azcommon.TemplateFuncs
	err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				// these fields are interpolated in the steps,
				// when more information is available
				"command_wrapper",
				"post_mount_commands",
				"pre_mount_commands",
				"mount_path",
			},
		},
	}, raws...)

	// defaults
	if b.config.OSDiskStorageAccountType == "" {
		b.config.OSDiskStorageAccountType = string(compute.PremiumLRS)
	}

	// checks, accumulate any errors or warnings
	var errs *packer.MultiError
	var warns []string

	if b.config.FromScratch {
		if b.config.OSDiskSizeGB == 0 {
			errs = packer.MultiErrorAppend(
				errs, errors.New("osdisk_size_gb is required with from_scratch"))
		}
		if len(b.config.PreMountCommands) == 0 {
			errs = packer.MultiErrorAppend(
				errs, errors.New("pre_mount_commands is required with from_scratch"))
		}
	}

	if err != nil {
		return nil, err
	}
	return warns, errs
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	if runtime.GOOS != "linux" {
		return nil, errors.New("the azure-chroot builder only works on Linux environments")
	}

	var azcli client.AzureClientSet

	wrappedCommand := func(command string) (string, error) {
		ictx := b.config.ctx
		ictx.Data = &struct{ Command string }{Command: command}
		return interpolate.Render(b.config.CommandWrapper, &ictx)
	}

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("hook", hook)
	state.Put("ui", ui)
	state.Put("wrappedCommand", chroot.CommandWrapper(wrappedCommand))

	info, err := azcli.MetadataClient().GetComputeInfo()
	if err != nil {
		log.Printf("MetadataClient().GetComputeInfo(): error: %+v", err)
		err := fmt.Errorf(
			"Error retrieving information ARM resource ID and location" +
				"of the VM that Packer is running on.\n" +
				"Please verify that Packer is running on a proper Azure VM.")
		ui.Error(err.Error())
		return nil, err
	}

	osDiskName := "PackerBuiltOsDisk"

	state.Put("instance", info)
	if err != nil {
		return nil, err
	}

	// Build the steps
	steps := []multistep.Step{
		//		&StepInstanceInfo{},
	}

	if !b.config.FromScratch {
		panic("Only from_scratch is currently implemented")
		// create disk from PIR / managed image (warn on non-linux images)
	} else {
		steps = append(steps,
			&StepCreateNewDisk{
				SubscriptionID:         info.SubscriptionID,
				ResourceGroup:          info.ResourceGroupName,
				DiskName:               osDiskName,
				DiskSizeGB:             b.config.OSDiskSizeGB,
				DiskStorageAccountType: b.config.OSDiskStorageAccountType,
			})
	}

	steps = append(steps,
		&StepAttachDisk{ // sets 'device' in stateBag
			SubscriptionID: info.SubscriptionID,
			ResourceGroup:  info.ResourceGroupName,
			DiskName:       osDiskName,
		},
		&chroot.StepPreMountCommands{
			Commands: b.config.PreMountCommands,
		},
		&StepMountDevice{
			MountOptions:   b.config.MountOptions,
			MountPartition: b.config.MountPartition,
			MountPath:      b.config.MountPath,
		},
		&chroot.StepPostMountCommands{
			Commands: b.config.PostMountCommands,
		},
		&chroot.StepMountExtra{
			ChrootMounts: b.config.ChrootMounts,
		},
		&chroot.StepCopyFiles{
			Files: b.config.CopyFiles,
		},
		&chroot.StepChrootProvision{},
		&chroot.StepEarlyCleanup{},
	)

	// Run!
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	return nil, nil
}

var _ packer.Builder = &Builder{}
