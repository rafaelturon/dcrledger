#!/bin/bash


DECRED_RELEASE=v1.3.0

DIST=decred-linux-arm-$DECRED_RELEASE
TARBALL=$DIST.tar.gz
DCRURL=https://github.com/decred/decred-binaries/releases/download/${DECRED_RELEASE}/${TARBALL}

# update distro to latest
sudo apt-get update
sudo apt-get -qy dist-upgrade

# install required packages
sudo apt-get -qy install rng-tools jq

# set up hardware RNG generator
# http://fios.sector16.net/hardware-rng-on-raspberry-pi/
echo 'HRNGDEVICE=/dev/hwrng' | sudo tee -a /etc/default/rng-tools
echo bcm2708_rng | sudo tee -a /etc/modules

# create & populate decred configuration directories
mkdir ~/.dcrd ~/.dcrwallet ~/.dcrctl

RPC_PASSWORD=$(openssl rand -hex 32)

tee .dcrd/dcrd.conf > /dev/null <<DCRD_CONF
rpcuser=rpc
rpcpass=$RPC_PASSWORD
DCRD_CONF

tee .dcrctl/dcrctl.conf > /dev/null <<DCRCTL_CONF
rpcuser=rpc
rpcpass=$RPC_PASSWORD
DCRCTL_CONF

tee .dcrwallet/dcrwallet.conf > /dev/null <<DCRWALLET_CONF
username=rpc
password=$RPC_PASSWORD
DCRWALLET_CONF

# set up Decred tools
echo 'export PATH=/home/pi/decred:$PATH' >> ~/.bashrc
. ~/.bashrc

echo Downloading Decred $DECRED_RELEASE binary
rm -f decred
wget -q $DCRURL && \
tar xzf $TARBALL && \
ln -s $DIST decred && \
rm -f $TARBALL

echo
echo Installation completed, you can disconnect your network cable now.
echo
echo -n Press any key to reboot:
read
echo Rebooting...
sudo reboot
