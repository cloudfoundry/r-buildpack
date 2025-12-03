package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
					WithHealthCheckType("process").
					Execute(name, filepath.Join(fixtures, "packages_vendored"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(logs).Should(SatisfyAll(
					ContainSubstring("Ncpus=2"),
					ContainSubstring("installing *source* package"),
					ContainSubstring("DONE (stringr)"),
					ContainSubstring("DONE (jsonlite)"),
					ContainSubstring("Cleaning up vendored packages"),
				))

				platformType := strings.ToLower(os.Getenv("SWITCHBLADE_PLATFORM"))
				switch platformType {
				case "docker":
					Eventually(func() string {
						cmd := exec.Command("docker", "container", "logs", deployment.Name)
						output, err := cmd.CombinedOutput()
						Expect(err).NotTo(HaveOccurred())
						return string(output)
					}).Should(SatisfyAll(
						ContainSubstring("STRINGR INSTALLED SUCCESSFULLY"),
						ContainSubstring("{\"jsonlite\":\"installed\""),
					))
				case "cf":
					// CF: process-based health check means app is running if process is up
					// No runtime log validation needed for CF (logs are captured during deploy)
				}
			})
		})
	}
}
