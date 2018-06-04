package finalize

import (
	"github.com/cloudfoundry/libbuildpack"
	"os"
	"path/filepath"
)

type Finalizer struct {
	BuildDir string
	DepDir   string
	Log      *libbuildpack.Logger
}

func Run(sf *Finalizer) error {
	//Delete vendored packages to optimize disk space
	if err := sf.CleanupVendorDir(); err != nil {
		sf.Log.Error("Error cleaning up vendored packages R: %v", err)
		return err
	}
	return nil
}

func (sf *Finalizer) CleanupVendorDir() error {
	rPackagesPath := filepath.Join(sf.BuildDir, "rPackages")
	if exists, _ := libbuildpack.FileExists(rPackagesPath); exists {
		return os.RemoveAll(rPackagesPath)
	}
	return nil
}
