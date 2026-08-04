#!/bin/bash
set -euo pipefail

INCUS_BRIDGE="incusbr0"
INCUS_SUBNET="10.162.89.0/24"

echo "Setting up NAT for Incus containers in WSL2..."

# Enable IP forwarding
sysctl -w net.ipv4.ip_forward=1

# Check if MASQUERADE rule already exists
if ! iptables -t nat -C POSTROUTING -s "$INCUS_SUBNET" ! -o "$INCUS_BRIDGE" -j MASQUERADE 2>/dev/null; then
    iptables -t nat -A POSTROUTING -s "$INCUS_SUBNET" ! -o "$INCUS_BRIDGE" -j MASQUERADE
    echo "Added MASQUERADE rule for $INCUS_SUBNET"
else
    echo "MASQUERADE rule already exists"
fi

# Check if FORWARD rules exist
if ! iptables -C FORWARD -i "$INCUS_BRIDGE" -j ACCEPT 2>/dev/null; then
    iptables -A FORWARD -i "$INCUS_BRIDGE" -j ACCEPT
    iptables -A FORWARD -o "$INCUS_BRIDGE" -m state --state RELATED,ESTABLISHED -j ACCEPT
    echo "Added FORWARD rules for $INCUS_BRIDGE"
else
    echo "FORWARD rules already exist"
fi

echo "NAT setup complete. Containers should have internet access."
