#!/bin/bash
set -e

# Register with Red Hat Subscription Manager if credentials are provided
if [ ! -z "$RHSM_USER" ] && [ ! -z "$RHSM_PASS" ]; then
    sudo subscription-manager register --username="$RHSM_USER" --password="$RHSM_PASS" --auto-attach
fi

# Enable EPEL repository
sudo dnf install -y https://dl.fedoraproject.org/pub/epel/epel-release-latest-9.noarch.rpm

# Install additional RHEL-specific packages
sudo dnf install -y \
    tuned \
    chrony \
    policycoreutils-python-utils \
    setools-console \
    audit

# Configure tuned for virtual machine
sudo tuned-adm profile virtual-guest

# Configure chronyd
sudo systemctl enable chronyd
sudo systemctl start chronyd

# Configure SELinux
sudo semanage port -a -t ssh_port_t -p tcp 22

# Clean up
sudo dnf clean all
sudo rm -rf /var/cache/dnf/*

# Unregister from RHSM if we registered
if [ ! -z "$RHSM_USER" ] && [ ! -z "$RHSM_PASS" ]; then
    sudo subscription-manager unregister
fi
