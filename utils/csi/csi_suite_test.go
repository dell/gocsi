package csi_test

import (
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

func TestCsiUtils(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	log.SetOutput(GinkgoWriter)
	RegisterFailHandler(Fail)
	RunSpecs(t, "CSI Utils Suite")
}

// Declarations for Ginkgo DSL
type (
	Done        ginkgo.Done
	Benchmarker ginkgo.Benchmarker
)

var (
	GinkgoWriter                          = ginkgo.GinkgoWriter
	GinkgoRandomSeed                      = ginkgo.GinkgoRandomSeed
	GinkgoParallelNode                    = ginkgo.GinkgoParallelNode
	GinkgoT                               = ginkgo.GinkgoT
	CurrentGinkgoTestDescription          = ginkgo.CurrentGinkgoTestDescription
	RunSpecs                              = ginkgo.RunSpecs
	RunSpecsWithDefaultAndCustomReporters = ginkgo.RunSpecsWithDefaultAndCustomReporters
	RunSpecsWithCustomReporters           = ginkgo.RunSpecsWithCustomReporters
	Skip                                  = ginkgo.Skip
	Fail                                  = ginkgo.Fail
	GinkgoRecover                         = ginkgo.GinkgoRecover
	Describe                              = ginkgo.Describe
	FDescribe                             = ginkgo.FDescribe
	PDescribe                             = ginkgo.PDescribe
	XDescribe                             = ginkgo.XDescribe
	Context                               = ginkgo.Context
	FContext                              = ginkgo.FContext
	PContext                              = ginkgo.PContext
	XContext                              = ginkgo.XContext
	It                                    = ginkgo.It
	FIt                                   = ginkgo.FIt
	PIt                                   = ginkgo.PIt
	XIt                                   = ginkgo.XIt
	Specify                               = ginkgo.Specify
	FSpecify                              = ginkgo.FSpecify
	PSpecify                              = ginkgo.PSpecify
	XSpecify                              = ginkgo.XSpecify
	By                                    = ginkgo.By
	Measure                               = ginkgo.Measure
	FMeasure                              = ginkgo.FMeasure
	PMeasure                              = ginkgo.PMeasure
	XMeasure                              = ginkgo.XMeasure
	BeforeSuite                           = ginkgo.BeforeSuite
	AfterSuite                            = ginkgo.AfterSuite
	SynchronizedBeforeSuite               = ginkgo.SynchronizedBeforeSuite
	SynchronizedAfterSuite                = ginkgo.SynchronizedAfterSuite
	BeforeEach                            = ginkgo.BeforeEach
	JustBeforeEach                        = ginkgo.JustBeforeEach
	AfterEach                             = ginkgo.AfterEach
)

// Declarations for Gomega DSL
var (
	RegisterFailHandler                   = gomega.RegisterFailHandler
	RegisterTestingT                      = gomega.RegisterTestingT
	InterceptGomegaFailures               = gomega.InterceptGomegaFailures
	Ω                                     = gomega.Ω
	Expect                                = gomega.Expect
	ExpectWithOffset                      = gomega.ExpectWithOffset
	Eventually                            = gomega.Eventually
	EventuallyWithOffset                  = gomega.EventuallyWithOffset
	Consistently                          = gomega.Consistently
	ConsistentlyWithOffset                = gomega.ConsistentlyWithOffset
	SetDefaultEventuallyTimeout           = gomega.SetDefaultEventuallyTimeout
	SetDefaultEventuallyPollingInterval   = gomega.SetDefaultEventuallyPollingInterval
	SetDefaultConsistentlyDuration        = gomega.SetDefaultConsistentlyDuration
	SetDefaultConsistentlyPollingInterval = gomega.SetDefaultConsistentlyPollingInterval
)

// Declarations for Gomega Matchers
var (
	Equal                = gomega.Equal
	BeEquivalentTo       = gomega.BeEquivalentTo
	BeIdenticalTo        = gomega.BeIdenticalTo
	BeNil                = gomega.BeNil
	BeTrue               = gomega.BeTrue
	BeFalse              = gomega.BeFalse
	HaveOccurred         = gomega.HaveOccurred
	Succeed              = gomega.Succeed
	MatchError           = gomega.MatchError
	BeClosed             = gomega.BeClosed
	Receive              = gomega.Receive
	BeSent               = gomega.BeSent
	MatchRegexp          = gomega.MatchRegexp
	ContainSubstring     = gomega.ContainSubstring
	HavePrefix           = gomega.HavePrefix
	HaveSuffix           = gomega.HaveSuffix
	MatchJSON            = gomega.MatchJSON
	MatchXML             = gomega.MatchXML
	MatchYAML            = gomega.MatchYAML
	BeEmpty              = gomega.BeEmpty
	HaveLen              = gomega.HaveLen
	HaveCap              = gomega.HaveCap
	BeZero               = gomega.BeZero
	ContainElement       = gomega.ContainElement
	ConsistOf            = gomega.ConsistOf
	HaveKey              = gomega.HaveKey
	HaveKeyWithValue     = gomega.HaveKeyWithValue
	BeNumerically        = gomega.BeNumerically
	BeTemporally         = gomega.BeTemporally
	BeAssignableToTypeOf = gomega.BeAssignableToTypeOf
	Panic                = gomega.Panic
	BeAnExistingFile     = gomega.BeAnExistingFile
	BeARegularFile       = gomega.BeARegularFile
	BeADirectory         = gomega.BeADirectory
	And                  = gomega.And
	SatisfyAll           = gomega.SatisfyAll
	Or                   = gomega.Or
	SatisfyAny           = gomega.SatisfyAny
	Not                  = gomega.Not
	WithTransform        = gomega.WithTransform
)
