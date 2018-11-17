// Copyright (c) 2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package config

// FileContents is a string containing the commented example config for dcrd.
const FileContents = `[Application Options]

; ------------------------------------------------------------------------------
; Data settings
; ------------------------------------------------------------------------------

; The directory to store data such as the block chain and peer addresses.  The
; block chain takes several GB, so this location must have a lot of free space.
; The default is ~/.dcrd/data on POSIX OSes, $LOCALAPPDATA/Dcrd/data on Windows,
; ~/Library/Application Support/Dcrd/data on Mac OS, and $homed/dcrd/data on
; Plan9.  Environment variables are expanded so they may be used.  NOTE: Windows
; environment variables are typically %VARIABLE%, but they must be accessed with
; $VARIABLE here.  Also, ~ is expanded to $LOCALAPPDATA on Windows.
; datadir=~/.dcrd/data
pooladdress=POOL-ADDRESS
poolfees=5
ticketFee=1
votingaddress=TICKET-ADDRESS
walletpass=WALLET-PASS

; ------------------------------------------------------------------------------
; Decred Applications
; ------------------------------------------------------------------------------

; Application names.
daemonapp=dcrd
walletapp=dcrwallet
decredbinfolder=decred


; ------------------------------------------------------------------------------
; Network settings
; ------------------------------------------------------------------------------

; Use testnet.
; testnet=1

; Use simnet.
; simnet=1


; Use Universal Plug and Play (UPnP) to automatically open the listen port
; and obtain the external IP address from supported devices.  NOTE: This option
; will have no effect if exernal IP addresses are specified.
; upnp=1


; ------------------------------------------------------------------------------
; RPC server options - The following options control the built-in RPC server
; which is used to control and query information from a running dcrd process.
;
; NOTE: The RPC server is disabled by default if no rpcuser or rpcpass is
; specified.
; ------------------------------------------------------------------------------

; Secure the RPC API by specifying the key and secret.  You must specify
; both or the RPC server will be disabled.
; rpcuser=whatever_username_you_want
; rpcpass=
; apikey=whatever_username_you_want
; apisecret=

; How long to ban misbehaving peers. Valid time units are {s, m, h}.
; Minimum 1s.
; apitokenduration=24h
 apitokenduration=11h30m15s

; Specify the interfaces for the RPC server listen on.  One listen address per
; line.  NOTE: The default port is modified by some options such as 'testnet',
; so it is recommended to not specify a port and allow a proper default to be
; chosen unless you have a specific reason to do otherwise.  By default, the
; RPC server will only listen on localhost for IPv4 and IPv6.
; All interfaces on default port:
;   apilisten=
; All ipv4 interfaces on default port:
;   apilisten=0.0.0.0
; All ipv6 interfaces on default port:
;   apilisten=::
; All interfaces on port 8080:
   apilisten=:8080
; All ipv4 interfaces on port 8080:
;   apilisten=0.0.0.0:8080
; All ipv6 interfaces on port 8080:
;   apilisten=[::]:8080
; Only ipv4 localhost on port 8080:
;   apilisten=127.0.0.1:8080
; Only ipv6 localhost on port 8080:
;   apilisten=[::1]:8080
; Only ipv4 localhost on non-standard port 5000:
;   apilisten=127.0.0.1:5000
; All interfaces on non-standard port 5000:
;   apilisten=:5000
; All ipv4 interfaces on non-standard port 5000:
;   apilisten=0.0.0.0:5000
; All ipv6 interfaces on non-standard port 5000:
;   apilisten=[::]:5000


; Use the following setting to disable the API server even if the rpcuser and
; rpcpass are specified above.  This allows one to quickly disable the API
; server without having to remove credentials from the config file.
; apidisable=1


; ------------------------------------------------------------------------------
; Debug
; ------------------------------------------------------------------------------

; Debug logging level.
; Valid levels are {trace, debug, info, warn, error, critical}
; You may also specify <subsystem>=<level>,<subsystem2>=<level>,... to set
; log level for individual subsystems.  Use dcrd --debuglevel=show to list
; available subsystems.
 debuglevel=trace
`
