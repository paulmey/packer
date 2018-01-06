package common

// Config for creating a managed disk. Used for temporary disk image
type DiskConfig struct {
	ResourceGroup string            `mapstructure:"disk_resourcegroup"`
	DiskName      string            `mapstructure:"disk_name"`
	Tags          map[string]string `mapstructure:"disk_tags"`
}

// Config for creating managed disk images.
type ImageConfig struct {
	ResourceGroup string            `mapstructure:"image_resourcegroup"`
	ImageName     string            `mapstructure:"image_name"`
	Tags          map[string]string `mapstructure:"image_tags"`
}
