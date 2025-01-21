package context

import (
	"context"
	"testing"
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
