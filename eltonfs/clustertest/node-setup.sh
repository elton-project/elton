#!/bin/bash
set -euvx

apt_install() {
    DEBIAN_FRONTEND=noninteractive apt install -y "$@"
}
setup_crash() {
    cat >~/.crashrc <<EOF
mod -s elton
bt
EOF
    cat >/usr/local/bin/last-crash <<'EOF'
#!/bin/bash
KERNEL=/usr/lib/debug/boot/vmlinux-$(uname -r)
DUMP=$(find /var/crash/ -type f -name 'dump.*' |sort |tail -n1)

echo "Kernel: $KERNEL"
echo "Dump: $DUMP"
crash "$KERNEL" "$DUMP"
EOF
    chmod +x /usr/local/bin/last-crash
    cat >/usr/local/bin/live-crash <<'EOF'
#!/bin/bash
KERNEL=/usr/lib/debug/boot/vmlinux-$(uname -r)

echo "Kernel: $KERNEL"
crash "$KERNEL"
EOF
    chmod +x /usr/local/bin/live-crash
}
install_kernel_debug_symbols() {
    cat >/etc/apt/sources.list.d/ddebs.list <<EOF
deb http://ddebs.ubuntu.com $(lsb_release -cs) main restricted universe multiverse
deb http://ddebs.ubuntu.com $(lsb_release -cs)-updates main restricted universe multiverse
deb http://ddebs.ubuntu.com $(lsb_release -cs)-proposed main restricted universe multiverse
EOF
    apt_install ubuntu-dbgsym-keyring
    apt update
    apt_install linux-image-$(uname -r)-dbgsym
}
install_go() {
    cd /usr/local/lib/
    wget https://dl.google.com/go/go1.13.7.linux-amd64.tar.gz
    tar xf go*.linux-amd64.tar.gz
    cd /usr/local/bin/
    ln -s ../lib/go/bin/* /usr/local/bin/
    cd
}


# Wait for all cloud-init tasks to complete.
until systemctl status cloud-final |grep -q 'Active: active (exited) since'; do
    sleep 1
    echo 'Waiting for all cloud-init tasks to complate ...'
done


# Change mirror server.
sed -i 's@//archive.ubuntu.com/@//jp.archive.ubuntu.com/@' /etc/apt/sources.list
sed -i 's@^\(\s\+\(primary\|security\):\) http://.*\.ubuntu\.com/ubuntu/\?$@\1 http://jp.archive.ubuntu.com/ubuntu@' /etc/cloud/cloud.cfg
apt update
apt upgrade -y

apt_install qemu-guest-agent
apt_install numactl
# Install and configure the kernel debug utilities.
# Disable writeback to prevent data lost when kernel panics.
apt_install kdump-tools crash gdb
# Use default settings.
# sed -i '/^#MAKEDUMP_ARGS/ i MAKEDUMP_ARGS=\"-c\"' /etc/default/kdump-tools
setup_crash
install_kernel_debug_symbols
sed -i 's/defaults/sync,noatime,nodiratime/' /etc/fstab
# Install the required packages for the elton.
apt_install build-essential automake libattr1-dev
install_go
# Install the required packages for the LTP test cases.
apt_install libaio-dev libnuma-dev libacl1-dev


# Setup git.
git config --global user.email "you@example.com"
git config --global user.name "Your Name"


# Clone latest elton source code.
# This directory used to manually debugging.
git clone --depth=1 https://gitlab.t-lab.cs.teu.ac.jp/yuuki/elton.git elton-base
# Download dependent modules to reduce execution time of "go build" command.
cd elton-base
go mod download
cd


# Install LTP
LTP_VERSION=20190930
[ -d ltp-full-$LTP_VERSION ] || curl -SsL https://github.com/linux-test-project/ltp/releases/download/$LTP_VERSION/ltp-full-$LTP_VERSION.tar.xz |xz -d |tar xv
ln -s ltp-full-$LTP_VERSION ltp
pushd ltp
make autotools
./configure
make -j`nproc` all
make install
popd


# Install SystemTAP from source
apt_install libdw-dev gettext
git clone --depth=1 git://sourceware.org/git/systemtap.git
pushd systemtap
./configure
make -j`nproc`
# Workaround for installation failure.
touch /root/systemtap/doc/SystemTap_Tapset_Reference/tapsets.pdf
make install
popd
wget https://sourceware.org/systemtap/examples/general/whythefail.stp


# Install golang debugger
go get -u github.com/go-delve/delve/cmd/dlv
ln -s ~/go/bin/dlv /usr/local/bin/


# Install docker
apt_install docker.io
# Download images
until docker version; do sleep 1; done
docker pull nginx
docker pull mysql:8.0
