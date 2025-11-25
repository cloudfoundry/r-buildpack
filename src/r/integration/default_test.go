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
					WithBuildpacks("r_buildpack").
					Execute(name, filepath.Join(fixtures, "default"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(logs).Should(
					ContainLines(MatchRegexp(`Installing r [\d\.]+`)),
				)

				Eventually(func() string {
					cmd := exec.Command("docker", "container", "logs", deployment.Name)
					output, err := cmd.CombinedOutput()
					Expect(err).NotTo(HaveOccurred())
					return string(output)
				}).Should(SatisfyAll(
					ContainSubstring("R program running"),
					ContainSubstring("[1] 16"),
				),
				)
			})
		})

		context("app that requires fortran support", func() {
			it("builds and runs the app", func() {
				deployment, logs, err := platform.Deploy.
					WithBuildpacks("r_buildpack").
					Execute(name, filepath.Join(fixtures, "fortran_required"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(logs).Should(SatisfyAll(
					ContainLines(MatchRegexp(`Installing r [\d\.]+`)),
					ContainSubstring("package 'hexbin' successfully unpacked and MD5 sums checked"),
				))

				Eventually(func() string {
					cmd := exec.Command("docker", "container", "logs", deployment.Name)
					output, err := cmd.CombinedOutput()
					Expect(err).NotTo(HaveOccurred())
					return string(output)
				}).Should(SatisfyAll(
					ContainSubstring("R program running with fortran"),
					ContainSubstring("[1] 64"),
					Not(MatchRegexp("installation of package .* had non-zero exit status")),
				))
			})
		})

		context("shiny web app", func() {
			it("builds and runs the app", func() {
				deployment, _, err := platform.Deploy.
					WithBuildpacks("r_buildpack").
					Execute(name, filepath.Join(fixtures, "shiny"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(deployment).Should(Serve(ContainSubstring("<title>Hello Shiny!</title>")))

				Eventually(func() string {
					cmd := exec.Command("docker", "container", "logs", deployment.Name)
					output, err := cmd.CombinedOutput()
					Expect(err).NotTo(HaveOccurred())
					return string(output)
				}).Should(
					ContainSubstring("library(shiny)"),
				)
			})
		})

		context("R app that requires plumber", func() {
			it("builds and runs the app", func() {
				deployment, _, err := platform.Deploy.
					WithBuildpacks("r_buildpack").
					Execute(name, filepath.Join(fixtures, "plumber"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(deployment).Should(Serve(
					ContainSubstring(`{"msg":["The message is: ''"]}`),
				))

				Eventually(func() string {
					cmd := exec.Command("docker", "container", "logs", deployment.Name)
					output, err := cmd.CombinedOutput()
					Expect(err).NotTo(HaveOccurred())
					return string(output)
				}).Should(
					ContainSubstring("library(plumber)"),
				)
			})
		})
	}
}
