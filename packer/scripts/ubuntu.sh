#!/bin/bash
set -e

# Install additional Ubuntu-specific packages
sudo apt-get install -y \
    software-properties-common \
    apt-transport-https \
    ca-certificates \
    gnupg \
    lsb-release \
    ubuntu-advantage-tools \
    chrony \
    apparmor \
    apparmor-utils

# Configure chrony
sudo systemctl enable chrony
sudo systemctl start chrony

# Enable automatic security updates
sudo apt-get install -y unattended-upgrades
sudo dpkg-reconfigure -f noninteractive unattended-upgrades

# Configure AppArmor
sudo aa-enforce /etc/apparmor.d/*

# Clean up
sudo apt-get clean
sudo apt-get autoremove -y
sudo rm -rf /var/lib/apt/lists/*
