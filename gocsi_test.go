package gocsi

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/user"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/dell/gocsi/mock/service"
	log "github.com/sirupsen/logrus"
)

func TestRun(t *testing.T) {
	originalOsExit := osExit
	defer func() { osExit = originalOsExit }()

	osExitCh := make(chan struct{})
	osExit = func(_ int) {
		close(osExitCh)
	}

	osUser, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	endpoint := fmt.Sprintf("unix://%s/csi.sock", wd)

	envVars := [][]string{
		{EnvVarDebug, "true"},
		{EnvVarLogLevel, "debug"},
		{EnvVarEndpoint, endpoint},
		{EnvVarEndpointPerms, "0777"},
		{EnvVarCredsCreateVol, "true"},
		{EnvVarCredsDeleteVol, "true"},
		{EnvVarCredsCtrlrPubVol, "true"},
		{EnvVarCredsCtrlrUnpubVol, "true"},
		{EnvVarCredsNodeStgVol, "true"},
		{EnvVarCredsNodePubVol, "true"},
		{EnvVarDisableFieldLen, "true"},
		{EnvVarRequireStagingTargetPath, "true"},
		{EnvVarRequireVolContext, "true"},
		{EnvVarCreds, "true"},
		{EnvVarSpecValidation, "false"},
		{EnvVarLoggingDisableVolCtx, "true"},
		{EnvVarPluginInfo, "true"},
		{EnvVarSerialVolAccessTimeout, "10s"},
		{EnvVarSpecReqValidation, "true"},
		{EnvVarSpecRepValidation, "true"},
		{EnvVarEndpointUser, osUser.Name},
		{EnvVarEndpointGroup, osUser.Gid},
		{EnvVarSerialVolAccessEtcdEndpoints, "http://127.0.0.1:2379"},
	}

	defer func() {
		for _, env := range envVars {
			if err := os.Unsetenv(env[0]); err != nil {
				t.Fatalf("failed to unset env var %s: %v", env[0], err)
			}
		}
	}()

	for _, env := range envVars {
		if err := os.Setenv(env[0], env[1]); err != nil {
			t.Fatalf("failed to set env var %s: %v", env[0], err)
		}
	}

	svc := service.NewServer()
	go Run(context.Background(), "Dell CSM Driver", "A Dell Container Storage Interface (CSI) Plugin", "", newMockStoragePlugin(svc, svc, svc))
	time.Sleep(5 * time.Second)

	if err := syscall.Kill(syscall.Getpid(), syscall.SIGINT); err != nil {
		t.Fatalf("failed to send SIGINT: %v", err)
	}
	// Wait until the server calls osExit() to exit
	<-osExitCh
}

func TestRunHelp(_ *testing.T) {
	originalOsExit := osExit
	originalOsArgs := os.Args
	defer func() {
		osExit = originalOsExit
		os.Args = originalOsArgs
	}()

	calledOsExit := make(chan struct{})
	osExit = func(code int) {
		calledOsExit <- struct{}{}
		if code == 1 {
			runtime.Goexit()
		}
	}

	svc := service.NewServer()
	os.Args = []string{"--?"}
	go Run(context.Background(), "Dell CSM Driver", "A Dell Container Storage Interface (CSI) Plugin", "", newMockStoragePlugin(svc, svc, svc))
	<-calledOsExit
}

func TestRunNoEndpoint(_ *testing.T) {
	originalOsExit := osExit

	calledOsExit := make(chan struct{})
	osExit = func(code int) {
		calledOsExit <- struct{}{}
		if code == 1 {
			runtime.Goexit()
		}
	}

	defer func() {
		osExit = originalOsExit
	}()

	svc := service.NewServer()
	go Run(context.Background(), "Dell CSM Driver", "A Dell Container Storage Interface (CSI) Plugin", "", newMockStoragePlugin(svc, svc, svc))

	<-calledOsExit
}

func TestRunFailListener(_ *testing.T) {
	originalOsExit := osExit

	calledOsExit := make(chan struct{})
	osExit = func(code int) {
		calledOsExit <- struct{}{}
		if code == 1 {
			runtime.Goexit()
		}
	}

	defer func() {
		osExit = originalOsExit
		os.Unsetenv(EnvVarEndpoint)
	}()

	os.Setenv(EnvVarEndpoint, "/bad/path/does/not/exist/gniro0$$")

	svc := service.NewServer()
	go Run(context.Background(), "Dell CSM Driver", "A Dell Container Storage Interface (CSI) Plugin", "", newMockStoragePlugin(svc, svc, svc))

	<-calledOsExit
}

func TestRunNoIdentityService(t *testing.T) {
	originalOsExit := osExit

	calledOsExit := make(chan struct{})
	osExit = func(code int) {
		calledOsExit <- struct{}{}
		if code == 1 {
			runtime.Goexit()
		}
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	endpoint := fmt.Sprintf("unix://%s/csi.sock", wd)

	defer func() {
		osExit = originalOsExit
		os.Unsetenv(EnvVarEndpoint)
	}()

	os.Setenv(EnvVarEndpoint, endpoint)

	svc := service.NewServer()
	go Run(context.Background(), "Dell CSM Driver", "A Dell Container Storage Interface (CSI) Plugin", "", newMockStoragePlugin(svc, nil, svc))
	<-calledOsExit
}

func TestRunNoControllerOrNodeService(t *testing.T) {
	originalOsExit := osExit

	calledOsExit := make(chan struct{})
	osExit = func(code int) {
		calledOsExit <- struct{}{}
		if code == 1 {
			runtime.Goexit()
		}
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	endpoint := fmt.Sprintf("unix://%s/csi.sock", wd)

	defer func() {
		osExit = originalOsExit
		os.Unsetenv(EnvVarEndpoint)
	}()

	os.Setenv(EnvVarEndpoint, endpoint)

	svc := service.NewServer()
	go Run(context.Background(), "Dell CSM Driver", "A Dell Container Storage Interface (CSI) Plugin", "", newMockStoragePlugin(nil, svc, nil))
	<-calledOsExit
}

// New returns a new Mock Storage Plug-in Provider.
// Due to cyclic imports with the mock/provider package, the mock provider is copied here.
func newMockStoragePlugin(controller csi.ControllerServer, identity csi.IdentityServer, node csi.NodeServer) StoragePluginProvider {
	return &StoragePlugin{
		Controller: controller,
		Identity:   identity,
		Node:       node,

		// BeforeServe allows the SP to participate in the startup
		// sequence. This function is invoked directly before the
		// gRPC server is created, giving the callback the ability to
		// modify the SP's interceptors, server options, or prevent the
		// server from starting by returning a non-nil error.
		BeforeServe: func(
			_ context.Context,
			_ *StoragePlugin,
			_ net.Listener,
		) error {
			log.WithField("service", service.Name).Debug("BeforeServe")
			return nil
		},

		EnvVars: []string{
			// Enable serial volume access.
			EnvVarSerialVolAccess + "=true",

			// Enable request and response validation.
			EnvVarSpecValidation + "=true",

			// Treat the following fields as required:
			//   * ControllerPublishVolumeResponse.PublishContext
			//   * NodeStageVolumeRequest.PublishContext
			//   * NodePublishVolumeRequest.PublishContext
			EnvVarRequirePubContext + "=true",
		},
	}
}
