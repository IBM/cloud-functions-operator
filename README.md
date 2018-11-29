This project provides a collection of Kubernetes operators for managing [Apache OpenWhisk](https://openwhisk.apache.org/) resources namely actions, packages, rules and triggers.

# Quick start

## Prerequisites

- A cluster running Kubernetes 1.11+ 
- `kubectl` installed and configured.
- [`kustomize`](https://github.com/kubernetes-sigs/kustomize) installed.

## Installing the operators

1. Install the CRDs using `kubectl`:

```sh
$ kubectl apply -f https://raw.github.com/IBM/openwhisk-operator/master/config/crds
```

2. Then install the operator:

```sh
$ kustomize build https://raw.github.com/IBM/openwhisk-operator/master/config/default | kubectl apply -f -
```

# Learn more

- [reference documentation](https://ibm.github.io/openwhisk-operator/)
- [contributions](./CONTRIBUTING.md)
