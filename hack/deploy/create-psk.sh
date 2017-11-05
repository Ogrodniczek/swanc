#!/usr/bin/env bash

VPN_PSK=$(dd if=/dev/urandom bs=128 count=1 2>/dev/null | base64 | tr -d "=+/" | dd bs=32 count=1 2>/dev/null)
DATA=" : PSK $VPN_PSK"
kubectl create secret generic -n kube-system swanc --from-literal=ipsec.secrets="$DATA"
