package service

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

func (s *service) Probe(
	_ context.Context,
	_ *csi.ProbeRequest) (
	*csi.ProbeResponse, error,
) {
	return &csi.ProbeResponse{
		Ready: &wrapperspb.BoolValue{Value: true},
	}, nil
}

func (s *serviceClient) Probe(
	_ context.Context,
	_ *csi.ProbeRequest,
	_ ...grpc.CallOption) (
	*csi.ProbeResponse, error,
) {
	return &csi.ProbeResponse{
		Ready: &wrapperspb.BoolValue{Value: true},
	}, nil
}

func (s *service) GetPluginInfo(
	_ context.Context,
	_ *csi.GetPluginInfoRequest) (
	*csi.GetPluginInfoResponse, error,
) {
	return &csi.GetPluginInfoResponse{
		Name:          Name,
		VendorVersion: VendorVersion,
		Manifest:      Manifest,
	}, nil
}

func (s *serviceClient) GetPluginInfo(
	_ context.Context,
	_ *csi.GetPluginInfoRequest,
	_ ...grpc.CallOption) (
	*csi.GetPluginInfoResponse, error,
) {
	return &csi.GetPluginInfoResponse{
		Name:          Name,
		VendorVersion: VendorVersion,
		Manifest:      Manifest,
	}, nil
}

func (s *service) GetPluginCapabilities(
	_ context.Context,
	_ *csi.GetPluginCapabilitiesRequest) (
	*csi.GetPluginCapabilitiesResponse, error,
) {
	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
			{
				Type: &csi.PluginCapability_VolumeExpansion_{
					VolumeExpansion: &csi.PluginCapability_VolumeExpansion{
						Type: csi.PluginCapability_VolumeExpansion_ONLINE,
					},
				},
			},
		},
	}, nil
}

func (s *serviceClient) GetPluginCapabilities(
	_ context.Context,
	_ *csi.GetPluginCapabilitiesRequest,
	_ ...grpc.CallOption) (
	*csi.GetPluginCapabilitiesResponse, error,
) {
	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
			{
				Type: &csi.PluginCapability_VolumeExpansion_{
					VolumeExpansion: &csi.PluginCapability_VolumeExpansion{
						Type: csi.PluginCapability_VolumeExpansion_ONLINE,
					},
				},
			},
		},
	}, nil
}
