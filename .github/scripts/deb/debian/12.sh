#!/bin/bash
set -e

#if [ -z "$UPLOAD_USER" ] || [ -z "$UPLOAD_PASS" ]; then
#    echo "Missing UPLOAD_USER or UPLOAD_PASS"
#    exit 1
#fi

export DEBIAN_FRONTEND=noninteractive
apt-get update -y && apt-get upgrade -y >/dev/null 2>&1
apt-get install -y build-essential git sudo wget curl zip unzip jq rsync >/dev/null 2>&1

cd "$GITHUB_WORKSPACE"
AGENT_VERSION=$(cat VERSION)
DEB_PACKAGE_NAME="raweb-agent"
DEB_ARCH="amd64"
DEB_DIST="$BUILD_CODE"
DEB_PACKAGE_FILE_NAME="${DEB_PACKAGE_NAME}_${AGENT_VERSION}_${DEB_DIST}_${DEB_ARCH}.deb"
DEB_REPO_URL="https://repo.julio.al/$UPLOAD_USER/$BUILD_REPO/${DEB_DIST}/"

echo "Checking if $DEB_PACKAGE_FILE_NAME exists at $DEB_REPO_URL..."
if curl -s "$DEB_REPO_URL" | grep -q "$DEB_PACKAGE_FILE_NAME"; then
    echo "✅ Package $DEB_PACKAGE_FILE_NAME already exists. Skipping build."
    exit 0
fi
cd /tmp
wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
rm -f go${GO_VERSION}.linux-amd64.tar.gz
export GOROOT=/usr/local/go
export PATH=$PATH:$GOROOT/bin
cd $GITHUB_WORKSPACE
go clean -cache -modcache -testcache -i
go mod tidy
go build -ldflags="-s -w" -o agent run.go

# --- Prepare DEB package structure ---
DEB_BUILD_DIR="$GITHUB_WORKSPACE/debbuild"
DEB_ROOT="$DEB_BUILD_DIR/${DEB_PACKAGE_NAME}_${AGENT_VERSION}_${DEB_ARCH}"

rm -rf "$DEB_BUILD_DIR"
mkdir -p "$DEB_ROOT/raweb/apps/agent"
mkdir -p "$DEB_ROOT/etc/systemd/system"
mkdir -p "$DEB_ROOT/DEBIAN"

cp "$GITHUB_WORKSPACE/agent" "$DEB_ROOT/raweb/apps/agent/"
cp "$GITHUB_WORKSPACE/config.json" "$DEB_ROOT/raweb/apps/agent/"
chmod +x "$DEB_ROOT/raweb/apps/agent/agent"

cat > "$DEB_ROOT/etc/systemd/system/raweb-agent.service" <<EOF
[Unit]
Description=Raweb Panel Agent
After=network.target
Wants=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/raweb/apps/agent
ExecStart=/raweb/apps/agent/agent --config=/raweb/apps/agent/config.json
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

cat > "$DEB_ROOT/DEBIAN/control" <<EOF
Package: $DEB_PACKAGE_NAME
Version: $AGENT_VERSION
Section: utils
Priority: optional
Architecture: $DEB_ARCH
Maintainer: Raweb Panel <cd@julio.al>
Description: Raweb Panel Agent for Debian 12 (Bookworm)
 Agent for Raweb Panel management system.
 Built for Debian 12 (Bookworm).
EOF

cat > "$DEB_ROOT/DEBIAN/postinst" <<'EOF'
#!/bin/bash
set -e
systemctl daemon-reload
systemctl enable raweb-agent.service
if ! systemctl is-active --quiet raweb-agent.service; then
    systemctl start raweb-agent.service
fi
echo "Raweb Agent installed and started"
EOF

cat > "$DEB_ROOT/DEBIAN/prerm" <<'EOF'
#!/bin/bash
set -e
if systemctl is-active --quiet raweb-agent.service; then
    systemctl stop raweb-agent.service
fi
if systemctl is-enabled --quiet raweb-agent.service; then
    systemctl disable raweb-agent.service
fi
EOF

cat > "$DEB_ROOT/DEBIAN/postrm" <<'EOF'
#!/bin/bash
set -e
systemctl daemon-reload
if [ -d "/raweb/apps/agent" ] && [ -z "$(ls -A /raweb/apps/agent)" ]; then
    rmdir /raweb/apps/agent
fi
if [ -d "/raweb/apps" ] && [ -z "$(ls -A /raweb/apps)" ]; then
    rmdir /raweb/apps
fi
echo "Raweb Agent removed successfully"
EOF

chmod 755 "$DEB_ROOT/DEBIAN/postinst" "$DEB_ROOT/DEBIAN/prerm" "$DEB_ROOT/DEBIAN/postrm"

DEB_PACKAGE_FILE="$DEB_BUILD_DIR/${DEB_PACKAGE_NAME}_${AGENT_VERSION}_${DEB_DIST}_${DEB_ARCH}.deb"
dpkg-deb --build "$DEB_ROOT" "$DEB_PACKAGE_FILE"

ls -la "$DEB_PACKAGE_FILE"
echo "$UPLOAD_PASS" > $GITHUB_WORKSPACE/.rsync
chmod 600 $GITHUB_WORKSPACE/.rsync
rsync -avz --password-file=$GITHUB_WORKSPACE/.rsync "$DEB_PACKAGE_FILE" rsync://$UPLOAD_USER@repo.julio.al/$BUILD_FOLDER/$BUILD_REPO/$BUILD_CODE/
