package lockprovider

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

	// test an error case
	if id == "" {
		return nil, fmt.Errorf("empty ID")
	}
	lock := &gosync.TryMutex{}
	return lock, nil
}

func (ml *MyType) GetLockWithName(ctx context.Context, name string) (gosync.TryLocker, error) {
	fmt.Println(ctx)

	// test an error case
	if name == "" {
		return nil, fmt.Errorf("empty name")
	}
	lock := &gosync.TryMutex{}
	return lock, nil
}

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

func TestVolumeLockerProvider(t *testing.T) {
	myType := &MyType{}
	err := testInterfaceMethods(myType, "testId", "testName")
	assert.NoError(t, err)

	// empty ID should return error
	err = testInterfaceMethods(myType, "", "testName")
	assert.ErrorContains(t, err, "empty ID")

	// empty name should return error
	err = testInterfaceMethods(myType, "testId", "")
	assert.ErrorContains(t, err, "empty name")
}
