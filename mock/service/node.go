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
	"path"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"golang.org/x/net/context"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

type ContextKey string

func (s *service) NodeStageVolume(
	_ context.Context,
	_ *csi.NodeStageVolumeRequest) (
	*csi.NodeStageVolumeResponse, error,
) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *service) NodeUnstageVolume(
	_ context.Context,
	_ *csi.NodeUnstageVolumeRequest) (
	*csi.NodeUnstageVolumeResponse, error,
) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *service) NodePublishVolume(
	_ context.Context,
	req *csi.NodePublishVolumeRequest) (
	*csi.NodePublishVolumeResponse, error,
) {
	device, ok := req.PublishContext["device"]
	if !ok {
		return nil, status.Error(
			codes.InvalidArgument,
			"publish volume info 'device' key required")
	}

	s.volsRWL.Lock()
	defer s.volsRWL.Unlock()

	i, v := s.findVolNoLock("id", req.VolumeId)
	if i < 0 {
		return nil, status.Error(codes.NotFound, req.VolumeId)
	}

	// nodeMntPathKey is the key in the volume's attributes that is set to a
	// mock mount path if the volume has been published by the node
	nodeMntPathKey := path.Join(s.nodeID, req.TargetPath)

	// Check to see if the volume has already been published.
	if v.VolumeContext[nodeMntPathKey] != "" {

		// Requests marked Readonly fail due to volumes published by
		// the Mock driver supporting only RW mode.
		if req.Readonly {
			return nil, status.Error(codes.AlreadyExists, req.VolumeId)
		}

		return &csi.NodePublishVolumeResponse{}, nil
	}

	// Publish the volume.
	v.VolumeContext[nodeMntPathKey] = device
	s.vols[i] = v

	return &csi.NodePublishVolumeResponse{}, nil
}

func (s *service) NodeUnpublishVolume(
	_ context.Context,
	req *csi.NodeUnpublishVolumeRequest) (
	*csi.NodeUnpublishVolumeResponse, error,
) {
	s.volsRWL.Lock()
	defer s.volsRWL.Unlock()

	i, v := s.findVolNoLock("id", req.VolumeId)
	if i < 0 {
		return nil, status.Error(codes.NotFound, req.VolumeId)
	}

	// nodeMntPathKey is the key in the volume's attributes that is set to a
	// mock mount path if the volume has been published by the node
	nodeMntPathKey := path.Join(s.nodeID, req.TargetPath)

	// Check to see if the volume has already been unpublished.
	if v.VolumeContext[nodeMntPathKey] == "" {
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}

	// Unpublish the volume.
	delete(v.VolumeContext, nodeMntPathKey)
	s.vols[i] = v

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (s *service) NodeGetInfo(
	_ context.Context,
	_ *csi.NodeGetInfoRequest) (
	*csi.NodeGetInfoResponse, error,
) {
	return &csi.NodeGetInfoResponse{
		NodeId: s.nodeID,
	}, nil
}

func (s *service) NodeGetCapabilities(
	_ context.Context,
	_ *csi.NodeGetCapabilitiesRequest) (
	*csi.NodeGetCapabilitiesResponse, error,
) {
	return &csi.NodeGetCapabilitiesResponse{}, nil
}

func (s *service) NodeGetVolumeStats(
	_ context.Context,
	req *csi.NodeGetVolumeStatsRequest) (
	*csi.NodeGetVolumeStatsResponse, error,
) {
	var f *csi.Volume
	for _, v := range s.vols {
		if v.VolumeId == req.VolumeId {
			/* #nosec G601 */
			f = &v
		}
	}
	if f == nil {
		return nil, status.Errorf(codes.NotFound, "No volume found with id %s", req.VolumeId)
	}

	return &csi.NodeGetVolumeStatsResponse{
		Usage: []*csi.VolumeUsage{
			{
				Available: int64(float64(f.CapacityBytes) * 0.6),
				Total:     f.CapacityBytes,
				Used:      int64(float64(f.CapacityBytes) * 0.4),
				Unit:      csi.VolumeUsage_BYTES,
			},
		},
	}, nil
}

func (s *service) NodeExpandVolume(
	_ context.Context,
	_ *csi.NodeExpandVolumeRequest) (
	*csi.NodeExpandVolumeResponse, error,
) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *serviceClient) NodeStageVolume(
	ctx context.Context,
	_ *csi.NodeStageVolumeRequest, _ ...grpc.CallOption) (
	*csi.NodeStageVolumeResponse, error,
) {
	if ctx.Value(ContextKey("returnError")) == "true" {
		return nil, status.Error(codes.InvalidArgument, "Returned error from mock NodeStageVolume")
	}

	return &csi.NodeStageVolumeResponse{}, nil
}

func (s *serviceClient) NodeUnstageVolume(
	ctx context.Context,
	_ *csi.NodeUnstageVolumeRequest, _ ...grpc.CallOption) (
	*csi.NodeUnstageVolumeResponse, error,
) {
	if ctx.Value(ContextKey("returnError")) == "true" {
		return nil, status.Error(codes.InvalidArgument, "Returned error from mock NodeUnstageVolume")
	}

	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (s *serviceClient) NodePublishVolume(
	ctx context.Context,
	_ *csi.NodePublishVolumeRequest, _ ...grpc.CallOption) (
	*csi.NodePublishVolumeResponse, error,
) {
	if ctx.Value(ContextKey("returnError")) == "true" {
		return nil, status.Error(codes.InvalidArgument, "Returned error from mock NodePublishVolume")
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

func (s *serviceClient) NodeUnpublishVolume(
	ctx context.Context,
	_ *csi.NodeUnpublishVolumeRequest, _ ...grpc.CallOption) (
	*csi.NodeUnpublishVolumeResponse, error,
) {
	if ctx.Value(ContextKey("returnError")) == "true" {
		return nil, status.Error(codes.InvalidArgument, "Returned error from mock NodeUnpublishVolume")
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (s *serviceClient) NodeGetInfo(
	ctx context.Context,
	_ *csi.NodeGetInfoRequest, _ ...grpc.CallOption) (
	*csi.NodeGetInfoResponse, error,
) {
	if ctx.Value(ContextKey("returnError")) == "true" {
		return nil, status.Error(codes.InvalidArgument, "Returned error from mock NodeGetInfo")
	}

	return &csi.NodeGetInfoResponse{}, nil
}

func (s *serviceClient) NodeGetCapabilities(
	ctx context.Context,
	_ *csi.NodeGetCapabilitiesRequest, _ ...grpc.CallOption) (
	*csi.NodeGetCapabilitiesResponse, error,
) {
	// if CTX has this key, we want to return error for UT
	if ctx.Value(ContextKey("returnError")) == "true" {
		return nil, status.Error(codes.InvalidArgument, "Returned error from mock NodeGetCapabilities")
	}

	// send back one capability
	nodeCapabalities := []*csi.NodeServiceCapability{
		{
			// Required for NodeExpandVolume
			Type: &csi.NodeServiceCapability_Rpc{
				Rpc: &csi.NodeServiceCapability_RPC{
					Type: csi.NodeServiceCapability_RPC_EXPAND_VOLUME,
				},
			},
		},
	}

	return &csi.NodeGetCapabilitiesResponse{Capabilities: nodeCapabalities}, nil
}

func (s *serviceClient) NodeGetVolumeStats(
	ctx context.Context,
	_ *csi.NodeGetVolumeStatsRequest, _ ...grpc.CallOption) (
	*csi.NodeGetVolumeStatsResponse, error,
) {
	if ctx.Value(ContextKey("returnError")) == "true" {
		return nil, status.Error(codes.InvalidArgument, "Returned error from mock NodeGetVolumeStats")
	}

	return &csi.NodeGetVolumeStatsResponse{}, nil
}

func (s *serviceClient) NodeExpandVolume(
	ctx context.Context,
	_ *csi.NodeExpandVolumeRequest, _ ...grpc.CallOption) (
	*csi.NodeExpandVolumeResponse, error,
) {
	// if CTX has this key, we want to return error for UT
	if ctx.Value(ContextKey("returnError")) == "true" {
		return nil, status.Error(codes.InvalidArgument, "Returned error from mock NodeExpandVolume")
	}

	return &csi.NodeExpandVolumeResponse{}, nil
}
