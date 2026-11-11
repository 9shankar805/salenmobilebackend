#!/bin/bash
# ==============================================================================
# Fix MediaMTX config on VPS: enable API, allow all stream paths without auth
# ==============================================================================

set -e

echo "⚙️ Updating /opt/mediamtx/mediamtx.yml..."
cat << 'EOF' > /opt/mediamtx/mediamtx.yml
logLevel: info
logDestinations: [stdout]

# RTMP Ingest from OBS
rtmp: yes
rtmpAddress: :1935

# HLS Output for Web & App
hls: yes
hlsAddress: :8888
hlsAlwaysRemux: no
hlsVariant: mpegts
hlsSegmentCount: 4
hlsSegmentDuration: 2s
hlsPartDuration: 500ms
hlsAllowOrigin: '*'

# WebRTC
webrtc: yes
webrtcAddress: :8889
webrtcAllowOrigin: '*'

# API for backend health check
api: yes
apiAddress: 127.0.0.1:9997

# Stream paths configuration (accepts both live/<key> and <key>)
paths:
  'live/~^.*$':
  '~^.*$':
  all:
EOF

echo "🔄 Restarting MediaMTX service..."
systemctl restart mediamtx
sleep 1
systemctl status mediamtx --no-pager
echo "✅ MediaMTX is now configured to accept any OBS stream key without authentication!"

