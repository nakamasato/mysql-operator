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

# Test
## Scorecard

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

## kuttl

https://kuttl.dev/docs/

Prerequisite:
- Kubernetes Cluster
- kubectl

Version:

    ```
    kubectl-kuttl -v
    kubectl-kuttl version 0.11.1
    ```

    ```
    KUTTL Version: version.Info{GitVersion:"0.11.1", GitCommit:"25776a2", BuildDate:"2021-08-09T15:18:32Z", GoVersion:"go1.16.6", Compiler:"gc", Platform:"darwin/amd64"}
    ```

1. Install kuttl plugin

    ```
    kubectl krew install kuttl
    ```

1. Run kuttl test

    Run tests against the Kubernetes cluster with default kubeconfig:

    ```
    kubectl kuttl test ./bundle/tests/scorecard/kuttl
    ```
    <details>

    ```
    kubectl kuttl test ./bundle/tests/scorecard/kuttl

    2021/09/20 13:07:17 running without a 'kuttl-test.yaml' configuration
    2021/09/20 13:07:17 kutt-test config testdirs is overridden with args: [ ./bundle/tests/scorecard/kuttl ]
    === RUN   kuttl
        harness.go:457: starting setup
        harness.go:248: running tests using configured kubeconfig.
        harness.go:285: Successful connection to cluster at: https://kubernetes.docker.internal:6443
        harness.go:353: running tests
        harness.go:74: going to run test suite with timeout of 30 seconds for each step
        harness.go:365: testsuite: ./bundle/tests/scorecard/kuttl has 1 tests
    === RUN   kuttl/harness
    === RUN   kuttl/harness/with-valid-mysql
    === PAUSE kuttl/harness/with-valid-mysql
    === CONT  kuttl/harness/with-valid-mysql
        logger.go:42: 13:07:19 | with-valid-mysql | Creating namespace: kuttl-test-becoming-liger
        logger.go:42: 13:07:19 | with-valid-mysql/0-mysql-deployment | starting test step 0-mysql-deployment
        logger.go:42: 13:07:20 | with-valid-mysql/0-mysql-deployment | Deployment:kuttl-test-becoming-liger/mysql created
        logger.go:42: 13:07:20 | with-valid-mysql/0-mysql-deployment | Service:kuttl-test-becoming-liger/mysql created
        logger.go:42: 13:07:22 | with-valid-mysql/0-mysql-deployment | test step completed 0-mysql-deployment
        logger.go:42: 13:07:22 | with-valid-mysql | with-valid-mysql events from ns kuttl-test-becoming-liger:
        logger.go:42: 13:07:22 | with-valid-mysql | 2021-09-20 13:07:20 +0900 JST   Normal  Pod mysql-5fd4b796b6-tr7wx      Binding      Scheduled       Successfully assigned kuttl-test-becoming-liger/mysql-5fd4b796b6-tr7wx to docker-desktop        default-scheduler
        logger.go:42: 13:07:22 | with-valid-mysql | 2021-09-20 13:07:20 +0900 JST   Normal  ReplicaSet.apps mysql-5fd4b796b6    SuccessfulCreate Created pod: mysql-5fd4b796b6-tr7wx
        logger.go:42: 13:07:22 | with-valid-mysql | 2021-09-20 13:07:20 +0900 JST   Normal  Deployment.apps mysql           ScalingReplicaSet    Scaled up replica set mysql-5fd4b796b6 to 1
        logger.go:42: 13:07:22 | with-valid-mysql | 2021-09-20 13:07:21 +0900 JST   Normal  Pod mysql-5fd4b796b6-tr7wx.spec.containers{mysql}            Pulled  Container image "mysql:5.7" already present on machine
        logger.go:42: 13:07:22 | with-valid-mysql | 2021-09-20 13:07:21 +0900 JST   Normal  Pod mysql-5fd4b796b6-tr7wx.spec.containers{mysql}            Created Created container mysql
        logger.go:42: 13:07:22 | with-valid-mysql | 2021-09-20 13:07:21 +0900 JST   Normal  Pod mysql-5fd4b796b6-tr7wx.spec.containers{mysql}            Started Started container mysql
        logger.go:42: 13:07:22 | with-valid-mysql | Deleting namespace: kuttl-test-becoming-liger
    === CONT  kuttl
        harness.go:399: run tests finished
        harness.go:508: cleaning up
        harness.go:563: removing temp folder: ""
    --- PASS: kuttl (5.00s)
        --- PASS: kuttl/harness (0.00s)
            --- PASS: kuttl/harness/with-valid-mysql (2.85s)
    PASS
    ```

    </details>

    Run tests against `kind` cluster:

    ```
    kubectl kuttl test --start-kind=true ./bundle/tests/scorecard/kuttl
    ```

kuttl in scorecard: (currently just create MySQL Deployment and Service)

```
operator-sdk scorecard ./bundle --selector=suite=kuttlsuite --wait-time 60s
```

Internally, it runs as follows:

1. run kuttl in [entrypoint](https://github.com/operator-framework/operator-sdk/blob/master/images/scorecard-test-kuttl/entrypoint)
    ```shell
    kubectl-kuttl test ${KUTTL_PATH} \
        --config=${KUTTL_CONFIG} \
        --namespace=${SCORECARD_NAMESPACE} \
        --report=JSON --artifacts-dir=/tmp > /tmp/kuttl.stdout 2> /tmp/kuttl.stderr
    ```
1. [main.go](https://github.com/operator-framework/operator-sdk/blob/master/images/scorecard-test-kuttl/main.go) converts the kuttl result into scorecard result (`v1alpha3.TestStatus`)

    ```shell
    [21-09-20 23:52:01] [docker-desktop] masato-naka at mac in ~/repos/nakamasato/operator-sdk/images/scorecard-test-kuttl on update-writing-kuttl-scorecard-tests ✔
    ± go run ./main.go
    {
        "results": [
            {
                "name": "with-valid-mysql",
                "state": "pass"
            }
        ]
    }
    ```

## e2e with kuttl

Prerequisite
- Kubernetes cluster
- Run `docker build -t mysql-operator .` to build `mysql-operator:latest`


```
kubectl kuttl test
```

<details>

```
=== RUN   kuttl
    harness.go:457: starting setup
    harness.go:248: running tests using configured kubeconfig.
    harness.go:285: Successful connection to cluster at: https://kubernetes.docker.internal:6443
    logger.go:42: 17:51:09 |  | running command: [make deploy IMG=mysql-operator VERSION=latest]
    logger.go:42: 17:51:09 |  | /Users/masato-naka/repos/nakamasato/mysql-operator/bin/controller-gen "crd:trivialVersions=true,preserveUnknownFields=false" rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
    logger.go:42: 17:51:10 |  | cd config/manager && /Users/masato-naka/repos/nakamasato/mysql-operator/bin/kustomize edit set image controller=mysql-operator
    logger.go:42: 17:51:10 |  | /Users/masato-naka/repos/nakamasato/mysql-operator/bin/kustomize build config/default | kubectl apply -f -
    logger.go:42: 17:51:11 |  | namespace/mysql-operator-system unchanged
    logger.go:42: 17:51:11 |  | customresourcedefinition.apiextensions.k8s.io/mysqls.mysql.nakamasato.com configured
    logger.go:42: 17:51:11 |  | customresourcedefinition.apiextensions.k8s.io/mysqlusers.mysql.nakamasato.com configured
    logger.go:42: 17:51:11 |  | serviceaccount/mysql-operator-controller-manager unchanged
    logger.go:42: 17:51:11 |  | role.rbac.authorization.k8s.io/mysql-operator-leader-election-role unchanged
    logger.go:42: 17:51:11 |  | clusterrole.rbac.authorization.k8s.io/mysql-operator-manager-role configured
    logger.go:42: 17:51:11 |  | clusterrole.rbac.authorization.k8s.io/mysql-operator-metrics-reader unchanged
    logger.go:42: 17:51:11 |  | clusterrole.rbac.authorization.k8s.io/mysql-operator-proxy-role unchanged
    logger.go:42: 17:51:11 |  | rolebinding.rbac.authorization.k8s.io/mysql-operator-leader-election-rolebinding unchanged
    logger.go:42: 17:51:11 |  | clusterrolebinding.rbac.authorization.k8s.io/mysql-operator-manager-rolebinding unchanged
    logger.go:42: 17:51:11 |  | clusterrolebinding.rbac.authorization.k8s.io/mysql-operator-proxy-rolebinding unchanged
    logger.go:42: 17:51:11 |  | configmap/mysql-operator-manager-config unchanged
    logger.go:42: 17:51:11 |  | service/mysql-operator-controller-manager-metrics-service unchanged
    logger.go:42: 17:51:11 |  | deployment.apps/mysql-operator-controller-manager configured
    harness.go:353: running tests
    harness.go:74: going to run test suite with timeout of 120 seconds for each step
    harness.go:365: testsuite: tests/e2e/ has 1 tests
=== RUN   kuttl/harness
=== RUN   kuttl/harness/with-valid-mysql
=== PAUSE kuttl/harness/with-valid-mysql
=== CONT  kuttl/harness/with-valid-mysql
2021/09/21 17:51:11 object detected with no GVK Kind for path /Users/masato-naka/repos/nakamasato/mysql-operator/tests/e2e/with-valid-mysql/01-assert.yaml
2021/09/21 17:51:11 object detected with no GVK Kind for path /Users/masato-naka/repos/nakamasato/mysql-operator/tests/e2e/with-valid-mysql/01-run-mysql-operator.yaml
    logger.go:42: 17:51:11 | with-valid-mysql | Skipping creation of user-supplied namespace: default
    logger.go:42: 17:51:11 | with-valid-mysql/0-mysql-deployment | starting test step 0-mysql-deployment
    logger.go:42: 17:51:11 | with-valid-mysql/0-mysql-deployment | Deployment:default/mysql updated
    logger.go:42: 17:51:12 | with-valid-mysql/0-mysql-deployment | Service:default/mysql updated
    logger.go:42: 17:51:12 | with-valid-mysql/0-mysql-deployment | test step completed 0-mysql-deployment
    logger.go:42: 17:51:12 | with-valid-mysql/1-run-mysql-operator | starting test step 1-run-mysql-operator
    logger.go:42: 17:51:12 | with-valid-mysql/1-run-mysql-operator | test step completed 1-run-mysql-operator
    logger.go:42: 17:51:12 | with-valid-mysql/2-create-mysql-user | starting test step 2-create-mysql-user
    logger.go:42: 17:51:12 | with-valid-mysql/2-create-mysql-user | running command: [kubectl apply -k ../../../config/samples-on-k8s --namespace default]
    logger.go:42: 17:51:13 | with-valid-mysql/2-create-mysql-user | service/mysql configured
    logger.go:42: 17:51:13 | with-valid-mysql/2-create-mysql-user | deployment.apps/mysql configured
    logger.go:42: 17:51:13 | with-valid-mysql/2-create-mysql-user | mysql.mysql.nakamasato.com/mysql-sample unchanged
    logger.go:42: 17:51:13 | with-valid-mysql/2-create-mysql-user | mysqluser.mysql.nakamasato.com/nakamasato unchanged
    logger.go:42: 17:51:14 | with-valid-mysql/2-create-mysql-user | test step completed 2-create-mysql-user
    logger.go:42: 17:51:14 | with-valid-mysql | with-valid-mysql events from ns default:
    logger.go:42: 17:51:14 | with-valid-mysql | 2021-09-21 16:54:31 +0900 JST       Warning Node docker-desktop             EvictionThresholdMet        Attempting to reclaim ephemeral-storage
    logger.go:42: 17:51:14 | with-valid-mysql | 2021-09-21 16:54:59 +0900 JST       Normal  Node docker-desktop             NodeHasDiskPressure Node docker-desktop status is now: NodeHasDiskPressure
    logger.go:42: 17:51:14 | with-valid-mysql | 2021-09-21 16:59:59 +0900 JST       Normal  Node docker-desktop             NodeHasNoDiskPressure       Node docker-desktop status is now: NodeHasNoDiskPressure
    logger.go:42: 17:51:14 | with-valid-mysql | 2021-09-21 17:18:35 +0900 JST       Warning MySQLUser.mysql.nakamasato.com nakamasato  ProcessingError  dial tcp: lookup mysql on 10.96.0.10:53: no such host
    logger.go:42: 17:51:14 | with-valid-mysql | Skipping deletion of user-supplied namespace: default
=== CONT  kuttl
    harness.go:399: run tests finished
    harness.go:508: cleaning up
    harness.go:563: removing temp folder: ""
--- PASS: kuttl (7.26s)
    --- PASS: kuttl/harness (0.00s)
        --- PASS: kuttl/harness/with-valid-mysql (3.17s)
PASS
```

</details>

What the e2e tests:
1. Deploy controller (local image if you're running in local).
1. Deploy mysql `Deployment` and `Service`.
1. Create `MySQL` and `MySQLUser` by `kubectl apply -k ../../../config/samples-on-k8s`.
1. Check `Secret` `mysql-mysql-sample-nakamasato`.

# OLM (ToDo)
# Reference
## Finalizers
- https://book.kubebuilder.io/reference/using-finalizers.html
- https://zdyxry.github.io/2019/09/13/Kubernetes-%E5%AE%9E%E6%88%98-Operator-Finalizers/
- https://sdk.operatorframework.io/docs/building-operators/golang/advanced-topics/

## Testing
- https://blog.bullgare.com/2021/02/mocking-for-unit-tests-and-e2e-tests-in-golang/
- https://int128.hatenablog.com/entry/2020/02/05/114940

## Managing errors:
https://cloud.redhat.com/blog/kubernetes-operators-best-practices
1. Return the error in the status of the object.
1. Generate an event describing the error.
