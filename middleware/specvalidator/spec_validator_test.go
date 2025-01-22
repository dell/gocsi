package specvalidator

import (
	"context"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"
)

func TestControllerValidateCreateVolumeRequest(t *testing.T) {
	interceptor := NewServerSpecValidator(
		WithRequestValidation(),
		WithResponseValidation(),
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
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return &csi.CreateVolumeResponse{
					Volume: &csi.Volume{
						VolumeId: "test-volume",
					},
				}, nil
			},
			wantErr: false,
		},
		{
			name:    "Missing Name",
			req:     &csi.CreateVolumeRequest{},
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
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
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
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
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

func TestControllerValidateDeleteVolumeRequest(t *testing.T) {
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
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
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
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
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

func TestControllerValidatePublishVolumeRequest(t *testing.T) {
	interceptor := NewServerSpecValidator(
		WithRequestValidation(),
		WithResponseValidation(),
		WithRequiresControllerPublishVolumeSecrets(),
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
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return &csi.ControllerPublishVolumeResponse{}, nil
			},
			wantErr: false,
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
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
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

func TestControllerValidateUnpublishVolumeRequest(t *testing.T) {
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
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
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
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
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

func TestControllerValidateVolumeCapabilitiesRequest(t *testing.T) {
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
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
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
