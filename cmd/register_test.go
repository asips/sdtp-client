package cmd

import (
	"testing"

	"github.com/asips/sdtp-client/internal"
	"github.com/stretchr/testify/assert"
)

func Test_doRegister(t *testing.T) {
	sdtp := createMockSDTP(t)

	ok := doRegister(t.Context(), sdtp)
	assert.True(t, ok)

	sdtp.err = internal.ErrExists
	ok = doRegister(t.Context(), sdtp)
	assert.False(t, ok)
}
