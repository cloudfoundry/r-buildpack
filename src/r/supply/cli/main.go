package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/r-buildpack/src/r/supply"
)

func main() {
	logger := libbuildpack.NewLogger(os.Stdout)

	buildpackDir, err := libbuildpack.GetBuildpackDir()
	if err != nil {
		logger.Error("Unable to determine buildpack directory: %s", err.Error())
		os.Exit(9)
	}

	manifest, err := libbuildpack.NewManifest(buildpackDir, logger, time.Now())
	if err != nil {
		logger.Error("Unable to load buildpack manifest: %s", err.Error())
		os.Exit(10)
	}

	installer := libbuildpack.NewInstaller(manifest)

	stager := libbuildpack.NewStager(os.Args[1:], logger, manifest)
	if err := stager.CheckBuildpackValid(); err != nil {
		os.Exit(11)
	}

	if err := installer.SetAppCacheDir(stager.CacheDir()); err != nil {
		logger.Error("Unable to setup appcache: %s", err)
		os.Exit(18)
	}

	if err = manifest.ApplyOverride(stager.DepsDir()); err != nil {
		logger.Error("Unable to apply override.yml files: %s", err)
		os.Exit(17)
	}

	if err := os.MkdirAll(filepath.Join(stager.DepDir(), "bin"), 0755); err != nil {
		logger.Error("Unable to create bin directory: %s", err.Error())
		os.Exit(13)
	}

	if err = stager.SetStagingEnvironment(); err != nil {
		logger.Error("Unable to setup environment variables: %s", err.Error())
		os.Exit(13)
	}

	command := &libbuildpack.Command{}

	supplier := supply.New(stager, command, manifest, installer, logger)

	if err := supplier.Run(); err != nil {
		os.Exit(14)
	}

	if err := stager.WriteConfigYml(nil); err != nil {
		logger.Error("Error writing config.yml: %s", err.Error())
		os.Exit(15)
	}

	if err = installer.CleanupAppCache(); err != nil {
		logger.Error("Unable to clean up app cache: %s", err)
		os.Exit(19)
	}
	stager.StagingComplete()
}
