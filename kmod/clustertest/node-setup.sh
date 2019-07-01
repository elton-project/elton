#!/bin/bash
set -euvx

apt_install() {
    until apt install -y "$@"; do
        sleep 1
    done
}


apt_install build-essential automake libattr1-dev

[ -d ltp-full-20190517 ] || curl -SsL https://github.com/linux-test-project/ltp/releases/download/20190517/ltp-full-20190517.tar.xz |xz -d |tar xv
ln -s ltp-full-20190517 ltp
pushd ltp
make autotools
./configure
make -j`nproc` all
make install
popd
