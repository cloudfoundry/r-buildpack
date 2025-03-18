package integration_test

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
)

func testDefault(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
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

		context("default simple R app", func() {
			it("builds and runs the app", func() {
				deployment, logs, err := platform.Deploy.
					Execute(name, filepath.Join(fixtures, "default"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(MatchRegexp(`Installing r [\d\.]+`)), logs.String())

				// model for next tests
				// Eventually(deployment).Should(Serve(ContainSubstring("XXX")))

				Eventually(func() string {
					cmd := exec.Command("docker", "container", "logs", deployment.Name)
					output, err := cmd.CombinedOutput()
					Expect(err).NotTo(HaveOccurred())
					return string(output)
				}).Should(
					And(
						ContainSubstring("R program running"),
						ContainSubstring("[1] 16"),
					),
				)
			})
		})
	}
}
