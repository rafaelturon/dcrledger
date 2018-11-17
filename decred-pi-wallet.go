package main

import (
	"os"
	"runtime"

	"github.com/rafaelturon/decred-pi-wallet/cmd/decredservices"
	"github.com/rafaelturon/decred-pi-wallet/cmd/muxservice"
	"github.com/rafaelturon/decred-pi-wallet/config"
)

var cfg *config.Config

func dcrpMain() error {
	// Load configuration and parse command line.  This function also
	// initializes logging and configures it accordingly.
	tcfg, _, err := config.LoadConfig()
	if err != nil {
		return err
	}
	cfg = tcfg
	defer func() {
		if config.LogRotator != nil {
			config.LogRotator.Close()
		}
	}()

	// Show version and home dir at startup.
	config.DcrpLog.Infof("Version %s (Go version %s)", config.Version(), runtime.Version())

	return nil
}
func main() {
	// Work around defer not working after os.Exit()
	if err := dcrpMain(); err != nil {
		os.Exit(1)
	}

	client, err := decredservices.Start(cfg)
	if err != nil {
		config.DcrpLog.Errorf("Error connecting to 'decredservices' %v", err)
	}

	muxservice.Start(cfg, client)
}
