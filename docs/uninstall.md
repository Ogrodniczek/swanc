# Uninstall Swanc
Please follow the steps below to uninstall Swanc:

1. Delete the various objects created for Swanc operator.
```console
$ ./hack/deploy/uninstall.sh
+ kubectl delete deployment -l app=swanc -n kube-system
deployment "swanc" deleted
+ kubectl delete service -l app=swanc -n kube-system
service "swanc" deleted
+ kubectl delete serviceaccount -l app=swanc -n kube-system
No resources found
+ kubectl delete clusterrolebindings -l app=swanc -n kube-system
No resources found
+ kubectl delete clusterrole -l app=swanc -n kube-system
No resources found
```

2. Now, wait several seconds for Swanc to stop running. To confirm that Swanc operator pod(s) have stopped running, run:
```console
$ kubectl get pods --all-namespaces -l app=swanc
```
