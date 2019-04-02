#!/bin/sh
SERVER=root@192.168.189.190

ssh -t $SERVER dmesg --follow
