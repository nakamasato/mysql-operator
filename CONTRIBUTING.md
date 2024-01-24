# 1. Code Style

[golangci-lint](https://golangci-lint.run)

Install:
```
brew install golangci-lint
```
Run lint:
```
golangci-lint run ./...
```

# 2. Run mysql-operator

## 2.1. Local

![](docs/run-local.drawio.svg)

1. Start kind cluster
    ```
    kind create cluster
    ```

1. Run MySQL with Docker.
    ```
    docker run -d -p 3306:3306 -e MYSQL_ROOT_PASSWORD=password --rm mysql:8
    ```
1. Install CRD and run the operator locally.
    ```
    make install run
    ```
1. Apply sample resources.
    ```
    kubectl apply -k config/samples
    ```
1. Check Custom resources

    ```
    kubectl get -k config/samples
    NAME                                      HOST        ADMINUSER   CONNECTED   USERCOUNT   DBCOUNT   REASON
    mysql.mysql.nakamasato.com/mysql-sample   localhost   root        true        1           1         Ping succeded and updated MySQLClients

    NAME                                     PHASE   REASON                          SCHEMAMIGRATION
    mysqldb.mysql.nakamasato.com/sample-db   Ready   Database successfully created   {"dirty":false,"version":1}

    NAME                                         MYSQLUSER   SECRET   PHASE   REASON
    mysqluser.mysql.nakamasato.com/sample-user   true        true     Ready   Both secret and mysql user are successfully created.
    ```

1. Confirm MySQL user is created in MySQL container.

    ```
    docker exec -it $(docker ps | grep mysql | head -1 |awk '{print $1}') mysql -uroot -ppassword -e "select User, Host, password_last_changed, password_expired, password_lifetime from mysql.user where User = 'sample-user';"
    ```

1. `Secret` `mysql-mysql-sample-sample-user` is created for the MySQL user.
    ```
    kubectl get secret mysql-mysql-sample-sample-user -o jsonpath='{.data.password}'
    ```
1. Confirm you can connect to MySQL with the generated user.
    ```
    docker exec -it $(docker ps | grep mysql | head -1 |awk '{print $1}') mysql -usample-user -p$(kubectl get secret mysql-mysql-sample-sample-user -o jsonpath='{.data.password}' | base64 --decode)
    ```

1. Delete all the resources.
    ```
    kubectl delete -k config/samples
    ```

    <details>

    ```
    1.6780545572555468e+09  INFO    [FetchMySQL] Not found  {"controller": "mysql", "controllerGroup": "mysql.nakamasato.com", "controllerKind": "MySQL", "mySQL": {"name":"mysql-sample","namespace":"default"}, "namespace": "default", "name": "mysql-sample", "reconcileID": "0b6db5c6-8b3b-43ce-b903-a4959d55064e", "mysql.Name": "", "mysql.Namespace": ""}
    1.678054557255548e+09   INFO    [FetchMySQLUser] Found. {"controller": "mysqluser", "controllerGroup": "mysql.nakamasato.com", "controllerKind": "MySQLUser", "mySQLUser": {"name":"nakamasato","namespace":"default"}, "namespace": "default", "name": "nakamasato", "reconcileID": "78d4a7cf-5be0-4d47-82c0-38c7fdcf675b", "name": "nakamasato", "mysqlUser.Namespace": "default"}
    1.678054557255587e+09   ERROR   [FetchMySQL] Failed     {"controller": "mysqluser", "controllerGroup": "mysql.nakamasato.com", "controllerKind": "MySQLUser", "mySQLUser": {"name":"nakamasato","namespace":"default"}, "namespace": "default", "name": "nakamasato", "reconcileID": "78d4a7cf-5be0-4d47-82c0-38c7fdcf675b", "error": "MySQL.mysql.nakamasato.com \"mysql-sample\" not found"}
    sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).Reconcile
            /Users/m.naka/go/pkg/mod/sigs.k8s.io/controller-runtime@v0.12.3/pkg/internal/controller/controller.go:121
    sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).reconcileHandler
            /Users/m.naka/go/pkg/mod/sigs.k8s.io/controller-runtime@v0.12.3/pkg/internal/controller/controller.go:320
    sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).processNextWorkItem
            /Users/m.naka/go/pkg/mod/sigs.k8s.io/controller-runtime@v0.12.3/pkg/internal/controller/controller.go:273
    sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).Start.func2.2
            /Users/m.naka/go/pkg/mod/sigs.k8s.io/controller-runtime@v0.12.3/pkg/internal/controller/controller.go:234
    ```

    When getting stuck:

    ```
    kubectl patch mysqluser nakamasato -p '{"metadata":{"finalizers": []}}' --type=merge
    ```

    </details>

    1. Secret is deleted.
        ```
        kubectl get secret
        ```
    1. MySQL user is deleted.
        ```
        docker exec -it $(docker ps | grep mysql | head -1 |awk '{print $1}') mysql -uroot -ppassword -e 'select User, Host from mysql.user;'
        ```

        <details><summary>details</summary>

        ```sql
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
docker rm -f $(docker ps | grep mysql | head -1 |awk '{print $1}')
```

## 2.2. Local kubernetes

![](docs/run-local-kubernetes.drawio.svg)

1. Deploy controller with [skaffold](https://skaffold.dev/)

    ```bash
    skaffold dev
    ```

    You can check the operator installed in `mysql-operator-system` namespace.

    ```bash
    kubectl get po -n mysql-operator-system
    NAME                                                 READY   STATUS    RESTARTS   AGE
    mysql-operator-controller-manager-5bc69f545b-fcxst   2/2     Running   0          51s
    ```

1. Create Custom Resources (`MySQL` and `MySQLUser` resources).

    ```bash
    kubectl apply -k config/samples-on-k8s
    ```

1. Check Custom resources

    ```bash
    kubectl get -k config/samples-on-k8s
    ```

    ```
    NAME                                      HOST            ADMINUSER   CONNECTED   USERCOUNT   DBCOUNT   REASON
    mysql.mysql.nakamasato.com/mysql-sample   mysql.default   root        true        1           1         Ping succeded and updated MySQLClients

    NAME                                     PHASE   REASON                          SCHEMAMIGRATION
    mysqldb.mysql.nakamasato.com/sample-db   Ready   Database successfully created   {"dirty":false,"version":0}

    NAME                                         MYSQLUSER   SECRET   PHASE   REASON
    mysqluser.mysql.nakamasato.com/sample-user   true        true     Ready   Both secret and mysql user are successfully created.
    ```

1. Confirm MySQL user is created in MySQL container.

    ```bash
    kubectl exec -it $(kubectl get po | grep mysql | head -1 | awk '{print $1}') -- mysql -uroot -ppassword -e 'select User, Host from mysql.user where User = "sample-user";'
    ```
    ```
    mysql: [Warning] Using a password on the command line interface can be insecure.
    +-------------+------+
    | User        | Host |
    +-------------+------+
    | sample-user | %    |
    +-------------+------+
    ```

1. `Secret` `mysql-mysql-sample-sample-user` is created for the MySQL user.

    ```
    kubectl get secret mysql-mysql-sample-sample-user -o jsonpath='{.data.password}'
    ```

1. Clean up the Custom Resources (`MySQL` and `MySQLUser` resources).

    ```bash
    kubectl delete -k config/samples-on-k8s
    ```

    <details><summary>If getting stuck in deletion</summary>

    ```
    kubectl exec -it $(kubectl get po | grep mysql | head -1 | awk '{print $1}') -- mysql -uroot -ppassword -e 'delete from mysql.user where User = "sample-user";'
    kubectl patch mysqluser sample-user -p '{"metadata":{"finalizers": []}}' --type=merge
    ```

    </details>

1. Stop the `skaffold dev` by `ctrl-c` -> will clean up the controller, CRDs, and installed resources.


## 2.3. Local with GCP Secret Manager

![](docs/run-local-with-gcp-secretmanager.drawio.svg)

1. Setup gcloud
    ```bash
    PROJECT_ID=<project_id>
    gcloud auth login
    gcloud config set project $PROJECT_ID
    gcloud auth application-default login
    gcloud services enable secretmanager.googleapis.com # only first time
    ```
1. Create secret `mysql-password`
    ```
    echo -n "password" | gcloud secrets create mysql-password --data-file=-
    ```

    Check secrets:

    ```
    gcloud secrets list
    ```

1. Run MySQL with docker
    ```
    docker run -d -p 3306:3306 -e MYSQL_ROOT_PASSWORD=password --rm mysql:8
    ```
1. Install and run operator
    ```
    make install
    PRJECT_ID=$PROJECT_ID go run main.go --cloud-secret-manager gcp
    ```
1. Create custom resources
    ```
    kubectl apply -k config/samples-wtih-gcp-secretmanager
    ```
1. Check
    ```
    kubectl get -k config/samples-wtih-gcp-secretmanager
    NAME                                      HOST        ADMINUSER   CONNECTED   USERCOUNT   DBCOUNT   REASON
    mysql.mysql.nakamasato.com/mysql-sample   localhost   root        true        1           0         Ping succeded and updated MySQLClients

    NAME                                     PHASE   REASON
    mysqldb.mysql.nakamasato.com/sample-db   Ready   Database successfully created

    NAME                                        MYSQLUSER   SECRET   PHASE   REASON
    mysqluser.mysql.nakamasato.com/nakamasato   true        true     Ready   Both secret and mysql user are successfully created.
    ```
1. Clean up

    1. Remove CR:
        ```
        kubectl delete -k config/samples-wtih-gcp-secretmanager
        ```
    1. Stop controller `ctrl+c`
    1. Uninstall
        ```
        make uninstall
        ```
1. Clean up GCP
    ```
    gcloud secrets delete mysql-password
    gcloud auth revoke
    gcloud auth application-default revoke
    gcloud config unset project
    ```

# 3. Monitoring

1. Prepare Prometheus with Prometheus Operator
    1. Monitor with Prometheus
        ```
        kubectl create -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/master/bundle.yaml
        ```
    1. Create `Prometheus`

        ```
        kubectl create ns monitoring
        kubectl apply -k https://github.com/nakamasato/kubernetes-training/contents/prometheus-operator
        ```
    1. port forward
        ```
        kubectl port-forward -n monitoring svc/prometheus-operated 9090:9090
        ```
1. Run mysql-operator with ServiceMonitor (a CRD of PrometheusOperator)
    1. Update `config/default/kustomization.yaml` to uncomment
        ```yaml
        - ../prometheus
        ```
    1. Run operator
        ```
        skaffold dev
        ```
1. Check Prometheus on http://localhost:9090/targets
    ![](docs/prometheus.png)

    You can see the graph with custom metrics with `{__name__=~"mysqloperator_.*"}`
    ![](docs/prometheus-graph.png)
1. Clean up
    1. Remove CRD
        ```
        kubectl delete -f config/samples-on-k8s/mysql_v1alpha1_mysqluser.yaml
        kubectl delete -f config/samples-on-k8s/mysql_v1alpha1_mysql.yaml
        ```
    1. Stop skaffold dev
    1. Remove Proemetheus and Prometheus Operator
        ```
        kubectl delete -k https://github.com/nakamasato/kubernetes-training/contents/prometheus-operator
        kubectl delete -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/master/bundle.yaml
        ```

# 4. Test

## 4.1. Tools

- Ginkgo
- Gomega

## 4.2. Controller Test

Run all

```
make test
```

Set `KUBEBUILDER_ASSETS`

```
export KUBEBUILDER_ASSETS="$(bin/setup-envtest use 1.26.0 -p path)"
```

Fail fast

```
bin/ginkgo -skip-package=e2e --fail-fast ./...
```

Run individual test

```
bin/ginkgo -skip-package=e2e --focus "Should have finalizer" --fail-fast ./...
```

## 4.3. e2e

moved to [docs/developer-guide/testing.md](docs/developer-guide/testing.md)

# 6. Tips
## 6.1. Error: `Operation cannot be fulfilled on mysqlusers.mysql.nakamasato.com \"john\": StorageError: invalid object, Code: 4, Key: /registry/mysql.nakamasato.com/mysqlusers/default/john, ResourceVersion: 0, AdditionalErrorMsg: Precondition failed: UID in precondition: cd9c94d1-992a-457d-8fab-489b21ed02e9, UID in object meta:`

```
[manager] 1.6781410047933352e+09        ERROR   Reconciler error        {"controller": "mysqluser", "controllerGroup": "mysql.nakamasato.com", "controllerKind": "MySQLUser", "mySQLUser": {"name":"john","namespace":"default"}, "namespace": "default", "name": "john", "reconcileID": "85fc0e64-f2b9-413f-af44-46ff1daad7f7", "error": "Operation cannot be fulfilled on mysqlusers.mysql.nakamasato.com \"john\": StorageError: invalid object, Code: 4, Key: /registry/mysql.nakamasato.com/mysqlusers/default/john, ResourceVersion: 0, AdditionalErrorMsg: Precondition failed: UID in precondition: cd9c94d1-992a-457d-8fab-489b21ed02e9, UID in object meta: "}
```

UID in precondition and UID in object meta are different?

https://github.com/kubernetes-sigs/controller-runtime/issues/2209

## 6.2. Slow build

```
time CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go  161.57s user 24.90s system 283% cpu 1:05.76 total
```

# 7. Reference
## 7.1. Finalizers
- https://book.kubebuilder.io/reference/using-finalizers.html
- https://zdyxry.github.io/2019/09/13/Kubernetes-%E5%AE%9E%E6%88%98-Operator-Finalizers/
- https://sdk.operatorframework.io/docs/building-operators/golang/advanced-topics/

## 7.2. Testing
- https://blog.bullgare.com/2021/02/mocking-for-unit-tests-and-e2e-tests-in-golang/
- https://int128.hatenablog.com/entry/2020/02/05/114940

## 7.3. Managing errors:
https://cloud.redhat.com/blog/kubernetes-operators-best-practices
1. Return the error in the status of the object. https://pkg.go.dev/github.com/shivanshs9/operator-utils@v1.0.1#section-readme
1. Generate an event describing the error.

## 7.4. MySQL
- http://go-database-sql.org/index.html
