package etcd_test

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

	csietcd "github.com/dell/gocsi/middleware/serialvolume/etcd"
	mwtypes "github.com/dell/gocsi/middleware/serialvolume/types"
	"go.etcd.io/etcd/client/pkg/v3/transport"
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
		os.Unsetenv(csietcd.EnvVarEndpoints)
		os.Unsetenv(csietcd.EnvVarAutoSyncInterval)
		os.Unsetenv(csietcd.EnvVarDialKeepAliveTimeout)
		os.Unsetenv(csietcd.EnvVarDialKeepAliveTime)
		os.Unsetenv(csietcd.EnvVarDialTimeout)
		os.Unsetenv(csietcd.EnvVarMaxCallRecvMsgSz)
		os.Unsetenv(csietcd.EnvVarMaxCallSendMsgSz)
		os.Unsetenv(csietcd.EnvVarTTL)
		os.Unsetenv(csietcd.EnvVarRejectOldCluster)
		os.Unsetenv(csietcd.EnvVarTLS)
		os.Unsetenv(csietcd.EnvVarTLSInsecure)
		os.Unsetenv(csietcd.EnvVarDialTimeout)
	}

	e, err := startEtcd(cert, key)
	if err != nil {
		log.Fatal(err)
	}
	<-e.Server.ReadyNotify()

	os.Setenv(csietcd.EnvVarEndpoints, "https://127.0.0.1:2379")
	os.Setenv(csietcd.EnvVarAutoSyncInterval, "10s")
	os.Setenv(csietcd.EnvVarDialKeepAliveTimeout, "10s")
	os.Setenv(csietcd.EnvVarDialKeepAliveTime, "10s")
	os.Setenv(csietcd.EnvVarDialTimeout, "1s")
	os.Setenv(csietcd.EnvVarDialTimeout, "10s")
	os.Setenv(csietcd.EnvVarMaxCallRecvMsgSz, "0")
	os.Setenv(csietcd.EnvVarMaxCallSendMsgSz, "0")
	os.Setenv(csietcd.EnvVarTTL, "10s")
	os.Setenv(csietcd.EnvVarRejectOldCluster, "false")
	os.Setenv(csietcd.EnvVarTLS, "true")
	os.Setenv(csietcd.EnvVarTLSInsecure, "true")

	if os.Getenv(csietcd.EnvVarEndpoints) == "" {
		os.Exit(0)
	}

	p, err = csietcd.New(context.TODO(), "/gocsi/etcd", 0, nil)
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
			m.(*csietcd.TryMutex).LockCtx = lockCtx

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
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
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
