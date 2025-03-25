package integration_test

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testInstallPackages(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
	return func(t *testing.T, context spec.G, it spec.S) {
		var (
			Expect     = NewWithT(t).Expect
			Eventually = NewWithT(t).Eventually

			name string
		)

		it.Before(func() {
			var err error
			name, err = switchblade.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(platform.Delete.Execute(name)).To(Succeed())
		})

		context("stringr package", func() {
			it("builds and runs", func() {
				deployment, _, err := platform.Deploy.
					Execute(name, filepath.Join(fixtures, "simple_package"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() string {
					cmd := exec.Command("docker", "container", "logs", deployment.Name)
					output, err := cmd.CombinedOutput()
					Expect(err).NotTo(HaveOccurred())
					return string(output)
				}).Should(SatisfyAll(
					ContainSubstring("R program running"),
					ContainSubstring("HELLO WORLD"),
				))
			})
		})

		context("packages vendored", func() {
			it("builds both packages parallely", func() {
				deployment, logs, err := platform.Deploy.
					Execute(name, filepath.Join(fixtures, "packages_vendored"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(logs).Should(SatisfyAll(
					ContainSubstring("Ncpus=2"),
					ContainSubstring("begin installing package stringr"),
					ContainSubstring("begin installing package jsonlite"),
					ContainSubstring("Cleaning up vendored packages"),
				))

				Eventually(func() string {
					cmd := exec.Command("docker", "container", "logs", deployment.Name)
					output, err := cmd.CombinedOutput()
					Expect(err).NotTo(HaveOccurred())
					return string(output)
				}).Should(SatisfyAll(
					ContainSubstring("STRINGR INSTALLED SUCCESSFULLY"),
					ContainSubstring("{\"jsonlite\":\"installed\""),
				))
			})
		})

		context("source missing for stringr", func() {
			it("fails", func() {
				_, logs, err := platform.Deploy.
					Execute(name, filepath.Join(fixtures, "simple_package_nosource"))
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("App staging failed")))

				Eventually(logs).Should(SatisfyAll(
					ContainSubstring("No source found for installing packages"),
					Not(ContainSubstring("STRINGR INSTALLED SUCCESSFULLY")),
				))
			})
		})

		context("R app that needs the Rscript bin for installation", func() {
			it("builds and runs", func() {
				deployment, _, err := platform.Deploy.
					Execute(name, filepath.Join(fixtures, "install_uses_rscript"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() string {
					cmd := exec.Command("docker", "container", "logs", deployment.Name)
					output, err := cmd.CombinedOutput()
					Expect(err).NotTo(HaveOccurred())
					return string(output)
				}).Should(SatisfyAll(
					ContainSubstring("R program running"),
					ContainSubstring("HELLO WORLD"),
				))
			})
		})
	}
}
