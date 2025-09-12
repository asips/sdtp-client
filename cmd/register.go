package cmd

import (
	"context"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/asips/sdtp-client/internal"
	"github.com/asips/sdtp-client/internal/log"
	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new client certificate with the server",
	Long:  "Register a new client certificate with the server",
	Run: func(cmd *cobra.Command, args []string) {
		apiUrl := parseApiUrl(strApiUrl)
		doRegister(apiUrl)
	},
}

func doRegister(apiUrl *url.URL) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	sdtp, err := internal.NewSDTP(apiUrl, certPath, keyPath, httpTimeout)
	if err != nil {
		log.Fatal("Failed to create SDTP client: %s", err)
	}

	if err := sdtp.Register(ctx); err != nil {
		log.Fatal("Registration failed: %s", err)
	}
	log.Printf("Registration successful. Contact your SDTP administrator to activate your account.")
}
