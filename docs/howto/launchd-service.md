# How to Deploy as a launchd Service

Run outrunner as a persistent service on macOS so it starts on login and restarts on failure.

## 1. Install the Binary

```bash
sudo cp outrunner /usr/local/bin/outrunner
```

## 2. Prepare Config and Directories

```bash
mkdir -p ~/.config/outrunner
cp outrunner.yml ~/.config/outrunner/config.yml
```

## 3. Create the Launch Agent

Create `~/Library/LaunchAgents/com.outrunner.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.outrunner</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/outrunner</string>
        <string>--url</string>
        <string>https://github.com/your/repo</string>
        <string>--token</string>
        <string>YOUR_TOKEN_HERE</string>
        <string>--config</string>
        <string>/Users/YOU/.config/outrunner/config.yml</string>
        <string>--max-runners</string>
        <string>2</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <dict>
        <key>SuccessfulExit</key>
        <false/>
    </dict>
    <key>ThrottleInterval</key>
    <integer>10</integer>
    <key>StandardOutPath</key>
    <string>/Users/YOU/.config/outrunner/outrunner.log</string>
    <key>StandardErrorPath</key>
    <string>/Users/YOU/.config/outrunner/outrunner.log</string>
</dict>
</plist>
```

Replace `YOUR_TOKEN_HERE` and `/Users/YOU/` with your actual values.

Note: For better token security, consider using `EnvironmentVariables` with a script that reads from the keychain, or store the token in a wrapper script with restricted permissions.

## 4. Load the Service

```bash
launchctl load ~/Library/LaunchAgents/com.outrunner.plist
```

## 5. Check Status

```bash
launchctl list | grep outrunner
tail -f ~/.config/outrunner/outrunner.log
```

## Managing the Service

```bash
# Stop
launchctl unload ~/Library/LaunchAgents/com.outrunner.plist

# Start
launchctl load ~/Library/LaunchAgents/com.outrunner.plist

# Restart (unload + load)
launchctl unload ~/Library/LaunchAgents/com.outrunner.plist
launchctl load ~/Library/LaunchAgents/com.outrunner.plist
```

## Launch Agent vs Launch Daemon

This guide uses a **Launch Agent** (runs as your user, starts on login). If you need outrunner to run before anyone logs in (e.g., on a headless Mac mini), use a **Launch Daemon** instead:

- Move the plist to `/Library/LaunchDaemons/com.outrunner.plist`
- Set `UserName` to the desired user
- Use `sudo launchctl load/unload`
- Tart requires a user session, so Launch Agent is usually the right choice for Tart

## Updating

```bash
launchctl unload ~/Library/LaunchAgents/com.outrunner.plist
sudo cp outrunner-new /usr/local/bin/outrunner
launchctl load ~/Library/LaunchAgents/com.outrunner.plist
```
