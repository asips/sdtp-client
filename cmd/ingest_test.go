package cmd

import (
	"context"
	"sync"
	"testing"

	"github.com/asips/sdtp-client/internal"
	"github.com/stretchr/testify/assert"
)

func Test_doIngest(t *testing.T) {
	listing := []internal.FileInfo{
		{ID: 0, Name: "file1.txt", Size: 1234, Tags: map[string]string{"stream": "test"}},
	}
	sdtp := createMockSDTP(t)
	sdtp.listing = listing

	downloadWorker = func(ctx context.Context, wg *sync.WaitGroup, sdtp internal.SDTPClient, files chan internal.FileInfo, noAck bool, destDir string) {
		for f := range files {
			t.Logf("Mock download worker processing file: %v", f)
		}
		wg.Done()
	}
	defer func() { downloadWorker = defaultDownloadWorker }()

	err := doIngest(t.Context(), sdtp, "dest/dir", map[string]string{"stream": "test"}, true, 10)

	assert.NoError(t, err)

}
