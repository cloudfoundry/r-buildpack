package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
)

func testEnvHook(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
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

		context("R app with an env hook", func() {
			it("builds and runs the hook", func() {
				_, logs, err := platform.Deploy.
					WithBuildpacks("r_buildpack").
					WithHealthCheckType("process").
					Execute(name, filepath.Join(fixtures, "env_hook"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(logs).Should(
					ContainLines(MatchRegexp("-----> Setting up R environment using r.env.sh")), logs.String(),
				)
			})
		})
	}
}
