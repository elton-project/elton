#!/bin/sh
SERVER=root@192.168.189.55

ssh -t $SERVER dmesg --follow
