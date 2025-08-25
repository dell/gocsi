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
