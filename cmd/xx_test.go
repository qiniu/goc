package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReturnHello(t *testing.T) {
	assert.Equal(t, ReturnHello(), "Hello")
}
