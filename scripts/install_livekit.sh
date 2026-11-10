#!/bin/bash
# ==============================================================================
# Salen Mobile Guru - LiveKit WebRTC Audio/Video Call Server Setup
# Target Server: 209.50.241.46 (Linux AMD64)
# ==============================================================================

set -e

echo "🚀 [1/5] Installing LiveKit Server..."
mkdir -p /opt/livekit
cd /opt/livekit

# Download official LiveKit binary using the official installer or direct release
if command -v curl >/dev/null 2>&1; then
    curl -sSL https://get.livekit.io | bash
fi

# Ensure /usr/local/bin/livekit-server exists
if [ ! -f /usr/local/bin/livekit-server ]; then
    echo "Fetching latest livekit-server from GitHub..."
    LATEST_TAG=$(curl -s https://api.github.com/repos/livekit/livekit/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    wget -q "https://github.com/livekit/livekit/releases/download/${LATEST_TAG}/livekit_${LATEST_TAG#v}_linux_amd64.tar.gz" -O livekit.tar.gz
    tar -xzf livekit.tar.gz -C /usr/local/bin/
    rm -f livekit.tar.gz
fi

echo "⚙️ [2/5] Creating LiveKit configuration (livekit.yaml)..."
cat << 'EOF' > /opt/livekit/livekit.yaml
# LiveKit Server Configuration for Salen Mobile Guru In-App Calling
port: 7880
bind_addresses:
  - ""

rtc:
  tcp_port: 7881
  port_range_start: 50000
  port_range_end: 50100
  use_external_ip: true

room:
  auto_create: true
  empty_timeout: 300 # 5 minutes

# Credentials for generating tokens from salenmobilebackend
keys:
  APIKeySalenGuru2026: SecretKeySalenGuruAudioCallLive2026Master
EOF

echo "🛠️ [3/5] Setting up systemd service for 24/7 background operation..."
cat << 'EOF' > /etc/systemd/system/livekit.service
[Unit]
Description=LiveKit WebRTC Call Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/livekit
ExecStart=/usr/local/bin/livekit-server --config /opt/livekit/livekit.yaml
Restart=always
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

echo "🔄 [4/5] Reloading and starting LiveKit service..."
systemctl daemon-reload
systemctl enable livekit
systemctl restart livekit

echo "🔥 [5/5] Configuring firewall ports for WebRTC audio calls..."
if command -v ufw >/dev/null 2>&1; then
    ufw allow 7880/tcp comment 'LiveKit Signaling WebSocket'
    ufw allow 7881/tcp comment 'LiveKit WebRTC TCP'
    ufw allow 50000:50100/udp comment 'LiveKit WebRTC Media UDP'
    ufw reload || true
fi

echo "=========================================================="
echo "✅ LiveKit WebRTC Server Installed & Running on VPS!"
echo "----------------------------------------------------------"
echo "WebSocket URL : ws://209.50.241.46:7880"
echo "API Key       : APIKeySalenGuru2026"
echo "API Secret    : SecretKeySalenGuruAudioCallLive2026Master"
echo "=========================================================="
systemctl status livekit --no-pager

