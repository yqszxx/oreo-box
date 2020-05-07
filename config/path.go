package config

const (
	Root              = "/var/lib/oreo-box/"
	ImagePath         = Root + "image/"
	ImageTempPath     = "/tmp/oreo-box/image/"
	BoxDataPath       = Root + "box/"
	InfoFileName      = "config.json"
	LogFileName       = "output.log"
	SignatureFileName = "signature.asc"
	ImageDataFileName = "image.tar"
	MountPath         = "rootfs/"
	WritableLayerPath = "writableLayer/"
	NetworkPath       = Root + "network/"
)
