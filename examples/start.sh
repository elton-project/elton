#!/bin/sh
set -e

OPTION="$1"
if [ "${OPTION}" = "backup" ]; then
  make backupconfig
  bin/elton server -f config.tml --backup
else
  make config
  bin/elton server -f config.tml
fi
