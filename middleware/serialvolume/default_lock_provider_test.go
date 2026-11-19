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
	"io"
	"testing"
)

func TestGetLockWithID(t *testing.T) {
	provider := &defaultLockProvider{
		volIDLocks:   make(map[string]*refCountedLock),
		volNameLocks: make(map[string]*refCountedLock),
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

	_, exists := provider.volIDLocks[id]
	if !exists {
		t.Errorf("lock not found for ID %s", id)
	}
}

func TestGetLockWithName(t *testing.T) {
	provider := &defaultLockProvider{
		volIDLocks:   make(map[string]*refCountedLock),
		volNameLocks: make(map[string]*refCountedLock),
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

	_, exists := provider.volNameLocks[name]
	if !exists {
		t.Errorf("lock not found for name %s", name)
	}
}

func TestDefaultLockProviderCleanup(t *testing.T) {
	t.Run("ID lock evicted after Close", func(t *testing.T) {
		provider := &defaultLockProvider{
			volIDLocks:   make(map[string]*refCountedLock),
			volNameLocks: make(map[string]*refCountedLock),
		}

		lock, err := provider.GetLockWithID(context.Background(), "vol-1")
		if err != nil {
			t.Fatal(err)
		}
		if len(provider.volIDLocks) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(provider.volIDLocks))
		}

		lock.(io.Closer).Close()
		if len(provider.volIDLocks) != 0 {
			t.Fatalf("expected 0 entries after Close, got %d", len(provider.volIDLocks))
		}
	})

	t.Run("Name lock evicted after Close", func(t *testing.T) {
		provider := &defaultLockProvider{
			volIDLocks:   make(map[string]*refCountedLock),
			volNameLocks: make(map[string]*refCountedLock),
		}

		lock, err := provider.GetLockWithName(context.Background(), "vol-name")
		if err != nil {
			t.Fatal(err)
		}
		if len(provider.volNameLocks) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(provider.volNameLocks))
		}

		lock.(io.Closer).Close()
		if len(provider.volNameLocks) != 0 {
			t.Fatalf("expected 0 entries after Close, got %d", len(provider.volNameLocks))
		}
	})

	t.Run("Concurrent references keep entry alive", func(t *testing.T) {
		provider := &defaultLockProvider{
			volIDLocks:   make(map[string]*refCountedLock),
			volNameLocks: make(map[string]*refCountedLock),
		}

		ctx := context.Background()
		lock1, _ := provider.GetLockWithID(ctx, "vol-1")
		lock2, _ := provider.GetLockWithID(ctx, "vol-1")

		if provider.volIDLocks["vol-1"].refCount != 2 {
			t.Fatalf("expected refCount 2, got %d", provider.volIDLocks["vol-1"].refCount)
		}

		lock1.(io.Closer).Close()
		if len(provider.volIDLocks) != 1 {
			t.Fatal("entry should still exist with one reference")
		}

		lock2.(io.Closer).Close()
		if len(provider.volIDLocks) != 0 {
			t.Fatal("entry should be evicted after all references released")
		}
	})

	t.Run("Close is idempotent", func(t *testing.T) {
		provider := &defaultLockProvider{
			volIDLocks:   make(map[string]*refCountedLock),
			volNameLocks: make(map[string]*refCountedLock),
		}

		lock, _ := provider.GetLockWithID(context.Background(), "vol-1")
		closer := lock.(io.Closer)
		closer.Close()
		closer.Close() // should not panic or double-decrement
		if len(provider.volIDLocks) != 0 {
			t.Fatal("expected 0 entries")
		}
	})
}
