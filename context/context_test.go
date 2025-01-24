package context

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRequestID(t *testing.T) {
	// Test case: GetRequestID should return that there is no available ID
	ctx := context.Background()
	requestID := uint64(0) // 0 is an invalid request ID

	actualID, available := GetRequestID(ctx)
	if available {
		t.Errorf("Expected request ID to be unavailable, got %v", available)
	}
	if actualID != requestID {
		t.Errorf("Expected request ID to be %d, got %d", requestID, actualID)
	}
	/*
	   // Test case: GetRequestID should return that there is an available ID
	   requestID = uint64(123)
	   ctxWithID := context.WithValue(ctx, ctxRequestIDKey, requestID)

	   actualID, available = GetRequestID(ctxWithID)

	   	if !available {
	   		t.Errorf("Expected request ID to be available, got %v", available)
	   	}

	   	if actualID != requestID {
	   		t.Errorf("Expected request ID to be %d, got %d", requestID, actualID)
	   	}
	*/
}

func TestWithEnviron(t *testing.T) {
	want := []string{"key=value"}

	ctx := context.Background()
	ctx = WithEnviron(ctx, want)

	got := ctx.Value(ctxOSEnviron)
	assert.Equal(t, want, got.([]string))
}

func TestWithLookupEnv(t *testing.T) {
	f := lookupEnvFunc(func(key string) (string, bool) {
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
	ctx = WithLookupEnv(ctx, func(s string) (string, bool) {
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
