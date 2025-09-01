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
