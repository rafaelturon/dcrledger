#!/bin/bash

WIFI_SSID=mySSID
WIFI_PASSWORD=myPassword

# install required packages
sudo apt-get -qy install wpasupplicant

echo Adding $WIFI_SSID SSID network
wpa_passphrase $WIFI_SSID $WIFI_PASSWORD >> /etc/wpa_supplicant/wpa_supplicant.conf

echo "@reboot screen -d -S decred-pi-wallet -m /home/pi/decred-pi-wallet" >> /var/spool/cron/crontabs/pi

echo
echo Configuration completed, you can disconnect your network cable now.
echo
echo -n Press any key to reboot:
read
echo Rebooting...
sudo reboot