sudo apt-get update
sudo apt-get install git --yes
git init

curl -o go1.11.4.linux-armv6l.tar.gz https://dl.google.com/go/go1.11.4.linux-armv6l.tar.gz
sudo tar -C /usr/local -xzf go1.11.4.linux-armv6l.tar.gz

mkdir -p ~/go/bin
echo PATH=/usr/local/go/bin:~/go/bin:\$PATH >> .profile
source .profile

go get -u github.com/constabulary/gb/...
curl https://pre-commit.com/install-local.py | python -
source .profile

sudo apt-get install libusb-1.0-0-dev --yes
sudo ldconfig

git clone https://github.com/Yubico/yubihsm-connector.git ~/yubihsm-connector
cd yubihsm-connector
pre-commit install
make
cd ~

git clone https://github.com/openssl/openssl openssl-3.0.0-dev
cd openssl-3.0.0-dev
./config
make
sudo make install
cd ~

curl https://bootstrap.pypa.io/get-pip.py -o get-pip.py
sudo python get-pip.py
sudo apt-get install python-dev --yes
sudo apt-get install libffi-dev --yes

sudo pip install yubihsm[http]

cat > 10-yubihsm.rules << EOF
# This udev file should be used with udev 188 and newer
ACTION!="add|change", GOTO="yubihsm2_connector_end"

# Yubico YubiHSM2
# The OWNER attribute here has to match the uid of the process running the Connector
SUBSYSTEM=="usb", ATTRS{idVendor}=="1050", ATTRS{idProduct}=="0030", OWNER="yubihsm-connector"

LABEL="yubihsm2_connector_end”
EOF
sudo cp 10-yubihsm.rules /etc/udev/rules.d
sudo udevadm control --reload-rules

sudo yubihsm-connector/bin/yubihsm-connector &
