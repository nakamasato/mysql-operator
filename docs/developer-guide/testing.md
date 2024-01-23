# Testing

## 1. Controller test

Tools:

- [envtest](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/envtest)

## 2. e2e test

Currently, there are two ways for e2e tests.

1. [kuttl](https://kuttl.dev/) (will be deleted)
1. [kind](https://kind.sigs.k8s.io/) + [skaffold](https://skaffold.dev/) + [Ginkgo](https://onsi.github.io/ginkgo/) + [Gomage](https://onsi.github.io/gomega/)

### 2.1. e2e with kind + skaffold + ginkgo + gomega

Prerequisite:
- [kind](https://kind.sigs.k8s.io/): local Kubernetes cluster
- [Skaffold](https://skaffold.dev/): workflow for building, pushing and deploying your application
- [ginkgo](https://onsi.github.io/ginkgo/)
- [gomega](https://onsi.github.io/gomega/)

Test steps:
1. BeforeSuite:
    1. Prepare `kind` cluster.
        1. Create `kind` cluster if not exist. Otherwise, recreate `kind` cluster if `lazymode` is false.
    1. Set up `k8sClient`.
    1. Delete `MySQLUser` and `MySQL` resources.
    1. Execute `skaffold run`.
        1. Deploy CRD and mysql-operator.
        1. Deploy MySQL with `Deployment`.
    1. Check `mysql-operator` is running.
1. Run test cases.
1. AfterSuite:
    1. Execute `skaffold delete`.
    1. Clean up `kind` cluster.

Run:

```
make e2e-with-ginkgo
```

<details><summary>If we want to debug with running each step with commands</summary>

1. Create a `kind` cluster:
    ```bash
    kind create cluster --name mysql-operator-e2e --kubeconfig e2e/kubeconfig --config e2e/kind-config.yml --wait 30s
    ```
1. Delete `MySQLUser` resources if exists.
    1. Delete the object:
        ```bash
        kubectl delete mysqluser john --kubeconfig e2e/kubeconfig
        ```
    1. Remove the finalizer if stuck:
        ```bash
        kubectl patch mysqluser john -p '{"metadata":{"finalizers": []}}' --type=merge --kubeconfig e2e/kubeconfig
        ```
1. Delete `MySQL` resources if exists.
    1. Delete the object:
        ```bash
        kubectl delete mysql mysql-sample --kubeconfig e2e/kubeconfig
        ```
    1. Remove the finalizer if stuck:
        ```
        kubectl patch mysql mysql-sample -p '{"metadata":{"finalizers": []}}' --type=merge --kubeconfig e2e/kubeconfig
        ```
1. Deploy `CRD`, `mysql-operator`, and MySQL with `Deployment`:
    ```
    cd e2e && skaffold run --kubeconfig kubeconfig --tail
    ```

</details>

If you encounter an error `Build Failed. Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Check if docker is running.`

```
sudo ln -s "$HOME/.docker/run/docker.sock" /var/run/docker.sock
```

Ref: https://github.com/GoogleContainerTools/skaffold/issues/7985
