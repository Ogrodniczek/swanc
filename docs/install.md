> New to Swanc? Please start [here](/docs/tutorial.md).

# Installation Guide

## Using YAML
Swanc can be installed using YAML files includes in the [/hack/deploy](/hack/deploy) folder.

```console
# Install without RBAC roles
$ curl https://raw.githubusercontent.com/appscode/swanc/5.0.0-alpha.0/hack/deploy/without-rbac.yaml \
  | kubectl apply -f -


# Install with RBAC roles
$ curl https://raw.githubusercontent.com/appscode/swanc/5.0.0-alpha.0/hack/deploy/with-rbac.yaml \
  | kubectl apply -f -
```

## Verify installation
To check if Swanc operator pods have started, run the following command:
```console
$ kubectl get pods --all-namespaces -l app=swanc --watch
```

Once the operator pods are running, you can cancel the above command by typing `Ctrl+C`.
