#!/bin/bash
yum install -y gcc git
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
git clone https://github.com/aws/clock-bound
cd /clock-bound/clock-bound-d/
/root/.cargo/bin/cargo build --release
echo '# Ref: https://github.com/aws/clock-bound/tree/main/clock-bound-d' >> /etc/chrony.d/clockbound.conf
echo 'maxclockerror 50' >> /etc/chrony.d/clockbound.conf
systemctl restart chronyd
systemctl status chronyd
cp -v /clock-bound/target/release/clockbound /usr/local/bin/clockbound
chown chrony:chrony /usr/local/bin/clockbound
mkdir -p /run/clockbound
chmod 775 /run/clockbound
chown chrony:chrony /run/clockbound

cat >/usr/lib/systemd/system/clockbound.service <<EOL
[Unit]
Description=ClockBound

[Service]
Type=simple
Restart=always
RestartSec=10
ExecStart=/usr/local/bin/clockbound --max-drift-rate 50 --disable-clock-disruption-support
RuntimeDirectory=clockbound
RuntimeDirectoryPreserve=yes
WorkingDirectory=/run/clockbound
User=chrony
Group=chrony

[Install]
WantedBy=multi-user.target
EOL

systemctl daemon-reload
systemctl enable clockbound
systemctl start clockbound
systemctl status clockbound

METADATA_TOKEN=$(curl -X PUT "http://169.254.169.254/latest/api/token" -H "X-aws-ec2-metadata-token-ttl-seconds: 21600")
INTERNAL_IP=$(curl -H "X-aws-ec2-metadata-token: $METADATA_TOKEN" http://169.254.169.254/latest/meta-data/local-ipv4)
mkdir -p /etc/hedge/ && echo -n "$INTERNAL_IP" > /etc/hedge/internal-ip

SAMPLE_VERSION=$(curl -s https://api.github.com/repos/flowerinthenight/hedge-cb/releases/latest | jq -r ".tag_name")
cd /tmp/ && wget https://github.com/flowerinthenight/hedge-cb/releases/download/$SAMPLE_VERSION/hedge-$SAMPLE_VERSION-x86_64-linux.tar.gz
tar xvzf hedge-$SAMPLE_VERSION-x86_64-linux.tar.gz
cp -v example /usr/local/bin/hedge
chown root:root /usr/local/bin/hedge

# NOTE: This is NOT recommended! You can use IAM role + Secrets Manager instead, for example.
# The password will be visible under 'Advanced details' in 'Launch Templates', as well as
# tools that can display command line arguments, like, htop.
cat >/usr/lib/systemd/system/hedge.service <<EOL
[Unit]
Description=Hedge

[Service]
Type=simple
Restart=always
RestartSec=10
ExecStart=/usr/local/bin/hedge -db postgres://postgres:pass@location.rds.amazonaws.com:5432/dbname

[Install]
WantedBy=multi-user.target
EOL

systemctl daemon-reload
systemctl enable hedge
systemctl start hedge
systemctl status hedge
