package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CF R Buildpack", func() {
	var app *cutlass.App

	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	Context("with a simple R app", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "simple"))
			Expect(app.PushNoStart()).To(Succeed())
		})

		It("Logs R buildpack version", func() {
			RunCF("set-health-check", app.Name, "process")
			Expect(app.Restart()).To(Succeed())
			Expect(app.ConfirmBuildpack(buildpackVersion)).To(Succeed())

			Eventually(app.Stdout.String).Should(ContainSubstring("R program running"))
			Eventually(app.Stdout.String).Should(ContainSubstring("[1] 16"))
		})
	})

	Context("with a simple R app that requires fortran", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "simple_fortran_required"))
			Expect(app.PushNoStart()).To(Succeed())
		})

		It("Logs R buildpack version and does not warn about package installation status", func() {
			RunCF("set-health-check", app.Name, "process")
			Expect(app.Restart()).To(Succeed())
			Expect(app.ConfirmBuildpack(buildpackVersion)).To(Succeed())

			Eventually(app.Stdout.String).Should(ContainSubstring("R program running with fortran"))
			Eventually(app.Stdout.String).Should(ContainSubstring("[1] 64"))

			Expect(app.Stdout.String()).ShouldNot(MatchRegexp("installation of package .* had non-zero exit status"))
		})
	})

	Context("with an R app that requires shiny", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "shiny"))
			Expect(app.PushNoStart()).To(Succeed())
		})

		It("runs without needing to download shiny", func() {
			Expect(app.Restart()).To(Succeed())
			Expect(app.ConfirmBuildpack(buildpackVersion)).To(Succeed())

			Eventually(app.Stdout.String).Should(ContainSubstring("library(shiny)"))
		})
	})

	Context("with an R app that requires plumber", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "plumber"))
			Expect(app.PushNoStart()).To(Succeed())
		})

		It("runs without needing to download plumber", func() {
			Expect(app.Restart()).To(Succeed())
			Expect(app.ConfirmBuildpack(buildpackVersion)).To(Succeed())
			Expect(app.GetBody("/?msg=hello")).To(ContainSubstring(`{"msg":["The message is: 'hello'"]}`))

			Eventually(app.Stdout.String).Should(ContainSubstring("library(plumber)"))
		})
	})
})
