## Decred wallet ledger services

### Features
* Blockchain services and wallet
  - Connect to network
  - Automatically starts wallet services
  - Enables access to more critical commands using 2FA
    * Device shutdown
    * Ticket buying
* Exposes a secure API service
  - Balance
    * Available Balance
  - Tickets stats
    * Own Mempool
    * Immature
    * Live
    * Total Subsidy (in DCR)


### Architecture
* Wallet agent (docker)
  - Security and connectivity health check
  - Private key generation
    * Exodus process (12+ seed words, password, recovery link)
    https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki
  - Address management
  - Transactions
    * Send transaction
    * Ticket buying
* Ledger services (cloud)
  - Accounts
  - Payments
    * Invoices
    * Fiat and Crypto Transactions History
    * Third party payment processors
* Pricing services (cloud)

### Installation

This is a guide for setting up a [Decred](https://www.decred.org) wallet.

1. Download the installer script and verify its SHA256 value:

````bash
wget https://raw.githubusercontent.com/rafaelturon/dcrledger/master/install.sh
sha256sum install.sh
2db3908d4e1d7325423b903e24ddd5b4d0181aa38f79ca474f56d373d4cc8ba8  install.sh

````

2. Run the install script that will update the system, install all the required packages and configure it.

````bash
./install.sh
````

