package types

import (
	"context"

	"github.com/akutz/gosync"
)

// VolumeLockerProvider is able to provide gosync.TryLocker objects for
// volumes by ID and name.
type VolumeLockerProvider interface {
	// GetLockWithID gets a lock for a volume with provided ID. If a lock
	// for the specified volume ID does not exist then a new lock is created
	// and returned.
	GetLockWithID(ctx context.Context, id string) (gosync.TryLocker, error)

	// GetLockWithName gets a lock for a volume with provided name. If a lock
	// for the specified volume name does not exist then a new lock is created
	// and returned.
	GetLockWithName(ctx context.Context, name string) (gosync.TryLocker, error)
}

// Method to test the interface methods above
func testInterfaceMethods(l VolumeLockerProvider, id, name string) error {
	ctx := context.Background()

	_, err := l.GetLockWithID(ctx, id)
	if err != nil {
		return err
	}
	_, err = l.GetLockWithName(ctx, name)
	if err != nil {
		return err
	}
	return nil
}
