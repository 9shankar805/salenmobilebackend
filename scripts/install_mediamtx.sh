#!/bin/bash
# ==============================================================================
# Salen Mobile Guru - MediaMTX Self-Hosted HLS Live Streaming Server Setup
# Target Server: 209.50.241.46 (Linux AMD64)
# ==============================================================================

set -e

echo "🚀 [1/5] Creating installation directory..."
mkdir -p /opt/mediamtx
cd /opt/mediamtx

echo "📥 [2/5] Downloading MediaMTX v1.9.3..."
wget -q -O mediamtx.tar.gz https://github.com/bluenviron/mediamtx/releases/download/v1.9.3/mediamtx_v1.9.3_linux_amd64.tar.gz
tar -xzf mediamtx.tar.gz
rm -f mediamtx.tar.gz

echo "⚙️ [3/5] Configuring MediaMTX (RTMP on 1935, HLS on 8888)..."
cat << 'EOF' > /opt/mediamtx/mediamtx.yml
# MediaMTX Configuration for Salen Mobile Guru Live Classes
logLevel: info
logDestinations: [stdout]

# RTMP input from OBS Studio
rtmp: yes
rtmpAddress: :1935

# Native HLS output for SalenMobileGuru12 app
hls: yes
hlsAddress: :8888
hlsAlwaysRemux: no
hlsVariant: mpegts
hlsSegmentCount: 4
hlsSegmentDuration: 2s
hlsPartDuration: 500ms
hlsAllowOrigin: '*'

# Stream paths configuration
paths:
  all:
    readUser: ""
    readPass: ""
EOF

echo "🛠️ [4/5] Setting up systemd background service (runs 24/7 on boot)..."
cat << 'EOF' > /etc/systemd/system/mediamtx.service
[Unit]
Description=MediaMTX Live Streaming Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/mediamtx
ExecStart=/opt/mediamtx/mediamtx /opt/mediamtx/mediamtx.yml
Restart=always
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable mediamtx
systemctl restart mediamtx

echo "🔥 [5/5] Configuring firewall ports (1935 RTMP, 8888 HLS)..."
if command -v ufw >/dev/null 2>&1; then
    ufw allow 1935/tcp comment 'MediaMTX RTMP OBS'
    ufw allow 8888/tcp comment 'MediaMTX HLS Playback'
    ufw reload || true
fi

echo "=========================================================="
echo "✅ MediaMTX Live Streaming Server Installed & Running!"
echo "----------------------------------------------------------"
echo "OBS Stream Server: rtmp://209.50.241.46:1935/live/"
echo "OBS Stream Key   : salen"
echo "Mobile HLS Stream: http://209.50.241.46:8888/live/salen/index.m3u8"
echo "=========================================================="
systemctl status mediamtx --no-pager

