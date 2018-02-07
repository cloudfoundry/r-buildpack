package supply

import (
	"bytes"
	"io/ioutil"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"
)

type Manifest interface {
	AllDependencyVersions(string) []string
	InstallDependency(libbuildpack.Dependency, string) error
}

type Stager interface {
	DepDir() string
	DepsIdx() string
	LinkDirectoryInDepDir(string, string) error
}

type Supplier struct {
	Stager   Stager
	Manifest Manifest
	Log      *libbuildpack.Logger
}

func New(stager Stager, manifest Manifest, logger *libbuildpack.Logger) *Supplier {
	return &Supplier{
		Stager:   stager,
		Manifest: manifest,
		Log:      logger,
	}
}

func (s *Supplier) Run() error {
	s.Log.BeginStep("Supplying R")

	if err := s.InstallR(); err != nil {
		s.Log.Error("Error installing R: %v", err)
		return err
	}

	if err := s.RewriteRHome(); err != nil {
		s.Log.Error("Error rewriting R_HOME: %v", err)
		return err
	}

	return nil
}

func (s *Supplier) RewriteRHome() error {
	path := filepath.Join(s.Stager.DepDir(), "r", "bin", "R")
	body, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	body = bytes.Replace(body, []byte("/usr/local/lib/R"), []byte(filepath.Join("$DEPS_DIR", s.Stager.DepsIdx(), "r")), -1)

	return ioutil.WriteFile(path, body, 0755)
}

func (s *Supplier) InstallR() error {
	versions := s.Manifest.AllDependencyVersions("r")
	ver, err := libbuildpack.FindMatchingVersion("x", versions)
	if err != nil {
		return err
	}

	if err := s.Manifest.InstallDependency(libbuildpack.Dependency{Name: "r", Version: ver}, filepath.Join(s.Stager.DepDir(), "r")); err != nil {
		return err
	}

	if err := s.Stager.LinkDirectoryInDepDir(filepath.Join(s.Stager.DepDir(), "r", "bin"), "bin"); err != nil {
		return err
	}
	return s.Stager.LinkDirectoryInDepDir(filepath.Join(s.Stager.DepDir(), "r", "lib"), "lib")
}
