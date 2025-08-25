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

package middleware_test

import (
	"context"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/dell/gocsi/utils/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestChainUnaryClient(t *testing.T) {
	// Test case: Empty interceptors
	var interceptors []grpc.UnaryClientInterceptor
	chain0 := middleware.ChainUnaryClient(interceptors...)
	assert.NotNil(t, chain0)

	// Test case: Invoking the non-interceptor chain
	ctx := context.Background()
	method := "TestMethod"
	req := "TestRequest"
	rep := "TestReply"
	cc := &grpc.ClientConn{}
	var opts []grpc.CallOption
	invoker := func(
		ctx context.Context,
		method string,
		req, rep interface{},
		cc *grpc.ClientConn,
		opts ...grpc.CallOption,
	) error {
		// Do something
		log.Info("ctx:", ctx)
		log.Info("req:", req)
		log.Info("rep:", rep)
		log.Info("cc:", cc)
		log.Info("opts:", opts)
		log.Info("method:", method)
		return nil
	}
	err := chain0(ctx, method, req, rep, cc, invoker, opts...)
	assert.NoError(t, err)

	// Test case: Single interceptor
	interceptors = []grpc.UnaryClientInterceptor{
		func(
			ctx context.Context,
			method string,
			req, rep interface{},
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			// Do something
			return invoker(ctx, method, req, rep, cc, opts...)
		},
	}
	chain := middleware.ChainUnaryClient(interceptors...)
	assert.NotNil(t, chain)

	// Test case: Multiple interceptors
	interceptors = []grpc.UnaryClientInterceptor{
		func(
			ctx context.Context,
			method string,
			req, rep interface{},
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			// Do something
			return invoker(ctx, method, req, rep, cc, opts...)
		},
		func(
			ctx context.Context,
			method string,
			req, rep interface{},
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			// Do something else
			return invoker(ctx, method, req, rep, cc, opts...)
		},
	}
	chainN := middleware.ChainUnaryClient(interceptors...)
	assert.NotNil(t, chainN)

	// Test case: Invoking the multi-interceptor chain
	ctx = context.Background()
	method = "TestMethod"
	req = "TestRequest"
	rep = "TestReply"
	cc = &grpc.ClientConn{}
	opts = []grpc.CallOption{}
	invoker = func(
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
	err = chainN(ctx, method, req, rep, cc, invoker, opts...)
	assert.NoError(t, err)
}

func TestChainUnaryServer(t *testing.T) {
	// Test case: Empty interceptors
	var interceptors []grpc.UnaryServerInterceptor
	chain0 := middleware.ChainUnaryServer(interceptors...)
	assert.NotNil(t, chain0)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		log.Info("ctx:", ctx)
		log.Info("req:", req)
		return "response", nil
	}
	resp, err := chain0(context.Background(), "request", nil, handler)
	assert.NoError(t, err)
	assert.Equal(t, "response", resp)

	// Test case: Single interceptor
	interceptors = []grpc.UnaryServerInterceptor{
		func(
			ctx context.Context,
			req interface{},
			info *grpc.UnaryServerInfo,
			handler grpc.UnaryHandler,
		) (interface{}, error) {
			// Do something
			log.Info("ctx:", ctx)
			log.Info("req:", req)
			log.Info("info:", info)
			return handler(ctx, req)
		},
	}
	chain := middleware.ChainUnaryServer(interceptors...)
	assert.NotNil(t, chain)

	// Test case: Multiple interceptors
	interceptors = []grpc.UnaryServerInterceptor{
		func(
			ctx context.Context,
			req interface{},
			info *grpc.UnaryServerInfo,
			handler grpc.UnaryHandler,
		) (interface{}, error) {
			// Do something
			log.Info("ctx:", ctx)
			log.Info("req:", req)
			log.Info("info:", info)
			return handler(ctx, req)
		},
		func(
			ctx context.Context,
			req interface{},
			info *grpc.UnaryServerInfo,
			handler grpc.UnaryHandler,
		) (interface{}, error) {
			// Do something else
			log.Info("ctx:", ctx)
			log.Info("req:", req)
			log.Info("info:", info)
			return handler(ctx, req)
		},
	}
	chainN := middleware.ChainUnaryServer(interceptors...)
	assert.NotNil(t, chainN)

	// Test case: Invoking the chain
	ctx := context.Background()
	req := "TestRequest"
	info := &grpc.UnaryServerInfo{}
	handler = func(
		ctx context.Context,
		req interface{},
	) (interface{}, error) {
		// Do something
		log.Info("ctx:", ctx)
		log.Info("req:", req)
		return "TestResponse", nil
	}

	rep, err := chainN(ctx, req, info, handler)
	assert.NoError(t, err)
	assert.Equal(t, "TestResponse", rep)
}

func TestIsNilResponse(t *testing.T) {
	tests := []struct {
		name string
		rep  interface{}
		want bool
	}{
		{
			name: "Nil Response",
			rep:  nil,
			want: true,
		},
		{
			name: "Non-Nil Response",
			rep:  &csi.CreateVolumeResponse{},
			want: false,
		},
		{
			name: "Nil Response Inside Interface",
			rep: func() interface{} {
				var rep *csi.CreateVolumeResponse
				return rep
			}(),
			want: true,
		},
		{
			name: "Non-Nil Response Inside Interface",
			rep: func() interface{} {
				rep := &csi.CreateVolumeResponse{}
				return rep
			}(),
			want: false,
		},
		{
			name: "Non-Response Type Inside Interface",
			rep: func() interface{} {
				var rep int
				return rep
			}(),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := middleware.IsNilResponse(tt.rep); got != tt.want {
				t.Errorf("IsNilResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}
