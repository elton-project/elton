#!/bin/bash
set -euvx

apt_install() {
    until apt install -y "$@"; do
        sleep 1
    done
}

apt_install qemu-guest-agent
# Install the required packages for the elton.
apt_install build-essential automake libattr1-dev
# Install the required packages for the LTP test cases.
apt_install libaio-dev libnuma-dev libacl1-dev


# Install LTP
[ -d ltp-full-20190517 ] || curl -SsL https://github.com/linux-test-project/ltp/releases/download/20190517/ltp-full-20190517.tar.xz |xz -d |tar xv
ln -s ltp-full-20190517 ltp
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
