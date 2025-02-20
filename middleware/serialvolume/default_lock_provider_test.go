package serialvolume

import (
	"context"
	"testing"

	"github.com/akutz/gosync"
)

func TestGetLockWithID(t *testing.T) {
	provider := &defaultLockProvider{
		volIDLocks:   make(map[string]gosync.TryLocker),
		volNameLocks: make(map[string]gosync.TryLocker),
	}

	ctx := context.Background()
	id := "test-id"

	lock, err := provider.GetLockWithID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}

	if lock == nil {
		t.Error("expected non-nil lock")
	}

	storedLock, exists := provider.volIDLocks[id]
	if !exists {
		t.Errorf("lock not found for ID %s", id)
	}

	if lock != storedLock {
		t.Errorf("expected lock %v, got %v", lock, storedLock)
	}
}

func TestGetLockWithName(t *testing.T) {
	provider := &defaultLockProvider{
		volIDLocks:   make(map[string]gosync.TryLocker),
		volNameLocks: make(map[string]gosync.TryLocker),
	}

	ctx := context.Background()
	name := "test-name"

	lock, err := provider.GetLockWithName(ctx, name)
	if err != nil {
		t.Fatal(err)
	}

	if lock == nil {
		t.Error("expected non-nil lock")
	}

	// Ensure the lock is stored in the map
	storedLock, exists := provider.volNameLocks[name]
	if !exists {
		t.Errorf("lock not found for name %s", name)
	}

	if lock != storedLock {
		t.Errorf("expected lock %v, got %v", lock, storedLock)
	}
}
