#!/bin/sh
set -euv
SERVER=root@192.168.189.190
DIR=samplefs-kmod

rsync -av --del ./ $SERVER:$DIR/
ssh -t $SERVER "cd $DIR && make"
