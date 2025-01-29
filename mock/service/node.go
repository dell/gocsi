package service

import (
	"path"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"golang.org/x/net/context"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

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
	// return nil, status.Error(codes.Unimplemented, "")
	return &csi.NodeExpandVolumeResponse{}, nil
}

func (s *serviceClient) NodeStageVolume(
	_ context.Context,
	_ *csi.NodeStageVolumeRequest, _ ...grpc.CallOption) (
	*csi.NodeStageVolumeResponse, error,
) {
	return &csi.NodeStageVolumeResponse{}, nil
}

func (s *serviceClient) NodeUnstageVolume(
	_ context.Context,
	_ *csi.NodeUnstageVolumeRequest, _ ...grpc.CallOption) (
	*csi.NodeUnstageVolumeResponse, error,
) {
	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (s *serviceClient) NodePublishVolume(
	_ context.Context,
	_ *csi.NodePublishVolumeRequest, _ ...grpc.CallOption) (
	*csi.NodePublishVolumeResponse, error,
) {
	return &csi.NodePublishVolumeResponse{}, nil
}

func (s *serviceClient) NodeUnpublishVolume(
	_ context.Context,
	_ *csi.NodeUnpublishVolumeRequest, _ ...grpc.CallOption) (
	*csi.NodeUnpublishVolumeResponse, error,
) {
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (s *serviceClient) NodeGetInfo(
	_ context.Context,
	_ *csi.NodeGetInfoRequest, _ ...grpc.CallOption) (
	*csi.NodeGetInfoResponse, error,
) {
	return &csi.NodeGetInfoResponse{}, nil
}

func (s *serviceClient) NodeGetCapabilities(
	_ context.Context,
	_ *csi.NodeGetCapabilitiesRequest, _ ...grpc.CallOption) (
	*csi.NodeGetCapabilitiesResponse, error,
) {
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
	_ context.Context,
	_ *csi.NodeGetVolumeStatsRequest, _ ...grpc.CallOption) (
	*csi.NodeGetVolumeStatsResponse, error,
) {
	return &csi.NodeGetVolumeStatsResponse{}, nil
}

func (s *serviceClient) NodeExpandVolume(
	_ context.Context,
	_ *csi.NodeExpandVolumeRequest, _ ...grpc.CallOption) (
	*csi.NodeExpandVolumeResponse, error,
) {
	return &csi.NodeExpandVolumeResponse{}, nil
}
