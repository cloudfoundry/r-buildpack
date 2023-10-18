package supply_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/r-buildpack/src/r/supply"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -source=supply.go --destination=mocks_test.go --package=supply_test

var _ = Describe("Supply", func() {
	var (
		depDir        string
		buildDir      string
		supplier      *supply.Supplier
		logger        *libbuildpack.Logger
		mockCtrl      *gomock.Controller
		mockStager    *MockStager
		mockManifest  *MockManifest
		mockInstaller *MockInstaller
		mockCommand   *MockCommand
		buffer        *bytes.Buffer
	)

	BeforeEach(func() {
		var err error
		buffer = new(bytes.Buffer)
		logger = libbuildpack.NewLogger(buffer)

		mockCtrl = gomock.NewController(GinkgoT())
		mockStager = NewMockStager(mockCtrl)
		mockManifest = NewMockManifest(mockCtrl)
		mockInstaller = NewMockInstaller(mockCtrl)
		mockCommand = NewMockCommand(mockCtrl)

		depDir, err = ioutil.TempDir("", "r.depdir")
		Expect(err).ToNot(HaveOccurred())

		buildDir, err = ioutil.TempDir("", "r.builddir")
		Expect(err).ToNot(HaveOccurred())

		mockStager.EXPECT().DepDir().AnyTimes().Return(depDir)
		mockStager.EXPECT().BuildDir().AnyTimes().Return(buildDir)

		supplier = supply.New(mockStager, mockCommand, mockManifest, mockInstaller, logger)
	})

	AfterEach(func() {
		mockCtrl.Finish()
		os.RemoveAll(depDir)
	})

	Describe("InstallR", func() {
		Context("A version of R is specified in buildpacks.yml", func() {
			const version string = "1.2.3"

			It("Installs that version of R", func() {
				buildpackYAMLString := fmt.Sprintf("r:\n  version: %s", version)
				Expect(ioutil.WriteFile(filepath.Join(buildDir, "buildpack.yml"), []byte(buildpackYAMLString), 0666)).To(Succeed())

				mockManifest.EXPECT().AllDependencyVersions("r").Return([]string{version, "3.4.3"})
				mockInstaller.EXPECT().InstallDependency(libbuildpack.Dependency{Name: "r", Version: version}, filepath.Join(depDir, "r"))
				mockStager.EXPECT().LinkDirectoryInDepDir(filepath.Join(depDir, "r", "bin"), "bin")
				mockStager.EXPECT().LinkDirectoryInDepDir(filepath.Join(depDir, "r", "lib"), "lib")

				Expect(supplier.InstallR()).To(Succeed())
			})
		})

		Context("A version of R is NOT specified in buildpacks.yml", func() {
			It("Installs that default version of R", func() {
				mockManifest.EXPECT().AllDependencyVersions("r").Return([]string{"1.2.3", "3.4.3"})
				mockInstaller.EXPECT().InstallDependency(libbuildpack.Dependency{Name: "r", Version: "3.4.3"}, filepath.Join(depDir, "r"))
				mockStager.EXPECT().LinkDirectoryInDepDir(filepath.Join(depDir, "r", "bin"), "bin")
				mockStager.EXPECT().LinkDirectoryInDepDir(filepath.Join(depDir, "r", "lib"), "lib")

				Expect(supplier.InstallR()).To(Succeed())
			})
		})
	})

	Describe("InstallPackages", func() {
		Context("There's a reasonable package name", func() {
			It("Suceeds", func() {
				mockStager.EXPECT().DepsDir().Return("/deps/dir")
				mockCommand.EXPECT().Run(gomock.Any()).Do(func(cmd *exec.Cmd) {
					Expect(cmd.Args).To(Equal([]string{
						"R",
						"--vanilla",
						"-e",
						"install.packages(c(\"good.PACKAGE.name1\"), repos=\"https://good.cran.mirror\", dependencies=TRUE, Ncpus=0)\n",
					}))
					Expect(cmd.Dir).To(Equal(buildDir))
					Expect(cmd.Env).To(ContainElement("DEPS_DIR=/deps/dir"))
				})
				Expect(supplier.InstallPackages(
					supply.Packages{
						[]supply.Source{
							{
								CranMirror: "https://good.cran.mirror",
								Packages: []supply.Package{
									{Name: "good.PACKAGE.name1"},
								}},
						}})).To(Succeed())
			})
		})
		Context("There's a malformed package name", func() {
			It("Returns an error", func() {
				Expect(supplier.InstallPackages(
					supply.Packages{
						[]supply.Source{
							{
								CranMirror: "https://good.cran.mirror",
								Packages: []supply.Package{
									{Name: `bad"package"name`},
								}},
						}})).ToNot(Succeed())
			})
		})
		Context("The dependencies argument is provided", func() {
			It("Succeeds", func() {
				mockStager.EXPECT().DepsDir().Return("/deps/dir")
				mockCommand.EXPECT().Run(gomock.Any()).Do(func(cmd *exec.Cmd) {
					Expect(cmd.Args).To(Equal([]string{
						"R",
						"--vanilla",
						"-e",
						"install.packages(c(\"good.PACKAGE.name1\"), repos=\"https://good.cran.mirror\", dependencies=c(\"Depends\", \"Imports\"), Ncpus=0)\n",
					}))
				})
				Expect(supplier.InstallPackages(
					supply.Packages{
						[]supply.Source{
							{
								CranMirror: "https://good.cran.mirror",
								Dependencies: []string{"Depends", "Imports"},
								Packages: []supply.Package{
									{Name: "good.PACKAGE.name1"},
								}},
						}})).To(Succeed())
			})
		})
	})
	Describe("RewriteRHome", func() {
		BeforeEach(func() {
			Expect(os.MkdirAll(filepath.Join(depDir, "r", "bin"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(depDir, "r", "bin", "R"), []byte(`#!/bin/bash
# Shell wrapper for R executable.

export R_HOME_DIR=/usr/local/lib/R
export R_SHARE_DIR=/usr/local/lib/R/share
export R_INCLUDE_DIR=/usr/local/lib/R/include
			`), 0755)).To(Succeed())

			mockStager.EXPECT().DepsIdx().AnyTimes().Return("3")
		})

		It("replaces compiled prefix dir with runtime installed dir", func() {
			Expect(supplier.RewriteRHome()).To(Succeed())

			body, err := ioutil.ReadFile(filepath.Join(depDir, "r", "bin", "R"))
			Expect(err).ToNot(HaveOccurred())
			Expect(string(body)).To(Equal(`#!/bin/bash
# Shell wrapper for R executable.

export R_HOME_DIR=$DEPS_DIR/3/r
export R_SHARE_DIR=$DEPS_DIR/3/r/share
export R_INCLUDE_DIR=$DEPS_DIR/3/r/include
			`))
		})
	})
})
