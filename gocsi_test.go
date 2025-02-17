package gocsi

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/dell/gocsi/mock/service"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
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
	go Run(context.Background(), "Dell CSM Driver", "A Dell Container Storage Interface (CSI) Plugin", "", newMockStoragePluginProvider(svc, svc, svc))
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
	go Run(context.Background(), "Dell CSM Driver", "A Dell Container Storage Interface (CSI) Plugin", "", newMockStoragePluginProvider(svc, svc, svc))
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
	go Run(context.Background(), "Dell CSM Driver", "A Dell Container Storage Interface (CSI) Plugin", "", newMockStoragePluginProvider(svc, svc, svc))

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
	go Run(context.Background(), "Dell CSM Driver", "A Dell Container Storage Interface (CSI) Plugin", "", newMockStoragePluginProvider(svc, svc, svc))

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
	go Run(context.Background(), "Dell CSM Driver", "A Dell Container Storage Interface (CSI) Plugin", "", newMockStoragePluginProvider(svc, nil, svc))
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
	go Run(context.Background(), "Dell CSM Driver", "A Dell Container Storage Interface (CSI) Plugin", "", newMockStoragePluginProvider(nil, svc, nil))
	<-calledOsExit
}

func TestInitEndpointOwner(t *testing.T) {
	// Create a new StoragePlugin instance
	svc := service.NewServer()
	sp := newMockStoragePlugin(nil, svc, nil)

	// Create a new listener
	lis, err := net.Listen("unix", "test.sock")
	if err != nil {
		t.Fatal(err)
	}
	defer lis.Close()

	// Set up the context
	ctx := context.Background()

	defer os.Unsetenv(EnvVarEndpointUser)
	defer os.Unsetenv(EnvVarEndpointGroup)

	tests := []struct {
		name      string
		user      string
		group     string
		expectErr string
	}{
		{
			name:      "Testinvalid user",
			user:      "testuser",
			group:     "",
			expectErr: "unknown user",
		},
		{
			name:      "Test user cannot be found",
			user:      "123",
			group:     "",
			expectErr: "unknown user",
		},
		{
			name:      "test valid user",
			user:      strconv.Itoa(os.Getuid()),
			group:     "",
			expectErr: "",
		},
		{
			name:      "Test invalid group",
			user:      "",
			group:     "testgroup",
			expectErr: "unknown group",
		},
		{
			name:      "Test group cannot be found",
			user:      "",
			group:     "123",
			expectErr: "unknown group",
		},
		{
			name:      "Test valid group",
			user:      "",
			group:     strconv.Itoa(os.Getgid()),
			expectErr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.user != "" {
				if err := os.Setenv(EnvVarEndpointUser, tt.user); err != nil {
					t.Fatal(err)
				}
			}
			if tt.group != "" {
				if err := os.Setenv(EnvVarEndpointGroup, tt.group); err != nil {
					t.Fatal(err)
				}
			}
			if err := sp.initEndpointOwner(ctx, lis); err != nil && !strings.Contains(err.Error(), tt.expectErr) {
				if tt.expectErr == "" {
					// change value for clearer logging when no error was expected, but we recieved one
					tt.expectErr = "no error"
				}
				t.Errorf("StoragePlugin.getPluginInfo() returned error = %s, but expected: %s", err.Error(), tt.expectErr)
			}
		})
	}
}

// TODO: Add test case for stop
func TestStop(t *testing.T) {
	// Create a new StoragePlugin instance
	svc := service.NewServer()
	sp := newMockStoragePlugin(nil, svc, nil)

	// sp.stopOnce = sync.Once{}
	// sp.server = &grpc.Server{}
	// sp.server.quit = make(chan struct{})

	// Create a new listener
	lis, err := net.Listen("unix", "test.sock")
	if err != nil {
		t.Fatal(err)
	}
	defer lis.Close()

	// Set up the context
	ctx := context.Background()

	// assert.NotPanics(t, sp.Stop(ctx))
	sp.Stop(ctx)
}

func TestGetPluginInfo(t *testing.T) {
	svc := service.NewServer()
	sp := newMockStoragePlugin(nil, svc, nil)

	// Create a new listener
	lis, err := net.Listen("unix", "test.sock")
	if err != nil {
		t.Fatal(err)
	}
	defer lis.Close()

	// Set up the context
	ctx := context.Background()

	// set up request
	req := &csi.GetPluginInfoRequest{}

	// set up handler
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		log.Info("ctx:", ctx)
		log.Info("req:", req)
		resp := &csi.GetPluginInfoResponse{
			Name:          "my-plugin",
			VendorVersion: "1.0.0",
			Manifest:      map[string]string{"key": "value"},
		}
		return resp, nil
	}

	tests := []struct {
		name       string
		pluginInfo csi.GetPluginInfoResponse
		info       *grpc.UnaryServerInfo
		expectErr  string
	}{
		{
			name: "Test happy path",
			pluginInfo: csi.GetPluginInfoResponse{
				Name:          "my-plugin",
				VendorVersion: "1.0.0",
				Manifest:      map[string]string{"key": "value"},
			},
			info: &grpc.UnaryServerInfo{
				FullMethod: "/csi.v1.Identity/GetPluginInfo",
				Server:     svc,
			},
			expectErr: "",
		},
		{
			name: "Test blank name",
			pluginInfo: csi.GetPluginInfoResponse{
				Name:          "",
				VendorVersion: "1.0.0",
				Manifest:      map[string]string{"key": "value"},
			},
			info: &grpc.UnaryServerInfo{
				FullMethod: "/csi.v1.Identity/GetPluginInfo",
				Server:     svc,
			},
			expectErr: "",
		},
		{
			name: "Test with unparsable method name",
			pluginInfo: csi.GetPluginInfoResponse{
				Name:          "my-plugin",
				VendorVersion: "1.0.0",
				Manifest:      map[string]string{"key": "value"},
			},
			info: &grpc.UnaryServerInfo{
				FullMethod: "test",
				Server:     svc,
			},
			expectErr: "ParseMethod",
		},
		{
			name: "Test with wrong method",
			pluginInfo: csi.GetPluginInfoResponse{
				Name:          "my-plugin",
				VendorVersion: "1.0.0",
				Manifest:      map[string]string{"key": "value"},
			},
			info: &grpc.UnaryServerInfo{
				FullMethod: "/csi.v1.Controller/CreateVolume",
				Server:     svc,
			},
			expectErr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp.pluginInfo = tt.pluginInfo
			if _, err := sp.getPluginInfo(ctx, req, tt.info, handler); err != nil && !strings.Contains(err.Error(), tt.expectErr) {
				if tt.expectErr == "" {
					// change value for clearer logging when no error was expected, but we recieved one
					tt.expectErr = "no error"
				}
				t.Errorf("StoragePlugin.getPluginInfo() returned error = %s, but expected: %s", err.Error(), tt.expectErr)
			}
		})
	}
}

// New returns a new Mock Storage Plug-in Provider.
// Due to cyclic imports with the mock/provider package, the mock provider is copied here.
func newMockStoragePluginProvider(controller csi.ControllerServer, identity csi.IdentityServer, node csi.NodeServer) StoragePluginProvider {
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

func newMockStoragePlugin(controller csi.ControllerServer, identity csi.IdentityServer, node csi.NodeServer) StoragePlugin {
	return StoragePlugin{
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

func TestStoragePlugin_Serve(t *testing.T) {
	svc := service.NewServer()

	// Create a new listener
	lis, err := net.Listen("unix", "test.sock")
	if err != nil {
		t.Fatal(err)
	}
	defer lis.Close()

	defer os.Unsetenv(EnvVarMode)

	// Set up the context
	ctx := context.Background()

	tests := []struct {
		name          string
		controllerSvc csi.ControllerServer
		nodeSvc       csi.NodeServer
		mode          string
		expectErr     string
	}{
		{
			name:          "Test missing controller service",
			controllerSvc: nil,
			nodeSvc:       svc,
			mode:          "controller",
			expectErr:     "controller service is required",
		},
		{
			name:          "Test missing node service",
			controllerSvc: svc,
			nodeSvc:       nil,
			mode:          "node",
			expectErr:     "node service is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := newMockStoragePlugin(tt.controllerSvc, svc, tt.nodeSvc)
			os.Setenv(EnvVarMode, tt.mode)
			err := sp.Serve(ctx, lis)
			assert.ErrorContains(t, err, tt.expectErr)
		})
	}
}
