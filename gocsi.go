/*
 *
 * Copyright © 2021-2024 Dell Inc. or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

//go:generate make

// Package gocsi provides a Container Storage Interface (CSI) library,
// client, and other helpful utilities.
package gocsi

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"os/user"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"text/template"

	"github.com/container-storage-interface/spec/lib/go/csi"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	csictx "github.com/dell/gocsi/context"
	utils "github.com/dell/gocsi/utils/csi"
	"github.com/dell/gocsi/utils/middleware"
)

var osExit = func(code int) {
	os.Exit(code)
}

// Run launches a CSI storage plug-in.
func Run(
	ctx context.Context,
	appName, appDescription, appUsage string,
	sp StoragePluginProvider,
) {
	// Check for the debug value.
	if v, ok := csictx.LookupEnv(ctx, EnvVarDebug); ok {
		/* #nosec G104 */
		if ok, _ := strconv.ParseBool(v); ok {
			_ = csictx.Setenv(ctx, EnvVarLogLevel, "debug")
			_ = csictx.Setenv(ctx, EnvVarReqLogging, "true")
			_ = csictx.Setenv(ctx, EnvVarRepLogging, "true")
		}
	}

	// Adjust the log level.
	lvl := log.InfoLevel
	if v, ok := csictx.LookupEnv(ctx, EnvVarLogLevel); ok {
		var err error
		if lvl, err = log.ParseLevel(v); err != nil {
			lvl = log.InfoLevel
		}
	}
	log.SetLevel(lvl)

	printUsage := func() {
		// app is the information passed to the printUsage function
		app := struct {
			Name        string
			Description string
			Usage       string
			BinPath     string
		}{
			appName,
			appDescription,
			appUsage,
			os.Args[0],
		}

		t, err := template.New("t").Parse(usage)
		if err != nil {
			log.WithError(err).Fatalln("failed to parse usage template")
		}
		if err := t.Execute(os.Stderr, app); err != nil {
			log.WithError(err).Fatalln("failed emitting usage")
		}
		return
	}

	// Check for a help flag.
	fs := flag.NewFlagSet("csp", flag.ExitOnError)
	fs.Usage = printUsage
	var help bool
	fs.BoolVar(&help, "?", false, "")
	err := fs.Parse(os.Args)
	if err == flag.ErrHelp || help {
		printUsage()
		osExit(1)
	}

	// If no endpoint is set then print the usage.
	if os.Getenv(EnvVarEndpoint) == "" {
		printUsage()
		osExit(1)
	}

	l, err := utils.GetCSIEndpointListener()
	if err != nil {
		log.WithError(err).Info("failed to listen")
		osExit(1)
	}

	// Define a lambda that can be used in the exit handler
	// to remove a potential UNIX sock file.
	var rmSockFileOnce sync.Once
	rmSockFile := func() {
		rmSockFileOnce.Do(func() {
			if l == nil || l.Addr() == nil {
				return
			}
			/* #nosec G104 */
			if l.Addr().Network() == netUnix {
				sockFile := l.Addr().String()
				_ = os.RemoveAll(sockFile)
				log.WithField("path", sockFile).Info("removed sock file")
			}
		})
	}

	trapSignals(func() {
		sp.GracefulStop(ctx)
		rmSockFile()
		log.Info("server stopped gracefully")
	})

	if err := sp.Serve(ctx, l); err != nil {
		rmSockFile()
		log.WithError(err).Info("grpc failed")
		osExit(1)
	}
}

// StoragePluginProvider is able to serve a gRPC endpoint that provides
// the CSI services: Controller, Identity, Node.
type StoragePluginProvider interface {
	// Serve accepts incoming connections on the listener lis, creating
	// a new ServerTransport and service goroutine for each. The service
	// goroutine read gRPC requests and then call the registered handlers
	// to reply to them. Serve returns when lis.Accept fails with fatal
	// errors.  lis will be closed when this method returns.
	// Serve always returns non-nil error.
	Serve(ctx context.Context, lis net.Listener) error

	// Stop stops the gRPC server. It immediately closes all open
	// connections and listeners.
	// It cancels all active RPCs on the server side and the corresponding
	// pending RPCs on the client side will get notified by connection
	// errors.
	Stop(ctx context.Context)

	// GracefulStop stops the gRPC server gracefully. It stops the server
	// from accepting new connections and RPCs and blocks until all the
	// pending RPCs are finished.
	GracefulStop(ctx context.Context)
}

// StoragePlugin is the collection of services and data used to server
// a new gRPC endpoint that acts as a CSI storage plug-in (SP).
type StoragePlugin struct {
	// Controller is the eponymous CSI service.
	Controller csi.ControllerServer

	// Identity is the eponymous CSI service.
	Identity csi.IdentityServer

	// Node is the eponymous CSI service.
	Node csi.NodeServer

	// ServerOpts is a list of gRPC server options used when serving
	// the SP. This list should not include a gRPC interceptor option
	// as one is created automatically based on the interceptor configuration
	// or provided list of interceptors.
	ServerOpts []grpc.ServerOption

	// Interceptors is a list of gRPC server interceptors to use when
	// serving the SP. This list should not include the interceptors
	// defined in the GoCSI package as those are configured by default
	// based on runtime configuration settings.
	Interceptors []grpc.UnaryServerInterceptor

	// BeforeServe is an optional callback that is invoked after the
	// StoragePlugin has been initialized, just prior to the creation
	// of the gRPC server. This callback may be used to perform custom
	// initialization logic, modify the interceptors and server options,
	// or prevent the server from starting by returning a non-nil error.
	BeforeServe func(context.Context, *StoragePlugin, net.Listener) error

	// EnvVars is a list of default environment variables and values.
	EnvVars []string

	// RegisterAdditionalServers allows the driver to register additional
	// grpc servers on the same grpc connection. These can be used
	// for proprietary extensions.
	RegisterAdditionalServers func(*grpc.Server)

	serveOnce sync.Once
	stopOnce  sync.Once
	server    *grpc.Server

	envVars    map[string]string
	pluginInfo csi.GetPluginInfoResponse
}

// Serve accepts incoming connections on the listener lis, creating
// a new ServerTransport and service goroutine for each. The service
// goroutine read gRPC requests and then call the registered handlers
// to reply to them. Serve returns when lis.Accept fails with fatal
// errors.  lis will be closed when this method returns.
// Serve always returns non-nil error.
func (sp *StoragePlugin) Serve(ctx context.Context, lis net.Listener) error {
	var err error
	sp.serveOnce.Do(func() {
		// Please note that the order of the below init functions is
		// important and should not be altered unless by someone aware
		// of how they work.

		// Adding this function to the context allows `csictx.LookupEnv`
		// to search this SP's default env vars for a value.
		ctx = csictx.WithLookupEnv(ctx, sp.lookupEnv)

		// Adding this function to the context allows `csictx.Setenv`
		// to set environment variables in this SP's env var store.
		ctx = csictx.WithSetenv(ctx, sp.setenv)

		// Initialize the storage plug-in's environment variables map.
		sp.initEnvVars(ctx)

		// Adjust the endpoint's file permissions.
		if err = sp.initEndpointPerms(ctx, lis); err != nil {
			return
		}

		// Adjust the endpoint's file ownership.
		if err = sp.initEndpointOwner(ctx, lis); err != nil {
			return
		}

		// Initialize the storage plug-in's info.
		sp.initPluginInfo(ctx)

		// Initialize the interceptors.
		sp.initInterceptors(ctx)

		// Invoke the SP's BeforeServe function to give the SP a chance
		// to perform any local initialization routines.
		if f := sp.BeforeServe; f != nil {
			if err = f(ctx, sp, lis); err != nil {
				return
			}
		}

		// Add the interceptors to the server if any are configured.
		if i := sp.Interceptors; len(i) > 0 {
			sp.ServerOpts = append(sp.ServerOpts,
				grpc.UnaryInterceptor(middleware.ChainUnaryServer(i...)))
		}

		// Initialize the gRPC server.
		sp.server = grpc.NewServer(sp.ServerOpts...)

		// Register the CSI services.
		// Always require the identity service.
		if sp.Identity == nil {
			err = errors.New("identity service is required")
			return
		}
		// Either a Controller or Node service should be supplied.
		if sp.Controller == nil && sp.Node == nil {
			err = errors.New(
				"either a controller or node service is required")
			return
		}

		// Always register the identity service.
		csi.RegisterIdentityServer(sp.server, sp.Identity)
		log.Info("identity service registered")

		// Determine which of the controller/node services to register
		mode := csictx.Getenv(ctx, EnvVarMode)
		if strings.EqualFold(mode, "controller") {
			mode = "controller"
		} else if strings.EqualFold(mode, "node") {
			mode = "node"
		} else {
			mode = ""
		}

		if mode == "" || mode == "controller" {
			if sp.Controller == nil {
				err = errors.New("controller service is required")
				return
			}
			csi.RegisterControllerServer(sp.server, sp.Controller)
			log.Info("controller service registered")
		}
		if mode == "" || mode == "node" {
			if sp.Node == nil {
				err = errors.New("node service is required")
				return
			}
			csi.RegisterNodeServer(sp.server, sp.Node)
			log.Info("node service registered")
		}

		// Register any additional servers required.
		if sp.RegisterAdditionalServers != nil {
			sp.RegisterAdditionalServers(sp.server)
		}

		endpoint := fmt.Sprintf(
			"%s://%s",
			lis.Addr().Network(), lis.Addr().String())
		log.WithField("endpoint", endpoint).Info("serving")

		// Start the gRPC server.
		err = sp.server.Serve(lis)
		return
	})
	return err
}

// Stop stops the gRPC server. It immediately closes all open
// connections and listeners.
// It cancels all active RPCs on the server side and the corresponding
// pending RPCs on the client side will get notified by connection
// errors.
func (sp *StoragePlugin) Stop(_ context.Context) {
	sp.stopOnce.Do(func() {
		if sp.server != nil {
			sp.server.Stop()
		}
		log.Info("stopped")
	})
}

// GracefulStop stops the gRPC server gracefully. It stops the server
// from accepting new connections and RPCs and blocks until all the
// pending RPCs are finished.
func (sp *StoragePlugin) GracefulStop(_ context.Context) {
	sp.stopOnce.Do(func() {
		if sp.server != nil {
			sp.server.GracefulStop()
		}
		log.Info("gracefully stopped")
	})
}

const netUnix = "unix"

func (sp *StoragePlugin) initEndpointPerms(
	ctx context.Context, lis net.Listener,
) error {
	if lis.Addr().Network() != netUnix {
		return nil
	}

	v, ok := csictx.LookupEnv(ctx, EnvVarEndpointPerms)
	if !ok || v == "0755" {
		return nil
	}
	u, err := strconv.ParseUint(v, 8, 32)
	if err != nil {
		return err
	}

	p := lis.Addr().String()
	m := os.FileMode(u)

	log.WithFields(map[string]interface{}{
		"path": p,
		"mode": m,
	}).Info("chmod csi endpoint")

	if err := os.Chmod(p, m); err != nil {
		return err
	}

	return nil
}

func (sp *StoragePlugin) initEndpointOwner(
	ctx context.Context, lis net.Listener,
) error {
	if lis.Addr().Network() != netUnix {
		return nil
	}

	var (
		usrName string
		grpName string

		uid  = os.Getuid()
		gid  = os.Getgid()
		puid = uid
		pgid = gid
	)

	if v, ok := csictx.LookupEnv(ctx, EnvVarEndpointUser); ok {
		m, err := regexp.MatchString(`^\d+$`, v)
		if err != nil {
			return err
		}
		usrName = v
		szUID := v
		if m {
			u, err := user.LookupId(v)
			if err != nil {
				return err
			}
			usrName = u.Username
		} else {
			u, err := user.Lookup(v)
			if err != nil {
				return err
			}
			szUID = u.Uid
		}
		iuid, err := strconv.Atoi(szUID)
		if err != nil {
			return err
		}
		uid = iuid
	}

	if v, ok := csictx.LookupEnv(ctx, EnvVarEndpointGroup); ok {
		m, err := regexp.MatchString(`^\d+$`, v)
		if err != nil {
			return err
		}
		grpName = v
		szGID := v
		if m {
			u, err := user.LookupGroupId(v)
			if err != nil {
				return err
			}
			grpName = u.Name
		} else {
			u, err := user.LookupGroup(v)
			if err != nil {
				return err
			}
			szGID = u.Gid
		}
		igid, err := strconv.Atoi(szGID)
		if err != nil {
			return err
		}
		gid = igid
	}

	if uid != puid || gid != pgid {
		f := lis.Addr().String()
		log.WithFields(map[string]interface{}{
			"uid":  usrName,
			"gid":  grpName,
			"path": f,
		}).Info("chown csi endpoint")
		if err := os.Chown(f, uid, gid); err != nil {
			return err
		}
	}

	return nil
}

func (sp *StoragePlugin) lookupEnv(key string) (string, bool) {
	val, ok := sp.envVars[key]
	return val, ok
}

func (sp *StoragePlugin) setenv(key, val string) error {
	sp.envVars[key] = val
	return nil
}

func (sp *StoragePlugin) getEnvBool(ctx context.Context, key string) bool {
	v, ok := csictx.LookupEnv(ctx, key)
	if !ok {
		return false
	}
	if b, err := strconv.ParseBool(v); err == nil {
		return b
	}
	return false
}

func trapSignals(onExit func()) {
	sigc := make(chan os.Signal, 1)
	sigs := []os.Signal{
		syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
	}
	signal.Notify(sigc, sigs...)
	go func() {
		for s := range sigc {
			log.WithField("signal", s).Info("received signal; shutting down")
			if onExit != nil {
				onExit()
			}
			osExit(0)
		}
	}()
}

type logger struct {
	f func(msg string, args ...interface{})
	w io.Writer
}

func newLogger(f func(msg string, args ...interface{})) *logger {
	l := &logger{f: f}
	r, w := io.Pipe()
	l.w = w
	go func() {
		scan := bufio.NewScanner(r)
		for scan.Scan() {
			f(scan.Text())
		}
	}()
	return l
}

func (l *logger) Write(data []byte) (int, error) {
	return l.w.Write(data)
}
