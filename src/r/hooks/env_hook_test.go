package hooks_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/ansicleaner"
	"github.com/cloudfoundry/r-buildpack/src/r/hooks"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EnvHook", func() {
	var (
		err      error
		buildDir string
		stager   *libbuildpack.Stager
		hook     libbuildpack.Hook
		buffer   *bytes.Buffer
	)

	BeforeEach(func() {
		buildDir, err = ioutil.TempDir("", "r-buildpack.build.")
		Expect(err).To(BeNil())

		buffer = new(bytes.Buffer)
		logger := libbuildpack.NewLogger(ansicleaner.New(buffer))

		args := []string{buildDir, "", "/tmp/not-exist", "9"}
		stager = libbuildpack.NewStager(args, logger, &libbuildpack.Manifest{})

		hook = &hooks.EnvHook{}
	})

	AfterEach(func() {
		Expect(os.RemoveAll(buildDir)).To(Succeed())
	})

	Describe("BeforeCompile", func() {
		BeforeEach(func() {
			Expect(os.Unsetenv("SOME_VAR")).To(Succeed())

			Expect(ioutil.WriteFile(filepath.Join(buildDir, "r.env.sh"), []byte(`#!/bin/bash
export SOME_VAR=some-value
`), 0755)).To(Succeed())
		})

		AfterEach(func() {
			Expect(os.Unsetenv("SOME_VAR")).To(Succeed())
		})

		It("executes the r.env.sh script inherits the environment", func() {
			Expect(hook.BeforeCompile(stager)).To(Succeed())
			Expect(os.Getenv("SOME_VAR")).To(Equal("some-value"))
		})

		Context("failures cases", func() {
			Context("when the script execution fails", func() {
				BeforeEach(func() {
					Expect(ioutil.WriteFile(filepath.Join(buildDir, "r.env.sh"), []byte(`#!/bin/bash
echo "oh no"
exit 1
`), 0755)).To(Succeed())
				})

				It("returns an error", func() {
					Expect(hook.BeforeCompile(stager)).To(MatchError(ContainSubstring("oh no")))
					Expect(hook.BeforeCompile(stager)).To(MatchError(ContainSubstring("exit status 1")))
				})
			})
		})
	})
})
