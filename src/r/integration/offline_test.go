package integration_test

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testOffline(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
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

		context("offline build", func() {
			it("builds both vendored packages parallely", func() {
				deployment, logs, err := platform.Deploy.
					WithoutInternetAccess().
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
	}
}
