# Overview
## Custom Resources
* `MySQL` - MySQL cluster or server.
* `MySQLUser` - MySQL user.
* `MySQLDB` - MySQL database.

## Contents

- Developer Guide
    - [Reconciliation](developer-guide/reconciliation.md)
    - [API Resource](developer-guide/api-resources.md)
    - [Debug](developer-guide/debug.md)

## Getting Started

1. Install CRD
    ```
    kubectl apply -k https://github.com/nakamasato/mysql-operator/config/install
    ```
1. (Optional) prepare MySQL.
    ```
    kubectl apply -k https://github.com/nakamasato/mysql-operator/config/mysql
    ```
1. Apply custom resources (`MySQL`, `MySQLUser`, `MySQLDB`).

    `mysql.yaml` credentials to connect to the MySQL:

    ```yaml
    apiVersion: mysql.nakamasato.com/v1alpha1
    kind: MySQL
    metadata:
      name: mysql-sample
    spec:
      host: mysql.default # need to include namespace if you use Kubernetes Service as an endpoint.
      admin_user:
        name: root
        type: raw
      admin_password:
        name: password
        type: raw
    ```

    `mysqluser.yaml`: MySQL user

    ```yaml
    apiVersion: mysql.nakamasato.com/v1alpha1
    kind: MySQLUser
    metadata:
      name: nakamasato
    spec:
      mysqlName: mysql-sample
      host: '%'
    ```

    `mysqldb.yaml`: MySQL database

    ```yaml
    apiVersion: mysql.nakamasato.com/v1alpha1
    kind: MySQLDB
    metadata:
      name: sample-db # this is not a name for MySQL database but just a Kubernetes object name
    spec:
      dbName: sample_db # this is MySQL database name
      mysqlName: mysql-sample
    ```

    ```
    kubectl apply -k https://github.com/nakamasato/mysql-operator/config/samples-on-k8s
    ```
1. Check `MySQLUser` and `Secret` for the MySQL user

    ```
    kubectl get mysqluser
    NAME         PHASE   REASON
    nakamasato   Ready   Both secret and mysql user are successfully created.
    ```

    ```
    kubectl get secret
    NAME                            TYPE     DATA   AGE
    mysql-mysql-sample-nakamasato   Opaque   1      10s
    ```
1. Connect to MySQL with the secret
    ```
    kubectl exec -it $(kubectl get po | grep mysql | head -1 | awk '{print $1}') -- mysql -unakamasato -p$(kubectl get secret mysql-mysql-sample-nakamasato -o jsonpath='{.data.password}' | base64 --decode)
    ```
1. Delete custom resources (`MySQL`, `MySQLUser`, `MySQLDB`).
    Example:
    ```
    kubectl delete -k https://github.com/nakamasato/mysql-operator/config/samples-on-k8s
    ```

    <details><summary>NOTICE</summary>

    custom resources might get stuck if MySQL is deleted before (to be improved). → Remove finalizers to forcifully delete the stuck objects:
    ```
    kubectl patch mysqluser <resource_name> -p '{"metadata":{"finalizers": []}}' --type=merge
    ```
    ```
    kubectl patch mysql <resource_name> -p '{"metadata":{"finalizers": []}}' --type=merge
    ```

    ```
    kubectl patch mysqldb <resource_name> -p '{"metadata":{"finalizers": []}}' --type=merge
    ```

    </details>

1. (Optional) Delete MySQL
    ```
    kubectl delete -k https://github.com/nakamasato/mysql-operator/config/mysql
    ```
1. Uninstall `mysql-operator`
    ```
    kubectl delete -k https://github.com/nakamasato/mysql-operator/config/install
    ```
