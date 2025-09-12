package cmd

import (
	"context"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/asips/sdtp-client/internal"
	"github.com/asips/sdtp-client/internal/log"
	"github.com/spf13/cobra"
)

var (
	destDir     string
	tags        map[string]string
	stream      string
	shortName   string
	mission     string
	noAckFlag   bool
	listFlag    bool
	concurrency uint
)

var ingestCmd = &cobra.Command{
	Use:   "ingest",
	Short: "Ingest data from SDTP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		apiUrl := parseApiUrl(strApiUrl)
		if _, err := os.Stat(destDir); os.IsNotExist(err) {
			log.Printf("creating destination directory: %s", destDir)
			os.MkdirAll(destDir, 0755)
		}
		if flags.Changed("stream") {
			tags["stream"] = stream
		}
		if flags.Changed("mission") {
			tags["mission"] = mission
		}
		if flags.Changed("short-name") {
			tags["ShortName"] = shortName
		}
		if checkCertExprFlag {
			checkCert(certPath, keyPath, checkCertDays)
		}

		if listFlag {
			doList(apiUrl, certPath, keyPath, tags, httpTimeout)
		}

		return doIngest(
			apiUrl,
			certPath,
			keyPath,
			destDir,
			tags,
			noAckFlag,
			httpTimeout,
		)
	},
}

func init() {
	flags := ingestCmd.Flags()

	flags.StringVarP(&destDir, "dest-dir", "d", "", "Local directory to ingest data to")
	flags.StringVar(&stream, "stream", "", "SDTP 'stream' field (query parameter)")
	flags.StringVar(&shortName, "short-name", "", "SDTP 'ShortName' field (query parameter)")
	flags.StringVar(&mission, "mission", "", "SDTP 'mission' field (query parameter)")
	flags.StringToStringVarP(&tags, "tag", "t", map[string]string{}, "<key>=<value> tags to filter by. May be specified multiple times or as a comma-separated list")
	flags.BoolVar(&noAckFlag, "no-ack", false, "Acknowledge files after successful ingest")
	flags.BoolVar(&listFlag, "list", false, "List available files, but do not download")
	flags.DurationVar(&httpTimeout, "http-timeout", time.Minute*5, "HTTP client timeout in seconds for list operations")
	flags.UintVar(&concurrency, "concurrency", 4, "Number of concurrent downloads")

	flags.MarkDeprecated("list", "use 'list' sub-command instead")
}

func doIngest(apiUrl *url.URL, certPath, keyPath, destDir string, tags map[string]string, noAck bool, timeout time.Duration) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	sdtpFactory := func() *internal.SDTP {
		sdtp, err := internal.NewSDTP(apiUrl, certPath, keyPath, timeout)
		if err != nil {
			log.Fatal("Failed to create SDTP client: %s", err)
		}
		return sdtp
	}
	sdtp := sdtpFactory()

	files, err := sdtp.List(ctx, tags)
	if err != nil {
		log.Fatal("Failed to list files: %s", err)
	}

	if len(files) == 0 {
		log.Printf("No files found")
		return nil
	}
	log.Printf("Found %d files:", len(files))

	wg := sync.WaitGroup{}
	filesCh := make(chan internal.FileInfo, concurrency)
	for i := 0; i < int(concurrency); i++ {
		go downloadWorker(ctx, &wg, filesCh, sdtpFactory, noAck, destDir)
		wg.Add(1)
	}

	for _, file := range files {
		filesCh <- file
	}
	close(filesCh)

	wg.Wait()

	return nil
}

func downloadWorker(ctx context.Context, wg *sync.WaitGroup, files chan internal.FileInfo, sdtpFactory func() *internal.SDTP, noAck bool, destDir string) {
	defer wg.Done()
	sdtp := sdtpFactory()
	for {
		select {
		case file, more := <-files:
			if !more {
				return
			}
			log.Printf("downloading fileid=%d(%s)", file.ID, file.Name)
			if err := sdtp.Download(ctx, file, destDir); err != nil {
				log.Printf("failed to download fileid=%d(%s), skipping ack; %s", file.ID, file.Name, err)
				continue
			}
			if !noAck {
				if err := sdtp.Ack(ctx, file); err != nil {
					log.Printf("failed to ack fileid=%d(%s); %s", file.ID, file.Name, err)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
