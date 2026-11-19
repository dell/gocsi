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
	"sync"

	"github.com/akutz/gosync"
)

// refCountedLock wraps a TryLocker with a reference count so it can be
// evicted from the map once no goroutine holds a reference.
type refCountedLock struct {
	gosync.TryLocker
	refCount int
}

type defaultLockProvider struct {
	volIDLocksL   sync.Mutex
	volNameLocksL sync.Mutex
	volIDLocks    map[string]*refCountedLock
	volNameLocks  map[string]*refCountedLock
}

// lockHandle is the per-call wrapper returned to callers. It implements
// gosync.TryLocker for normal lock operations and io.Closer to decrement
// the reference count and evict the map entry when no longer in use.
type lockHandle struct {
	gosync.TryLocker
	releaseFunc func()
	once        sync.Once
}

// Close decrements the reference count of the underlying lock entry and
// removes it from the map when no goroutine holds a reference.
func (h *lockHandle) Close() error {
	h.once.Do(h.releaseFunc)
	return nil
}

func (i *defaultLockProvider) GetLockWithID(
	_ context.Context, id string,
) (gosync.TryLocker, error) {
	i.volIDLocksL.Lock()
	defer i.volIDLocksL.Unlock()
	entry := i.volIDLocks[id]
	if entry == nil {
		entry = &refCountedLock{TryLocker: &gosync.TryMutex{}}
		i.volIDLocks[id] = entry
	}
	entry.refCount++
	return &lockHandle{
		TryLocker: entry.TryLocker,
		releaseFunc: func() {
			i.volIDLocksL.Lock()
			defer i.volIDLocksL.Unlock()
			entry.refCount--
			if entry.refCount == 0 {
				delete(i.volIDLocks, id)
			}
		},
	}, nil
}

func (i *defaultLockProvider) GetLockWithName(
	_ context.Context, name string,
) (gosync.TryLocker, error) {
	i.volNameLocksL.Lock()
	defer i.volNameLocksL.Unlock()
	entry := i.volNameLocks[name]
	if entry == nil {
		entry = &refCountedLock{TryLocker: &gosync.TryMutex{}}
		i.volNameLocks[name] = entry
	}
	entry.refCount++
	return &lockHandle{
		TryLocker: entry.TryLocker,
		releaseFunc: func() {
			i.volNameLocksL.Lock()
			defer i.volNameLocksL.Unlock()
			entry.refCount--
			if entry.refCount == 0 {
				delete(i.volNameLocks, name)
			}
		},
	}, nil
}
