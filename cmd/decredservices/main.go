package decredservices

import (
	"errors"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/decred/dcrd/rpcclient"
	"github.com/rafaelturon/decred-pi-wallet/config"
)

var (
	ServiceConfig *config.Config
	Log           = config.DsvcLog
)

func getScreenCommand(appDir string, appName string) (string, error) {
	var homeDir string
	// Get the OS specific home directory via the Go standard lib.
	usr, err := user.Current()
	if err == nil {
		homeDir = usr.HomeDir
	}

	// Fall back to standard HOME environment variable that works
	// for most POSIX OSes if the directory from the Go standard
	// lib failed.
	if err != nil || homeDir == "" {
		homeDir = os.Getenv("HOME")
	}

	if homeDir != "" {
		return filepath.Join(homeDir, appDir, appName), nil
	}

	return "", errors.New("usr home: app not found")
}

func executeBashCommand(argCmd string) (string, error) {
	bashCmd := exec.Command("bash", "-c", argCmd)
	Log.Debugf("Executing bash ~$ %s", argCmd)
	bashOut, err := bashCmd.Output()
	if err != nil {
		return "", err
	}
	return string(bashOut), nil
}

func startDaemonAndWallet(defaultDaemonFilename, defaultWalletFilename, defaultDecredDirname string) error {
	var dcrdPs string
	var dcrdArg string
	var dcrwalletPs string
	var dcrwalletArg string

	dcrdPs = "pidof " + defaultDaemonFilename
	dcrdPsOut, err := executeBashCommand(dcrdPs)
	if len(dcrdPsOut) == 0 && err != nil {
		Log.Info("Starting decred daemon...")
		dcrdExec, err := getScreenCommand(defaultDecredDirname, defaultDaemonFilename)
		if err != nil {
			return err
		}
		dcrdArg = "screen -d -S dcrd -m " + dcrdExec
		dcrdOut, err := executeBashCommand(dcrdArg)
		if err != nil {
			Log.Errorf("Error initializing 'dcrd': %s", dcrdOut)
			return err
		}
	} else {
		Log.Debugf("Decred daemon process already running with PID: %s", strings.TrimSpace(dcrdPsOut))
	}

	dcrwalletPs = "pidof " + defaultWalletFilename
	dcrwalletPsOut, err := executeBashCommand(dcrwalletPs)
	if len(dcrwalletPsOut) == 0 && err != nil {
		Log.Info("Starting decred wallet...")
		dcrwalletExec, err := getScreenCommand(defaultDecredDirname, defaultWalletFilename)
		if err != nil {
			return err
		}
		dcrwalletArg = "screen -d -S dcrwallet -m " + dcrwalletExec
		dcrwalletOut, err := executeBashCommand(dcrwalletArg)
		if err != nil {
			Log.Errorf("Error initializing 'dcrwallet': %s", dcrwalletOut)
			return err
		}
	} else {
		Log.Debugf("Decred wallet process already running with PID: %s", strings.TrimSpace(dcrwalletPsOut))
	}

	return nil
}

// Start Decred wallet and daemon services
func Start(tcfg *config.Config) (*rpcclient.Client, error) {
	ServiceConfig = tcfg
	config.InitLogRotator(ServiceConfig.LogFile)
	UseLogger(Log)
	Log.Debugf("Decred daemon app name: %s", ServiceConfig.DaemonApp)
	Log.Debugf("Decred wallet app name: %s", ServiceConfig.WalletApp)
	Log.Debugf("Decred bin folder: %s", ServiceConfig.DecredBinFolder)

	err := startDaemonAndWallet(ServiceConfig.DaemonApp, ServiceConfig.WalletApp, ServiceConfig.DecredBinFolder)
	if err != nil {
		Log.Errorf("Error starting 'Daemon and Wallet': %v", err)
		return nil, err
	}

	client, err := newRPCService()
	if err != nil {
		Log.Errorf("Error starting 'rpcclient': %v", err)
		return nil, err
	}

	return client, nil
}
