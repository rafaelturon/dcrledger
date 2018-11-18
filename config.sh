#!/bin/bash

echo "@reboot screen -d -S dcrledger -m /home/user/dcrledger" >> /var/spool/cron/crontabs/pi

echo
echo Configuration completed, you can disconnect your network cable now.
echo
echo -n Press any key to reboot:
read
echo Rebooting...
sudo reboot