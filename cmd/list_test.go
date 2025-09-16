package cmd

import (
	"testing"

	"github.com/asips/sdtp-client/internal"
	"github.com/stretchr/testify/assert"
)

func Test_doList(t *testing.T) {
	listing := []internal.FileInfo{
		{ID: 0, Name: "file1.txt", Size: 1234, Tags: map[string]string{"stream": "test"}},
		{ID: 1, Name: "file2.txt", Size: 1234, Tags: map[string]string{"stream": "test"}},
		{ID: 2, Name: "file3.txt", Size: 1234, Tags: map[string]string{"stream": "test"}},
		{ID: 3, Name: "file4.txt", Size: 1234, Tags: map[string]string{"stream": "test"}},
	}
	sdtp := createMockSDTP(t)
	sdtp.listing = listing

	count, err := doList(t.Context(), sdtp, map[string]string{"stream": "test"})

	assert.NoError(t, err)
	assert.Equal(t, 4, count)
}
