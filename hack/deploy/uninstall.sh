#!/bin/bash
set -x

kubectl delete deployment -l app=swanc -n kube-system

# Delete RBAC objects, if --rbac flag was used.
kubectl delete serviceaccount -l app=swanc -n kube-system
kubectl delete clusterrolebindings -l app=swanc -n kube-system
kubectl delete clusterrole -l app=swanc -n kube-system
