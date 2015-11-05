#!/bin/sh
set -e

export GOMAXPROCS=4

OPTION="$1"
if [ "${OPTION}" = "master" ]; then
  bin/elton master -f examples/master.tml
elif [ "${OPTION}" = "backup" ]; then
  bin/elton slave -f examples/slave.tml --backup
else
  bin/elton slave -f examples/slave.tml
fi
