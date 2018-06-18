package finalize

import (
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"
)

type Finalizer struct {
	BuildDir string
	DepDir   string
	Log      *libbuildpack.Logger
}

func Run(f *Finalizer) error {
	//Delete vendored packages to optimize disk space
	if err := f.CleanupVendorDir(); err != nil {
		f.Log.Error("Error cleaning up vendored packages R: %v", err)
		return err
	}
	return nil
}

func (f *Finalizer) CleanupVendorDir() error {
	vendorPath := filepath.Join(f.BuildDir, "vendor_r")
	if exists, _ := libbuildpack.FileExists(vendorPath); exists {
		f.Log.Info("Cleaning up vendored packages")
		return os.RemoveAll(vendorPath)
	}
	return nil
}
