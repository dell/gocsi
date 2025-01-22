package requestid

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func Test_newRequestIDInjector(t *testing.T) {
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

func Test_interceptor_handleServer(t *testing.T) {
	type fields struct {
		id uint64
	}
	type args struct {
		ctx     context.Context
		req     interface{}
		in2     *grpc.UnaryServerInfo
		handler grpc.UnaryHandler
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "Test case 1",
			fields: fields{
				id: 123,
			},
			args: args{
				ctx: context.Background(),
				req: "test request",
				in2: &grpc.UnaryServerInfo{},
				handler: func(ctx context.Context, req interface{}) (interface{}, error) {
					return "test response", nil
				},
			},
			want:    "test response",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &interceptor{
				id: tt.fields.id,
			}
			got, err := s.handleServer(tt.args.ctx, tt.args.req, tt.args.in2, tt.args.handler)
			if (err != nil) != tt.wantErr {
				t.Errorf("interceptor.handleServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("interceptor.handleServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_interceptor_handleClient(t *testing.T) {
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
				invoker: func(ctx context.Context, method string, req, rep interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
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
