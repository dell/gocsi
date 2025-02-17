package logging

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	csictx "github.com/dell/gocsi/context"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestClientLogger(t *testing.T) {
	cLogger := NewClientLogger()

	invoker := func(
		ctx context.Context,
		method string,
		req, rep interface{},
		cc *grpc.ClientConn,
		opts ...grpc.CallOption,
	) error {
		log.Info("ctx:", ctx)
		log.Info("req:", req)
		log.Info("rep:", rep)
		log.Info("cc:", cc)
		log.Info("opts:", opts)
		log.Info("method:", method)
		return nil
	}

	err := cLogger(
		context.Background(),
		"TestMethod",
		&csi.CreateVolumeRequest{},
		&csi.CreateVolumeResponse{},
		&grpc.ClientConn{},
		invoker,
		grpc.EmptyCallOption{},
	)
	assert.NoError(t, err)
}

func TestServerLogger(t *testing.T) {
	sLogger := NewServerLogger()

	res, err := sLogger(
		context.Background(),
		&csi.CreateVolumeRequest{},
		&grpc.UnaryServerInfo{},
		// grpc.UnaryHandler handler
		func(ctx context.Context, req interface{}) (interface{}, error) {
			log.Info("ctx:", ctx)
			log.Info("req:", req)
			return &csi.CreateVolumeResponse{}, nil
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, &csi.CreateVolumeResponse{}, res)
}

func TestHandle(t *testing.T) {
	w := &bytes.Buffer{}

	defaultCtx := context.Background()

	// Create a mock method
	method := "example.ExampleMethod"

	// Mock error to be returned by next function
	defaultErr := errors.New("example error")
	defaultReq := &csi.CreateVolumeResponse{}
	defaultRes := &csi.CreateVolumeResponse{}

	defaultNext := func() (interface{}, error) {
		return defaultRes, nil
	}

	tests := []struct {
		name    string
		i       *interceptor
		req     interface{}
		next    func() (interface{}, error)
		getCtx  func() context.Context
		wantRes interface{}
		wantErr bool
	}{
		{
			name: "nil request",
			i:    &interceptor{},
			req:  nil,
			next: func() (interface{}, error) {
				return nil, defaultErr
			},
			wantRes: nil,
			wantErr: true,
		},
		{
			name:    "request and response disabled",
			i:       &interceptor{},
			req:     defaultReq,
			next:    defaultNext,
			wantRes: defaultRes,
			wantErr: false,
		},
		{
			name: "with request logging",
			i:    newLoggingInterceptor(WithRequestLogging(w)),
			req:  defaultReq,
			next: defaultNext,
			getCtx: func() context.Context {
				md := metadata.Pairs(
					csictx.RequestIDKey, "123",
				)
				return metadata.NewIncomingContext(context.Background(), md)
			},
			wantRes: defaultRes,
			wantErr: false,
		},
		{
			name:    "with response logging",
			i:       newLoggingInterceptor(WithResponseLogging(w)),
			req:     defaultReq,
			next:    defaultNext,
			wantRes: defaultRes,
			wantErr: false,
		},
		{
			name: "log failed response",
			i:    newLoggingInterceptor(WithResponseLogging(w)),
			req:  defaultReq,
			next: func() (interface{}, error) {
				return nil, defaultErr
			},
			wantRes: nil,
			wantErr: true,
		},
		{
			name: "with request and response logging",
			i: newLoggingInterceptor(WithRequestLogging(nil),
				WithResponseLogging(nil), WithDisableLogVolumeContext()),
			// special object type to trigger all checks
			req: &struct {
				// field called "Secrets"
				Secrets string
				// field called "VolumeContext"
				VolumeContext struct{}
				// unexported field from another package
				md metadata.MD
			}{},
			next: defaultNext,
			getCtx: func() context.Context {
				md := metadata.Pairs(
					csictx.RequestIDKey, "234",
				)
				return metadata.NewIncomingContext(context.Background(), md)
			},
			wantRes: defaultRes,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the handle function
			ctx := defaultCtx
			if tt.getCtx != nil {
				ctx = tt.getCtx()
			}
			resp, err := tt.i.handle(ctx, method, tt.req, tt.next)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantRes, resp)
		})
	}
}
