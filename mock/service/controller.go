package service

import (
	"fmt"
	"math"
	"path"
	"strconv"
	"strings"

	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi"
	utils "github.com/dell/gocsi/utils/csi"
)

func (s *serviceClient) CreateVolume(
	ctx context.Context,
	req *csi.CreateVolumeRequest, _ ...grpc.CallOption) (
	*csi.CreateVolumeResponse, error,
) {
	// if CTX has this key, we want to return error for UT
	if ctx.Value(ContextKey("returnError")) == "true" {
		return nil, status.Error(codes.InvalidArgument, "Returned error from mock CreateVolume")
	}
	return s.service.CreateVolume(ctx, req)
}

func (s *service) CreateVolume(
	ctx context.Context,
	req *csi.CreateVolumeRequest) (
	*csi.CreateVolumeResponse, error,
) {
	if len(req.Name) > 128 {
		return nil, status.Errorf(codes.InvalidArgument,
			"exceeds size limit: Name: max=128, size=%d", len(req.Name))
	}

	for k, v := range req.Parameters {
		if len(k) > 128 {
			return nil, status.Errorf(codes.InvalidArgument,
				"exceeds size limit: Parameters[%s]: max=128, size=%d", k, len(k))
		}

		if len(v) > 128 {
			return nil, status.Errorf(codes.InvalidArgument,
				"exceeds size limit: Parameters[%s]: max=128, size=%d", k, len(v))
		}
	}

	// Check to see if the volume already exists.
	if i, v := s.findVolByName(ctx, req.Name); i >= 0 {
		return &csi.CreateVolumeResponse{Volume: &v}, nil
	}

	// If no capacity is specified then use 100GiB
	capacity := utils.Gib100
	if cr := req.CapacityRange; cr != nil {
		if rb := cr.RequiredBytes; rb > 0 {
			capacity = rb
		}
		if lb := cr.LimitBytes; lb > 0 {
			capacity = lb
		}
	}

	// Create the volume and add it to the service's in-mem volume slice.
	v := s.newVolume(req.Name, capacity)
	s.volsRWL.Lock()
	defer s.volsRWL.Unlock()
	s.vols = append(s.vols, v)

	return &csi.CreateVolumeResponse{Volume: &v}, nil
}

func (s *serviceClient) DeleteVolume(
	ctx context.Context,
	req *csi.DeleteVolumeRequest, _ ...grpc.CallOption) (
	*csi.DeleteVolumeResponse, error,
) {
	// if CTX has this key, we want to return error for UT
	if ctx.Value(ContextKey("returnError")) == "true" {
		return nil, status.Error(codes.InvalidArgument, "Returned error from mock DeleteVolume")
	}
	return s.service.DeleteVolume(ctx, req)
}

func (s *service) DeleteVolume(
	_ context.Context,
	req *csi.DeleteVolumeRequest) (
	*csi.DeleteVolumeResponse, error,
) {
	s.volsRWL.Lock()
	defer s.volsRWL.Unlock()

	// If the volume does not exist then return an idempotent response.
	i, _ := s.findVolNoLock("id", req.VolumeId)
	if i < 0 {
		return &csi.DeleteVolumeResponse{}, nil
	}

	// This delete logic preserves order and prevents potential memory
	// leaks. The slice's elements may not be pointers, but the structs
	// themselves have fields that are.
	copy(s.vols[i:], s.vols[i+1:])
	s.vols[len(s.vols)-1] = csi.Volume{}
	s.vols = s.vols[:len(s.vols)-1]
	log.WithField("volumeID", req.VolumeId).Debug("mock delete volume")
	return &csi.DeleteVolumeResponse{}, nil
}

func (s *serviceClient) ControllerPublishVolume(
	ctx context.Context,
	req *csi.ControllerPublishVolumeRequest, _ ...grpc.CallOption) (
	*csi.ControllerPublishVolumeResponse, error,
) {
	// if CTX has this key, we want to return error for UT
	if ctx.Value(ContextKey("returnError")) == "true" {
		return nil, status.Error(codes.InvalidArgument, "Returned error from mock ControllerPublishVolume")
	}
	return s.service.ControllerPublishVolume(ctx, req)
}

func (s *service) ControllerPublishVolume(
	_ context.Context,
	req *csi.ControllerPublishVolumeRequest) (
	*csi.ControllerPublishVolumeResponse, error,
) {
	s.volsRWL.Lock()
	defer s.volsRWL.Unlock()

	i, v := s.findVolNoLock("id", req.VolumeId)
	if i < 0 {
		return nil, status.Error(codes.NotFound, req.VolumeId)
	}

	// devPathKey is the key in the volume's attributes that is set to a
	// mock device path if the volume has been published by the controller
	// to the specified node.
	devPathKey := path.Join(req.NodeId, "dev")

	// Check to see if the volume is already published.
	if device := v.VolumeContext[devPathKey]; device != "" {
		return &csi.ControllerPublishVolumeResponse{
			PublishContext: map[string]string{
				"device": device,
			},
		}, nil
	}

	// Publish the volume.
	device := "/dev/mock"
	v.VolumeContext[devPathKey] = device
	s.vols[i] = v

	return &csi.ControllerPublishVolumeResponse{
		PublishContext: map[string]string{
			"device": device,
		},
	}, nil
}

func (s *serviceClient) ControllerUnpublishVolume(
	ctx context.Context,
	req *csi.ControllerUnpublishVolumeRequest, _ ...grpc.CallOption) (
	*csi.ControllerUnpublishVolumeResponse, error,
) {
	// if CTX has this key, we want to return error for UT
	if ctx.Value(ContextKey("returnError")) == "true" {
		return nil, status.Error(codes.InvalidArgument, "Returned error from mock ControllerUnpublishVolume")
	}
	return s.service.ControllerUnpublishVolume(ctx, req)
}

func (s *service) ControllerUnpublishVolume(
	_ context.Context,
	req *csi.ControllerUnpublishVolumeRequest) (
	*csi.ControllerUnpublishVolumeResponse, error,
) {
	s.volsRWL.Lock()
	defer s.volsRWL.Unlock()

	i, v := s.findVolNoLock("id", req.VolumeId)
	if i < 0 {
		return nil, status.Error(codes.NotFound, req.VolumeId)
	}

	// devPathKey is the key in the volume's attributes that is set to a
	// mock device path if the volume has been published by the controller
	// to the specified node.
	devPathKey := path.Join(req.NodeId, "dev")

	// if NodeID is not blank, unpublish from just that node
	if req.NodeId != "" {
		// Check to see if the volume is already unpublished.
		if v.VolumeContext[devPathKey] == "" {
			return &csi.ControllerUnpublishVolumeResponse{}, nil
		}

		// Unpublish the volume.
		delete(v.VolumeContext, devPathKey)
	} else {
		// NodeID is blank, unpublish from all nodes, which can be identified by
		// ending with "/dev"
		for k := range v.VolumeContext {
			if strings.HasSuffix(k, devPathKey) {
				delete(v.VolumeContext, k)
			}
		}
	}
	s.vols[i] = v

	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

func (s *serviceClient) ValidateVolumeCapabilities(
	ctx context.Context,
	req *csi.ValidateVolumeCapabilitiesRequest, _ ...grpc.CallOption) (
	*csi.ValidateVolumeCapabilitiesResponse, error,
) {
	// if CTX has this key, we want to return error for UT
	if ctx.Value(ContextKey("returnError")) == "true" {
		return nil, status.Error(codes.InvalidArgument, "Returned error from mock ValidateVolumeCapabilities")
	}
	return s.service.ValidateVolumeCapabilities(ctx, req)
}

func (s *service) ValidateVolumeCapabilities(
	_ context.Context,
	req *csi.ValidateVolumeCapabilitiesRequest) (
	*csi.ValidateVolumeCapabilitiesResponse, error,
) {
	return &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{
			VolumeContext:      req.GetVolumeContext(),
			VolumeCapabilities: req.GetVolumeCapabilities(),
			Parameters:         req.GetParameters(),
		},
	}, nil
}

func (s *serviceClient) ListVolumes(
	ctx context.Context,
	req *csi.ListVolumesRequest, _ ...grpc.CallOption) (
	*csi.ListVolumesResponse, error,
) {
	// if CTX has this key, we want to return error for UT
	if ctx.Value(ContextKey("returnError")) == "true" {
		return nil, status.Error(codes.InvalidArgument, "Returned error from mock ListVolumes")
	}
	return s.service.ListVolumes(ctx, req)
}

func (s *service) ListVolumes(
	_ context.Context,
	req *csi.ListVolumesRequest) (
	*csi.ListVolumesResponse, error,
) {
	// Copy the mock volumes into a new slice in order to avoid
	// locking the service's volume slice for the duration of the
	// ListVolumes RPC.
	var vols []csi.Volume
	func() {
		s.volsRWL.RLock()
		defer s.volsRWL.RUnlock()
		vols = make([]csi.Volume, len(s.vols))
		copy(vols, s.vols)
	}()

	var (
		ulenVols      = int64(len(vols))
		maxEntries    = req.MaxEntries
		startingToken int64
	)

	if v := req.StartingToken; v != "" {
		i, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return nil, status.Errorf(
				codes.InvalidArgument,
				"startingToken=%d !< int32=%d",
				startingToken, math.MaxUint32)
		}
		startingToken = i
	}

	if startingToken > ulenVols {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"startingToken=%d > len(vols)=%d",
			startingToken, ulenVols)
	}

	// Discern the number of remaining entries.
	// #nosec G115
	rem := int32(ulenVols - startingToken)

	// If maxEntries is 0 or greater than the number of remaining entries then
	// set maxEntries to the number of remaining entries.
	if maxEntries == 0 || maxEntries > rem {
		maxEntries = rem
	}

	var (
		i       int
		j       = startingToken
		entries = make(
			[]*csi.ListVolumesResponse_Entry,
			maxEntries)
	)

	for i = 0; i < len(entries); i++ {
		entries[i] = &csi.ListVolumesResponse_Entry{
			Volume: &vols[j],
		}
		j++
	}

	var nextToken string
	if n := startingToken + int64(i); n < ulenVols {
		nextToken = fmt.Sprintf("%d", n)
	}

	return &csi.ListVolumesResponse{
		Entries:   entries,
		NextToken: nextToken,
	}, nil
}

func (s *serviceClient) GetCapacity(
	ctx context.Context,
	req *csi.GetCapacityRequest, _ ...grpc.CallOption) (
	*csi.GetCapacityResponse, error,
) {
	// if CTX has this key, we want to return error for UT
	if ctx.Value(ContextKey("returnError")) == "true" {
		return nil, status.Error(codes.InvalidArgument, "Returned error from mock GetCapacity")
	}
	return s.service.GetCapacity(ctx, req)
}

func (s *service) GetCapacity(
	_ context.Context,
	_ *csi.GetCapacityRequest) (
	*csi.GetCapacityResponse, error,
) {
	return &csi.GetCapacityResponse{
		AvailableCapacity: utils.Tib100,
	}, nil
}

func (s *serviceClient) ControllerGetCapabilities(
	ctx context.Context,
	req *csi.ControllerGetCapabilitiesRequest, _ ...grpc.CallOption) (
	*csi.ControllerGetCapabilitiesResponse, error,
) {
	// if CTX has this key, we want to return error for UT
	if ctx.Value(ContextKey("returnError")) == "true" {
		return nil, status.Error(codes.InvalidArgument, "Returned error from mock ControllerGetCapabilities")
	}
	return s.service.ControllerGetCapabilities(ctx, req)
}

func (s *service) ControllerGetCapabilities(
	_ context.Context,
	_ *csi.ControllerGetCapabilitiesRequest) (
	*csi.ControllerGetCapabilitiesResponse, error,
) {
	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: []*csi.ControllerServiceCapability{
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
					},
				},
			},
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
					},
				},
			},
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_LIST_VOLUMES,
					},
				},
			},
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_GET_CAPACITY,
					},
				},
			},
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
					},
				},
			},
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
					},
				},
			},
		},
	}, nil
}

func (s *serviceClient) CreateSnapshot(
	ctx context.Context,
	req *csi.CreateSnapshotRequest, _ ...grpc.CallOption) (
	*csi.CreateSnapshotResponse, error,
) {
	// if CTX has this key, we want to return error
	// this allows for unit tests to force an error via ctx as needed
	if ctx.Value(ContextKey("returnError")) == "true" {
		return nil, status.Error(codes.InvalidArgument, "Returned error from mock CreateSnapshot")
	}
	return s.service.CreateSnapshot(ctx, req)
}

func (s *service) CreateSnapshot(
	_ context.Context,
	req *csi.CreateSnapshotRequest) (
	*csi.CreateSnapshotResponse, error,
) {
	snap := s.newSnapshot(req.Name, utils.Tib)
	s.snapsRWL.Lock()
	defer s.snapsRWL.Unlock()
	s.snaps = append(s.snaps, snap)

	return &csi.CreateSnapshotResponse{
		Snapshot: &snap,
	}, nil
}

func (s *serviceClient) DeleteSnapshot(
	ctx context.Context,
	req *csi.DeleteSnapshotRequest, _ ...grpc.CallOption) (
	*csi.DeleteSnapshotResponse, error,
) {
	// if CTX has this key, we want to return error for UT
	if ctx.Value(ContextKey("returnError")) == "true" {
		return nil, status.Error(codes.InvalidArgument, "Returned error from mock DeleteSnapshot")
	}
	return s.service.DeleteSnapshot(ctx, req)
}

func (s *service) DeleteSnapshot(
	_ context.Context,
	req *csi.DeleteSnapshotRequest) (
	*csi.DeleteSnapshotResponse, error,
) {
	if req.SnapshotId == "" {
		return nil, status.Error(codes.InvalidArgument, "required: SnapshotID")
	}

	return &csi.DeleteSnapshotResponse{}, nil
}

func (s *serviceClient) ListSnapshots(
	ctx context.Context,
	req *csi.ListSnapshotsRequest, _ ...grpc.CallOption) (
	*csi.ListSnapshotsResponse, error,
) {
	// if CTX has this key, we want to return error for UT
	if ctx.Value(ContextKey("returnError")) == "true" {
		return nil, status.Error(codes.InvalidArgument, "Returned error from mock ListSnapshots")
	}
	return s.service.ListSnapshots(ctx, req)
}

func (s *service) ListSnapshots(
	_ context.Context,
	req *csi.ListSnapshotsRequest) (
	*csi.ListSnapshotsResponse, error,
) {
	// Copy the mock snapshots into a new slice in order to avoid
	// locking the service's snapshot slice for the duration of the
	// ListSnapshots RPC.
	var snaps []csi.Snapshot
	func() {
		s.snapsRWL.RLock()
		defer s.snapsRWL.RUnlock()
		snaps = make([]csi.Snapshot, len(s.snaps))
		copy(snaps, s.snaps)
	}()

	var (
		ulensnaps     = int64(len(snaps))
		maxEntries    = req.MaxEntries
		startingToken int64
	)

	if s := req.StartingToken; s != "" {
		i, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return nil, status.Errorf(
				codes.InvalidArgument,
				"startingToken=%d !< int32=%d",
				startingToken, math.MaxUint32)
		}
		startingToken = i
	}

	if startingToken > ulensnaps {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"startingToken=%d > len(snaps)=%d",
			startingToken, ulensnaps)
	}

	// Discern the number of remaining entries.
	// #nosec G115
	rem := int32(ulensnaps - startingToken)

	// If maxEntries is 0 or greater than the number of remaining entries then
	// set maxEntries to the number of remaining entries.
	if maxEntries == 0 || maxEntries > rem {
		maxEntries = rem
	}

	var (
		i       int
		j       = startingToken
		entries = make(
			[]*csi.ListSnapshotsResponse_Entry,
			maxEntries)
	)

	log.WithField("entries", entries).WithField("rem", rem).WithField("maxEntries", maxEntries).Debug("KEK")
	for i = 0; i < len(entries); i++ {
		log.WithField("i", i).WithField("j", j).WithField("maxEntries", maxEntries).Debugf("rem: %d\n", rem)
		entries[i] = &csi.ListSnapshotsResponse_Entry{
			Snapshot: &snaps[j],
		}
		j++
	}

	var nextToken string
	if n := startingToken + int64(i); n < ulensnaps {
		nextToken = fmt.Sprintf("%d", n)
	}

	log.WithField("nextToken", nextToken).Debugf("Entries: %#v\n", entries)
	return &csi.ListSnapshotsResponse{
		Entries:   entries,
		NextToken: nextToken,
	}, nil
}

func (s *serviceClient) ControllerExpandVolume(
	ctx context.Context,
	req *csi.ControllerExpandVolumeRequest, _ ...grpc.CallOption) (
	*csi.ControllerExpandVolumeResponse, error,
) {
	// if CTX has this key, we want to return error for UT
	if ctx.Value(ContextKey("returnError")) == "true" {
		return nil, status.Error(codes.InvalidArgument, "Returned error from mock ControllerExpandVolume")
	}
	return s.service.ControllerExpandVolume(ctx, req)
}

func (s *service) ControllerExpandVolume(
	_ context.Context,
	req *csi.ControllerExpandVolumeRequest) (
	*csi.ControllerExpandVolumeResponse, error,
) {
	s.volsRWL.Lock()
	defer s.volsRWL.Unlock()

	i, v := s.findVolNoLock("id", req.VolumeId)
	if i < 0 {
		return nil, status.Error(codes.NotFound, req.VolumeId)
	}

	var capacity int64

	if cr := req.CapacityRange; cr != nil {
		if rb := cr.RequiredBytes; rb > 0 {
			capacity = rb
		}
		if lb := cr.LimitBytes; lb > 0 {
			capacity = lb
		}
	}

	if capacity < v.CapacityBytes {
		return nil, status.Error(codes.OutOfRange, "requested new capacity smaller than existing")
	}

	v.CapacityBytes = capacity

	return &csi.ControllerExpandVolumeResponse{
		CapacityBytes:         v.CapacityBytes,
		NodeExpansionRequired: false,
	}, nil
}

func (s *serviceClient) ControllerGetVolume(
	ctx context.Context,
	req *csi.ControllerGetVolumeRequest, _ ...grpc.CallOption) (
	*csi.ControllerGetVolumeResponse, error,
) {
	// if CTX has this key, we want to return error for UT
	if ctx.Value(ContextKey("returnError")) == "true" {
		return nil, status.Error(codes.InvalidArgument, "Returned error from mock ControllerGetVolume")
	}
	return s.service.ControllerGetVolume(ctx, req)
}

func (s *service) ControllerGetVolume(
	_ context.Context,
	_ *csi.ControllerGetVolumeRequest) (
	*csi.ControllerGetVolumeResponse, error,
) {
	return nil, status.Error(codes.Unimplemented, "")
}
