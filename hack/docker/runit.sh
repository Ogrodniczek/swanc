#!/bin/bash

export > /etc/envvars

[[ $DEBUG == true ]] && set -x

echo "Mounting ipsec.conf ..."
cmd="swanc run --init-only $@"
echo $cmd
$cmd
rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi

echo "Starting runit..."
exec /usr/sbin/runsvdir-start
