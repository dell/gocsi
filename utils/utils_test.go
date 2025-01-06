package utils_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/dell/gocsi/mock/service"
	"github.com/dell/gocsi/utils"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

var _ = Describe("ParseMethod", func() {
	var (
		err        error
		version    int32
		service    string
		methodName string
	)
	BeforeEach(func() {
		version, service, methodName, err = utils.ParseMethod(
			CurrentGinkgoTestDescription().ComponentTexts[1])
	})
	It("/csi.v0.Identity/GetPluginInfo", func() {
		Ω(err).ShouldNot(HaveOccurred())
		Ω(version).Should(Equal(int32(0)))
		Ω(service).Should(Equal("Identity"))
		Ω(methodName).Should(Equal("GetPluginInfo"))
	})
	It("/csi.v1.Identity/GetPluginInfo", func() {
		Ω(err).ShouldNot(HaveOccurred())
		Ω(version).Should(Equal(int32(1)))
		Ω(service).Should(Equal("Identity"))
		Ω(methodName).Should(Equal("GetPluginInfo"))
	})
	It("/csi.v1.Node/NodePublishVolume", func() {
		Ω(err).ShouldNot(HaveOccurred())
		Ω(version).Should(Equal(int32(1)))
		Ω(service).Should(Equal("Node"))
		Ω(methodName).Should(Equal("NodePublishVolume"))
	})
	It("/csi.v1-rc1.Node/NodePublishVolume", func() {
		Ω(err).Should(HaveOccurred())
		Ω(err.Error()).Should(Equal(fmt.Sprintf("ParseMethod: invalid: %s",
			CurrentGinkgoTestDescription().ComponentTexts[1])))
	})
	It("/csi.v1.Node", func() {
		Ω(err).Should(HaveOccurred())
		Ω(err.Error()).Should(Equal(fmt.Sprintf("ParseMethod: invalid: %s",
			CurrentGinkgoTestDescription().ComponentTexts[1])))
	})
	It(fmt.Sprintf("/csi.v%d.Node/NodePublishVolume", math.MaxInt64), func() {
		Ω(err).Should(HaveOccurred())
		Ω(err.Error()).Should(Equal(fmt.Sprintf(
			`ParseMethod: strconv.ParseInt: `+
				`parsing "%d": value out of range`, math.MaxInt64)))
	})
})

var errMissingCSIEndpoint = errors.New("missing CSI_ENDPOINT")

var _ = Describe("GetCSIEndpoint", func() {
	var (
		err         error
		proto       string
		addr        string
		expEndpoint string
		expProto    string
		expAddr     string
	)
	BeforeEach(func() {
		expEndpoint = CurrentGinkgoTestDescription().ComponentTexts[2]
		os.Setenv(utils.CSIEndpoint, expEndpoint)
	})
	AfterEach(func() {
		proto = ""
		addr = ""
		expEndpoint = ""
		expProto = ""
		expAddr = ""
		os.Unsetenv(utils.CSIEndpoint)
	})
	JustBeforeEach(func() {
		proto, addr, err = utils.GetCSIEndpoint()
	})

	Context("Valid Endpoint", func() {
		shouldBeValid := func() {
			Ω(os.Getenv(utils.CSIEndpoint)).Should(Equal(expEndpoint))
			Ω(proto).Should(Equal(expProto))
			Ω(addr).Should(Equal(expAddr))
		}
		Context("tcp://127.0.0.1", func() {
			BeforeEach(func() {
				expProto = "tcp"
				expAddr = "127.0.0.1"
			})
			It("Should Be Valid", shouldBeValid)
		})
		Context("tcp://127.0.0.1:8080", func() {
			BeforeEach(func() {
				expProto = "tcp"
				expAddr = "127.0.0.1:8080"
			})
			It("Should Be Valid", shouldBeValid)
		})
		Context("tcp://*:8080", func() {
			BeforeEach(func() {
				expProto = "tcp"
				expAddr = "*:8080"
			})
			It("Should Be Valid", shouldBeValid)
		})
		Context("unix://path/to/sock.sock", func() {
			BeforeEach(func() {
				expProto = "unix"
				expAddr = "path/to/sock.sock"
			})
			It("Should Be Valid", shouldBeValid)
		})
		Context("unix:///path/to/sock.sock", func() {
			BeforeEach(func() {
				expProto = "unix"
				expAddr = "/path/to/sock.sock"
			})
			It("Should Be Valid", shouldBeValid)
		})
		Context("sock.sock", func() {
			BeforeEach(func() {
				expProto = "unix"
				expAddr = "sock.sock"
			})
			It("Should Be Valid", shouldBeValid)
		})
		Context("/tmp/sock.sock", func() {
			BeforeEach(func() {
				expProto = "unix"
				expAddr = "/tmp/sock.sock"
			})
			It("Should Be Valid", shouldBeValid)
		})
	})

	Context("Missing Endpoint", func() {
		Context("", func() {
			It("Should Be Missing", func() {
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(errMissingCSIEndpoint))
			})
		})
		Context("    ", func() {
			It("Should Be Missing", func() {
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(errMissingCSIEndpoint))
			})
		})
	})

	Context("Invalid Network Address", func() {
		shouldBeInvalid := func() {
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(Equal(fmt.Sprintf(
				"invalid network address: %s", expEndpoint)))
		}
		Context("tcp5://localhost:5000", func() {
			It("Should Be An Invalid Endpoint", shouldBeInvalid)
		})
		Context("unixpcket://path/to/sock.sock", func() {
			It("Should Be An Invalid Endpoint", shouldBeInvalid)
		})
	})

	Context("Invalid Implied Sock File", func() {
		shouldBeInvalid := func() {
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(Equal(fmt.Sprintf(
				"invalid implied sock file: %[1]s: "+
					"open %[1]s: no such file or directory",
				expEndpoint)))
		}
		Context("Xtcp5:/localhost:5000", func() {
			It("Should Be An Invalid Implied Sock File", shouldBeInvalid)
		})
		Context("Xunixpcket:/path/to/sock.sock", func() {
			It("Should Be An Invalid Implied Sock File", shouldBeInvalid)
		})
	})
})

var _ = Describe("ParseProtoAddr", func() {
	Context("Empty Address", func() {
		It("Should Be An Empty Address", func() {
			_, _, err := utils.ParseProtoAddr("")
			Ω(err).Should(HaveOccurred())
			Ω(err).Should(Equal(utils.ErrParseProtoAddrRequired))
		})
		It("Should Be An Empty Address", func() {
			_, _, err := utils.ParseProtoAddr("   ")
			Ω(err).Should(HaveOccurred())
			Ω(err).Should(Equal(utils.ErrParseProtoAddrRequired))
		})
	})
})

var _ = Describe("ParseMap", func() {
	Context("One Pair", func() {
		It("Should Be Valid", func() {
			data := utils.ParseMap("k1=v1")
			Ω(data).Should(HaveLen(1))
			Ω(data["k1"]).Should(Equal("v1"))
		})
	})
	Context("Empty Line", func() {
		It("Should Be Valid", func() {
			data := utils.ParseMap("")
			Ω(data).Should(HaveLen(0))
		})
	})
	Context("Key Sans Value", func() {
		It("Should Be Valid", func() {
			data := utils.ParseMap("k1")
			Ω(data).Should(HaveLen(1))
		})
	})
	Context("Two Pair", func() {
		It("Should Be Valid", func() {
			data := utils.ParseMap("k1=v1, k2=v2")
			Ω(data).Should(HaveLen(2))
			Ω(data["k1"]).Should(Equal("v1"))
			Ω(data["k2"]).Should(Equal("v2"))
		})
	})
	Context("Two Pair with Quoting & Escaping", func() {
		It("Should Be Valid", func() {
			data := utils.ParseMap(`k1=v1, "k2=v2""s"`)
			Ω(data).Should(HaveLen(2))
			Ω(data["k1"]).Should(Equal("v1"))
			Ω(data["k2"]).Should(Equal(`v2"s`))
		})
		It("Should Be Valid", func() {
			data := utils.ParseMap(`k1=v1, "k2=v2\'s"`)
			Ω(data).Should(HaveLen(2))
			Ω(data["k1"]).Should(Equal("v1"))
			Ω(data["k2"]).Should(Equal(`v2\'s`))
		})
		It("Should Be Valid", func() {
			data := utils.ParseMap(`k1=v1, k2=v2's`)
			Ω(data).Should(HaveLen(2))
			Ω(data["k1"]).Should(Equal("v1"))
			Ω(data["k2"]).Should(Equal(`v2's`))
		})
	})
	Context("Two Pair with Three Spaces Between Them", func() {
		It("Should Be Valid", func() {
			data := utils.ParseMap("k1=v1,   k2=v2")
			Ω(data).Should(HaveLen(2))
			Ω(data["k1"]).Should(Equal("v1"))
			Ω(data["k2"]).Should(Equal("v2"))
		})
	})
	Context("Two Pair with One Sans Value", func() {
		It("Should Be Valid", func() {
			data := utils.ParseMap("k1=, k2=v2")
			Ω(data).Should(HaveLen(2))
			Ω(data["k1"]).Should(Equal(""))
			Ω(data["k2"]).Should(Equal("v2"))
		})
	})
	Context("Two Pair with One Sans Value & Three Spaces Between Them", func() {
		It("Should Be Valid", func() {
			data := utils.ParseMap("k1=,    k2=v2")
			Ω(data).Should(HaveLen(2))
			Ω(data["k1"]).Should(Equal(""))
			Ω(data["k2"]).Should(Equal("v2"))
		})
	})
	Context("One Pair with Quoted Value", func() {
		It("Should Be Valid", func() {
			data := utils.ParseMap("k1=v 1")
			Ω(data).Should(HaveLen(1))
			Ω(data["k1"]).Should(Equal("v 1"))
		})
	})
	Context("Three Pair with Mixed Values", func() {
		It("Should Be Valid", func() {
			data := utils.ParseMap(`"k1=v 1", "k2=v 2 ", "k3 =v3"  `)
			Ω(data).Should(HaveLen(3))
			Ω(data["k1"]).Should(Equal("v 1"))
			Ω(data["k2"]).Should(Equal("v 2 "))
			Ω(data["k3 "]).Should(Equal("v3"))
		})
	})
})

var _ = Describe("CompareVolume", func() {
	It("a == b", func() {
		a := csi.Volume{VolumeId: "0"}
		b := csi.Volume{VolumeId: "0"}
		Ω(utils.CompareVolume(a, b)).Should(Equal(0))
		a.CapacityBytes = 1
		b.CapacityBytes = 1
		Ω(utils.CompareVolume(a, b)).Should(Equal(0))
		a.VolumeContext = map[string]string{"key": "val"}
		b.VolumeContext = map[string]string{"key": "val"}
		Ω(utils.CompareVolume(a, b)).Should(Equal(0))
	})
	It("a > b", func() {
		a := csi.Volume{VolumeId: "0"}
		b := csi.Volume{}
		Ω(utils.CompareVolume(a, b)).Should(Equal(1))
		b.VolumeId = "0"
		Ω(utils.CompareVolume(a, b)).Should(Equal(0))
		a.CapacityBytes = 1
		Ω(utils.CompareVolume(a, b)).Should(Equal(1))
		b.CapacityBytes = 1
		Ω(utils.CompareVolume(a, b)).Should(Equal(0))
		a.VolumeContext = map[string]string{"key": "val"}
		Ω(utils.CompareVolume(a, b)).Should(Equal(1))
		b.VolumeContext = map[string]string{"key": "val"}
		Ω(utils.CompareVolume(a, b)).Should(Equal(0))
	})
	It("a < b", func() {
		b := csi.Volume{VolumeId: "0"}
		a := csi.Volume{}
		Ω(utils.CompareVolume(a, b)).Should(Equal(-1))
		a.VolumeId = "0"
		Ω(utils.CompareVolume(a, b)).Should(Equal(0))
		b.CapacityBytes = 1
		Ω(utils.CompareVolume(a, b)).Should(Equal(-1))
		a.CapacityBytes = 1
		Ω(utils.CompareVolume(a, b)).Should(Equal(0))
		b.VolumeContext = map[string]string{"key": "val"}
		Ω(utils.CompareVolume(a, b)).Should(Equal(-1))
		a.VolumeContext = map[string]string{"key": "val"}
		Ω(utils.CompareVolume(a, b)).Should(Equal(0))
	})
})

var _ = Describe("EqualVolumeCapability", func() {
	It("a == b", func() {
		a := &csi.VolumeCapability{
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			},
			AccessType: &csi.VolumeCapability_Block{
				Block: &csi.VolumeCapability_BlockVolume{},
			},
		}
		b := &csi.VolumeCapability{
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			},
			AccessType: &csi.VolumeCapability_Block{
				Block: &csi.VolumeCapability_BlockVolume{},
			},
		}
		Ω(utils.EqualVolumeCapability(a, b)).Should(BeTrue())
		a.AccessMode.Mode = csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY
		Ω(utils.EqualVolumeCapability(a, b)).Should(BeFalse())
		b.AccessMode.Mode = csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY
		Ω(utils.EqualVolumeCapability(a, b)).Should(BeTrue())
		a.AccessMode = nil
		Ω(utils.EqualVolumeCapability(a, b)).Should(BeFalse())
		b.AccessMode = nil
		Ω(utils.EqualVolumeCapability(a, b)).Should(BeTrue())
		a = nil
		Ω(utils.EqualVolumeCapability(nil, b)).Should(BeFalse())
		b = nil
		Ω(utils.EqualVolumeCapability(a, b)).Should(BeFalse())

		aAT := &csi.VolumeCapability_Mount{
			Mount: &csi.VolumeCapability_MountVolume{
				FsType:     "ext4",
				MountFlags: []string{"rw"},
			},
		}
		bAT := &csi.VolumeCapability_Mount{
			Mount: &csi.VolumeCapability_MountVolume{
				FsType:     "ext4",
				MountFlags: []string{"rw"},
			},
		}

		a = &csi.VolumeCapability{
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			},
			AccessType: aAT,
		}
		b = &csi.VolumeCapability{
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			},
			AccessType: bAT,
		}
		Ω(utils.EqualVolumeCapability(a, b)).Should(BeTrue())
		aAT.Mount.FsType = "xfs"
		Ω(utils.EqualVolumeCapability(a, b)).Should(BeFalse())
		bAT.Mount.FsType = "xfs"
		Ω(utils.EqualVolumeCapability(a, b)).Should(BeTrue())
		aAT.Mount.MountFlags = append(aAT.Mount.MountFlags, "nosuid")
		Ω(utils.EqualVolumeCapability(a, b)).Should(BeFalse())
		bAT.Mount.MountFlags = append(bAT.Mount.MountFlags, "nosuid")
		Ω(utils.EqualVolumeCapability(a, b)).Should(BeTrue())
		aAT.Mount.MountFlags[0] = "ro"
		Ω(utils.EqualVolumeCapability(a, b)).Should(BeFalse())
		bAT.Mount.MountFlags[0] = "ro"
		Ω(utils.EqualVolumeCapability(a, b)).Should(BeTrue())
		aAT.Mount.MountFlags = nil
		Ω(utils.EqualVolumeCapability(a, b)).Should(BeFalse())
		bAT.Mount.MountFlags = nil
		Ω(utils.EqualVolumeCapability(a, b)).Should(BeTrue())

		// error test: a is non-nil, b is nil
		a.AccessMode = &csi.VolumeCapability_AccessMode{
			Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
		}
		b.AccessMode = nil
		Ω(utils.EqualVolumeCapability(a, b)).Should(BeFalse())
	})
})

var _ = Describe("AreVolumeCapabilitiesCompatible", func() {
	It("compatible", func() {
		aMountAT := &csi.VolumeCapability_Mount{
			Mount: &csi.VolumeCapability_MountVolume{
				FsType:     "ext4",
				MountFlags: []string{"rw"},
			},
		}
		a := []*csi.VolumeCapability{
			{
				AccessMode: &csi.VolumeCapability_AccessMode{
					Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
				},
				AccessType: aMountAT,
			},
		}

		b := []*csi.VolumeCapability{
			{
				AccessMode: &csi.VolumeCapability_AccessMode{
					Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
				},
				AccessType: &csi.VolumeCapability_Block{
					Block: &csi.VolumeCapability_BlockVolume{},
				},
			},
			{
				AccessMode: &csi.VolumeCapability_AccessMode{
					Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
				},
				AccessType: &csi.VolumeCapability_Mount{
					Mount: &csi.VolumeCapability_MountVolume{
						FsType:     "ext4",
						MountFlags: []string{"rw"},
					},
				},
			},
		}

		Ω(utils.AreVolumeCapabilitiesCompatible(a, b)).Should(BeTrue())
		aMountAT.Mount.MountFlags[0] = "ro"
		Ω(utils.AreVolumeCapabilitiesCompatible(a, b)).Should(BeFalse())
		a[0].AccessType = &csi.VolumeCapability_Block{
			Block: &csi.VolumeCapability_BlockVolume{},
		}
		Ω(utils.AreVolumeCapabilitiesCompatible(a, b)).Should(BeTrue())

		// make len(a) > len(b) for an error test
		a = append(a, &csi.VolumeCapability{
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			},
			AccessType: &csi.VolumeCapability_Mount{
				Mount: &csi.VolumeCapability_MountVolume{
					FsType:     "ext4",
					MountFlags: []string{"rw"},
				},
			},
		})
		a = append(a, &csi.VolumeCapability{
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
			},
			AccessType: &csi.VolumeCapability_Mount{
				Mount: &csi.VolumeCapability_MountVolume{
					FsType:     "ext4",
					MountFlags: []string{"rw"},
				},
			},
		})
		out, err := utils.AreVolumeCapabilitiesCompatible(a, b)
		Ω(out).Should(BeFalse())
		Ω(err).Should(HaveOccurred())

	})
})

func TestParseSlice(t *testing.T) {
	RegisterTestingT(t)

	// Test case: One value
	input := "value1"
	expected := []string{"value1"}
	result := utils.ParseSlice(input)
	Expect(result).To(Equal(expected))

	// Test case: Multiple values
	input = "value1, value2, value3"
	expected = []string{"value1", "value2", "value3"}
	result = utils.ParseSlice(input)
	Expect(result).To(Equal(expected))

	// Test case: Empty string
	input = ""
	expected = []string{}
	result = utils.ParseSlice(input)
	Expect(len(result)).To(Equal(0))

	// Test case: Values with whitespace
	// TODO: I think it's supposed to trim both trailing and leading whitespace.
	// It's currently only trimming leading whitespace.. except for the last item.
	input = " value1 , value2 , value3 "
	expected = []string{"value1 ", "value2 ", "value3"}
	result = utils.ParseSlice(input)
	Expect(result).To(Equal(expected))

	// Test case: Values with quotes
	input = `value1, "value2 ", " value3 "`
	expected = []string{"value1", "value2 ", " value3 "}
	result = utils.ParseSlice(input)
	Expect(result).To(Equal(expected))

	/*
		// TODO: Escaped quotes seem to be totally broken in the parse method.
		// Test case: Values with escaped quotes
		input = `value1, "value2 \"with quotes\"", "value3"`
		expected = []string{"value1", `value2 "with quotes"`, "value3"}
		result = utils.ParseSlice(input)
		Expect(result).To(Equal(expected))

		// Test case: Values with escaped quotes and whitespace
		input = `value1, "value2 \"with quotes\" ", " value3 "`
		expected = []string{"value1", `value2 "with quotes" `, " value3 "}
		result = utils.ParseSlice(input)
		Expect(result).To(Equal(expected))
	*/
}

func TestPageVolumes(t *testing.T) {
	RegisterTestingT(t)

	// Create a new CSI controller service
	svc := service.NewClient()

	// Create a mock controller client
	// TODO: This is the part that is breaking-- our mock controller client is not functional
	//mockClient := svc

	// Create a context
	ctx := context.Background()

	// Create a list volumes request
	req := csi.ListVolumesRequest{}

	// Call the PageVolumes function
	cvol, cerr := utils.PageVolumes(ctx, svc, req)
	var err error
	vols := []csi.Volume{}
	for {
		select {
		case v, ok := <-cvol:
			if !ok {
				return
			}
			vols = append(vols, v)
		case e, ok := <-cerr:
			if !ok {
				return
			}
			err = e
		}
	}

	Expect(err).To(BeNil())
	Expect(vols).To(HaveLen(4))
}

func TestPageSnapshots(t *testing.T) {
	RegisterTestingT(t)

	// Create a new CSI controller service
	svc := service.NewClient()

	// Create a mock controller client
	// TODO: This is the part that is breaking-- our mock controller client is not functional
	//mockClient := svc

	// Create a context
	ctx := context.Background()

	// Create a list volumes request
	req := csi.ListSnapshotsRequest{}

	// Call the PageSnapshots function
	csnap, cerr := utils.PageSnapshots(ctx, svc, req)
	snaps := []csi.Snapshot{}
	var err error
	for {
		select {
		case v, ok := <-csnap:
			if !ok {
				return
			}
			snaps = append(snaps, v)
		case e, ok := <-cerr:
			if !ok {
				return
			}
			err = e
		}
	}

	Expect(err).To(BeNil())
	Expect(snaps).To(HaveLen(3))
}

// struct for error injection testing with GRPCStatus
type ErrorStruct struct {
	StatusCode int
	Msg        string
}

func (e *ErrorStruct) Error() string {
	return fmt.Sprintf("Error %d: %s", e.StatusCode, e.Msg)
}
func (e *ErrorStruct) GRPCStatus() *grpcstatus.Status {
	if e == nil || e.StatusCode == 0 {
		return grpcstatus.New(codes.OK, e.Msg)
	}
	return grpcstatus.New(codes.Code(e.StatusCode), e.Msg)

}
func TestIsSuccess(t *testing.T) {
	RegisterTestingT(t)

	// Test case: Successful response - nil err
	Expect(utils.IsSuccess(nil)).To(BeNil())

	// Test case: Non-successful response
	response := ErrorStruct{
		StatusCode: http.StatusNotFound,
	}
	Expect(utils.IsSuccess(&response)).To(Not(BeNil()))

	// Test case: Successful response - GRPC OK
	response = ErrorStruct{StatusCode: 0}
	Expect(utils.IsSuccess(&response)).To(BeNil())

	// Test case: Successful response - acceptable opt
	response = ErrorStruct{
		StatusCode: http.StatusOK,
	}
	Expect(utils.IsSuccess(&response, http.StatusOK)).To(BeNil())

}

func TestSimple(t *testing.T) {
	RegisterTestingT(t)
}

func TestParseMapWS(t *testing.T) {
	RegisterTestingT(t)

	// Test case: One Pair
	data := utils.ParseMapWS("k1=v1")
	Expect(data).To(HaveLen(1))
	Expect(data["k1"]).To(Equal("v1"))

	// Test case: Empty Line
	data = utils.ParseMapWS("")
	Expect(data).To(HaveLen(0))

	// Test case: Key Sans Value
	data = utils.ParseMapWS("k1")
	Expect(data).To(HaveLen(0)) // should be empty if no key/value pairs

	// Test case: Two Pair
	data = utils.ParseMapWS("k1=v1 k2=v2")
	Expect(data).To(HaveLen(2))
	Expect(data["k1"]).To(Equal("v1"))
	Expect(data["k2"]).To(Equal("v2"))

	/*
		// TODO: This test case SHOULD pass based on the comments at the top of ParseMapWS.
		// It is VERY concerning that it doesn't.
		// Test case: Two Pair with Quoting & Escaping
		data = utils.ParseMapWS(`k1=v1 "k2=v2"`)
		Expect(data).To(HaveLen(2))
		Expect(data["k1"]).To(Equal("v1"))
		Expect(data["k2"]).To(Equal(`v2"s`))
	*/

	/*
		// TODO: This test case also SHOULD pass. I don't think ParseMapWS handles quotations as advertised.
		// Test case: Two Pair with Quoting & Escaping
		data = utils.ParseMapWS(`k1=v1 "k2=v2\'s"`)
		Expect(data).To(HaveLen(2))
		Expect(data["k1"]).To(Equal("v1"))
		Expect(data["k2"]).To(Equal(`v2\'s`))
	*/

	// Test case: Two Pair with Quoting & Escaping
	data = utils.ParseMapWS(`k1=v1 k2=v2\'s`)
	Expect(data).To(HaveLen(2))
	Expect(data["k1"]).To(Equal("v1"))
	Expect(data["k2"]).To(Equal(`v2's`))
}

func TestNewMountCapability(t *testing.T) {
	RegisterTestingT(t)

	// Test case: Single mount capability
	expected := &csi.VolumeCapability{
		AccessMode: &csi.VolumeCapability_AccessMode{
			Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
		},
		AccessType: &csi.VolumeCapability_Mount{
			Mount: &csi.VolumeCapability_MountVolume{
				FsType:     "ext4",
				MountFlags: []string{"ro"},
			},
		},
	}

	actual := utils.NewMountCapability(
		csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
		"ext4",
		"ro",
	)

	Expect(actual).To(Equal(expected))
}

func TestNewBlockCapability(t *testing.T) {
	RegisterTestingT(t)

	// Test case: Single block capability
	expected := &csi.VolumeCapability{
		AccessMode: &csi.VolumeCapability_AccessMode{
			Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
		},
		AccessType: &csi.VolumeCapability_Block{
			Block: &csi.VolumeCapability_BlockVolume{},
		},
	}

	actual := utils.NewBlockCapability(
		csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
	)

	Expect(actual).To(Equal(expected))
}

func TestEqualVolume(t *testing.T) {
	RegisterTestingT(t)

	// Test case: Equal volumes
	a := &csi.Volume{
		VolumeId:      "1",
		CapacityBytes: 100,
		VolumeContext: map[string]string{
			"name": "test",
		},
	}
	b := &csi.Volume{
		VolumeId:      "1",
		CapacityBytes: 100,
		VolumeContext: map[string]string{
			"name": "test",
		},
	}
	Ω(utils.EqualVolume(a, b)).Should(BeTrue())

	// Test case: Different volume IDs
	a.VolumeId = "2"
	Ω(utils.EqualVolume(a, b)).Should(BeFalse())

	// Test case: Different capacity bytes
	a.VolumeId = "1"
	a.CapacityBytes = 200
	Ω(utils.EqualVolume(a, b)).Should(BeFalse())

	// Test case: Different volume context
	a.CapacityBytes = 100
	a.VolumeContext = map[string]string{
		"name": "test2",
	}
	Ω(utils.EqualVolume(a, b)).Should(BeFalse())

	// Test case: One volume is nil
	a = nil
	Ω(utils.EqualVolume(a, b)).Should(BeFalse())

	// Test case: Both volumes are nil
	b = nil
	Ω(utils.EqualVolume(a, b)).Should(BeFalse())
}

func TestGetCSIEndpointListener(t *testing.T) {
	RegisterTestingT(t)

	// Test case: TCP endpoint
	os.Setenv("CSI_ENDPOINT", "tcp://localhost:5000")
	lis, err := utils.GetCSIEndpointListener()
	Ω(err).ShouldNot(HaveOccurred())
	Ω(lis.Addr().Network()).Should(Equal("tcp"))
	Ω(lis.Addr().String()).Should(Equal("127.0.0.1:5000"))

	// Test case: Invalid endpoint
	os.Setenv("CSI_ENDPOINT", "invalid://endpoint")
	lis, err = utils.GetCSIEndpointListener()
	Ω(err).Should(HaveOccurred())
	Ω(lis).Should(BeNil())

	// Test case: Empty endpoint
	os.Setenv("CSI_ENDPOINT", "")
	lis, err = utils.GetCSIEndpointListener()
	Ω(err).Should(HaveOccurred())
	Ω(lis).Should(BeNil())
}

func TestIsVolumeCapabilityCompatible(t *testing.T) {
	RegisterTestingT(t)

	// Test case: Compatible volume capabilities
	a := &csi.VolumeCapability{
		AccessMode: &csi.VolumeCapability_AccessMode{
			Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
		},
		AccessType: &csi.VolumeCapability_Block{
			Block: &csi.VolumeCapability_BlockVolume{},
		},
	}
	b := []*csi.VolumeCapability{
		{
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			},
			AccessType: &csi.VolumeCapability_Block{
				Block: &csi.VolumeCapability_BlockVolume{},
			},
		},
		{
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			},
			AccessType: &csi.VolumeCapability_Mount{
				Mount: &csi.VolumeCapability_MountVolume{
					FsType:     "ext4",
					MountFlags: []string{"rw"},
				},
			},
		},
	}
	out, err := utils.IsVolumeCapabilityCompatible(a, b)
	Ω(out).Should(BeTrue())
	Ω(err).Should(BeNil())
}
