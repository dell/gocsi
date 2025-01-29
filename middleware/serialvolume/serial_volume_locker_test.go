package serialvolume

import (
	"context"
	"testing"
	"time"

	"github.com/akutz/gosync"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateVolume(t *testing.T) {
	locker := &defaultLockProvider{
		volIDLocks:   map[string]gosync.TryLocker{},
		volNameLocks: map[string]gosync.TryLocker{},
	}
	interceptor := New(WithTimeout(1*time.Second), WithLockProvider(locker))

	handler := func(_ context.Context, _ interface{}) (interface{}, error) {
		return &csi.CreateVolumeResponse{}, nil
	}

	req := &csi.CreateVolumeRequest{Name: "test-volume"}
	info := &grpc.UnaryServerInfo{}

	_, err := interceptor(context.Background(), req, info, handler)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCreateVolumeTimeout(t *testing.T) {
	locker := &MockVolumeLockerProvider{locks: make(map[string]bool)}
	interceptor := New(WithLockProvider(locker), WithTimeout(1*time.Millisecond))

	handler := func(_ context.Context, _ interface{}) (interface{}, error) {
		return &csi.CreateVolumeResponse{}, nil
	}

	req := &csi.CreateVolumeRequest{Name: "test-volume"}
	info := &grpc.UnaryServerInfo{}

	// Lock the volume to simulate a timeout
	lock, _ := locker.GetLockWithName(context.Background(), req.Name)
	lock.TryLock(0)

	_, err := interceptor(context.Background(), req, info, handler)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if status.Code(err) != codes.Aborted {
		t.Fatalf("expected Aborted error, got %v", err)
	}
}

func TestControllerPublishVolume(t *testing.T) {
	interceptor := New(WithTimeout(1 * time.Second))

	handler := func(_ context.Context, _ interface{}) (interface{}, error) {
		return &csi.ControllerPublishVolumeResponse{}, nil
	}

	req := &csi.ControllerPublishVolumeRequest{VolumeId: "test-volume"}
	info := &grpc.UnaryServerInfo{}

	_, err := interceptor(context.Background(), req, info, handler)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestControllerPublishVolumeTimeout(t *testing.T) {
	locker := &MockVolumeLockerProvider{locks: make(map[string]bool)}
	interceptor := New(WithLockProvider(locker), WithTimeout(1*time.Millisecond))

	handler := func(_ context.Context, _ interface{}) (interface{}, error) {
		return &csi.ControllerPublishVolumeResponse{}, nil
	}

	req := &csi.ControllerPublishVolumeRequest{VolumeId: "test-volume"}
	info := &grpc.UnaryServerInfo{}

	// Lock the volume to simulate a timeout
	lock, _ := locker.GetLockWithID(context.Background(), req.VolumeId)
	lock.TryLock(0)

	_, err := interceptor(context.Background(), req, info, handler)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if status.Code(err) != codes.Aborted {
		t.Fatalf("expected Aborted error, got %v", err)
	}
}

func TestControllerUnpublishVolume(t *testing.T) {
	interceptor := New(WithTimeout(1 * time.Second))

	handler := func(_ context.Context, _ interface{}) (interface{}, error) {
		return &csi.ControllerUnpublishVolumeResponse{}, nil
	}

	req := &csi.ControllerUnpublishVolumeRequest{VolumeId: "test-volume"}
	info := &grpc.UnaryServerInfo{}

	_, err := interceptor(context.Background(), req, info, handler)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestControllerUnpublishVolumeTimeout(t *testing.T) {
	locker := &MockVolumeLockerProvider{locks: make(map[string]bool)}
	interceptor := New(WithLockProvider(locker), WithTimeout(1*time.Millisecond))

	handler := func(_ context.Context, _ interface{}) (interface{}, error) {
		return &csi.ControllerUnpublishVolumeResponse{}, nil
	}

	req := &csi.ControllerUnpublishVolumeRequest{VolumeId: "test-volume"}
	info := &grpc.UnaryServerInfo{}

	// Lock the volume to simulate a timeout
	lock, _ := locker.GetLockWithID(context.Background(), req.VolumeId)
	lock.TryLock(0)

	_, err := interceptor(context.Background(), req, info, handler)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if status.Code(err) != codes.Aborted {
		t.Fatalf("expected Aborted error, got %v", err)
	}
}

func TestDeleteVolume(t *testing.T) {
	interceptor := New(WithTimeout(1 * time.Second))

	handler := func(_ context.Context, _ interface{}) (interface{}, error) {
		return &csi.DeleteVolumeResponse{}, nil
	}

	req := &csi.DeleteVolumeRequest{VolumeId: "test-volume"}
	info := &grpc.UnaryServerInfo{}

	_, err := interceptor(context.Background(), req, info, handler)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDeleteVolumeTimeout(t *testing.T) {
	locker := &MockVolumeLockerProvider{locks: make(map[string]bool)}
	interceptor := New(WithLockProvider(locker), WithTimeout(1*time.Millisecond))

	handler := func(_ context.Context, _ interface{}) (interface{}, error) {
		return &csi.DeleteVolumeResponse{}, nil
	}

	req := &csi.DeleteVolumeRequest{VolumeId: "test-volume"}
	info := &grpc.UnaryServerInfo{}

	// Lock the volume to simulate a timeout
	lock, _ := locker.GetLockWithID(context.Background(), req.VolumeId)
	lock.TryLock(0)

	_, err := interceptor(context.Background(), req, info, handler)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if status.Code(err) != codes.Aborted {
		t.Fatalf("expected Aborted error, got %v", err)
	}
}

func TestNodePublishVolume(t *testing.T) {
	interceptor := New(WithTimeout(1 * time.Second))

	handler := func(_ context.Context, _ interface{}) (interface{}, error) {
		return &csi.NodePublishVolumeResponse{}, nil
	}

	req := &csi.NodePublishVolumeRequest{VolumeId: "test-volume"}
	info := &grpc.UnaryServerInfo{}

	_, err := interceptor(context.Background(), req, info, handler)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestNodePublishVolumeTimeout(t *testing.T) {
	locker := &MockVolumeLockerProvider{locks: make(map[string]bool)}
	interceptor := New(WithLockProvider(locker), WithTimeout(1*time.Millisecond))

	handler := func(_ context.Context, _ interface{}) (interface{}, error) {
		return &csi.NodePublishVolumeResponse{}, nil
	}

	req := &csi.NodePublishVolumeRequest{VolumeId: "test-volume"}
	info := &grpc.UnaryServerInfo{}

	// Lock the volume to simulate a timeout
	lock, _ := locker.GetLockWithID(context.Background(), req.VolumeId)
	lock.TryLock(0)

	_, err := interceptor(context.Background(), req, info, handler)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if status.Code(err) != codes.Aborted {
		t.Fatalf("expected Aborted error, got %v", err)
	}
}

func TestNodeUnpublishVolume(t *testing.T) {
	interceptor := New(WithTimeout(1 * time.Second))

	handler := func(_ context.Context, _ interface{}) (interface{}, error) {
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}

	req := &csi.NodeUnpublishVolumeRequest{VolumeId: "test-volume"}
	info := &grpc.UnaryServerInfo{}

	_, err := interceptor(context.Background(), req, info, handler)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestNodeUnpublishVolumeTimeout(t *testing.T) {
	locker := &MockVolumeLockerProvider{locks: make(map[string]bool)}
	interceptor := New(WithLockProvider(locker), WithTimeout(1*time.Millisecond))

	handler := func(_ context.Context, _ interface{}) (interface{}, error) {
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}

	req := &csi.NodeUnpublishVolumeRequest{VolumeId: "test-volume"}
	info := &grpc.UnaryServerInfo{}

	// Lock the volume to simulate a timeout
	lock, _ := locker.GetLockWithID(context.Background(), req.VolumeId)
	lock.TryLock(0)

	_, err := interceptor(context.Background(), req, info, handler)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if status.Code(err) != codes.Aborted {
		t.Fatalf("expected Aborted error, got %v", err)
	}
}

// MockVolumeLockerProvider is a mock implementation of the VolumeLockerProvider interface.
type MockVolumeLockerProvider struct {
	locks map[string]bool
}

func (m *MockVolumeLockerProvider) GetLockWithID(_ context.Context, id string) (gosync.TryLocker, error) {
	if _, exists := m.locks[id]; !exists {
		m.locks[id] = false
	}
	return &MockLock{id: id, locks: m.locks}, nil
}

func (m *MockVolumeLockerProvider) GetLockWithName(_ context.Context, name string) (gosync.TryLocker, error) {
	if _, exists := m.locks[name]; !exists {
		m.locks[name] = false
	}
	return &MockLock{name: name, locks: m.locks}, nil
}

// MockLock is a mock implementation of the gosync.TryLocker interface.
type MockLock struct {
	id    string
	name  string
	locks map[string]bool
}

func (m *MockLock) TryLock(_ time.Duration) bool {
	if m.locks[m.id] {
		return false
	}
	m.locks[m.id] = true
	return true
}

func (m *MockLock) Lock() {
	m.locks[m.id] = true
}

func (m *MockLock) Unlock() {
	m.locks[m.id] = false
}

func (m *MockLock) Close() error {
	return nil
}
