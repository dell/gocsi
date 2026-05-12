/*
 *
 * Copyright © 2026 Dell Inc. or its subsidiaries. All Rights Reserved.
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
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *serviceClient) CreateVolumeGroupSnapshot(
	ctx context.Context,
	req *csi.CreateVolumeGroupSnapshotRequest, _ ...grpc.CallOption) (
	*csi.CreateVolumeGroupSnapshotResponse, error,
) {
	// if CTX has this key, we want to return error for UT
	if ctx.Value(ContextKey("returnError")) == "true" {
		return nil, status.Error(codes.InvalidArgument, "Returned error from mock ControllerGetVolume")
	}
	return s.service.CreateVolumeGroupSnapshot(ctx, req)
}

func (s *service) CreateVolumeGroupSnapshot(
	_ context.Context,
	req *csi.CreateVolumeGroupSnapshotRequest) (
	*csi.CreateVolumeGroupSnapshotResponse, error,
) {
	// Validate name
	if len(req.Name) == 0 {
		return nil, status.Errorf(codes.InvalidArgument,
			"exceeds size limit: Name: max=128, size=0")
	}
	if len(req.Name) > 128 {
		return nil, status.Errorf(codes.InvalidArgument,
			"exceeds size limit: Name: max=128, size=%d", len(req.Name))
	}

	// Validate source volume IDs
	if len(req.SourceVolumeIds) == 0 || req.SourceVolumeIds == nil {
		return nil, status.Error(codes.InvalidArgument, "required: SourceVolumeIds")
	}

	groupSnapshot := s.newGroupSnapshot(req.Name, req.SourceVolumeIds)
	s.groupSnapsRWL.Lock()
	defer s.groupSnapsRWL.Unlock()

	// Now we can safely append since we're working with pointers
	s.groupSnaps = append(s.groupSnaps, &groupSnapshot)

	return &csi.CreateVolumeGroupSnapshotResponse{
		GroupSnapshot: &groupSnapshot,
	}, nil
}

func (s *serviceClient) DeleteVolumeGroupSnapshot(
	ctx context.Context,
	req *csi.DeleteVolumeGroupSnapshotRequest, _ ...grpc.CallOption) (
	*csi.DeleteVolumeGroupSnapshotResponse, error,
) {
	// if CTX has this key, we want to return error for UT
	if ctx.Value(ContextKey("returnError")) == "true" {
		return nil, status.Error(codes.InvalidArgument, "Returned error from mock DeleteVolumeGroupSnapshot")
	}
	return s.service.DeleteVolumeGroupSnapshot(ctx, req)
}

func (s *service) DeleteVolumeGroupSnapshot(
	_ context.Context,
	req *csi.DeleteVolumeGroupSnapshotRequest) (
	*csi.DeleteVolumeGroupSnapshotResponse, error,
) {
	s.groupSnapsRWL.RLock()
	defer s.groupSnapsRWL.RUnlock()
	index := -1
	for i, groupSnapshot := range s.groupSnaps {
		if strings.EqualFold(groupSnapshot.GroupSnapshotId, req.GroupSnapshotId) {
			index = i
			break
		}
	}

	if index < 0 {
		return nil, status.Error(codes.NotFound, "Group snapshot not found")
	}

	// This delete logic preserves order and prevents potential memory
	// leaks. With pointers, we can simply set the last element to nil
	copy(s.groupSnaps[index:], s.groupSnaps[index+1:])
	s.groupSnaps[len(s.groupSnaps)-1] = nil
	s.groupSnaps = s.groupSnaps[:len(s.groupSnaps)-1]
	log.WithField("volumeGroupSnapshotID", req.GroupSnapshotId).Debug("mock delete volume")

	return &csi.DeleteVolumeGroupSnapshotResponse{}, nil
}

func (s *serviceClient) GetVolumeGroupSnapshot(
	ctx context.Context,
	req *csi.GetVolumeGroupSnapshotRequest, _ ...grpc.CallOption) (
	*csi.GetVolumeGroupSnapshotResponse, error,
) {
	// if CTX has this key, we want to return error for UT
	if ctx.Value(ContextKey("returnError")) == "true" {
		return nil, status.Error(codes.InvalidArgument, "Returned error from mock GetVolumeGroupSnapshot")
	}
	return s.service.GetVolumeGroupSnapshot(ctx, req)
}

func (s *service) GetVolumeGroupSnapshot(
	_ context.Context,
	req *csi.GetVolumeGroupSnapshotRequest) (
	*csi.GetVolumeGroupSnapshotResponse, error,
) {
	s.groupSnapsRWL.RLock()
	defer s.groupSnapsRWL.RUnlock()
	for _, groupSnapshot := range s.groupSnaps {
		if groupSnapshot.GroupSnapshotId == req.GroupSnapshotId {
			return &csi.GetVolumeGroupSnapshotResponse{
				GroupSnapshot: groupSnapshot,
			}, nil
		}
	}
	return nil, status.Error(codes.NotFound, "Group snapshot not found")
}

func (s *serviceClient) GroupControllerGetCapabilities(
	ctx context.Context,
	req *csi.GroupControllerGetCapabilitiesRequest, _ ...grpc.CallOption) (
	*csi.GroupControllerGetCapabilitiesResponse, error,
) {
	// if CTX has this key, we want to return error for UT
	if ctx.Value(ContextKey("returnError")) == "true" {
		return nil, status.Error(codes.InvalidArgument, "Returned error from mock ControllerGetCapabilities")
	}
	return s.service.GroupControllerGetCapabilities(ctx, req)
}

func (s *service) GroupControllerGetCapabilities(
	_ context.Context,
	_ *csi.GroupControllerGetCapabilitiesRequest) (
	*csi.GroupControllerGetCapabilitiesResponse, error,
) {
	return &csi.GroupControllerGetCapabilitiesResponse{
		Capabilities: []*csi.GroupControllerServiceCapability{
			{
				Type: &csi.GroupControllerServiceCapability_Rpc{
					Rpc: &csi.GroupControllerServiceCapability_RPC{
						Type: csi.GroupControllerServiceCapability_RPC_CREATE_DELETE_GET_VOLUME_GROUP_SNAPSHOT,
					},
				},
			},
		},
	}, nil
}

func (s *service) newGroupSnapshot(groupSnapID string, sourceVolIDs []string) csi.VolumeGroupSnapshot {
	return csi.VolumeGroupSnapshot{
		GroupSnapshotId: groupSnapID,
		CreationTime:    timestamppb.Now(),
		ReadyToUse:      true,
		Snapshots:       s.newSnapshotsForGroup(sourceVolIDs),
	}
}

func (s *service) newSnapshotsForGroup(sourceVolIDs []string) []*csi.Snapshot {
	ts := timestamppb.Now()
	var snapshots []*csi.Snapshot
	for _, sourceVolID := range sourceVolIDs {
		snapshot := &csi.Snapshot{
			SnapshotId:     sourceVolID,
			SourceVolumeId: sourceVolID,
			SizeBytes:      1024,
			CreationTime:   ts,
			ReadyToUse:     true,
		}
		snapshots = append(snapshots, snapshot)
	}
	return snapshots
}
