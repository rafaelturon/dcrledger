## Decred wallet on the Raspberry Pi

### Features
* Automatically starts services
  - Decred daemon and wallet
* Exposes a secure API service
  - Balance
    * Available Balance
  - Tickets stats
    * Own Mempool
    * Immature
    * Live
    * Total Subsidy (in DCR)
* Enables access to more critical commands using 2FA
  - Device shutdown
  - Ticket buying

### Installation

This is a guide for setting up a [Decred](https://www.decred.org) wallet on the Raspberry Pi.


1. Get the [Raspbian Lite image](https://www.raspberrypi.org/downloads/raspbian/) and [flash it onto a USB stick or SD card](https://www.raspberrypi.org/documentation/installation/installing-images/README.md)


2. If you want to perform a headless installation (e.g. don't have a keyboard connected to the Pi) you can [enable SSH](https://www.raspberrypi.org/documentation/remote-access/ssh/) before booting the Pi for the first time.  Note that this way you won't have a 100% cold wallet.

3. Log into the Pi for the first time and change your user password.  Set up SSH access if you need to log in remotely.

4. Build 'Decred Pi Wallet' using ARM
 - $: git clone https://github.com/rafaelturon/decred-pi-wallet.git
 - $: go get -d ./...
 - $: env GOOS=linux GOARCH=arm go build -v github.com/rafaelturon/decred-pi-wallet
 - Check ARM $: file ./decred-pi-wallet
   * ELF 32-bit LSB executable, ARM, EABI5 version 1 (SYSV), statically linked, not stripped 

5. Copy 'Decred Pi Wallet' files
 - scp ~/decred-pi-wallet pi@raspberrypi.local:~/

6. Download the installer script and verify its SHA256 value:

````bash
wget https://raw.githubusercontent.com/rafaelturon/decred-pi-wallet/master/install.sh
sha256sum install.sh
2db3908d4e1d7325423b903e24ddd5b4d0181aa38f79ca474f56d373d4cc8ba8  install.sh

````

7. Run the install script that will update the system, install all the required packages and configure the Pi's [hardware random number generator (RNG)](http://fios.sector16.net/hardware-rng-on-raspberry-pi/).  After the upgrade and package installation is completed, it will ask you to confirm the kernel upgrade - answer *Yes*.  Once the upgrade is finished, the Pi will reboot.

````bash
./install.sh
````

8. If you want to run a cold wallet, you can now disconnect the network cable and carry on with creating your wallet offline.

9. After the reboot, log back in and proceed with creating your wallet: see [Offline wallets](https://github.com/chappjc/dcrwallet/blob/master/docs/offline_wallets.md) for more information.






Donate if you like the project: `DshQnZKBvxJzJVPF15qUUAPj7pCEGtRzgaD`
