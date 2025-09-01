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

package specvalidator

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestControllerCreateVolume(t *testing.T) {
	interceptor := NewServerSpecValidator(
		WithRequestValidation(),
		WithResponseValidation(),
		WithRequiresControllerCreateVolumeSecrets(),
	)

	info := &grpc.UnaryServerInfo{FullMethod: "/csi.v1.Controller/CreateVolume"}

	tests := []struct {
		name    string
		req     *csi.CreateVolumeRequest
		handler func(ctx context.Context, req interface{}) (interface{}, error)
		wantErr bool
	}{
		{
			name: "Valid Request",
			req: &csi.CreateVolumeRequest{
				Name:    "test-volume",
				Secrets: map[string]string{"foo": "bar"},
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: &csi.VolumeCapability_MountVolume{},
						},
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					},
				},
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.CreateVolumeResponse{
					Volume: &csi.Volume{
						VolumeId: "test-volume",
					},
				}, nil
			},
			wantErr: false,
		},
		{
			name: "Missing Name",
			req: &csi.CreateVolumeRequest{
				Secrets: map[string]string{"foo": "bar"},
			},
			wantErr: true,
		},
		{
			name: "Missing Secrets",
			req: &csi.CreateVolumeRequest{
				Name: "test-volume",
			},
			wantErr: true,
		},
		{
			name: "Missing Volume Response",
			req: &csi.CreateVolumeRequest{
				Name: "test-volume",
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: &csi.VolumeCapability_MountVolume{},
						},
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					},
				},
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.CreateVolumeResponse{
					Volume: nil,
				}, nil
			},
			wantErr: true,
		},
		{
			name: "Missing Volume ID Response",
			req: &csi.CreateVolumeRequest{
				Name: "test-volume",
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: &csi.VolumeCapability_MountVolume{},
						},
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					},
				},
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.CreateVolumeResponse{
					Volume: &csi.Volume{},
				}, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := interceptor(context.Background(), tt.req, info, tt.handler)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreateVolumeRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestControllerDeleteVolume(t *testing.T) {
	interceptor := NewServerSpecValidator(
		WithRequestValidation(),
		WithResponseValidation(),
		WithRequiresControllerDeleteVolumeSecrets(),
	)

	info := &grpc.UnaryServerInfo{FullMethod: "/csi.v1.Controller/DeleteVolume"}

	tests := []struct {
		name    string
		req     *csi.DeleteVolumeRequest
		handler func(ctx context.Context, req interface{}) (interface{}, error)
		wantErr bool
	}{
		{
			name: "Valid Request",
			req: &csi.DeleteVolumeRequest{
				VolumeId: "test-volume",
				Secrets:  map[string]string{"key": "value"},
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.DeleteVolumeResponse{}, nil
			},
			wantErr: false,
		},
		{
			name:    "Missing ID",
			req:     &csi.DeleteVolumeRequest{},
			wantErr: true,
		},
		{
			name: "Missing Secret",
			req: &csi.DeleteVolumeRequest{
				VolumeId: "test-volume",
				Secrets:  map[string]string{},
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.DeleteVolumeResponse{}, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := interceptor(context.Background(), tt.req, info, tt.handler)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDeleteVolumeRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestControllerPublishVolume(t *testing.T) {
	interceptor := NewServerSpecValidator(
		WithRequestValidation(),
		WithResponseValidation(),
		WithRequiresControllerPublishVolumeSecrets(),
		WithRequiresVolumeContext(),
	)

	info := &grpc.UnaryServerInfo{FullMethod: "/csi.v1.Controller/ControllerPublishVolume"}

	tests := []struct {
		name    string
		req     *csi.ControllerPublishVolumeRequest
		handler func(ctx context.Context, req interface{}) (interface{}, error)
		wantErr bool
	}{
		{
			name: "Valid Request",
			req: &csi.ControllerPublishVolumeRequest{
				VolumeId: "test-volume",
				NodeId:   "test-node",
				Secrets:  map[string]string{"key": "value"},
				VolumeCapability: &csi.VolumeCapability{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
				VolumeContext: map[string]string{"key": "value"},
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.ControllerPublishVolumeResponse{}, nil
			},
			wantErr: false,
		},
		{
			name: "Missing Volume Context",
			req: &csi.ControllerPublishVolumeRequest{
				VolumeId: "test-volume",
				NodeId:   "test-node",
				Secrets:  map[string]string{"key": "value"},
				VolumeCapability: &csi.VolumeCapability{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
				VolumeContext: map[string]string{},
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.ControllerPublishVolumeResponse{}, nil
			},
			wantErr: true,
		},
		{
			name: "Missing Node ID",
			req: &csi.ControllerPublishVolumeRequest{
				VolumeId: "test-volume",
				Secrets:  map[string]string{"key": "value"},
				VolumeCapability: &csi.VolumeCapability{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
				VolumeContext: map[string]string{"key": "value"},
			},

			wantErr: true,
		},
		{
			name: "Missing Secret",
			req: &csi.ControllerPublishVolumeRequest{
				VolumeId: "test-volume",
				NodeId:   "test-node",
				VolumeCapability: &csi.VolumeCapability{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
				},
				VolumeContext: map[string]string{"key": "value"},
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.ControllerPublishVolumeResponse{}, nil
			},
			wantErr: true,
		},
		{
			name: "Missing Volume Capability",
			req: &csi.ControllerPublishVolumeRequest{
				VolumeId:         "test-volume",
				NodeId:           "test-node",
				Secrets:          map[string]string{"key": "value"},
				VolumeCapability: nil,
				VolumeContext:    map[string]string{"key": "value"},
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.ControllerPublishVolumeResponse{}, nil
			},
			wantErr: true,
		},
		{
			name: "Missing Access Type No Access Mode",
			req: &csi.ControllerPublishVolumeRequest{
				VolumeId: "test-volume",
				NodeId:   "test-node",
				Secrets:  map[string]string{"key": "value"},
				VolumeCapability: &csi.VolumeCapability{
					AccessType: nil,
				},
				VolumeContext: map[string]string{"key": "value"},
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.ControllerPublishVolumeResponse{}, nil
			},
			wantErr: true,
		},
		{
			name: "Missing Access Type",
			req: &csi.ControllerPublishVolumeRequest{
				VolumeId: "test-volume",
				NodeId:   "test-node",
				Secrets:  map[string]string{"key": "value"},
				VolumeCapability: &csi.VolumeCapability{
					AccessType: nil,
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
				VolumeContext: map[string]string{"key": "value"},
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.ControllerPublishVolumeResponse{}, nil
			},
			wantErr: true,
		},
		{
			name: "Missing Access Mode Mount",
			req: &csi.ControllerPublishVolumeRequest{
				VolumeId: "test-volume",
				NodeId:   "test-node",
				Secrets:  map[string]string{"key": "value"},
				VolumeCapability: &csi.VolumeCapability{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: nil,
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
				VolumeContext: map[string]string{"key": "value"},
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.ControllerPublishVolumeResponse{}, nil
			},
			wantErr: true,
		},
		{
			name: "Missing Access Mode Block",
			req: &csi.ControllerPublishVolumeRequest{
				VolumeId: "test-volume",
				NodeId:   "test-node",
				Secrets:  map[string]string{"key": "value"},
				VolumeCapability: &csi.VolumeCapability{
					AccessType: &csi.VolumeCapability_Block{
						Block: nil,
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
				VolumeContext: map[string]string{"key": "value"},
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.ControllerPublishVolumeResponse{}, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := interceptor(context.Background(), tt.req, info, tt.handler)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePublishVolumeRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestControllerUnpublishVolume(t *testing.T) {
	interceptor := NewServerSpecValidator(
		WithRequestValidation(),
		WithResponseValidation(),
		WithRequiresControllerUnpublishVolumeSecrets(),
	)

	info := &grpc.UnaryServerInfo{FullMethod: "/csi.v1.Controller/ControllerUnpublishVolume"}

	tests := []struct {
		name    string
		req     *csi.ControllerUnpublishVolumeRequest
		handler func(ctx context.Context, req interface{}) (interface{}, error)
		wantErr bool
	}{
		{
			name: "Valid Request",
			req: &csi.ControllerUnpublishVolumeRequest{
				VolumeId: "test-volume",
				NodeId:   "test-node",
				Secrets:  map[string]string{"key": "value"},
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.ControllerPublishVolumeResponse{}, nil
			},
			wantErr: false,
		},
		{
			name: "Missing Secret",
			req: &csi.ControllerUnpublishVolumeRequest{
				VolumeId: "test-volume",
				NodeId:   "test-node",
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.ControllerUnpublishVolumeResponse{}, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := interceptor(context.Background(), tt.req, info, tt.handler)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUnpublishVolumeRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestControllerValidateVolumeCapabilities(t *testing.T) {
	interceptor := NewServerSpecValidator(
		WithRequestValidation(),
		WithResponseValidation(),
	)

	info := &grpc.UnaryServerInfo{FullMethod: "/csi.v1.Controller/ValidateVolumeCapabilities"}

	tests := []struct {
		name    string
		req     *csi.ValidateVolumeCapabilitiesRequest
		handler func(ctx context.Context, req interface{}) (interface{}, error)
		wantErr bool
	}{
		{
			name: "Valid Request",
			req: &csi.ValidateVolumeCapabilitiesRequest{
				VolumeId: "test-volume",
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: &csi.VolumeCapability_MountVolume{},
						},
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					},
				},
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.ControllerPublishVolumeResponse{}, nil
			},
			wantErr: false,
		},
		{
			name: "Missing Volume Capabilities",
			req: &csi.ValidateVolumeCapabilitiesRequest{
				VolumeId:           "test-volume",
				VolumeCapabilities: []*csi.VolumeCapability{},
			},
			wantErr: true,
		},
		{
			name: "Missing Access Type",
			req: &csi.ValidateVolumeCapabilitiesRequest{
				VolumeId: "test-volume",
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessType: nil,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Missing Access Mode Block",
			req: &csi.ValidateVolumeCapabilitiesRequest{
				VolumeId: "test-volume",
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessType: &csi.VolumeCapability_Block{
							Block: nil,
						},
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Missing Access Mode Mount",
			req: &csi.ValidateVolumeCapabilitiesRequest{
				VolumeId: "test-volume",
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: nil,
						},
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := interceptor(context.Background(), tt.req, info, tt.handler)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVolumeCapabilitiesRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestControllerGetCapacity(t *testing.T) {
	interceptor := NewServerSpecValidator(
		WithRequestValidation(),
		WithResponseValidation(),
	)

	info := &grpc.UnaryServerInfo{FullMethod: "/csi.v1.Controller/GetCapacity"}

	tests := []struct {
		name    string
		req     *csi.GetCapacityRequest
		handler func(ctx context.Context, req interface{}) (interface{}, error)
		wantErr bool
	}{
		{
			name: "Valid Request",
			req: &csi.GetCapacityRequest{
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: &csi.VolumeCapability_MountVolume{},
						},
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					},
				},
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.GetCapacityResponse{}, nil
			},
			wantErr: false,
		},
		{
			name: "Missing Access Type",
			req: &csi.GetCapacityRequest{
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessType: nil,
					},
				},
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.GetCapacityResponse{}, nil
			},
			wantErr: true,
		},
		{
			name: "Missing Access Mode Block",
			req: &csi.GetCapacityRequest{
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessType: &csi.VolumeCapability_Block{
							Block: nil,
						},
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					},
				},
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.GetCapacityResponse{}, nil
			},
			wantErr: true,
		},
		{
			name: "Missing Access Mode Mount",
			req: &csi.GetCapacityRequest{
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: nil,
						},
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					},
				},
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.GetCapacityResponse{}, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := interceptor(context.Background(), tt.req, info, tt.handler)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGetCapacityRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestControllerValidateListVolumesResponse(t *testing.T) {
	interceptor := newSpecValidator()

	tests := []struct {
		name    string
		resp    *csi.ListVolumesResponse
		wantErr bool
	}{
		{
			name: "Valid Response",
			resp: &csi.ListVolumesResponse{
				Entries: []*csi.ListVolumesResponse_Entry{
					{
						Volume: &csi.Volume{
							VolumeId: "test-volume",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Missing Volume",
			resp: &csi.ListVolumesResponse{
				Entries: []*csi.ListVolumesResponse_Entry{
					{
						Volume: nil,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Missing Volume ID",
			resp: &csi.ListVolumesResponse{
				Entries: []*csi.ListVolumesResponse_Entry{
					{
						Volume: &csi.Volume{},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Missing Volume Context",
			resp: &csi.ListVolumesResponse{
				Entries: []*csi.ListVolumesResponse_Entry{
					{
						Volume: &csi.Volume{
							VolumeId:      "test-volume",
							VolumeContext: map[string]string{},
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := interceptor.validateResponse(context.Background(), "", tt.resp)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateListVolumesResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestControllerValidatControllerGetCapabilitiesResponse(t *testing.T) {
	interceptor := newSpecValidator()

	tests := []struct {
		name    string
		resp    *csi.ControllerGetCapabilitiesResponse
		wantErr bool
	}{
		{
			name: "Valid Response",
			resp: &csi.ControllerGetCapabilitiesResponse{
				Capabilities: []*csi.ControllerServiceCapability{
					{
						Type: &csi.ControllerServiceCapability_Rpc{
							Rpc: &csi.ControllerServiceCapability_RPC{
								Type: csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Missing Capability",
			resp: &csi.ControllerGetCapabilitiesResponse{
				Capabilities: []*csi.ControllerServiceCapability{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := interceptor.validateResponse(context.Background(), "", tt.resp)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateControllerGetCapabilitiesResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeStageVolume(t *testing.T) {
	interceptor := NewServerSpecValidator(
		WithRequestValidation(),
		WithResponseValidation(),
		WithRequiresNodeStageVolumeSecrets(),
		WithRequiresPublishContext(),
	)

	info := &grpc.UnaryServerInfo{FullMethod: "/csi.v1.Node/NodeStageVolume"}

	tests := []struct {
		name    string
		req     *csi.NodeStageVolumeRequest
		handler func(ctx context.Context, req interface{}) (interface{}, error)
		wantErr bool
	}{
		{
			name: "Valid Request",
			req: &csi.NodeStageVolumeRequest{
				VolumeId:          "test-volume",
				StagingTargetPath: "/tmp",
				Secrets:           map[string]string{"key": "value"},
				VolumeCapability: &csi.VolumeCapability{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
				PublishContext: map[string]string{"key": "value"},
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.NodeStageVolumeResponse{}, nil
			},
			wantErr: false,
		},
		{
			name: "Missing Publish Context",
			req: &csi.NodeStageVolumeRequest{
				VolumeId:          "test-volume",
				StagingTargetPath: "/tmp",
				Secrets:           map[string]string{"key": "value"},
				VolumeCapability: &csi.VolumeCapability{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
				PublishContext: map[string]string{},
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.NodeStageVolumeResponse{}, nil
			},
			wantErr: true,
		},
		{
			name: "Missing Staging Target Path",
			req: &csi.NodeStageVolumeRequest{
				VolumeId: "test-volume",
				Secrets:  map[string]string{"key": "value"},
				VolumeCapability: &csi.VolumeCapability{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
				PublishContext: map[string]string{"key": "value"},
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.NodeStageVolumeResponse{}, nil
			},
			wantErr: true,
		},
		{
			name: "Missing Secrets",
			req: &csi.NodeStageVolumeRequest{
				VolumeId:          "test-volume",
				StagingTargetPath: "/tmp",
				VolumeCapability: &csi.VolumeCapability{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
				PublishContext: map[string]string{"key": "value"},
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.NodeStageVolumeResponse{}, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := interceptor(context.Background(), tt.req, info, tt.handler)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNodeStageVolumeRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeUnstageVolume(t *testing.T) {
	interceptor := NewServerSpecValidator(
		WithRequestValidation(),
		WithResponseValidation(),
		WithRequiresNodeStageVolumeSecrets(),
	)

	info := &grpc.UnaryServerInfo{FullMethod: "/csi.v1.Node/NodeUnstageVolume"}

	tests := []struct {
		name    string
		req     *csi.NodeUnstageVolumeRequest
		handler func(ctx context.Context, req interface{}) (interface{}, error)
		wantErr bool
	}{
		{
			name: "Valid Request",
			req: &csi.NodeUnstageVolumeRequest{
				VolumeId:          "test-volume",
				StagingTargetPath: "/tmp",
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.NodeUnstageVolumeResponse{}, nil
			},
			wantErr: false,
		},
		{
			name: "Missing Staging Target Path",
			req: &csi.NodeUnstageVolumeRequest{
				VolumeId: "test-volume",
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.NodeUnstageVolumeResponse{}, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := interceptor(context.Background(), tt.req, info, tt.handler)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNodeUnstageVolumeRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodePublishVolume(t *testing.T) {
	interceptor := NewServerSpecValidator(
		WithRequestValidation(),
		WithResponseValidation(),
		WithRequiresNodePublishVolumeSecrets(),
		WithRequiresStagingTargetPath(),
	)

	info := &grpc.UnaryServerInfo{FullMethod: "/csi.v1.Node/NodePublishVolume"}

	tests := []struct {
		name    string
		req     *csi.NodePublishVolumeRequest
		handler func(ctx context.Context, req interface{}) (interface{}, error)
		wantErr bool
	}{
		{
			name: "Valid Request",
			req: &csi.NodePublishVolumeRequest{
				VolumeId:          "test-volume",
				StagingTargetPath: "/tmp",
				TargetPath:        "/tmp",
				Secrets:           map[string]string{"key": "value"},
				VolumeCapability: &csi.VolumeCapability{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.NodePublishVolumeResponse{}, nil
			},
			wantErr: false,
		},
		{
			name: "Missing Staging Target Path",
			req: &csi.NodePublishVolumeRequest{
				VolumeId:   "test-volume",
				TargetPath: "/tmp",
				Secrets:    map[string]string{"key": "value"},
				VolumeCapability: &csi.VolumeCapability{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.NodePublishVolumeResponse{}, nil
			},
			wantErr: true,
		},
		{
			name: "Missing Staging Target Path",
			req: &csi.NodePublishVolumeRequest{
				VolumeId:          "test-volume",
				StagingTargetPath: "/tmp",
				Secrets:           map[string]string{"key": "value"},
				VolumeCapability: &csi.VolumeCapability{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.NodePublishVolumeResponse{}, nil
			},
			wantErr: true,
		},
		{
			name: "Missing Secrets",
			req: &csi.NodePublishVolumeRequest{
				VolumeId:          "test-volume",
				Secrets:           map[string]string{},
				StagingTargetPath: "/tmp",
				TargetPath:        "/tmp",
				VolumeCapability: &csi.VolumeCapability{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.NodePublishVolumeResponse{}, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := interceptor(context.Background(), tt.req, info, tt.handler)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNodePublishVolumeRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeUnpublishVolume(t *testing.T) {
	interceptor := NewServerSpecValidator(
		WithRequestValidation(),
		WithResponseValidation(),
	)

	info := &grpc.UnaryServerInfo{FullMethod: "/csi.v1.Node/NodeUnpublishVolume"}

	tests := []struct {
		name    string
		req     *csi.NodeUnpublishVolumeRequest
		handler func(ctx context.Context, req interface{}) (interface{}, error)
		wantErr bool
	}{
		{
			name: "Valid Request",
			req: &csi.NodeUnpublishVolumeRequest{
				VolumeId:   "test-volume",
				TargetPath: "/tmp",
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.NodeUnpublishVolumeResponse{}, nil
			},
			wantErr: false,
		},
		{
			name: "Missing Target Path",
			req: &csi.NodeUnpublishVolumeRequest{
				VolumeId: "test-volume",
			},
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return &csi.NodeUnpublishVolumeResponse{}, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := interceptor(context.Background(), tt.req, info, tt.handler)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNodePublishVolumeRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeGetInfoResponse(t *testing.T) {
	interceptor := newSpecValidator()

	tests := []struct {
		name    string
		resp    *csi.NodeGetInfoResponse
		wantErr bool
	}{
		{
			name: "Valid Response",
			resp: &csi.NodeGetInfoResponse{
				NodeId: "test-node",
			},
			wantErr: false,
		},
		{
			name:    "Missing Node ID",
			resp:    &csi.NodeGetInfoResponse{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := interceptor.validateResponse(context.Background(), "", tt.resp)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNodeGetInfoResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeGetCapabilitiesResponse(t *testing.T) {
	interceptor := newSpecValidator()

	tests := []struct {
		name    string
		resp    *csi.NodeGetCapabilitiesResponse
		wantErr bool
	}{
		{
			name: "Valid Response",
			resp: &csi.NodeGetCapabilitiesResponse{
				Capabilities: []*csi.NodeServiceCapability{
					{
						Type: &csi.NodeServiceCapability_Rpc{
							Rpc: &csi.NodeServiceCapability_RPC{
								Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Missing Capability",
			resp: &csi.NodeGetCapabilitiesResponse{
				Capabilities: []*csi.NodeServiceCapability{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := interceptor.validateResponse(context.Background(), "", tt.resp)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNodeGetCapabilitiesResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeGetPluginInfoResponse(t *testing.T) {
	interceptor := newSpecValidator()

	tests := []struct {
		name    string
		resp    *csi.GetPluginInfoResponse
		wantErr bool
	}{
		{
			name: "Valid Response",
			resp: &csi.GetPluginInfoResponse{
				Name:          "test.com",
				VendorVersion: "v1.0.0",
				Manifest:      map[string]string{"key": "value"},
			},
			wantErr: false,
		},
		{
			name: "Missing Name",
			resp: &csi.GetPluginInfoResponse{
				VendorVersion: "v1.0.0",
				Manifest:      map[string]string{"key": "value"},
			},
			wantErr: true,
		},
		{
			name: "Invalid Name",
			resp: &csi.GetPluginInfoResponse{
				Name:          "test",
				VendorVersion: "v1.0.0",
				Manifest:      map[string]string{"key": "value"},
			},
			wantErr: true,
		},
		{
			name: "Name Too Long",
			resp: &csi.GetPluginInfoResponse{
				Name:          "pcswfUbbExPd3s6Om7HnbBaPsSDREIkb4g3TvwoHyVLHS4YjmZu9gcSrmHdQme6C",
				VendorVersion: "v1.0.0",
				Manifest:      map[string]string{"key": "value"},
			},
			wantErr: true,
		},
		{
			name: "Empty Vendor Version",
			resp: &csi.GetPluginInfoResponse{
				Name:     "test.com",
				Manifest: map[string]string{"key": "value"},
			},
			wantErr: true,
		},
		{
			name: "Invalid Vendor Version",
			resp: &csi.GetPluginInfoResponse{
				Name:          "test.com",
				VendorVersion: "test",
				Manifest:      map[string]string{"key": "value"},
			},
			wantErr: true,
		},
		{
			name: "Missing Manifest",
			resp: &csi.GetPluginInfoResponse{
				Name:          "test.com",
				VendorVersion: "test",
				Manifest:      map[string]string{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := interceptor.validateResponse(context.Background(), "", tt.resp)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNodeGetPluginInfoResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetPathLimit(t *testing.T) {
	// Test case: Default value
	assert.Equal(t, setPathLimit(10), 10)

	// Test case: Custom value
	os.Setenv(maxPathLimit, "20")
	assert.Equal(t, setPathLimit(10), 20)

	// Test case: Invalid value
	os.Setenv(maxPathLimit, "invalid")
	assert.Equal(t, setPathLimit(10), 10)

	// Test case: Empty value
	os.Setenv(maxPathLimit, "")
	assert.Equal(t, setPathLimit(10), 10)

	// Test case: Value less than default
	os.Setenv(maxPathLimit, "5")
	assert.Equal(t, setPathLimit(10), 10)
}

func TestValidateFieldSizes(t *testing.T) {
	type TestStruct struct {
		Name        string
		Description string
		NodeID      string
		Map         map[string]string
	}

	largeMap := generateLargeMap()

	tests := []struct {
		name    string
		msg     TestStruct
		wantErr bool
	}{
		{
			name: "Valid Field Sizes",
			msg: TestStruct{
				Name:        "test",
				Description: "description",
				NodeID:      "node",
				Map:         map[string]string{"key": "value"},
			},
			wantErr: false,
		},
		{
			name: "Valid Map Path Key Length",
			msg: TestStruct{
				Name:        "test",
				Description: "description",
				NodeID:      "node",
				Map:         map[string]string{"Path": "value"},
			},
			wantErr: false,
		},
		{
			name: "Exceeds Max Field String Length",
			msg: TestStruct{
				Name:        strings.Repeat("a", maxFieldString+1),
				Description: "description",
				NodeID:      "node",
				Map:         map[string]string{"key": "value"},
			},
			wantErr: true,
		},
		{
			name: "Exceeds Max Field NodeID Length",
			msg: TestStruct{
				Name:        "test",
				Description: "description",
				NodeID:      strings.Repeat("a", maxFieldNodeID+1),
				Map:         map[string]string{"key": "value"},
			},
			wantErr: true,
		},
		{
			name: "Exceeds Max Map Value Length",
			msg: TestStruct{
				Name:        "test",
				Description: "description",
				NodeID:      "node",
				Map:         map[string]string{"key": strings.Repeat("a", maxFieldMap+1)},
			},
			wantErr: true,
		},
		{
			name: "Exceeds Max Map Key Length",
			msg: TestStruct{
				Name:        "test",
				Description: "description",
				NodeID:      "node",
				Map:         map[string]string{strings.Repeat("a", maxFieldMap+1): "value"},
			},
			wantErr: true,
		},
		{
			name: "Exceeds Max Aggregated Map Size",
			msg: TestStruct{
				Name:        "test",
				Description: "description",
				NodeID:      "node",
				Map:         largeMap,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFieldSizes(&tt.msg)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, codes.InvalidArgument, status.Code(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// helper function that generates a map[string]string with keys and values of length maxFieldString
// and ensures that the total length of all keys and values in the map exceeds maxFieldMap
func generateLargeMap() map[string]string {
	result := make(map[string]string)
	kch := 'a'

	totalSize := 0
	for {
		// Generate keys and values of length maxFieldString
		key := strings.Repeat(string(kch), maxFieldString)
		value := strings.Repeat("0", maxFieldString)
		result[key] = value
		totalSize += maxFieldString * 2
		// Interrupt the loop if the total size exceeds maxFieldMap
		if totalSize > maxFieldMap {
			break
		}
		kch++ // unprintable character are fine too in this case
	}

	return result
}

func TestHandle(t *testing.T) {
	// Create a new instance of the interceptor
	// interceptor := newSpecValidator()

	// Create a mock context
	ctx := context.Background()

	// Create a mock method
	method := "/csi.v1.Controller/CreateVolume"

	// Create a mock request
	req := &csi.CreateVolumeRequest{
		Name:    "test-volume",
		Secrets: map[string]string{"foo": "bar"},
		VolumeCapabilities: []*csi.VolumeCapability{
			{
				AccessType: &csi.VolumeCapability_Mount{
					Mount: &csi.VolumeCapability_MountVolume{},
				},
				AccessMode: &csi.VolumeCapability_AccessMode{
					Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
				},
			},
		},
	}

	simpleResp := &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId: "test-volume",
		},
	}
	// Create a mock handler function
	handler := func() (interface{}, error) {
		return simpleResp, nil
	}

	// Test table
	tests := []struct {
		name    string
		opts    []Option
		req     interface{}
		next    func() (interface{}, error)
		want    interface{}
		wantErr bool
	}{
		{
			name: "Test case 1",
			opts: []Option{
				WithRequestValidation(),
				WithResponseValidation(),
			},
			req:     req,
			next:    handler,
			want:    simpleResp,
			wantErr: false,
		},
		{
			name: "Nil request, handler resp",
			opts: []Option{
				WithRequestValidation(),
				WithResponseValidation(),
			},
			req:     nil,
			next:    handler,
			want:    simpleResp,
			wantErr: false,
		},
		{
			name: "Test case 3",
			opts: []Option{
				WithRequestValidation(),
				WithResponseValidation(),
			},
			req: &csi.CreateVolumeRequest{
				Name:    "test-volume",
				Secrets: map[string]string{"foo": "bar"},
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: &csi.VolumeCapability_MountVolume{},
						},
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					},
				},
			},
			next:    handler,
			want:    simpleResp,
			wantErr: false,
		},
		{
			name: "Test case 4",
			opts: []Option{
				WithRequestValidation(),
				WithResponseValidation(),
			},
			req: &csi.CreateVolumeRequest{
				Name:    "test-volume",
				Secrets: map[string]string{"foo": "bar"},
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: &csi.VolumeCapability_MountVolume{},
						},
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					},
				},
			},
			next: func() (interface{}, error) {
				return nil, status.Error(codes.Internal, "Internal server error")
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Request with long fields",
			opts: []Option{
				WithRequestValidation(),
				WithResponseValidation(),
			},
			req: &csi.CreateVolumeRequest{
				Name: strings.Repeat("n", maxFieldString+1),
			},
			next:    handler,
			want:    nil,
			wantErr: true,
		},
		{
			name: "Request with empty GetVolumeId()",
			opts: []Option{
				WithRequestValidation(),
				WithResponseValidation(),
			},
			req: &csi.DeleteVolumeRequest{
				VolumeId: "",
			},
			next:    handler,
			want:    nil,
			wantErr: true,
		},
		{
			name: "Request with empty GetVolumeContext()",
			opts: []Option{
				WithRequestValidation(),
				WithRequiresVolumeContext(),
			},
			req: &csi.ControllerPublishVolumeRequest{
				VolumeId:      "volume-id",
				VolumeContext: nil,
			},
			next:    handler,
			want:    nil,
			wantErr: true,
		},
		{
			name: "Response with long fields",
			opts: []Option{
				WithResponseValidation(),
			},
			req: &csi.NodeGetInfoRequest{},
			next: func() (interface{}, error) {
				return &csi.NodeGetInfoResponse{
					NodeId: strings.Repeat("d", maxFieldNodeID+1),
				}, nil
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Response with empty GetVolumeId()",
			opts: []Option{
				WithRequestValidation(),
				WithRequiresVolumeContext(),
			},
			req: &csi.ControllerPublishVolumeRequest{
				VolumeId:      "volume-id",
				VolumeContext: nil,
			},
			next:    handler,
			want:    nil,
			wantErr: true,
		},
	}

	// Run the tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new interceptor with the options from the test
			i := newSpecValidator(tt.opts...)

			// Call the handle function
			resp, err := i.handle(ctx, method, tt.req, tt.next)
			// Assert the response and error
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Expected no error, but got: %v", err)
				}
				return
			}

			if tt.wantErr {
				t.Errorf("Expected an error, but got nil")
				return
			}

			// Assert the response
			if !reflect.DeepEqual(resp, tt.want) {
				t.Errorf("Expected response: %v, got: %v", tt.want, resp)
			}
		})
	}
}
