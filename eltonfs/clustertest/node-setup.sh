#!/bin/bash
set -euvx

apt_install() {
    DEBIAN_FRONTEND=noninteractive apt install -y "$@"
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
# Install the required packages for the elton.
apt_install build-essential automake libattr1-dev
# Install the required packages for the LTP test cases.
apt_install libaio-dev libnuma-dev libacl1-dev


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
