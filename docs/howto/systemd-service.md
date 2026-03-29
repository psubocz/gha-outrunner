# How to Deploy as a systemd Service

Run outrunner as a persistent service on Linux so it starts on boot and restarts on failure.

## 1. Install the Binary

```bash
sudo cp outrunner /usr/local/bin/outrunner
sudo chmod +x /usr/local/bin/outrunner
```

## 2. Create a Config File

```bash
sudo mkdir -p /etc/outrunner
sudo cp outrunner.yml /etc/outrunner/config.yml
```

## 3. Store the Token

Create an environment file that systemd will load. This keeps the token out of the unit file:

```bash
sudo tee /etc/outrunner/env > /dev/null <<'EOF'
OUTRUNNER_TOKEN=ghp_YOUR_TOKEN_HERE
EOF
sudo chmod 600 /etc/outrunner/env
```

## 4. Create a Service User

```bash
sudo useradd --system --no-create-home outrunner
```

If using Docker, add the user to the `docker` group:

```bash
sudo usermod -aG docker outrunner
```

If using libvirt, add to the `libvirt` group:

```bash
sudo usermod -aG libvirt outrunner
```

## 5. Create the Unit File

```bash
sudo tee /etc/systemd/system/outrunner.service > /dev/null <<'EOF'
[Unit]
Description=outrunner - ephemeral GitHub Actions runners
After=network-online.target docker.service libvirtd.service
Wants=network-online.target

[Service]
Type=simple
User=outrunner
EnvironmentFile=/etc/outrunner/env
ExecStart=/usr/local/bin/outrunner \
    --url https://github.com/your/repo \
    --token ${OUTRUNNER_TOKEN} \
    --config /etc/outrunner/config.yml \
    --max-runners 2
Restart=on-failure
RestartSec=10

# Security hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/tmp
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF
```

If using libvirt with overlay files outside `/tmp`, add the overlay directory to `ReadWritePaths`.

## 6. Enable and Start

```bash
sudo systemctl daemon-reload
sudo systemctl enable outrunner
sudo systemctl start outrunner
```

## 7. Check Status

```bash
sudo systemctl status outrunner
sudo journalctl -u outrunner -f
```

## Updating

To update outrunner:

```bash
sudo systemctl stop outrunner
sudo cp outrunner-new /usr/local/bin/outrunner
sudo systemctl start outrunner
```

## Token Rotation

Update the token in `/etc/outrunner/env` and restart:

```bash
sudo systemctl restart outrunner
```
