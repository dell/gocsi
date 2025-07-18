package rpcs_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/dell/gocsi/utils/rpcs"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

func TestRpcsUtils(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	log.SetOutput(ginkgo.GinkgoWriter)
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "RPCS Utils Suite")
}

var _ = ginkgo.Describe("ParseMethod", func() {
	var (
		err        error
		version    int32
		service    string
		methodName string
	)
	ginkgo.BeforeEach(func() {
		version, service, methodName, err = rpcs.ParseMethod(
			ginkgo.CurrentGinkgoTestDescription().ComponentTexts[1])
	})
	ginkgo.It("/csi.v0.Identity/GetPluginInfo", func() {
		gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
		gomega.Ω(version).Should(gomega.Equal(int32(0)))
		gomega.Ω(service).Should(gomega.Equal("Identity"))
		gomega.Ω(methodName).Should(gomega.Equal("GetPluginInfo"))
	})
	ginkgo.It("/csi.v1.Identity/GetPluginInfo", func() {
		gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
		gomega.Ω(version).Should(gomega.Equal(int32(1)))
		gomega.Ω(service).Should(gomega.Equal("Identity"))
		gomega.Ω(methodName).Should(gomega.Equal("GetPluginInfo"))
	})
	ginkgo.It("/csi.v1.Node/NodePublishVolume", func() {
		gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
		gomega.Ω(version).Should(gomega.Equal(int32(1)))
		gomega.Ω(service).Should(gomega.Equal("Node"))
		gomega.Ω(methodName).Should(gomega.Equal("NodePublishVolume"))
	})
	ginkgo.It("/csi.v1-rc1.Node/NodePublishVolume", func() {
		gomega.Ω(err).Should(gomega.HaveOccurred())
		gomega.Ω(err.Error()).Should(gomega.Equal(fmt.Sprintf("ParseMethod: invalid: %s",
			ginkgo.CurrentGinkgoTestDescription().ComponentTexts[1])))
	})
	ginkgo.It("/csi.v1.Node", func() {
		gomega.Ω(err).Should(gomega.HaveOccurred())
		gomega.Ω(err.Error()).Should(gomega.Equal(fmt.Sprintf("ParseMethod: invalid: %s",
			ginkgo.CurrentGinkgoTestDescription().ComponentTexts[1])))
	})
	ginkgo.It(fmt.Sprintf("/csi.v%d.Node/NodePublishVolume", math.MaxInt64), func() {
		gomega.Ω(err).Should(gomega.HaveOccurred())
		gomega.Ω(err.Error()).Should(gomega.Equal(fmt.Sprintf(
			`ParseMethod: strconv.ParseInt: `+
				`parsing "%d": value out of range`, math.MaxInt64)))
	})
})
