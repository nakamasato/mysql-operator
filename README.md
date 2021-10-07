# MySQL Operator

This is a go-based Kubernetes operator built with [operator-sdk](https://sdk.operatorframework.io/docs/building-operators/golang/), which manages MySQL databases, schema, users, permissions for existing MySQL clusters. This operator DOES NOT manage MySQL cluster like other MySQL operator.

## Versions

- Go: 1.16
## Components

- `MySQL`: MySQL cluster
- `MySQLUser`: MySQL user (`mysqlName` and `host`)
- `MySQLDB`: MySQL database including schema repository (ToDo)

## Getting Started

1. Install `mysql-operator-controller-manager` in `mysql-operator-system` namespace.

    ```
    make deploy IMG="ghcr.io/nakamasato/mysql-operator"
    ```
1. Apply custom resources (`MySQL`, `MySQLUser`).

    Example: apply MySQL Deployment and Service and `MySQL` `MySQLUser`:
    ```
    kubectl apply -k config/samples-on-k8s
    ```

1. Delete custom resources (`MySQL`, `MySQLUser`).
    Example:
    ```
    kubectl delete -k config/samples-on-k8s
    ```

    NOTICE: custom resources might get stuck if MySQL is deleted before (to be improved). → Remove finalizers to forcifully delete the stuck objects
    `kubectl patch mysqluser <resource_name> -p '{"metadata":{"finalizers": []}}' --type=merge` or `kubectl patch mysql <resource_name> -p '{"metadata":{"finalizers": []}}' --type=merge`

1. Uninstall
    ```
    make undeploy
    ```
## Contributing

[CONTRIBUTING](CONTRIBUTING.md)
