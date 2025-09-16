package cmd

import (
	"context"
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
		flags := cmd.Flags()
		certPath, err := flags.GetString("cert")
		cobra.CheckErr(err)
		keyPath, err := flags.GetString("key")
		cobra.CheckErr(err)
		httpTimeout, err := flags.GetDuration("http-timeout")
		cobra.CheckErr(err)

		strApiUrl, err := flags.GetString("api-url")
		cobra.CheckErr(err)
		apiUrl := parseApiUrl(strApiUrl)
		sdtp, err := internal.NewDefaultSDTP(apiUrl, certPath, keyPath, httpTimeout)
		if err != nil {
			log.Fatal("Failed to create SDTP client: %s", err)
		}

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		if ok := doRegister(ctx, sdtp); !ok {
			log.Fatal("Failed to register: %s", err)
		}
	},
}

func doRegister(ctx context.Context, sdtp internal.SDTPClient) bool {
	err := sdtp.Register(ctx)
	if err == internal.ErrExists {
		log.Printf("Registration already exists. Contact your SDTP administrator to activate your account.")
		return false
	} else if err != nil {
		log.Printf("Registration failed: %s", err)
		return false
	}
	log.Printf("Registration successful. Contact your SDTP administrator to activate your account.")

	return true
}
