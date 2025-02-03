package types

import (
	"context"
	"fmt"
	"testing"

	"github.com/akutz/gosync"
	"github.com/stretchr/testify/assert"
)

// MyType implements the VolumeLockerProvider interface
type MyType struct{}

// Methods for MyType to implement the VolumeLockerProvider interface
func (ml *MyType) GetLockWithID(ctx context.Context, id string) (gosync.TryLocker, error) {
	fmt.Println(ctx)
	fmt.Println(id)
	lock := &gosync.TryMutex{}
	return lock, nil
}

func (ml *MyType) GetLockWithName(ctx context.Context, name string) (gosync.TryLocker, error) {
	fmt.Println(ctx)
	fmt.Println(name)
	lock := &gosync.TryMutex{}
	return lock, nil
}

// Method to test the interface methods defined above
func testInterfaceMethods(l VolumeLockerProvider) error {
	ctx := context.Background()

	_, err := l.GetLockWithID(ctx, "testId")
	if err != nil {
		return err
	}
	_, err = l.GetLockWithName(ctx, "testName")
	if err != nil {
		return err
	}
	return nil
}

func TestVolumeLockerProvider(t *testing.T) {
	myType := &MyType{}
	err := testInterfaceMethods(myType)
	assert.NoError(t, err)
}
