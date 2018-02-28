package integration_test

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/net/websocket"
)

var _ = Describe("CF R Buildpack", func() {
	var app *cutlass.App

	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	Describe("R as a supply buildpack", func() {
		BeforeEach(func() {
			if !ApiHasMultiBuildpack() {
				Skip("Multi buildpack support is required")
			}
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "rserve_supply"))
			app.Buildpacks = []string{"r_buildpack", "python_buildpack"}
			app.Disk = "512M"
		})

		It("pythons uses Rserve", func() {
			PushAppAndConfirm(app)

			Expect(app.GetBody("/")).To(ContainSubstring("two(9) == 18.0"))
		})
	})

	Describe("R as a final buildpack", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "rserve"))
		})

		It("rserve websocket can be reached", func() {
			PushAppAndConfirm(app)

			origin, err := app.GetUrl("")
			Expect(err).ToNot(HaveOccurred())
			url := strings.Replace(origin, "http://", "ws://", 1) // Handle https as well
			ws, err := websocket.Dial(url, "", origin)
			if err != nil {
				log.Fatal(err)
			}

			msg := make([]byte, 512)
			_, err = ws.Read(msg)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(msg)).To(ContainSubstring("Rsrv0103QAP1"))

			// Send 7 * 2 ;; See https://github.com/jakutis/rserve-client/blob/master/lib/rserve.js # mkp_str
			message := []byte{0x03, 0x00, 0x00, 0x00, 0x0c, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0x06, 0x00, 0x00, 0x37, 0x20, 0x2a, 0x20, 0x32, 0x00, 0x01, 0x01}
			_, err = ws.Write(message)
			Expect(err).ToNot(HaveOccurred())

			msg = make([]byte, 512)
			_, err = ws.Read(msg)
			Expect(err).ToNot(HaveOccurred())

			// Returned packet with 14.0 as the response
			expect := []byte{1, 0, 1, 0, 16, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 12, 0, 0, 33, 8, 0, 0, 0, 0, 0, 0, 0, 0, 44, 64}
			Expect(msg[0:len(expect)]).To(Equal(expect))
		})
	})
})
