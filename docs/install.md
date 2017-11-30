> New to Swanc? Please start [here](/docs/tutorial.md).

# Installation Guide

## Using YAML
Swanc can be installed using YAML files includes in the [/hack/deploy](/hack/deploy) folder.

First, create a Secret called `swanc` to hold pre-shared key for securing communication:

```console
curl https://raw.githubusercontent.com/pharmer/swanc/5.0.0-alpha.2/hack/deploy/create-psk.sh | bash
```

Now, deploy `swanc` controller running the appropriate command depending on whether the cluster uses RBAC or not:

```console
# Install without RBAC roles
$ kubectl apply -f https://raw.githubusercontent.com/pharmer/swanc/5.0.0-alpha.2/hack/deploy/without-rbac.yaml


# Install with RBAC roles
$ kubectl apply -f https://raw.githubusercontent.com/pharmer/swanc/5.0.0-alpha.2/hack/deploy/with-rbac.yaml
```

## Verify installation
To check if Swanc operator pods have started, run the following command:
```console
$ kubectl get pods --all-namespaces -l app=swanc --watch
```

Once the operator pods are running, you can cancel the above command by typing `Ctrl+C`.
