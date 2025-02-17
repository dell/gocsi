package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	// main() does not return any errors or values, so verify panic wasn't hit
	assert.NotPanics(t, func() { main() })
}
