#!/bin/bash
NODES=${NODES:-}

for NODE in $NODES
do
  NAME=`echo $NODE | cut -d ':' -f1`
  HOST=`echo $NODE | cut -d ':' -f2`
  cat << EOF >> /etc/munin/munin.conf
[$NAME]
  address $HOST
  use_node_name yes

EOF
done

crond
/usr/sbin/munin-node > /dev/null 2>&1 &
su - munin -s /bin/bash -c '/usr/bin/munin-cron'
echo "Using the following munin nodes:"
echo $NODES
/usr/sbin/httpd -DFOREGROUND
