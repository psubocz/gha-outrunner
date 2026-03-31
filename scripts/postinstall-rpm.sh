#!/bin/sh
set -e

# Reload systemd
if [ -d /run/systemd/system ]; then
    systemctl daemon-reload >/dev/null || true
fi
