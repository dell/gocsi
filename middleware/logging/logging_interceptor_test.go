package logging

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	csictx "github.com/dell/gocsi/context"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestWithRequestLogging(t *testing.T) {
	w := &bytes.Buffer{}
	tests := []struct {
		name  string
		input io.Writer
		want  opts
	}{
		{
			name: "non nil input",
			want: opts{
				reqw:             w,
				repw:             nil,
				disableLogVolCtx: false,
			},
			input: w,
		},
		{
			name: "nil input",
			want: opts{
				reqw:             os.Stdout,
				repw:             nil,
				disableLogVolCtx: false,
			},
			input: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := WithRequestLogging(tt.input)

			// Use interceptor type to check the expected opts
			i := newLoggingInterceptor(opt)

			// Check interceptor type opts to verify WithRequestLogging set them as expected
			if i.opts != tt.want {
				t.Errorf("WithRequestLogging() returned function with parameters = %v, want function with parameters %v", i.opts, tt.want)
			}
		})
	}
}

func TestWithResponseLogging(t *testing.T) {
	w := &bytes.Buffer{}
	tests := []struct {
		name  string
		input io.Writer
		want  opts
	}{
		{
			name: "non nil input",
			want: opts{
				reqw:             nil,
				repw:             w,
				disableLogVolCtx: false,
			},
			input: w,
		},
		{
			name: "nil input",
			want: opts{
				reqw:             nil,
				repw:             os.Stdout,
				disableLogVolCtx: false,
			},
			input: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := WithResponseLogging(tt.input)

			// Use interceptor type to check the expected opts
			i := newLoggingInterceptor(opt)

			// Check interceptor type opts to verify WithResponseLogging set them as expected
			if i.opts != tt.want {
				t.Errorf("WithResponseLogging() returned function with parameters = %v, want function with parameters %v", i.opts, tt.want)
			}
		})
	}
}

func TestWithDisableLogVolumeContext(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{
			name: "switch to true",
			want: true,
		},
		{
			name: "test default",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use interceptor type to check if WithDisableLogVolumeContext is able to switch value from false to true
			i := newLoggingInterceptor()

			// if want = true, use WithDisableLogVolumeContext to switch disableLogVolCtx from false to true
			if tt.want {
				i = newLoggingInterceptor(WithDisableLogVolumeContext())
			}
			// Check interceptor type opts to verify WithDisableLogVolumeContext switched or didn't switch
			if i.opts.disableLogVolCtx != tt.want {
				t.Errorf("WithResponseLogging() returned function with parameters = %v, want function with parameters %v", i.opts, tt.want)
			}
		})
	}
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
			i:    newLoggingInterceptor(WithRequestLogging(w), WithResponseLogging(w)),
			req:  defaultReq,
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
