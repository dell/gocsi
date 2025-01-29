package cmd

import (
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

func TestDocTypeArg_String(t *testing.T) {
	// Create an instance of the docTypeArg struct
	s := &docTypeArg{}

	// Call the String() method
	result := s.String()

	// Assert the expected value
	expected := "md"
	if result != expected {
		t.Errorf("String() = %v, want %v", result, expected)
	}
}

func TestDocTypeArg_Type(t *testing.T) {
	// Create an instance of the docTypeArg struct
	s := &docTypeArg{}

	// Call the Type() method
	result := s.Type()

	// Assert the expected value
	expected := "md|man|rst"
	if result != expected {
		t.Errorf("Type() = %v, want %v", result, expected)
	}
}

func TestDocTypeArg_Set(t *testing.T) {
	// Create an instance of the docTypeArg struct
	s := &docTypeArg{}

	// Test case: Valid input
	err := s.Set("md")
	if err != nil {
		t.Errorf("Set() returned an error: %v", err)
	}
	if s.val != "md" {
		t.Errorf("Set() did not set the correct value: got %s, want %s", s.val, "md")
	}
	err = s.Set("man")
	if err != nil {
		t.Errorf("Set() returned an error: %v", err)
	}
	if s.val != "man" {
		t.Errorf("Set() did not set the correct value: got %s, want %s", s.val, "man")
	}
	err = s.Set("rst")
	if err != nil {
		t.Errorf("Set() returned an error: %v", err)
	}
	if s.val != "rst" {
		t.Errorf("Set() did not set the correct value: got %s, want %s", s.val, "rst")
	}

	// Test case: Invalid input
	err = s.Set("invalid")
	if err == nil {
		t.Errorf("Set() did not return an error for invalid input")
	}

	// Test case: Empty input
	err = s.Set("")
	if err == nil {
		t.Errorf("Set() did not return an error for empty input")
	}
}

func TestLogLevelArg_Type(t *testing.T) {
	// Create an instance of the logLevelArg struct
	a := &logLevelArg{}

	// Call the Type() method
	result := a.Type()

	// Assert the expected value
	expected := "PANIC|FATAL|ERROR|WARN|INFO|DEBUG"
	if result != expected {
		t.Errorf("Type() = %v, want %v", result, expected)
	}
}

func TestVolumeCapabilitySliceArg_Set(t *testing.T) {
	// Create an instance of the volumeCapabilitySliceArg struct
	s := &volumeCapabilitySliceArg{}

	// Test case: Valid input
	// err := s.Set("mode,type,fstype,mntflags")
	err := s.Set("1,1,1,1")
	if err != nil {
		t.Errorf("Set() returned an error: %v", err)
	}
	if len(s.data) != 1 {
		t.Errorf("Set() did not set the correct number of volume capabilities: got %d, want %d", len(s.data), 1)
	}
	capacity := s.data[0]
	if capacity.AccessMode.Mode != csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER {
		t.Errorf("Set() did not set the correct access mode: got %v, want %v", capacity.AccessMode.Mode, csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER)
	}

	// test case: mount type capability
	err = s.Set("1,2,2,2")
	if err != nil {
		t.Errorf("Set() returned an error: %v", err)
	}
	if len(s.data) != 2 {
		t.Errorf("Set() did not set the correct number of volume capabilities: got %d, want %d", len(s.data), 1)
	}
	capacity = s.data[1]
	if capacity.AccessMode.Mode != csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER {
		t.Errorf("Set() did not set the correct access mode: got %v, want %v", capacity.AccessMode.Mode, csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER)
	}

	// Test case: Invalid input
	err = s.Set("invalid")
	if err == nil {
		t.Errorf("Set() did not return an error for invalid input")
	}

	// Test case: Empty input
	err = s.Set("")
	if err == nil {
		t.Errorf("Set() did not return an error for empty input")
	}
}

func TestMapOfStringArg_Set(t *testing.T) {
	// Create an instance of the mapOfStringArg struct
	s := &mapOfStringArg{}

	// Test case: Valid input
	err := s.Set("key1=value1,key2=value2")
	if err != nil {
		t.Errorf("Set() returned an error: %v", err)
	}
	if len(s.data) != 2 {
		t.Errorf("Set() did not set the correct number of key-value pairs: got %d, want %d", len(s.data), 2)
	}
	if s.data["key1"] != "value1" {
		t.Errorf("Set() did not set the correct value for key1: got %s, want %s", s.data["key1"], "value1")
	}
	if s.data["key2"] != "value2" {
		t.Errorf("Set() did not set the correct value for key2: got %s, want %s", s.data["key2"], "value2")
	}
}

func TestMapOfStringArg_Type(t *testing.T) {
	// Create an instance of the mapOfStringArg struct
	s := &mapOfStringArg{}

	// Call the Type() method
	result := s.Type()

	// Assert the expected value
	expected := "key=val[,key=val,...]"
	if result != expected {
		t.Errorf("Type() = %v, want %v", result, expected)
	}
}

func TestVolumeCapabilitySliceArg_Type(t *testing.T) {
	// Create an instance of the volumeCapabilitySliceArg struct
	s := &volumeCapabilitySliceArg{}

	// Call the Type() method
	result := s.Type()

	// Assert the expected value
	expected := "mode,type[,fstype,mntflags]"
	if result != expected {
		t.Errorf("Type() = %v, want %v", result, expected)
	}
}
