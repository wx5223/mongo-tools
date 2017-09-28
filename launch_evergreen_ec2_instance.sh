#!/bin/bash
# Script launch an AWS EC2 instance in Evergreen from the current variant.
# See _usage_ for how this script should be invoked.

set -o errexit

# Default options
tag_name="Evergreen AMI"

# _usage_: Provides usage infomation
function _usage_ {
  cat << EOF
usage: $0 options
This script supports the following parameters for Linux platforms:
  -k <ssh_key_id>,     [REQUIRED] The ssh key id used to access the new AWS EC2 instance.
  -y <aws_ec2_yml>,    [REQUIRED] YAML file name where to store the new AWS EC2 instance
                       information. This file will be used in Evergreen for macro
                       expansion of variables used in other functions.
  -a <ami>,            [OPTIONAL] If not specifed defaults to the same as this host.
  -d <device>,         [OPTIONAL] EBS block device, specified as "name/size(GiB)/type/IOPS",
                       i.e., "xvdf/500/io1/1000" or "xvdg/3072/gp2/0". To specify more than
                       one device, invoke this option each time.
  -s <security_group>, [OPTIONAL] The security group to be used for the new AWS EC2 instance.
                       To specify more than one group, invoke this option each time.
  -i <instance_type>,  [OPTIONAL] If not specifed defaults to the same as this host.
  -p <key_prefix>,     [OPTIONAL] The key prefix for keys stored in the YAML file, i.e.,
                       'client_', will store keys as 'client_instance_id', etc.
  -t <tag_name>,       [OPTIONAL] The tag name of the new AWS EC2 instance.
EOF
}


# Parse command line options
while getopts "a:d:i:k:p:s:t:y:?" option
do
   case $option in
     a)
        ami=$OPTARG
        ;;
     d)
        devices="$devices $OPTARG"
        ;;
     i)
        instance_type=$OPTARG
        ;;
     k)
        ssh_key_id=$OPTARG
        ;;
     p)
        key_prefix=$OPTARG
        ;;
     s)
        sec_groups="$sec_groups $OPTARG"
        ;;
     t)
        tag_name=$OPTARG
        ;;
     y)
        aws_ec2_yml=$OPTARG
        ;;
     \?|*)
        _usage_
        exit 0
        ;;
    esac
done

if [ -z $aws_ec2_yml ]; then
  echo "Must specify aws_ec2_yml file"
  exit 1
fi

if [ -z $ssh_key_id ]; then
  echo "Must specify ssh_key_id"
  exit 1
fi

# Get the AMI information on the current host so we can launch a similar EC2 instance.
# See http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html#instancedata-data-retrieval
aws_metadata_url="http://169.254.169.254/latest/meta-data"
if [ -z $ami ]; then
  ami=$(curl -s $aws_metadata_url/ami-id)
fi
if [ -z $instance_type ]; then
  instance_type=$(curl -s $aws_metadata_url/instance-type)
fi

for device in $devices
do
  block_devices="$block_devices --blockDevice $(echo $device | tr -s '/' ' ')"
done

for sec_group in $sec_groups
do
  security_groups="$security_groups --securityGroup $sec_group"
done
echo "AMI EC2 info: $ami $instance_type"

# Launch a new instance.
aws_ec2=$(python aws_ec2.py                \
          --ami $ami                       \
          --instanceType $instance_type    \
          $block_devices                   \
          --keyName $ssh_key_id            \
          --mode create                    \
          $security_groups                 \
          --tagName "$tag_name"            \
          --tagOwner "$USER")
echo "Spawned new AMI EC2 instance: $aws_ec2"

# Get new instance ID & ip_address
instance_id=$(echo $aws_ec2 | sed -e "s/.*instance_id: //; s/ .*//")
ip_address=$(echo $aws_ec2 | sed -e "s/.*private_ip_address: //; s/ .*//")

# Save AWS information on spawned EC2 instance to be used as expansion macros.
echo "${key_prefix}instance_id: $instance_id" > $aws_ec2_yml
echo "${key_prefix}ami: $ami" >> $aws_ec2_yml
echo "${key_prefix}instance_type: $instance_type" >> $aws_ec2_yml
echo "${key_prefix}ip_address: $ip_address" >> $aws_ec2_yml
