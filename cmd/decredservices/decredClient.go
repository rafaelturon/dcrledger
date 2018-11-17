package decredservices

import (
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrd/rpcclient"
)

// ConnectDaemon start a rpc client connection
func newRPCService() (*rpcclient.Client, error) {
	Log.Info("Starting Decred RPC Client")

	// Only override the handlers for notifications you care about.
	// Also note most of the handlers will only be called if you register
	// for notifications.  See the documentation of the rpcclient
	// NotificationHandlers type for more details about each handler.
	ntfnHandlers := rpcclient.NotificationHandlers{
		OnAccountBalance: func(account string, balance dcrutil.Amount, confirmed bool) {
			Log.Tracef("New balance for account %s: %v", account,
				balance)
		},
	}

	// Connect to local dcrwallet RPC server using websockets.
	dcrdHomeDir := dcrutil.AppDataDir("dcrwallet", false)
	certs, err := ioutil.ReadFile(filepath.Join(dcrdHomeDir, "rpc.cert"))
	if err != nil {
		Log.Errorf("Error reading certificate %v", err)
	}
	connCfg := &rpcclient.ConnConfig{
		Host:         "localhost:9110",
		Endpoint:     "ws",
		User:         ServiceConfig.RPCUser,
		Pass:         ServiceConfig.RPCPass,
		Certificates: certs,
	}

LOOP:
	for {
		client1, err := rpcclient.New(connCfg, &ntfnHandlers)
		if err != nil {
			Log.Warnf("Failed to connect to client. Waiting a minute to try again... %v", err)
			time.Sleep(time.Minute)
			continue LOOP
		}
		client1.ListUnspent()
		break
	}
	// Register for block connect and disconnect notifications.
	client, err := rpcclient.New(connCfg, &ntfnHandlers)
	if err != nil {
		Log.Errorf("Error connecting to rcp client %v", err)
		return nil, err
	}
	Log.Trace("Wallet: Registration Complete")

	return client, nil
}
