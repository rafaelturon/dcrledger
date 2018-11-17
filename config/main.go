// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/btcsuite/btclog"
	"github.com/decred/dcrd/dcrutil"
	flags "github.com/jessevdk/go-flags"
)

const (
	defaultConfigFilename = "dcrwalletpi.conf"
	defaultDataDirname    = "data"
	defaultLogLevel       = "info"
	defaultLogDirname     = "logs"
	defaultLogFilename    = "dcrwalletpi.log"
)

var (
	defaultHomeDir    = dcrutil.AppDataDir("dcrwalletpi", false)
	defaultConfigFile = filepath.Join(defaultHomeDir, defaultConfigFilename)
	defaultDataDir    = filepath.Join(defaultHomeDir, defaultDataDirname)
	defaultLogDir     = filepath.Join(defaultHomeDir, defaultLogDirname)
)

// Config defines the configuration options for dcrd.
//
// See loadConfig for details on the configuration load process.
type Config struct {
	DataDir          string          `short:"b" long:"datadir" description:"Directory to store data"`
	WalletPass       string          `long:"walletpass" default-mask:"-" description:"The public wallet password -- Only required if the wallet was created with one"`
	PoolAddress      dcrutil.Address `long:"pooladdress" description:"The ticket pool address where ticket fees will go to"`
	PoolFees         float64         `long:"poolfees" description:"The per-ticket fee mandated by the ticket pool as a percent (e.g. 1.00 for 1.00% fee)"`
	TicketFee        float64         `long:"ticketfee" description:"Ticket fee is the DCR/kB rate you’ll pay to have your ticket purchase be included in a block by a miner"`
	VotingAddress    dcrutil.Address `long:"votingaddress" description:"Purchase tickets with voting rights assigned to this address"`
	DaemonApp        string          `long:"daemonapp" description:"Decred Daemon Name"`
	WalletApp        string          `long:"walletapp" description:"Decred Wallet Name"`
	DecredBinFolder  string          `long:"decredbinfolder" description:"Decred binaries location"`
	HomeDir          string          `short:"A" long:"appdata" description:"Path to application home directory"`
	ShowVersion      bool            `short:"V" long:"version" description:"Display version information and exit"`
	ConfigFile       string          `short:"C" long:"configfile" description:"Path to configuration file"`
	LogDir           string          `long:"logdir" description:"Directory to log output."`
	LogFile          string          `long:"logfile" description:"File to log output."`
	NoFileLogging    bool            `long:"nofilelogging" description:"Disable file logging."`
	DebugLevel       string          `short:"d" long:"debuglevel" description:"Logging level for all subsystems {trace, debug, info, warn, error, critical} -- You may also specify <subsystem>=<level>,<subsystem2>=<level>,... to set the log level for individual subsystems -- Use show to list available subsystems"`
	SimNet           bool            `long:"simnet" description:"Use the simulation test network"`
	RPCUser          string          `short:"u" long:"rpcuser" description:"Username for RPC connections"`
	RPCPass          string          `short:"P" long:"rpcpass" default-mask:"-" description:"Password for RPC connections"`
	APIKey           string          `short:"k" long:"apikey" description:"Key for API connections"`
	APISecret        string          `short:"S" long:"apisecret" default-mask:"-" description:"Secret for API connections"`
	APIListen        string          `long:"apilisten" description:"API server will only listen on localhost"`
	APITokenDuration time.Duration   `long:"apitokenduration" description:"How long to token be valid.  Valid time units are {s, m, h}.  Minimum 1 second"`
	APIDisable       string          `long:"apidisable" description:"This allows one to quickly disable the API Server"`
}

// cleanAndExpandPath expands environment variables and leading ~ in the
// passed path, cleans the result, and returns it.
func cleanAndExpandPath(path string) string {
	// Expand initial ~ to OS specific home directory.
	if strings.HasPrefix(path, "~") {
		homeDir := filepath.Dir(defaultHomeDir)
		path = strings.Replace(path, "~", homeDir, 1)
	}

	// NOTE: The os.ExpandEnv doesn't work with Windows-style %VARIABLE%,
	// but they variables can still be expanded via POSIX-style $VARIABLE.
	return filepath.Clean(os.ExpandEnv(path))
}

// validLogLevel returns whether or not logLevel is a valid debug log level.
func validLogLevel(logLevel string) bool {
	_, ok := btclog.LevelFromString(logLevel)
	return ok
}

// supportedSubsystems returns a sorted slice of the supported subsystems for
// logging purposes.
func supportedSubsystems() []string {
	// Convert the subsystemLoggers map keys to a slice.
	subsystems := make([]string, 0, len(subsystemLoggers))
	for subsysID := range subsystemLoggers {
		subsystems = append(subsystems, subsysID)
	}

	// Sort the subsystems for stable display.
	sort.Strings(subsystems)
	return subsystems
}

// parseAndSetDebugLevels attempts to parse the specified debug level and set
// the levels accordingly.  An appropriate error is returned if anything is
// invalid.
func parseAndSetDebugLevels(debugLevel string) error {
	// When the specified string doesn't have any delimters, treat it as
	// the log level for all subsystems.
	if !strings.Contains(debugLevel, ",") && !strings.Contains(debugLevel, "=") {
		// Validate debug log level.
		if !validLogLevel(debugLevel) {
			str := "the specified debug level [%v] is invalid"
			return fmt.Errorf(str, debugLevel)
		}

		// Change the logging level for all subsystems.
		setLogLevels(debugLevel)

		return nil
	}

	// Split the specified string into subsystem/level pairs while detecting
	// issues and update the log levels accordingly.
	for _, logLevelPair := range strings.Split(debugLevel, ",") {
		if !strings.Contains(logLevelPair, "=") {
			str := "the specified debug level contains an invalid " +
				"subsystem/level pair [%v]"
			return fmt.Errorf(str, logLevelPair)
		}

		// Extract the specified subsystem and log level.
		fields := strings.Split(logLevelPair, "=")
		subsysID, logLevel := fields[0], fields[1]

		// Validate subsystem.
		if _, exists := subsystemLoggers[subsysID]; !exists {
			str := "the specified subsystem [%v] is invalid -- " +
				"supported subsytems %v"
			return fmt.Errorf(str, subsysID, supportedSubsystems())
		}

		// Validate log level.
		if !validLogLevel(logLevel) {
			str := "the specified debug level [%v] is invalid"
			return fmt.Errorf(str, logLevel)
		}

		setLogLevel(subsysID, logLevel)
	}

	return nil
}

// filesExists reports whether the named file or directory exists.
func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// newConfigParser returns a new command line flags parser.
func newConfigParser(cfg *Config, options flags.Options) *flags.Parser {
	parser := flags.NewParser(cfg, options)
	return parser
}

// createDefaultConfig copies the file sample-dcrd.conf to the given destination path,
// and populates it with some randomly generated API key and password.
func createDefaultConfigFile(destPath string) error {
	// Create the destination directory if it does not exist.
	err := os.MkdirAll(filepath.Dir(destPath), 0700)
	if err != nil {
		return err
	}

	// Generate a random key and secret for the API server credentials.
	randomBytes := make([]byte, 20)
	_, err = rand.Read(randomBytes)
	if err != nil {
		return err
	}
	generatedAPIKey := base64.StdEncoding.EncodeToString(randomBytes)
	apiKeyLine := fmt.Sprintf("apikey=%v", generatedAPIKey)

	_, err = rand.Read(randomBytes)
	if err != nil {
		return err
	}
	generatedAPISecret := base64.StdEncoding.EncodeToString(randomBytes)
	apiSecretLine := fmt.Sprintf("apisecret=%v", generatedAPISecret)

	// Replace the apikey and apisecret lines in the sample configuration
	// file contents with their generated values.
	apiKeyRE := regexp.MustCompile(`(?m)^;\s*apikey=[^\s]*$`)
	apiSecretRE := regexp.MustCompile(`(?m)^;\s*apisecret=[^\s]*$`)
	s := apiKeyRE.ReplaceAllString(FileContents, apiKeyLine)
	s = apiSecretRE.ReplaceAllString(s, apiSecretLine)

	// Create config file at the provided path.
	dest, err := os.OpenFile(destPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0600)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = dest.WriteString(s)
	return err
}

// loadConfig initializes and parses the config using a config file and command
// line options.
//
// The configuration proceeds as follows:
// 	1) Start with a default config with sane settings
// 	2) Pre-parse the command line to check for an alternative config file
// 	3) Load configuration file overwriting defaults with any specified options
// 	4) Parse CLI options and overwrite/add any specified options
//
// The above results in dcrd functioning properly without any config settings
// while still allowing the user to override settings with config files and
// command line options.  Command line options always take precedence.
func LoadConfig() (*Config, []string, error) {
	// Default config.
	cfg := Config{
		HomeDir:    defaultHomeDir,
		ConfigFile: defaultConfigFile,
		DebugLevel: defaultLogLevel,
		DataDir:    defaultDataDir,
		LogDir:     defaultLogDir,
	}

	// Pre-parse the command line options to see if an alternative config
	// file or the version flag was specified.  Any errors aside from the
	// help message error can be ignored here since they will be caught by
	// the final parse below.
	preCfg := cfg
	preParser := newConfigParser(&preCfg, flags.HelpFlag)
	_, err := preParser.Parse()
	if err != nil {
		if e, ok := err.(*flags.Error); ok && e.Type != flags.ErrHelp {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		} else if ok && e.Type == flags.ErrHelp {
			fmt.Fprintln(os.Stdout, err)
			os.Exit(0)
		}
	}

	// Show the version and exit if the version flag was specified.
	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	usageMessage := fmt.Sprintf("Use %s -h to show usage", appName)
	if preCfg.ShowVersion {
		fmt.Printf("%s version %s (Go version %s)\n", appName, Version(), runtime.Version())
		os.Exit(0)
	}

	// Update the home directory for dcrd if specified. Since the home
	// directory is updated, other variables need to be updated to
	// reflect the new changes.
	if preCfg.HomeDir != "" {
		cfg.HomeDir, _ = filepath.Abs(preCfg.HomeDir)

		if preCfg.ConfigFile == defaultConfigFile {
			defaultConfigFile = filepath.Join(cfg.HomeDir,
				defaultConfigFilename)
			preCfg.ConfigFile = defaultConfigFile
			cfg.ConfigFile = defaultConfigFile
		} else {
			cfg.ConfigFile = preCfg.ConfigFile
		}
		if preCfg.DataDir == defaultDataDir {
			cfg.DataDir = filepath.Join(cfg.HomeDir, defaultDataDirname)
		} else {
			cfg.DataDir = preCfg.DataDir
		}
		if preCfg.LogDir == defaultLogDir {
			cfg.LogDir = filepath.Join(cfg.HomeDir, defaultLogDirname)
		} else {
			cfg.LogDir = preCfg.LogDir
		}
	}

	// Create a default config file when one does not exist and the user did
	// not specify an override.
	if !preCfg.SimNet && preCfg.ConfigFile == defaultConfigFile &&
		!fileExists(preCfg.ConfigFile) {

		err := createDefaultConfigFile(preCfg.ConfigFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating a default "+
				"config file: %v\n", err)
		}
	}

	// Load additional config from file.
	var configFileError error
	parser := newConfigParser(&cfg, flags.Default)
	if !cfg.SimNet || preCfg.ConfigFile != defaultConfigFile {
		err := flags.NewIniParser(parser).ParseFile(preCfg.ConfigFile)
		if err != nil {
			if _, ok := err.(*os.PathError); !ok {
				fmt.Fprintf(os.Stderr, "Error parsing config "+
					"file: %v\n", err)
				fmt.Fprintln(os.Stderr, usageMessage)
				return nil, nil, err
			}
			configFileError = err
		}
	}

	// Parse command line options again to ensure they take precedence.
	remainingArgs, err := parser.Parse()
	if err != nil {
		if e, ok := err.(*flags.Error); !ok || e.Type != flags.ErrHelp {
			fmt.Fprintln(os.Stderr, usageMessage)
		}
		return nil, nil, err
	}

	// Create the home directory if it doesn't already exist.
	funcName := "loadConfig"
	err = os.MkdirAll(cfg.HomeDir, 0700)
	if err != nil {
		// Show a nicer error message if it's because a symlink is
		// linked to a directory that does not exist (probably because
		// it's not mounted).
		if e, ok := err.(*os.PathError); ok && os.IsExist(err) {
			if link, lerr := os.Readlink(e.Path); lerr == nil {
				str := "is symlink %s -> %s mounted?"
				err = fmt.Errorf(str, e.Path, link)
			}
		}

		str := "%s: failed to create home directory: %v"
		err := fmt.Errorf(str, funcName, err)
		fmt.Fprintln(os.Stderr, err)
		return nil, nil, err
	}

	// Append the network type to the data directory so it is "namespaced"
	// per network.  In addition to the block database, there are other
	// pieces of data that are saved to disk such as address manager state.
	// All data is specific to a network, so namespacing the data directory
	// means each individual piece of serialized data does not have to
	// worry about changing names per network and such.
	//
	// Make list of old versions of testnet directories here since the
	// network specific DataDir will be used after this.
	cfg.DataDir = cleanAndExpandPath(cfg.DataDir)
	var oldTestNets []string
	oldTestNets = append(oldTestNets, filepath.Join(cfg.DataDir, "testnet"))
	LogRotator = nil
	if !cfg.NoFileLogging {
		// Append the network type to the log directory so it is "namespaced"
		// per network in the same fashion as the data directory.
		cfg.LogDir = cleanAndExpandPath(cfg.LogDir)

		// Initialize log rotation.  After log rotation has been initialized, the
		// logger variables may be used.
		cfg.LogFile = filepath.Join(cfg.LogDir, defaultLogFilename)
		InitLogRotator(cfg.LogFile)
	}

	// Special show command to list supported subsystems and exit.
	if cfg.DebugLevel == "show" {
		fmt.Println("Supported subsystems", supportedSubsystems())
		os.Exit(0)
	}

	// Parse, validate, and set debug log level(s).
	if err := parseAndSetDebugLevels(cfg.DebugLevel); err != nil {
		err := fmt.Errorf("%s: %v", funcName, err.Error())
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// Warn about missing config file only after all other configuration is
	// done.  This prevents the warning on help messages and invalid
	// options.  Note this should go directly before the return.
	if configFileError != nil {
		DcrpLog.Warnf("%v", configFileError)
	}

	return &cfg, remainingArgs, nil
}
