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

package etcd

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	mwtypes "github.com/dell/gocsi/middleware/serialvolume/lockprovider"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	etcd "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
)

var p mwtypes.VolumeLockerProvider

func TestMain(m *testing.M) {
	log.SetLevel(log.InfoLevel)

	cert, key, err := generateCertificate()
	if err != nil {
		log.Fatal(err)
	}

	// can't user defer since this func uses os.Exit
	cleanup := func() {
		os.Remove(cert)
		os.Remove(key)
		os.Unsetenv(EnvVarEndpoints)
		os.Unsetenv(EnvVarAutoSyncInterval)
		os.Unsetenv(EnvVarDialKeepAliveTimeout)
		os.Unsetenv(EnvVarDialKeepAliveTime)
		os.Unsetenv(EnvVarDialTimeout)
		os.Unsetenv(EnvVarMaxCallRecvMsgSz)
		os.Unsetenv(EnvVarMaxCallSendMsgSz)
		os.Unsetenv(EnvVarTTL)
		os.Unsetenv(EnvVarRejectOldCluster)
		os.Unsetenv(EnvVarTLS)
		os.Unsetenv(EnvVarTLSInsecure)
		os.Unsetenv(EnvVarDialTimeout)
	}

	e, err := startEtcd(cert, key)
	if err != nil {
		log.Fatal(err)
	}
	<-e.Server.ReadyNotify()

	os.Setenv(EnvVarEndpoints, "https://127.0.0.1:2379")
	os.Setenv(EnvVarAutoSyncInterval, "10s")
	os.Setenv(EnvVarDialKeepAliveTimeout, "10s")
	os.Setenv(EnvVarDialKeepAliveTime, "10s")
	os.Setenv(EnvVarDialTimeout, "1s")
	os.Setenv(EnvVarDialTimeout, "10s")
	os.Setenv(EnvVarMaxCallRecvMsgSz, "0")
	os.Setenv(EnvVarMaxCallSendMsgSz, "0")
	os.Setenv(EnvVarTTL, "10s")
	os.Setenv(EnvVarRejectOldCluster, "false")
	os.Setenv(EnvVarTLS, "true")
	os.Setenv(EnvVarTLSInsecure, "true")

	if os.Getenv(EnvVarEndpoints) == "" {
		os.Exit(0)
	}

	p, err = New(context.TODO(), "/gocsi/etcd", 0, nil)
	if err != nil {
		log.Fatalln(err)
	}
	exitCode := m.Run()
	p.(io.Closer).Close()
	cleanup()
	os.Exit(exitCode)
}

func TestTryMutex_Lock(t *testing.T) {
	var (
		i     int
		id    = t.Name()
		wait  sync.WaitGroup
		ready = make(chan struct{}, 5)
		mu    sync.Mutex // Mutex to protect access to i
	)

	// Wait for the goroutines with the other mutexes to finish, otherwise
	// those mutexes won't unlock and close their concurrency sessions to etcd.
	wait.Add(5)
	defer wait.Wait()

	// The context used when creating new locks and their concurrency sessions.
	ctx := context.Background()

	// The context used for the Lock functions.
	lockCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	m, err := p.GetLockWithID(ctx, id)
	if err != nil {
		t.Error(err)
		return
	}
	m.Lock()

	// Unlock m and close its session before exiting the test.
	defer m.(io.Closer).Close()
	defer m.Unlock()

	// Start five goroutines that all attempt to lock m and increment i.
	for j := 0; j < 5; j++ {
		go func() {
			defer wait.Done()

			m, err := p.GetLockWithID(ctx, id)
			if err != nil {
				t.Error(err)
				ready <- struct{}{}
				return
			}

			defer m.(io.Closer).Close()
			m.(*TryMutex).LockCtx = lockCtx

			ready <- struct{}{}
			m.Lock()
			mu.Lock()
			i++
			mu.Unlock()
		}()
	}

	// Give the above loop enough time to start the goroutines.
	<-ready
	time.Sleep(time.Duration(3) * time.Second)

	// Assert that i should have only been incremented once since only
	// one lock should have been obtained.
	if i > 0 {
		t.Errorf("i != 1: %d", i)
	}
}

func ExampleTryMutex_TryLock() {
	const lockName = "ExampleTryMutex_TryLock"

	// The context used when creating new locks and their concurrency sessions.
	ctx := context.Background()

	// Assign a TryMutex to m1 and then lock m1.
	m1, err := p.GetLockWithName(ctx, lockName)
	if err != nil {
		log.Error(err)
		return
	}
	defer m1.(io.Closer).Close()
	m1.Lock()

	// Start a goroutine that sleeps for one second and then
	// unlocks m1. This makes it possible for the TryLock
	// call below to lock m2.
	go func() {
		time.Sleep(time.Duration(1) * time.Second)
		m1.Unlock()
	}()

	// Try for three seconds to lock m2.
	m2, err := p.GetLockWithName(ctx, lockName)
	if err != nil {
		log.Error(err)
		return
	}
	defer m2.(io.Closer).Close()
	if m2.TryLock(time.Duration(3) * time.Second) {
		fmt.Println("lock obtained")
	}
	m2.Unlock()

	// Output: lock obtained
}

func ExampleTryMutex_TryLock_timeout() {
	const lockName = "ExampleTryMutex_TryLock_timeout"

	// The context used when creating new locks and their concurrency sessions.
	ctx := context.Background()

	// Assign a TryMutex to m1 and then lock m1.
	m1, err := p.GetLockWithName(ctx, lockName)
	if err != nil {
		log.Error(err)
		return
	}
	defer m1.(io.Closer).Close()
	defer m1.Unlock()
	m1.Lock()

	// Try for three seconds to lock m2.
	m2, err := p.GetLockWithName(ctx, lockName)
	if err != nil {
		log.Error(err)
		return
	}
	defer m2.(io.Closer).Close()
	if !m2.TryLock(time.Duration(3) * time.Second) {
		fmt.Println("lock not obtained")
	}

	// Output: lock not obtained
}

func startEtcd(cert string, key string) (*embed.Etcd, error) {
	cfg := embed.NewConfig()
	cfg.Dir = "/tmp/etcd-data"
	cfg.ListenClientUrls = []url.URL{{Scheme: "https", Host: "127.0.0.1:2379"}}
	cfg.ClientTLSInfo = transport.TLSInfo{
		CertFile: cert,
		KeyFile:  key,
	}
	cfg.PeerTLSInfo = transport.TLSInfo{
		CertFile: cert,
		KeyFile:  key,
	}
	cfg.ClientAutoTLS = false
	cfg.PeerAutoTLS = false

	e, err := embed.StartEtcd(cfg)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func generateCertificate() (string, string, error) {
	cert := "cert.pem"
	key := "key.pem"

	// Generate a private key
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", err
	}

	// Create a template for the certificate
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Dell"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Create the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return "", "", err
	}

	// Save the certificate to a file
	certOut, err := os.Create(cert)
	if err != nil {
		return "", "", err
	}
	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	if err != nil {
		return "", "", err
	}
	certOut.Close()

	// Save the private key to a file
	keyOut, err := os.Create(key)
	if err != nil {
		return "", "", err
	}
	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return "", "", err
	}
	pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})
	keyOut.Close()

	return cert, key, nil
}

func TestInitConfig(t *testing.T) {
	ctx := context.Background()

	validVars := map[string]string{
		EnvVarEndpoints:            "127.0.0.1:2379",
		EnvVarAutoSyncInterval:     "10s",
		EnvVarDialKeepAliveTime:    "10s",
		EnvVarDialKeepAliveTimeout: "10s",
		EnvVarDialTimeout:          "10s",
		EnvVarMaxCallSendMsgSz:     "2097152",
		EnvVarMaxCallRecvMsgSz:     "32",
		EnvVarUsername:             "user1name",
		EnvVarPassword:             "pass7word",
		EnvVarTLS:                  "true",
		EnvVarTLSInsecure:          "true",
		EnvVarRejectOldCluster:     "true",
	}

	cloneVarsMap := func(key string, value string) map[string]string {
		cloned := make(map[string]string)
		for k, v := range validVars {
			cloned[k] = v
		}
		if key != "" {
			cloned[key] = value
		}
		return cloned
	}

	tests := []struct {
		name          string
		envVarName    string
		envVars       map[string]string
		expectedError bool
		configChecker func(t *testing.T, cfg etcd.Config)
	}{
		{
			name:          "Valid EnvVars",
			envVars:       cloneVarsMap("", ""),
			expectedError: false,
			configChecker: func(t *testing.T, gotConfig etcd.Config) {
				assert.Equal(t, []string{"127.0.0.1:2379"}, gotConfig.Endpoints)
				assert.Equal(t, time.Second*10, gotConfig.AutoSyncInterval)
				assert.Equal(t, time.Second*10, gotConfig.DialKeepAliveTime)
				assert.Equal(t, time.Second*10, gotConfig.DialKeepAliveTimeout)
				assert.Equal(t, time.Second*10, gotConfig.DialTimeout)
				assert.Equal(t, 2097152, gotConfig.MaxCallSendMsgSize)
				assert.Equal(t, 32, gotConfig.MaxCallRecvMsgSize)
				assert.Equal(t, "user1name", gotConfig.Username)
				assert.Equal(t, "pass7word", gotConfig.Password)
				assert.NotNil(t, gotConfig.TLS)
				assert.Equal(t, true, gotConfig.TLS.InsecureSkipVerify)
				assert.Equal(t, true, gotConfig.RejectOldCluster)
			},
		},
		{
			name:          "Invalid AutoSyncInterval",
			envVars:       cloneVarsMap(EnvVarAutoSyncInterval, "split second"),
			expectedError: true,
		},
		{
			name:          "Invalid DialKeepAliveTime",
			envVars:       cloneVarsMap(EnvVarDialKeepAliveTime, "often"),
			expectedError: true,
		},
		{
			name:          "Invalid DialKeepAliveTimeout",
			envVars:       cloneVarsMap(EnvVarDialKeepAliveTimeout, "shortly"),
			expectedError: true,
		},
		{
			name:          "Invalid DialTimeout",
			envVars:       cloneVarsMap(EnvVarDialTimeout, "nevergiveup"),
			expectedError: true,
		},
		{
			name:          "Invalid MaxCallSendMsgSz",
			envVars:       cloneVarsMap(EnvVarMaxCallSendMsgSz, "bad"),
			expectedError: true,
		},
		{
			name:          "Invalid MaxCallRecvMsgSz",
			envVars:       cloneVarsMap(EnvVarMaxCallRecvMsgSz, "wrong"),
			expectedError: true,
		},
		{
			name:          "Invalid TLS",
			envVars:       cloneVarsMap(EnvVarTLS, "troo"),
			expectedError: true,
		},
		{
			name:          "Invalid TLSInsecure",
			envVars:       cloneVarsMap(EnvVarTLSInsecure, "!"),
			expectedError: true,
		},
		{
			name:          "Invalid RejectOldCluster",
			envVars:       cloneVarsMap(EnvVarRejectOldCluster, "maybe"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				_ = os.Setenv(k, v)
			}
			defer func() {
				for k := range tt.envVars {
					_ = os.Unsetenv(k)
				}
			}()

			config, err := initConfig(ctx, make(map[string]interface{}))

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				tt.configChecker(t, config)
			}
		})
	}
}
