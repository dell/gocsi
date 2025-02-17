package context

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestGetRequestID(t *testing.T) {
	tests := []struct {
		name          string
		ctx           context.Context
		wantID        uint64
		wantAvailable bool
	}{
		{
			name:          "Negative test: no ID in context",
			ctx:           context.Background(),
			wantID:        0,
			wantAvailable: false,
		},
		{
			name: "Get request ID from incoming context",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.MD{
				RequestIDKey: []string{"41"},
			}),
			wantID:        41,
			wantAvailable: true,
		},
		{
			name: "Get request ID from outgoing context",
			ctx: metadata.NewOutgoingContext(context.Background(), metadata.MD{
				RequestIDKey: []string{"102"},
			}),
			wantID:        102,
			wantAvailable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualID, actualAvailable := GetRequestID(tt.ctx)
			assert.Equal(t, tt.wantID, actualID)
			assert.Equal(t, tt.wantAvailable, actualAvailable)
		})
	}
}

func TestWithEnviron(t *testing.T) {
	want := []string{"key=value"}

	ctx := context.Background()
	ctx = WithEnviron(ctx, want)

	got := ctx.Value(ctxOSEnviron)
	assert.Equal(t, want, got.([]string))
}

func TestWithLookupEnv(t *testing.T) {
	f := lookupEnvFunc(func(_ string) (string, bool) {
		return "test", true
	})

	ctx := context.Background()
	ctx = WithLookupEnv(ctx, f)

	got := ctx.Value(ctxOSLookupEnvKey).(lookupEnvFunc)
	gotString, gotBool := got("")

	assert.Equal(t, "test", gotString)
	assert.Equal(t, true, gotBool)
}

func TestWithSetenv(t *testing.T) {
	f := setenvFunc(func(string, string) error {
		return nil
	})

	ctx := context.Background()
	ctx = WithSetenv(ctx, f)

	got := ctx.Value(ctxOSSetenvKey).(setenvFunc)
	gotErr := got("", "")
	assert.Nil(t, gotErr)
}

func TestGetenv(t *testing.T) {
	ctx := context.Background()
	ctx = WithEnviron(ctx, []string{"key=value"})

	v := Getenv(ctx, "key")
	assert.Equal(t, "value", v)

	ctx = context.Background()
	ctx = WithLookupEnv(ctx, func(_ string) (string, bool) {
		return "value", true
	})

	v = Getenv(ctx, "key")
	assert.Equal(t, "value", v)

	ctx = context.Background()
	os.Setenv("key", "value")
	defer os.Unsetenv("key")

	v = Getenv(ctx, "key")
	assert.Equal(t, "value", v)
}

func TestSetenv(t *testing.T) {
	f := setenvFunc(func(string, string) error {
		return nil
	})

	ctx := context.Background()
	ctx = WithSetenv(ctx, f)
	err := Setenv(ctx, "key", "value")
	assert.Nil(t, err)

	ctx = context.Background()
	err = Setenv(ctx, "key", "value")
	defer os.Unsetenv("key")
	assert.Nil(t, err)

	assert.Equal(t, "value", os.Getenv("key"))
}
