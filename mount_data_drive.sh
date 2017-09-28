#!/bin/bash
# Script used to mount /data on a separate drive from root.
# This script must be invoked either by a root user or with sudo.
# Currently only supports Linux.
# See _usage_ for how this script should be invoked.

set -o errexit

# Default options
fs_type=xfs
user_group=$USER:$(id -Gn $USER | cut -f1 -d ' ')

# _usage_: Provides usage infomation
function _usage_ {
  cat << EOF
usage: $0 options
This script supports the following parameters for Windows & Linux platforms:
  -d <deviceName>,     REQUIRED, Devices to mount /data on, i.e., "/dev/xvdb".
  -t <fsType>,         File system type, defaults to '$fs_type'.
  -o <mountOptions>,   File system mount options, i.e., "-m crc=0,finobt=0".
  -u <user:group>,     User:Group to make owner of /data. Defaults to '$user_group'.
EOF
}


# Parse command line options
while getopts "d:o:t:u:?" option
do
   case $option in
     d)
        device_name=$OPTARG
        ;;
     o)
        mount_options=$OPTARG
        ;;
     t)
        fs_type=$OPTARG
        ;;
     u)
        user_group=$OPTARG
        ;;
     \?|*)
        _usage_
        exit 0
        ;;
    esac
done

if [ -z $device_name ]; then
    echo "Must specify the device_name"
    _usage_
    exit 1
fi

# Unmount the current devices, if already mounted
umount /mnt || true
umount $device_name || true

# Mount the /data drive(s)
mkfs -t $fs_type $mount_options -f $device_name
echo "$device_name /data auto noatime 0 0" | tee -a /etc/fstab
mount -t $fs_type $device_name /data
mkdir -p /data/db || true
chown -R $user_group /data
mkdir /data/tmp
chmod 1777 /data/tmp
