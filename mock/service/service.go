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

package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/container-storage-interface/spec/lib/go/csi"
	utils "github.com/dell/gocsi/utils/csi"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// Name is the name of the CSI plug-in.
	Name = "mock.gocsi.rexray.com"

	// VendorVersion is the version returned by GetPluginInfo.
	VendorVersion = "1.1.0"
)

// Manifest is the SP's manifest.
var Manifest = map[string]string{
	"url": "https://github.com/dell/gocsi/tree/master/mock",
}

// Service is the CSI Mock service provider.
type MockServer interface {
	csi.ControllerServer
	csi.IdentityServer
	csi.NodeServer
}

type MockClient interface {
	csi.ControllerClient
	csi.NodeClient
	csi.IdentityClient
}

type service struct {
	sync.Mutex
	nodeID   string
	vols     []csi.Volume
	snaps    []csi.Snapshot
	volsRWL  sync.RWMutex
	snapsRWL sync.RWMutex
	volsNID  uint64
	snapsNID uint64
}

type serviceClient struct {
	service MockServer
}

// New returns a new Service.
func NewServer() MockServer {
	s := &service{nodeID: Name}

	// add some mock volumes to start with
	s.vols = []csi.Volume{
		s.newVolume("Mock Volume 1", utils.Gib100),
		s.newVolume("Mock Volume 2", utils.Gib100),
		s.newVolume("Mock Volume 3", utils.Gib100),
	}

	// add some mock snapshots to start with, too
	s.snaps = []csi.Snapshot{
		s.newSnapshot("Mock Snapshot 1", utils.Gib100),
		s.newSnapshot("Mock Snapshot 2", utils.Gib100),
		s.newSnapshot("Mock Snapshot 3", utils.Gib100),
	}
	return s
}

func NewClient() MockClient {
	return &serviceClient{
		service: NewServer(),
	}
}

func (s *service) newVolume(name string, capcity int64) csi.Volume {
	return csi.Volume{
		VolumeId:      fmt.Sprintf("%d", atomic.AddUint64(&s.volsNID, 1)),
		VolumeContext: map[string]string{"name": name},
		CapacityBytes: capcity,
	}
}

func (s *service) newSnapshot(_ string, size int64) csi.Snapshot {
	return csi.Snapshot{
		// We set the id to "<volume-id>:<snapshot-id>" since during delete requests
		// we are not given the parent volume id
		SnapshotId:     "12",
		SourceVolumeId: "4",
		SizeBytes:      size,
		CreationTime:   timestamppb.Now(),
		ReadyToUse:     true,
	}
}

func (s *service) findVol(k, v string) (volIdx int, volInfo csi.Volume) {
	s.volsRWL.RLock()
	defer s.volsRWL.RUnlock()
	return s.findVolNoLock(k, v)
}

func (s *service) findVolNoLock(k, v string) (volIdx int, volInfo csi.Volume) {
	volIdx = -1

	for i, vi := range s.vols {
		switch k {
		case "id":
			if strings.EqualFold(v, vi.VolumeId) {
				return i, vi
			}
		case "name":
			if n, ok := vi.VolumeContext["name"]; ok && strings.EqualFold(v, n) {
				return i, vi
			}
		}
	}

	return
}

func (s *service) findVolByName(
	_ context.Context, name string,
) (int, csi.Volume) {
	return s.findVol("name", name)
}
