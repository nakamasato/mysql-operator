# Overview

This is a go-based Kubernetes operator built with [operator-sdk](https://sdk.operatorframework.io/docs/building-operators/golang/), which manages MySQL databases, schema, users, permissions in existing MySQL servers. This operator DOES NOT manage MySQL cluster like other MySQL operators such as [vitess](https://github.com/vitessio/vitess), [mysql/mysql-operator](https://github.com/mysql/mysql-operator).

## Motivation

Reduce human operations:

1. **User management**: When creating a MySQL user for an application running on Kubernetes, it's necessary to create a MySQL user and create a Secret manually or with a script, which can be replaced with a Kubernetes operator. The initial idea is from KafkaUser and KafkaTopic in [Strimzi Kafka Operator](https://github.com/strimzi/strimzi-kafka-operator). With a custom resource for MySQL user, we can manage MySQL users with Kubernetes manifest files as a part of dependent application.
    Benefits from such a custom resource and operator:
    1. Kubernetes manifest files for an application and its dependent resources (including MySQL user) can be managed together with Kustomize or Helm chart, with which we can easily duplicate whole environment.
    1. There's no chance to require someone to check the raw password as it's stored directly to Secret by the operator, and read by the dependent application from the Secret.
1. **Database migration**: Reduce manual operations but keep changelog. When any schema migration or database operation is required, we needed a human operation, which has potential risk of human errors that should be avoided. With a Kubernetes operator, we can execute each database operation in the standard way with traceable changlog.

## Custom Resources
* `MySQL` - MySQL cluster or server.
* `MySQLUser` - MySQL user.
* `MySQLDB` - MySQL database.

## Contents

- Developer Guide
    - [Reconciliation](developer-guide/reconciliation.md)
    - [API Resource](developer-guide/api-resources.md)
    - [Debug](developer-guide/debug.md)
    - [Helm](developer-guide/helm.md)
    - [Testing](developer-guide/testing.md)
    - [Tools](developer-guide/tools.md)
- Usage
    - [Run on GKE and manage Cloud SQL (MySQL) with GCP SecretManager](usage/gcp-secretmanager.md)
    - [Schema Migration](usage/schema-migration.md)
    - [Install with Helm](usage/install-with-helm.md)


## Getting Started

1. Install CRD
    ```
    kubectl apply -k https://github.com/nakamasato/mysql-operator/config/install
    ```
1. (Optional) prepare MySQL.
    ```
    kubectl apply -k https://github.com/nakamasato/mysql-operator/config/mysql
    ```

1. Configure MySQL credentials for the operator using the custom resources `MySQL`.

    `mysql.yaml` credentials to connect to the MySQL: **This user is used to manage MySQL users and databases, which is ususally an admin user.**

    ```yaml
    apiVersion: mysql.nakamasato.com/v1alpha1
    kind: MySQL
    metadata:
      name: mysql-sample
    spec:
      host: mysql.default # need to include namespace if you use Kubernetes Service as an endpoint.
      adminUser:
        name: root
        type: raw
      adminPassword:
        name: password
        type: raw
    ```

    If you installed mysql sample with the command above, the password for the root user is `password`. You can apply `MySQL` with the following command.

    ```
    kubectl apply -f https://raw.githubusercontent.com/nakamasato/mysql-operator/main/config/samples-on-k8s/mysql_v1alpha1_mysql.yaml
    ```

    You can check the `MySQL` object and status:

    ```
    kubectl get mysql
    NAME           HOST            ADMINUSER   CONNECTED   USERCOUNT   DBCOUNT   REASON
    mysql-sample   mysql.default   root        true        0           0         Ping succeded and updated MySQLClients
    ```

1. Create a new MySQL user with custom resource `MySQLUser`.

    `mysqluser.yaml`: MySQL user

    ```yaml
    apiVersion: mysql.nakamasato.com/v1alpha1
    kind: MySQLUser
    metadata:
      name: sample-user
    spec:
      mysqlName: mysql-sample
      host: '%'
    ```

    1. Create a new MySQL user `sample-user`

        ```
        kubectl apply -f https://raw.githubusercontent.com/nakamasato/mysql-operator/main/config/samples-on-k8s/mysql_v1alpha1_mysqluser.yaml
        ```

    1. You can check the status of `MySQLUser` object

        ```
        kubectl get mysqluser
        NAME         MYSQLUSER   SECRET   PHASE   REASON
        sample-user  true        true     Ready   Both secret and mysql user are successfully created.
        ```

    1. You can also confirm the Secret for the new MySQL user is created.

        ```
        kubectl get secret
        NAME                            TYPE     DATA   AGE
        mysql-mysql-sample-sample-user  Opaque   1      4m3s
        ```

    1. Connect to MySQL with the newly created user

        ```
        kubectl exec -it $(kubectl get po | grep mysql | head -1 | awk '{print $1}') -- mysql -usample-user -p$(kubectl get secret mysql-mysql-sample-sample-user -o jsonpath='{.data.password}' | base64 --decode)
        ```

1. Create a new MySQL database with custom resource `MySQLDB`.

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
    kubectl apply -f https://raw.githubusercontent.com/nakamasato/mysql-operator/main/config/samples-on-k8s/mysql_v1alpha1_mysqldb.yaml
    ```

    ```
    kubectl get mysqldb
    NAME        PHASE   REASON                          SCHEMAMIGRATION
    sample-db   Ready   Database successfully created   {"dirty":false,"version":0}
    ```

1. Grant all priviledges of the created db (`sample_db`) to the create user (`sample-user`) (TODO: Currently there's no way to manage user permissions with operator.)

    ```
    kubectl exec -it $(kubectl get po | grep mysql | head -1 | awk '{print $1}') -- mysql -uroot -ppassword
    ```

    ```sql
    GRANT ALL PRIVILEGES ON sample_db.* TO 'sample-user'@'%';
    ```

    Now the created user got the permission to use `sample_db`.

    ```
    ubectl exec -it $(kubectl get po | grep mysql | head -1 | awk '{print $1}') -- mysql -usample-user -p$(kubectl get secret mysql-mysql-sample-sample-user -o jsonpath='{.data.password}' | base64 --decode)
    ```

    ```
    mysql> show databases;
    +--------------------+
    | Database           |
    +--------------------+
    | information_schema |
    | performance_schema |
    | sample_db          |
    +--------------------+
    3 rows in set (0.00 sec)
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
