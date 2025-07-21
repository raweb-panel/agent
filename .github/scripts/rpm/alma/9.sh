
if [ -z "$UPLOAD_USER" ] || [ -z "$UPLOAD_PASS" ]; then
    echo "Missing UPLOAD_USER or UPLOAD_PASS"
    exit 1
fi

dnf install -y epel-release >/dev/null 2>&1
dnf install -y gcc git sudo wget curl zip unzip jq rsync rpm-build >/dev/null 2>&1

cd "$GITHUB_WORKSPACE"
AGENT_VERSION=$(cat VERSION)
RPM_PACKAGE_NAME="raweb-agent"
RPM_ARCH="x86_64"
RPM_DIST="$BUILD_CODE"

RPM_BUILD_DIR="$GITHUB_WORKSPACE/rpmbuild"
RPM_ROOT="$RPM_BUILD_DIR/BUILDROOT/${RPM_PACKAGE_NAME}-${AGENT_VERSION}-${RPM_DIST}.${RPM_ARCH}"
RPM_PACKAGE_FILE_NAME="${RPM_PACKAGE_NAME}-${AGENT_VERSION}-${BUILD_CODE}.x86_64.rpm"
RPM_REPO_URL="https://repo.julio.al/$UPLOAD_USER/$BUILD_REPO/${BUILD_CODE}/"

if curl -s "$RPM_REPO_URL" | grep -q "$RPM_PACKAGE_FILE_NAME"; then
    echo "✅ Package $RPM_PACKAGE_FILE_NAME already exists. Skipping build."
    exit 0
fi

rm -rf "$RPM_BUILD_DIR"
mkdir -p "$RPM_BUILD_DIR"/{BUILD,BUILDROOT,RPMS,SOURCES,SPECS,SRPMS}
mkdir -p "$RPM_ROOT/raweb/apps/agent"
mkdir -p "$RPM_ROOT/etc/systemd/system"

cd /tmp
wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz >/dev/null 2>&1
tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz >/dev/null 2>&1
rm -f go${GO_VERSION}.linux-amd64.tar.gz
export GOROOT=/usr/local/go
export PATH=$PATH:$GOROOT/bin

cd $GITHUB_WORKSPACE
go clean -cache -modcache -testcache -i
go mod tidy
go build -ldflags="-s -w" -o agent run.go

cp "$GITHUB_WORKSPACE/agent" "$RPM_ROOT/raweb/apps/agent/"
cp "$GITHUB_WORKSPACE/config.json" "$RPM_ROOT/raweb/apps/agent/"

cat > "$RPM_ROOT/etc/systemd/system/raweb-agent.service" <<SERVICE
[Unit]
Description=Raweb Panel Agent
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/raweb/apps/agent
ExecStart=/raweb/apps/agent/agent --config=/raweb/apps/agent/config.json
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
SERVICE

cat > "$RPM_BUILD_DIR/SPECS/${RPM_PACKAGE_NAME}.spec" <<EOF
Name: $RPM_PACKAGE_NAME
Version: $AGENT_VERSION
Release: $BUILD_CODE
Summary: Raweb Panel Agent for AlmaLinux $BUILD_CODE
License: https://github.com/raweb-panel/agent?tab=License-1-ov-file
BuildArch: x86_64
Requires: systemd

%description
Raweb Panel Agent for AlmaLinux $BUILD_CODE

%post
systemctl daemon-reload
systemctl enable raweb-agent.service
systemctl restart raweb-agent.service || :

%preun
if [ \$1 -eq 0 ]; then
  systemctl stop raweb-agent.service || :
  systemctl disable raweb-agent.service || :
fi

%postun
systemctl daemon-reload

%files
/raweb/apps/agent/agent
/raweb/apps/agent/config.json
/etc/systemd/system/raweb-agent.service

%changelog
* $(date +"%a %b %d %Y") Raweb Panel <cd@julio.al> - $AGENT_VERSION-$BUILD_CODE
- Initial RPM release
EOF

rpmbuild --define "_topdir $RPM_BUILD_DIR" --buildroot "$RPM_ROOT" -bb "$RPM_BUILD_DIR/SPECS/${RPM_PACKAGE_NAME}.spec"

RPM_PACKAGE_FILE="$RPM_BUILD_DIR/RPMS/x86_64/${RPM_PACKAGE_NAME}-${AGENT_VERSION}-${BUILD_CODE}.x86_64.rpm"
ls -la "$RPM_PACKAGE_FILE"

echo "$UPLOAD_PASS" > $GITHUB_WORKSPACE/.rsync
chmod 600 $GITHUB_WORKSPACE/.rsync
rsync -avz --password-file=$GITHUB_WORKSPACE/.rsync "$RPM_PACKAGE_FILE" rsync://$UPLOAD_USER@repo.julio.al/$BUILD_FOLDER/$BUILD_REPO/$BUILD_CODE/
