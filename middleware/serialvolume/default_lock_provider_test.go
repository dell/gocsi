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
