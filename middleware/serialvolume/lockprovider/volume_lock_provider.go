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
