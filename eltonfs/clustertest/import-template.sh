#!/bin/sh
# Import the ubuntu cloud image to specified VM.
#
# 1. Create VM.
# 2. Delete exists HDD from the VM.
# 3. Import the latest ubuntu cloud image to the VM.
#    `./update-template.sh [vmid]`
# 4. Attach the imported disk.

set -euv
target_vmid=$1

cd /var/tmp
rm -f disco-server-cloudimg-amd64.img
wget https://cloud-images.ubuntu.com/disco/current/disco-server-cloudimg-amd64.img
qm importdisk "$target_vmid" ./disco-server-cloudimg-amd64.img local --format qcow2
