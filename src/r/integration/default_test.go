package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
			println(name)
		})

		// it.After(func() {
		// 	Expect(platform.Delete.Execute(name)).To(Succeed())
		// })

		context("default simple R app", func() {
			it("builds and runs the app", func() {
				deployment, logs, err := platform.Deploy.
					WithBuildpacks("r_buildpack").
					WithHealthCheckType("process").
					Execute(name, filepath.Join(fixtures, "default"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(logs).Should(
					ContainLines(MatchRegexp(`Installing r [\d\.]+`)),
				)

				// Verify runtime output - platform-specific approach
				platformType := strings.ToLower(os.Getenv("SWITCHBLADE_PLATFORM"))
				if platformType == "docker" {
					// Docker: check container logs directly
					Eventually(func() string {
						cmd := exec.Command("docker", "container", "logs", deployment.Name)
						output, err := cmd.CombinedOutput()
						if err != nil {
							return ""
						}
						return string(output)
					}).Should(SatisfyAll(
						ContainSubstring("R program running"),
						ContainSubstring("[1] 16"),
					))
				} else if platformType == "cf" {
					// CF: retry cf logs --recent with delays for log aggregation
					time.Sleep(5 * time.Second) // Initial delay
					Eventually(func() string {
						cmd := exec.Command("cf", "logs", name, "--recent")
						output, err := cmd.CombinedOutput()
						if err != nil {
							return ""
						}
						return string(output)
					}, "120s", "5s").Should(SatisfyAll(
						ContainSubstring("R program running"),
						ContainSubstring("[1] 16"),
					))
				}
			})
		})

		context("app that requires fortran support", func() {
			it("builds and runs the app", func() {
				deployment, logs, err := platform.Deploy.
					WithBuildpacks("r_buildpack").
					WithHealthCheckType("process").
					Execute(name, filepath.Join(fixtures, "fortran_required"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(logs).Should(SatisfyAll(
					ContainLines(MatchRegexp(`Installing r [\d\.]+`)),
					ContainSubstring("package 'hexbin' successfully unpacked and MD5 sums checked"),
					Not(MatchRegexp("installation of package .* had non-zero exit status")),
				))

				// Verify runtime output - platform-specific approach
				platformType := strings.ToLower(os.Getenv("SWITCHBLADE_PLATFORM"))
				if platformType == "docker" {
					// Docker: check container logs directly
					Eventually(func() string {
						cmd := exec.Command("docker", "container", "logs", deployment.Name)
						output, err := cmd.CombinedOutput()
						if err != nil {
							return ""
						}
						return string(output)
					}).Should(SatisfyAll(
						ContainSubstring("R program running with fortran"),
						ContainSubstring("[1] 64"),
						Not(MatchRegexp("installation of package .* had non-zero exit status")),
					))
				} else if platformType == "cf" {
					// CF: retry cf logs --recent with delays for log aggregation
					time.Sleep(5 * time.Second) // Initial delay
					Eventually(func() string {
						cmd := exec.Command("cf", "logs", name, "--recent")
						output, err := cmd.CombinedOutput()
						if err != nil {
							return ""
						}
						return string(output)
					}, "120s", "5s").Should(SatisfyAll(
						ContainSubstring("R program running with fortran"),
						ContainSubstring("[1] 64"),
						Not(MatchRegexp("installation of package .* had non-zero exit status")),
					))
				}
			})
		})

		// context("shiny web app", func() {
		// 	it("builds and runs the app", func() {
		// 		deployment, _, err := platform.Deploy.
		// 			WithBuildpacks("r_buildpack").
		// 			Execute(name, filepath.Join(fixtures, "shiny"))
		// 		Expect(err).NotTo(HaveOccurred())

		// 		Eventually(deployment).Should(Serve(ContainSubstring("<title>Hello Shiny!</title>")))
		// 	})
		// })

		// context("R app that requires plumber", func() {
		// 	it("builds and runs the app", func() {
		// 		deployment, _, err := platform.Deploy.
		// 			WithBuildpacks("r_buildpack").
		// 			WithHealthCheckType("process").
		// 			Execute(name, filepath.Join(fixtures, "plumber"))
		// 		Expect(err).NotTo(HaveOccurred())

		// 		Eventually(deployment).Should(Serve(
		// 			ContainSubstring(`{"msg":["The message is: ''"]}`),
		// 		))

		// 		Eventually(func() string {
		// 			cmd := exec.Command("docker", "container", "logs", deployment.Name)
		// 			output, err := cmd.CombinedOutput()
		// 			Expect(err).NotTo(HaveOccurred())
		// 			return string(output)
		// 		}).Should(
		// 			ContainSubstring("library(plumber)"),
		// 		)
		// 	})
		// })
	}
}
