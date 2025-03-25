package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
)

func testPushAppSecondTime(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
	return func(t *testing.T, context spec.G, it spec.S) {
		var (
			Expect     = NewWithT(t).Expect
			Eventually = NewWithT(t).Eventually

			name   string
			source string
		)

		it.Before(func() {
			var err error
			name, err = switchblade.RandomName()
			Expect(err).NotTo(HaveOccurred())

			source, err = switchblade.Source(filepath.Join(fixtures, "default"))
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(platform.Delete.Execute(name)).To(Succeed())
		})

		Regexp := `.*(linux_noarch_cflinuxfs4_.*-)?[\da-f]+\.tgz`
		DownloadRegexp := "Download " + Regexp
		CopyRegexp := "Copy " + Regexp

		context("push an app twice", func() {
			it("uses the cache for manifest dependencies", func() {
				_, logs, err := platform.Deploy.
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred())
				Eventually(logs).Should(SatisfyAll(
					ContainLines(MatchRegexp(DownloadRegexp)),
					Not(ContainLines(MatchRegexp(CopyRegexp))),
				))

				_, logs, err = platform.Deploy.
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred())
				Eventually(logs).Should(SatisfyAll(
					Not(ContainLines(MatchRegexp(DownloadRegexp))),
					ContainLines(MatchRegexp(CopyRegexp)),
				))
			})
		})
	}
}
