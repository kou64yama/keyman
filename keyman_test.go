package keyman_test

import (
	"keyman"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	v := keyman.Version()
	assert.Equal(t, "0.0.0", v)
}
