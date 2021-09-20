# Code Style

- [golangci-lint](https://golangci-lint.run)

# How to develop in local

1. Run MySQL with Docker.
    ```
    docker run -d -p 3306:3306 -e MYSQL_ROOT_PASSWORD=password --rm mysql:5.7
    ```
1. `kubectl` is configured to a Kubernetes cluster.
1. Install CRD and run the operator locally.
    ```
    make install run
    ```
1. Apply sample resources.
    ```
    kubectl apply -k config/samples
    ```
    1. MySQL user is created in MySQL container.

        ```
        docker exec -it $(docker ps | grep mysql | head -1 |awk '{print $1}') mysql -uroot -ppassword
        ```

        <details><summary>details</summary>

        ```sql
        mysql> select User, Host from mysql.user;
        +---------------+-----------+
        | User          | Host      |
        +---------------+-----------+
        | nakamasato    | %         |
        | root          | %         |
        | mysql.session | localhost |
        | mysql.sys     | localhost |
        | root          | localhost |
        +---------------+-----------+
        5 rows in set (0.00 sec)
        ```

        </details>

    1. `Secret` `mysql-mysql-sample-nakamasato` is created for MySQL user.
        ```
        kubectl get secret mysql-mysql-sample-nakamasato -o yaml
        ```
    1. You can connect to MySQL with the generated user.
        ```
        docker exec -it $(docker ps | grep mysql | head -1 |awk '{print $1}') mysql -unakamasato -p$(kubectl get secret mysql-mysql-sample-nakamasato -o jsonpath='{.data.password}' | base64 --decode)
        ```

1. Delete `MySQLUser`
    ```
    kubectl delete -k config/samples
    ```
    1. Secret is deleted.
        ```
        kubectl get secret
        ```
    1. MySQL user is deleted.
        ```
        docker exec -it $(docker ps | grep mysql | head -1 |awk '{print $1}') mysql -uroot -ppassword
        ```

        <details><summary>details</summary>

        ```sql
        mysql> select User, Host from mysql.user;
        +---------------+-----------+
        | User          | Host      |
        +---------------+-----------+
        | root          | %         |
        | mysql.session | localhost |
        | mysql.sys     | localhost |
        | root          | localhost |
        +---------------+-----------+
        5 rows in set (0.00 sec)
        ```

        </details>

1. Clean up

```
make uninstall
```

# Scorecard

Statically validate your operator bundle.

```
operator-sdk scorecard ./bundle --wait-time 60s
```

Default tests:
- basic-check-spec-test
- olm-bundle-validation-test
- olm-crds-have-validation-test
- olm-crds-have-resources-test
- olm-spec-descriptors-test
- olm-status-descriptors-test

More:
- [custom test example](https://github.com/operator-framework/operator-sdk/blob/09c3aa14625965af9f22f513cd5c891471dbded2/images/custom-scorecard-tests/main.go)
- [Writing Custom Scorecard Tests](https://sdk.operatorframework.io/docs/testing-operators/scorecard/custom-tests/)

# OLM (ToDo)
# Reference
## Finalizers
- https://book.kubebuilder.io/reference/using-finalizers.html
- https://zdyxry.github.io/2019/09/13/Kubernetes-%E5%AE%9E%E6%88%98-Operator-Finalizers/
- https://sdk.operatorframework.io/docs/building-operators/golang/advanced-topics/

## Testing
- https://blog.bullgare.com/2021/02/mocking-for-unit-tests-and-e2e-tests-in-golang/

## Managing errors:
https://cloud.redhat.com/blog/kubernetes-operators-best-practices
1. Return the error in the status of the object.
1. Generate an event describing the error.
