# Overview
## Resources
* `MySQL` - MySQL cluster or server.
* `MySQLUser` - MySQL user.

## Contents

- Developer Guide
    - [Reconciliation](developer-guide/reconciliation.md)
    - [Debug](developer-guide/debug.md)

## Getting Started

### 1. Install `mysql-operator-controller-manager` in `mysql-operator-system` namespace.

```
make deploy IMG="ghcr.io/nakamasato/mysql-operator"
```

### 2. Apply custom resources (`MySQL`, `MySQLUser`).

Example: apply MySQL Deployment and Service and `MySQL` `MySQLUser`:

```
kubectl apply -k config/samples-on-k8s
```

### 3. Delete custom resources (`MySQL`, `MySQLUser`).
Example:

```
kubectl delete -k config/samples-on-k8s
```

NOTICE: custom resources might get stuck if MySQL is deleted before (to be improved). → Remove finalizers to forcifully delete the stuck objects
`kubectl patch mysqluser <resource_name> -p '{"metadata":{"finalizers": []}}' --type=merge` or `kubectl patch mysql <resource_name> -p '{"metadata":{"finalizers": []}}' --type=merge`

### 4. Uninstall

```
make undeploy
```
