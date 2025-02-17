package requestid

import (
	"errors"
	"reflect"
	"testing"

	csictx "github.com/dell/gocsi/context"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestNewRequestIDInjector(t *testing.T) {
	tests := []struct {
		name string
		want *interceptor
	}{
		{
			name: "Test case",
			want: &interceptor{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newRequestIDInjector(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newRequestIDInjector() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInterceptorHandleServer(t *testing.T) {
	tests := []struct {
		name    string
		id      uint64
		getCtx  func() context.Context
		req     interface{}
		handler grpc.UnaryHandler
		want    interface{}
		wantErr bool
	}{
		{
			name:   "Basic positive",
			id:     123,
			getCtx: func() context.Context { return context.Background() },
			req:    "test request",
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return "test response", nil
			},
			want:    "test response",
			wantErr: false,
		},
		{
			name: "With good request ID",
			id:   123,
			getCtx: func() context.Context {
				md := metadata.Pairs(
					csictx.RequestIDKey, "2452",
				)
				return metadata.NewIncomingContext(context.Background(), md)
			},
			req: "test request",
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return "response2452", nil
			},
			want:    "response2452",
			wantErr: false,
		},
		{
			name: "Basic negative",
			id:   1023,
			getCtx: func() context.Context {
				md := metadata.Pairs(
					csictx.RequestIDKey, "non-uint-id",
				)
				return metadata.NewIncomingContext(context.Background(), md)
			},
			req: "test request",
			handler: func(_ context.Context, _ interface{}) (interface{}, error) {
				return "", errors.New("internal error")
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &interceptor{
				id: tt.id,
			}
			got, err := s.handleServer(tt.getCtx(), tt.req, nil, tt.handler)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInterceptorHandleClient(t *testing.T) {
	type fields struct {
		id uint64
	}
	type args struct {
		ctx     context.Context
		method  string
		req     interface{}
		rep     interface{}
		cc      *grpc.ClientConn
		invoker grpc.UnaryInvoker
		opts    []grpc.CallOption
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test case 1",
			fields: fields{
				id: 123,
			},
			args: args{
				ctx:    context.Background(),
				method: "exampleMethod",
				req:    "exampleRequest",
				rep:    "exampleResponse",
				cc:     &grpc.ClientConn{},
				invoker: func(_ context.Context, _ string, _, _ interface{}, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
					return nil
				},
				opts: []grpc.CallOption{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &interceptor{
				id: tt.fields.id,
			}
			if err := s.handleClient(tt.args.ctx, tt.args.method, tt.args.req, tt.args.rep, tt.args.cc, tt.args.invoker, tt.args.opts...); (err != nil) != tt.wantErr {
				t.Errorf("interceptor.handleClient() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
