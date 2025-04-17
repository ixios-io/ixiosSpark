#!/bin/bash

# Check if the built binary exists
if [ ! -f "./build/bin/ixiosSpark" ]; then
    echo "Error: ixiosSpark binary not found."
    echo "First run ./build.sh"
    exit 1
fi

# Check if ixiosSpark validator is already installed
if systemctl list-unit-files | grep -q ixios-validator.service; then
    echo "Existing validator service detected. Please use the validator installation script to update."
    exit 1
fi

# Check if ixiosSpark node is already installed
if systemctl list-unit-files | grep -q ixios-node.service; then
    echo "Existing service detected. Stopping for update..."
    sudo systemctl stop ixios-node.service
fi

# Create ixios user if it doesn't exist
if ! id "ixios" &>/dev/null; then
    echo "Creating ixios user..."

    if getent group ixios &>/dev/null; then
        sudo useradd -m -s /bin/bash -g ixios ixios
    else
        sudo useradd -m -s /bin/bash -U ixios
    fi
fi

# Proceed with installation
sudo cp ./build/bin/ixiosSpark /usr/bin/ixiosSpark

sudo mkdir -p /var/lib/ixios
sudo chown ixios:ixios /var/lib/ixios

# Create the systemd service file
cat << 'EOF' | sudo tee /etc/systemd/system/ixios-node.service
[Unit]
Description=IxiosSpark Node
After=network.target
StartLimitIntervalSec=0
StartLimitBurst=0

[Service]
Type=simple
User=ixios
Group=ixios
ExecStart=/usr/bin/ixiosSpark
Restart=always
RestartSec=1s

[Install]
WantedBy=multi-user.target
EOF

# Enable and start the service
sudo systemctl daemon-reload
sudo systemctl enable ixios-node.service
sudo systemctl start ixios-node.service

echo "ixiosSpark has been successfully installed and started."
