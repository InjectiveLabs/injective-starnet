#!/bin/bash

#### Startup script that triggers on machine startup, put any logic here ###

# Install dependencies
sudo apt-get update
sudo apt-get install -y build-essential git curl

# Build raid only if it doesn't exist
if [ ! -e "/dev/md0" ]; then
    # Build raid
    sudo mdadm --create --verbose /dev/md0 --level=0 --raid-devices=4 /dev/nvme0n1 /dev/nvme0n2 /dev/nvme0n3 /dev/nvme0n4
    sudo mkfs.xfs /dev/md0
    sudo mount /dev/md0 /home/injectived/.injectived
    sudo chown -R injectived:injectived /home/injectived/.injectived
    sleep 30
    sudo cp -R /home/injectived/artifacts/* /home/injectived/.injectived/
    sudo chown -R injectived:injectived /home/injectived/.injectived
fi


