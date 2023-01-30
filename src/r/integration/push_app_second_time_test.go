package integration_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("pushing an app a second time", func() {
	var app *cutlass.App
	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	BeforeEach(func() {
		if cutlass.Cached {
			Skip("but running cached tests")
		}

		app = cutlass.New(filepath.Join(bpDir, "fixtures", "simple"))
		app.Buildpacks = []string{"r_buildpack"}
	})

	Regexp := fmt.Sprintf(`.*(linux_noarch_%s_.*-)?[\da-f]+\.tgz`, os.Getenv("CF_STACK"))
	DownloadRegexp := "Download " + Regexp
	CopyRegexp := "Copy " + Regexp

	It("uses the cache for manifest dependencies", func() {
		app.Disk = "2G"
		app.HealthCheck = "process"
		Expect(app.Push()).To(Succeed())
		Expect(app.ConfirmBuildpack(buildpackVersion)).To(Succeed())
		Expect(app.Stdout.String()).To(MatchRegexp(DownloadRegexp))
		Expect(app.Stdout.String()).ToNot(MatchRegexp(CopyRegexp))

		app.Stdout.Reset()
		Expect(app.Push()).To(Succeed())
		Expect(app.ConfirmBuildpack(buildpackVersion)).To(Succeed())
		Expect(app.Stdout.String()).To(MatchRegexp(CopyRegexp))
		Expect(app.Stdout.String()).ToNot(MatchRegexp(DownloadRegexp))
	})
})
