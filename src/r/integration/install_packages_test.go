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

	Context("with the stringr package", func() {

		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "simple_packages"))
			app.Disk = "1028M"
			app.Memory = "1028M"
		})

		It("Logs R buildpack version", func() {
			Expect(app.Push()).ToNot(Succeed())
			Expect(app.ConfirmBuildpack(buildpackVersion)).To(Succeed())

			Eventually(app.Stdout.String).Should(ContainSubstring("R program running"))
			Eventually(app.Stdout.String).Should(ContainSubstring("HELLO WORLD"))
		})
	})

	Context("with the vendored stringr package", func() {

		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "simple_packages_vendored"))
			app.Disk = "1028M"
			app.Memory = "1028M"
		})

		It("Installs stringr successfully", func() {
			Expect(app.Push()).ToNot(Succeed())
			Expect(app.ConfirmBuildpack(buildpackVersion)).To(Succeed())

			Eventually(app.Stdout.String).Should(ContainSubstring("STRINGR INSTALLED SUCCESSFULLY"))
			Eventually(app.Stdout.String).Should(ContainSubstring("Cleaning up vendored packages"))
		})

		It("Installs stringr and jsonlite parallely", func() {
			Expect(app.Push()).ToNot(Succeed())
			Expect(app.ConfirmBuildpack(buildpackVersion)).To(Succeed())

			Eventually(app.Stdout.String).Should(ContainSubstring("STRINGR INSTALLED SUCCESSFULLY"))
			Eventually(app.Stdout.String).Should(ContainSubstring("{\"jsonlite\":\"installed\""))
			Eventually(app.Stdout.String).Should(ContainSubstring("Ncpus=2"))
			Eventually(app.Stdout.String).Should(MatchRegexp(`begin installing package.+\n.*begin installing package`))
		})

	})

	Context("with the source missing for stringr package", func() {

		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "simple_packages_nosource"))
			app.Disk = "1028M"
			app.Memory = "1028M"
		})

		It("stringr installation fails", func() {
			Expect(app.Push()).ToNot(Succeed())
			Expect(app.ConfirmBuildpack(buildpackVersion)).To(Succeed())

			Eventually(app.Stdout.String).Should(ContainSubstring("No source found for installing packages"))
			Eventually(app.Stdout.String).ShouldNot(ContainSubstring("STRINGR INSTALLED SUCCESSFULLY"))
		})
	})

	Context("with an R app that needs the Rscript bin for installation", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "install_uses_rscript"))
			app.Memory = "2G"
			app.Disk = "2G"
		})

		It("Logs R buildpack version", func() {
			Expect(app.Push()).ToNot(Succeed())
			Expect(app.ConfirmBuildpack(buildpackVersion)).To(Succeed())

			Eventually(app.Stdout.String).Should(ContainSubstring("R program running"))
			Eventually(app.Stdout.String).Should(ContainSubstring("HELLO WORLD"))
		})
	})
})
