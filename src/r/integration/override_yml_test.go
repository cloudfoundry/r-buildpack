package integration_test

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testOverrideYml(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
	return func(t *testing.T, context spec.G, it spec.S) {
		var (
			Expect = NewWithT(t).Expect

			name string
		)

		it.Before(func() {
			var err error
			name, err = switchblade.RandomName()
			Expect(err).NotTo(HaveOccurred())
			println(name)
		})

		it.After(func() {
			Expect(platform.Delete.Execute(name)).To(Succeed())
		})

		context("override final buildpack", func() {
			it("Forces R from override buildpack", func() {
				_, logs, err := platform.Deploy.
					WithBuildpacks(
						"override_buildpack",
						"r_buildpack",
					).
					Execute(name, filepath.Join(fixtures, "default"))
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("App staging failed")))

				// In CF API v3, staging logs are not in the logs buffer when staging fails.
				// We need to target the app's org/space and fetch them using `cf logs --recent`.
				// Switchblade creates a new org/space with the same name as the app.
				var recentLogsStr string
				for i := 0; i < 3; i++ {
					targetCmd := exec.Command("cf", "target", "-o", name, "-s", name)
					_ = targetCmd.Run()

					recentLogs := bytes.NewBuffer(nil)
					cmd := exec.Command("cf", "logs", name, "--recent")
					cmd.Stdout = recentLogs
					cmd.Stderr = recentLogs
					_ = cmd.Run()

					recentLogsStr = recentLogs.String()
					if recentLogsStr != "" && !bytes.Contains(recentLogs.Bytes(), []byte("not found")) {
						break
					}
					time.Sleep(100 * time.Millisecond)
				}

				// Verify staging logs contain expected strings
				Expect(recentLogsStr).To(ContainSubstring("-----> OverrideYML Buildpack"))
				Expect(recentLogsStr).To(ContainSubstring("-----> Installing r"))
				Expect(recentLogsStr).To(MatchRegexp(`Copy .*/r.tgz`))
				Expect(recentLogsStr).To(ContainSubstring("Error installing R: dependency sha256 mismatch: expected sha256 062d906c87839d03b243e2821e10653c89b4c92878bfe2bf995dec231e117bfc, actual sha256 b56b58ac21f9f42d032e1e4b8bf8b8823e69af5411caa15aee2b140bc756962f"))

				_ = logs // Original logs buffer only contains setup output
			})
		})
	}
}
