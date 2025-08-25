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

type defaultLockProvider struct {
	volIDLocksL   sync.Mutex
	volNameLocksL sync.Mutex
	volIDLocks    map[string]gosync.TryLocker
	volNameLocks  map[string]gosync.TryLocker
}

func (i *defaultLockProvider) GetLockWithID(
	_ context.Context, id string,
) (gosync.TryLocker, error) {
	i.volIDLocksL.Lock()
	defer i.volIDLocksL.Unlock()
	lock := i.volIDLocks[id]
	if lock == nil {
		lock = &gosync.TryMutex{}
		i.volIDLocks[id] = lock
	}
	return lock, nil
}

func (i *defaultLockProvider) GetLockWithName(
	_ context.Context, name string,
) (gosync.TryLocker, error) {
	i.volNameLocksL.Lock()
	defer i.volNameLocksL.Unlock()
	lock := i.volNameLocks[name]
	if lock == nil {
		lock = &gosync.TryMutex{}
		i.volNameLocks[name] = lock
	}
	return lock, nil
}
