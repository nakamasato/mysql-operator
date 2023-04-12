# Schema Migration

Schema migration feature uses https://github.com/golang-migrate/migrate but the supported feature in mysql-operator is limited.
Currently, only [GitHub source](https://github.com/golang-migrate/migrate/tree/master/source/github) is supported.

## Usage

1. Prepare schema files.
    Example:
    ```sql
    CREATE TABLE test_table (id int, name varchar(10));

    ```

    ```sql
    DROP TABLE test_table;
    ```

    https://github.com/nakamasato/mysql-operator/tree/96dc1eeaf00c8afb42f1c9b63859ff57c440e584/config/sample-migrations

1. Create MySQLDB yaml with `schemaMigrationFromGitHub`

    ```yaml
    apiVersion: mysql.nakamasato.com/v1alpha1
    kind: MySQLDB
    metadata:
      labels:
        app.kubernetes.io/name: mysqldb
        app.kubernetes.io/instance: mysqldb-sample
        app.kubernetes.io/part-of: mysql-operator
        app.kubernetes.io/managed-by: kustomize
        app.kubernetes.io/created-by: mysql-operator
      name: sample-db # this is not a name for MySQL database but just a Kubernetes object name
    spec:
      dbName: sample_db # this is MySQL database name
      mysqlName: mysql-sample
      schemaMigrationFromGitHub:
        owner: nakamasato
        repo: mysql-operator
        path: config/sample-migrations
        ref: 96dc1eeaf00c8afb42f1c9b63859ff57c440e584 # (optional) you can write branch, tag, sha
    ```

    This configuration will generate `"github://nakamasato/mysql-operator/config/sample-migrations#96dc1eeaf00c8afb42f1c9b63859ff57c440e584"` as `sourceUrl` for [source/github](https://github.com/golang-migrate/migrate/tree/master/source/github)

1. Run mysql & mysql-operator

    ```
    docker run -d -p 3306:3306 -e MYSQL_ROOT_PASSWORD=password --rm mysql:8
    ```

    ```bash
    make install run # to be updated with helm command
    ```
1. Create resources

    ```
    kubectl apply -k config/samples
    ```
1. Check `test_table` is created.

    ```
    docker exec -it $(docker ps | grep mysql | head -1 |awk '{print $1}') mysql -uroot -ppassword
    ```

    ```sql
    mysql> use sample_db;
    Reading table information for completion of table and column names
    You can turn off this feature to get a quicker startup with -A

    Database changed
    mysql> show tables;
    +---------------------+
    | Tables_in_sample_db |
    +---------------------+
    | schema_migrations   |
    | test_table          |
    +---------------------+
    2 rows in set (0.00 sec)
    ```
1. Clean up

    ```
    kubectl delete -k config/samples
    ```
