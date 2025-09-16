package cmd

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/asips/sdtp-client/internal"
	"github.com/asips/sdtp-client/internal/log"
	"github.com/spf13/cobra"
)

var ingestCmd = &cobra.Command{
	Use:   "ingest",
	Short: "Ingest data from SDTP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		certPath, err := flags.GetString("cert")
		cobra.CheckErr(err)
		keyPath, err := flags.GetString("key")
		cobra.CheckErr(err)
		httpTimeout, err := flags.GetDuration("http-timeout")
		cobra.CheckErr(err)
		checkCertDays, err := flags.GetInt("check-cert-days")
		cobra.CheckErr(err)
		checkCertExprFlag, err := flags.GetBool("check-cert-expr")
		cobra.CheckErr(err)

		mustValidateCert(certPath, keyPath, checkCertDays)

		strApiUrl, err := flags.GetString("api-url")
		cobra.CheckErr(err)
		apiUrl := parseApiUrl(strApiUrl)
		sdtp, err := internal.NewDefaultSDTP(apiUrl, certPath, keyPath, httpTimeout)
		if err != nil {
			log.Fatal("Failed to create SDTP client: %s", err)
		}

		destDir, err := flags.GetString("dest-dir")
		cobra.CheckErr(err)
		if _, err := os.Stat(destDir); os.IsNotExist(err) {
			log.Printf("creating destination directory: %s", destDir)
			os.MkdirAll(destDir, 0755)
		}

		tags, err := flags.GetStringToString("tag")
		cobra.CheckErr(err)

		stream, err := flags.GetString("stream")
		cobra.CheckErr(err)
		if flags.Changed("stream") {
			tags["stream"] = stream
		}
		mission, err := flags.GetString("mission")
		cobra.CheckErr(err)
		if flags.Changed("mission") {
			tags["mission"] = mission
		}
		shortName, err := flags.GetString("short-name")
		cobra.CheckErr(err)
		if flags.Changed("short-name") {
			tags["ShortName"] = shortName
		}
		if checkCertExprFlag {
			mustValidateCert(certPath, keyPath, checkCertDays)
		}

		noAckFlag, err := flags.GetBool("no-ack")
		cobra.CheckErr(err)

		concurrency, err := flags.GetUint("concurrency")
		cobra.CheckErr(err)

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		return doIngest(
			ctx,
			sdtp,
			destDir,
			tags,
			noAckFlag,
			concurrency,
		)
	},
}

func init() {
	flags := ingestCmd.Flags()

	flags.StringP("dest-dir", "d", "", "Local directory to ingest data to")
	flags.String("stream", "", "SDTP 'stream' field (query parameter)")
	flags.String("short-name", "", "SDTP 'ShortName' field (query parameter)")
	flags.String("mission", "", "SDTP 'mission' field (query parameter)")
	flags.StringToStringP("tag", "t", map[string]string{}, "<key>=<value> tags to filter by. May be specified multiple times or as a comma-separated list")
	flags.Bool("no-ack", false, "Skip acknowledgment after successful ingest")
	flags.Bool("list", false, "List available files, but do not download")
	flags.Duration("http-timeout", time.Minute*5, "HTTP client timeout in seconds for list operations")
	flags.Uint("concurrency", 4, "Number of concurrent downloads")

	flags.MarkDeprecated("list", "use 'list' sub-command instead")
}

func doIngest(ctx context.Context, sdtp internal.SDTPClient, destDir string, tags map[string]string, noAck bool, concurrency uint) error {
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
		go downloadWorker(ctx, &wg, sdtp, filesCh, noAck, destDir)
		wg.Add(1)
	}

	for _, file := range files {
		filesCh <- file
	}
	close(filesCh)

	wg.Wait()

	return nil
}

func defaultDownloadWorker(ctx context.Context, wg *sync.WaitGroup, sdtp internal.SDTPClient, files chan internal.FileInfo, noAck bool, destDir string) {
	defer wg.Done()

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

var downloadWorker = defaultDownloadWorker
