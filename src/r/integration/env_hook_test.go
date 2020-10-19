package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Env Hook", func() {
	var app *cutlass.App

	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	Context("with a simple R app", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "env_hook"))
			app.Disk = "1G"
			Expect(app.PushNoStart()).To(Succeed())
		})

		It("logs that the env hook is running", func() {
			RunCF("set-health-check", app.Name, "process")
			Expect(app.Restart()).To(Succeed())
			Expect(app.ConfirmBuildpack(buildpackVersion)).To(Succeed())

			Eventually(app.Stdout.String).Should(ContainSubstring("-----> Setting up R environment using r.env.sh"))
		})
	})
})
