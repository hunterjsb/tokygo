#!/bin/bash

# Prepare application directory with correct permissions

set -e

echo "Setting ownership of /opt/tokygo to ec2-user..."

# Change ownership to ec2-user so scripts can write to it
chown -R ec2-user:ec2-user /opt/tokygo

echo "Permissions updated successfully"

exit 0
