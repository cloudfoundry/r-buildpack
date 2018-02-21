package supply_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"r/supply"

	"bytes"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/golang/mock/gomock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -source=supply.go --destination=mocks_test.go --package=supply_test

var _ = Describe("Supply", func() {
	var (
		depDir       string
		supplier     *supply.Supplier
		logger       *libbuildpack.Logger
		mockCtrl     *gomock.Controller
		mockStager   *MockStager
		mockManifest *MockManifest
		mockCommand  *MockCommand
		buffer       *bytes.Buffer
	)

	BeforeEach(func() {
		var err error
		buffer = new(bytes.Buffer)
		logger = libbuildpack.NewLogger(buffer)

		mockCtrl = gomock.NewController(GinkgoT())
		mockStager = NewMockStager(mockCtrl)
		mockManifest = NewMockManifest(mockCtrl)
		mockCommand = NewMockCommand(mockCtrl)
		depDir, err = ioutil.TempDir("", "r.depdir")
		Expect(err).ToNot(HaveOccurred())
		mockStager.EXPECT().DepDir().AnyTimes().Return(depDir)
		supplier = supply.New(mockStager, mockCommand, mockManifest, logger)
	})

	AfterEach(func() {
		mockCtrl.Finish()
		os.RemoveAll(depDir)
	})

	Describe("InstallR", func() {
		It("installs and links r", func() {
			mockManifest.EXPECT().AllDependencyVersions("r").Return([]string{"3.4.3"})
			mockManifest.EXPECT().InstallDependency(libbuildpack.Dependency{Name: "r", Version: "3.4.3"}, filepath.Join(depDir, "r"))
			mockStager.EXPECT().LinkDirectoryInDepDir(filepath.Join(depDir, "r", "bin"), "bin")
			mockStager.EXPECT().LinkDirectoryInDepDir(filepath.Join(depDir, "r", "lib"), "lib")

			Expect(supplier.InstallR()).To(Succeed())
		})
	})

	Describe("InstallPackages", func() {
		Context("There's a reasonable package name", func() {
			It("Suceeds", func() {
				mockStager.EXPECT().DepsDir().Return("/deps/dir")
				mockStager.EXPECT().BuildDir().Return("/build/dir")
				mockCommand.EXPECT().Run(gomock.Any()).Do(func(cmd *exec.Cmd) {
					Expect(cmd.Args).To(Equal([]string{
						"R",
						"--vanilla",
						"-e",
						"install.packages(c(\"good.PACKAGE.name1\"), repos=\"https://good.cran.mirror\", dependencies=TRUE)\n",
					}))
					Expect(cmd.Dir).To(Equal("/build/dir"))
					Expect(cmd.Env).To(ContainElement("DEPS_DIR=/deps/dir"))
				})
				Expect(supplier.InstallPackages(
					supply.Packages{
						[]supply.Source{
							supply.Source{
								CranMirror: "https://good.cran.mirror",
								Packages: []supply.Package{
									supply.Package{Name: "good.PACKAGE.name1"},
								}},
						}})).To(Succeed())
			})
		})
		Context("There's a malformed package name", func() {
			It("Returns an error", func() {
				Expect(supplier.InstallPackages(
					supply.Packages{
						[]supply.Source{
							supply.Source{
								CranMirror: "https://good.cran.mirror",
								Packages: []supply.Package{
									supply.Package{Name: `bad"package"name`},
								}},
						}})).ToNot(Succeed())
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
