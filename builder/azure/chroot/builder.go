package chroot

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/hashicorp/packer/builder/amazon/chroot"
	azcommon "github.com/hashicorp/packer/builder/azure/common"
	"github.com/hashicorp/packer/builder/azure/common/client"
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

	OSDiskSizeGB             int32  `mapstructure:"os_disk_size_gb"`
	OSDiskStorageAccountType string `mapstructure:"os_disk_storage_account_type"`
	OSDiskCacheType          string `mapstructure:"os_disk_cache_type"`

	ImageResourceID string `mapstructure:"image_resource_id"`
	ImageOSState    string `mapstructure:"image_os_state"`

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

	// Defaults
	if b.config.ChrootMounts == nil {
		b.config.ChrootMounts = make([][]string, 0)
	}

	if len(b.config.ChrootMounts) == 0 {
		b.config.ChrootMounts = [][]string{
			{"proc", "proc", "/proc"},
			{"sysfs", "sysfs", "/sys"},
			{"bind", "/dev", "/dev"},
			{"devpts", "devpts", "/dev/pts"},
			{"binfmt_misc", "binfmt_misc", "/proc/sys/fs/binfmt_misc"},
		}
	}

	// set default copy file if we're not giving our own
	if b.config.CopyFiles == nil {
		if !b.config.FromScratch {
			b.config.CopyFiles = []string{"/etc/resolv.conf"}
		}
	}

	if b.config.CommandWrapper == "" {
		b.config.CommandWrapper = "{{.Command}}"
	}

	if b.config.MountPath == "" {
		b.config.MountPath = "/mnt/packer-azure-chroot-disks/{{.Device}}"
	}

	if b.config.MountPartition == "" {
		b.config.MountPartition = "1"
	}

	if b.config.OSDiskStorageAccountType == "" {
		b.config.OSDiskStorageAccountType = string(compute.PremiumLRS)
	}

	if b.config.OSDiskCacheType == "" {
		b.config.OSDiskCacheType = string(compute.CachingTypesReadOnly)
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

	//	OsState: compute.OperatingSystemStateTypes(s.ImageOSState),
	//	StorageAccountType: compute.StorageAccountTypes(s.OSDiskStorageAccountType),

	if err := checkOSState(b.config.ImageOSState); err != nil {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("image_os_state: %v", err))
	}
	if err := checkDiskCacheType(b.config.OSDiskCacheType); err != nil {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("os_disk_cache_type: %v", err))
	}
	if err := checkStorageAccountType(b.config.OSDiskStorageAccountType); err != nil {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("os_disk_storage_account_type: %v", err))
	}

	if err != nil {
		return nil, err
	}

	packer.LogSecretFilter.Set(b.config.AccessKey, b.config.SecretKey, b.config.Token)
	return warns, errs
}

func checkOSState(s string) interface{} {
	for _, v := range compute.PossibleOperatingSystemStateTypesValues() {
		if compute.OperatingSystemStateTypes(s) == v {
			return nil
		}
	}
	return fmt.Errorf("%q is not a valid value (%v)",
		s, compute.PossibleOperatingSystemStateTypesValues())
}

func checkDiskCacheType(s string) interface{} {
	for _, v := range compute.PossibleCachingTypesValues() {
		if compute.CachingTypes(s) == v {
			return nil
		}
	}
	return fmt.Errorf("%q is not a valid value (%v)",
		s, compute.PossibleCachingTypesValues())
}

func checkStorageAccountType(s string) interface{} {
	for _, v := range compute.PossibleStorageAccountTypesValues() {
		if compute.StorageAccountTypes(s) == v {
			return nil
		}
	}
	return fmt.Errorf("%q is not a valid value (%v)",
		s, compute.PossibleStorageAccountTypesValues())
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	if runtime.GOOS != "linux" {
		return nil, errors.New("the azure-chroot builder only works on Linux environments")
	}

	// todo: instantiate Azure client
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
	state.Put("azureclient", azcli)
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
		&StepAttachDisk{}, // uses os_disk_resource_id and sets 'device' in stateBag
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
		&StepCreateImage{
			ImageResourceID:          b.config.ImageResourceID,
			ImageOSState:             b.config.ImageOSState,
			OSDiskCacheType:          b.config.OSDiskCacheType,
			OSDiskStorageAccountType: b.config.OSDiskStorageAccountType,
		},
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
