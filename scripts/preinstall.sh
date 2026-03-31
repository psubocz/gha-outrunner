#!/bin/sh
set -e

# Create system user/group before package files are installed,
# so /etc/outrunner can be owned by the outrunner user.
if ! getent group outrunner >/dev/null; then
    groupadd --system outrunner
fi
if ! getent passwd outrunner >/dev/null; then
    useradd --system \
        --gid outrunner \
        --no-create-home \
        --shell /bin/false \
        --comment "outrunner service" \
        outrunner
fi
