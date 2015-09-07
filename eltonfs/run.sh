#!/bin/bash

/usr/sbin/munin-node > /dev/null 2>&1 &
/elton/bin/eltonfs elton1:12345 --upperdir=upper --lowerdir=lower --host=elton_eltonfs_1 mountpoint
