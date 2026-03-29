package openings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	c := New()

	assert.NotNil(t, c, "New() should not return nil")
}
