package integration_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
	"golang.org/x/net/websocket"
)

func testRServe(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
	return func(t *testing.T, context spec.G, it spec.S) {
		var (
			Expect     = NewWithT(t).Expect
			Eventually = NewWithT(t).Eventually
			name       string
		)

		it.Before(func() {
			var err error
			name, err = switchblade.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(platform.Delete.Execute(name)).To(Succeed())
		})

		context("Rserve TCP server app", func() {
			it("builds, runs and rserve websocket is reachable", func() {
				deployment, logs, err := platform.Deploy.
					Execute(name, filepath.Join(fixtures, "rserve"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(MatchRegexp(`Installing r [\d\.]+`)), logs.String())

				origin := deployment.ExternalURL
				url := strings.Replace(origin, "http://", "ws://", 1)

				var ws *websocket.Conn
				Eventually(func() error {
					ws, err = websocket.Dial(url, "", origin)
					return err
				}).Should(Succeed())

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
	}
}
