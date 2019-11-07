#!/bin/bash
set -euvx

apt_install() {
    DEBIAN_FRONTEND=noninteractive apt install -y "$@"
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
    wget https://dl.google.com/go/go1.13.4.linux-amd64.tar.gz
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
apt update
apt upgrade -y

apt_install qemu-guest-agent
# Install the kernel debug utilities.
# Disable writeback to prevent data lost when kernel panics.
apt_install kdump-tools crash gdb
install_kernel_debug_symbols
sed -i 's/defaults/sync,noatime,nodiratime/' /etc/fstab
# Install the required packages for the elton.
apt_install build-essential automake libattr1-dev
install_go
# Install the required packages for the LTP test cases.
apt_install libaio-dev libnuma-dev libacl1-dev


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


# Install docker
apt_install docker.io
# Download images
until docker version; do sleep 1; done
docker pull nginx
docker pull mysql:8.0
