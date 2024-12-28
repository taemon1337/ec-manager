#!/bin/bash
set -e

# Update package lists
if [ -f /etc/debian_version ]; then
    # Ubuntu/Debian
    sudo apt-get update
    sudo apt-get upgrade -y
    sudo apt-get install -y \
        curl \
        wget \
        git \
        vim \
        unzip
elif [ -f /etc/redhat-release ]; then
    # RHEL/CentOS
    echo "Installing base packages..."
    sudo dnf install -y --skip-broken \
        curl \
        wget \
        git \
        vim \
        unzip
fi

# Install AWS CLI
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install
rm -rf aws awscliv2.zip

# Set timezone to UTC
sudo timedatectl set-timezone UTC

# Configure sysctl for better network performance
cat << EOF | sudo tee /etc/sysctl.d/99-network-performance.conf
net.core.somaxconn = 1024
net.core.netdev_max_backlog = 5000
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
net.ipv4.tcp_wmem = 4096 12582912 16777216
net.ipv4.tcp_rmem = 4096 12582912 16777216
net.ipv4.tcp_max_syn_backlog = 8096
net.ipv4.tcp_slow_start_after_idle = 0
net.ipv4.tcp_tw_reuse = 1
EOF

sudo sysctl -p /etc/sysctl.d/99-network-performance.conf

# Set up CloudWatch agent
wget https://s3.amazonaws.com/amazoncloudwatch-agent/linux/amd64/latest/AmazonCloudWatchAgent.zip
unzip AmazonCloudWatchAgent.zip
sudo ./install.sh
rm -rf AmazonCloudWatchAgent*

# Basic security configurations
sudo sed -i 's/#ClientAliveInterval.*/ClientAliveInterval 300/' /etc/ssh/sshd_config
sudo sed -i 's/#ClientAliveCountMax.*/ClientAliveCountMax 2/' /etc/ssh/sshd_config
sudo systemctl restart sshd

# Clean up
if [ -f /etc/debian_version ]; then
    sudo apt-get clean
    sudo apt-get autoremove -y
elif [ -f /etc/redhat-release ]; then
    sudo dnf clean all
fi
