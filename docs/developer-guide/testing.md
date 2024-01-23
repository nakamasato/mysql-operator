# Testing

## 1. Controller test

Tools:

- [envtest](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/envtest)

## 2. e2e test

Currently, there are two ways for e2e tests.

1. [kuttl](https://kuttl.dev/) (will be deleted)
1. [kind](https://kind.sigs.k8s.io/) + [skaffold](https://skaffold.dev/) + [Ginkgo](https://onsi.github.io/ginkgo/) + [Gomage](https://onsi.github.io/gomega/)

### 2.1. e2e with kind + skaffold + ginkgo + gomega

#### 2.1.1 Prerequisite**

- [kind](https://kind.sigs.k8s.io/): local Kubernetes cluster
- [Skaffold](https://skaffold.dev/): workflow for building, pushing and deploying your application
- [ginkgo](https://onsi.github.io/ginkgo/)
- [gomega](https://onsi.github.io/gomega/)

#### 2.1.2. Test steps**

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

#### 2.1.3. Run

```
make e2e-with-ginkgo
```

#### 2.1.4 Manually run all the steps

if we want to debug with running each step with commands

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

1. Run individual test

    ```
    bin/ginkgo --focus "Successfully Create MySQL database" ./e2e
    ```


### 2.2. e2e with kuttl

#### 2.2.1. Prerequisite

- [kind](https://kind.sigs.k8s.io/): local Kubernetes cluster
- [krew](https://krew.sigs.k8s.io/): kubectl plugin manager
    ```bash
    (
      set -x; cd "$(mktemp -d)" &&
      OS="$(uname | tr '[:upper:]' '[:lower:]')" &&
      ARCH="$(uname -m | sed -e 's/x86_64/amd64/' -e 's/\(arm\)\(64\)\?.*/\1\2/' -e 's/aarch64$/arm64/')" &&
      KREW="krew-${OS}_${ARCH}" &&
      curl -fsSLO "https://github.com/kubernetes-sigs/krew/releases/latest/download/${KREW}.tar.gz" &&
      tar zxvf "${KREW}.tar.gz" &&
      ./"${KREW}" install krew
    )
    ```
    Add `export PATH="${KREW_ROOT:-$HOME/.krew}/bin:$PATH"` to `~/.bashrc` or `~/.zshrc`
- [kuttl](https://kuttl.dev/): The KUbernetes Test TooL (kuttl)
    ```
    brew tap kudobuilder/tap
    brew install kuttl-cli
    ```

#### 2.2.2. Test scenario

1. MySQL `Deployment` and `Service`. -> Assert MySQL replica is 1.
1. Apply `config/samples-on-k8s`. -> `Secret` `mysql-mysql-sample-nakamasato` exists.


#### 2.2.3. e2e steps

1. Build a docker image `mysql-operator` with the latest codes.
1. Deploy controller (local image if you're running in local).
1. Deploy mysql `Deployment` and `Service`.
1. Create `MySQL` and `MySQLUser` by `kubectl apply -k ../../../config/samples-on-k8s`.
1. Check `Secret` `mysql-mysql-sample-nakamasato`.

#### 2.2.4. Run

```
make e2e-with-kuttl
```


<details>

```
docker build -t ghcr.io/nakamasato/mysql-operator:latest .
[+] Building 1.1s (17/17) FINISHED
 => [internal] load build definition from Dockerfile                                    0.0s
 => => transferring dockerfile: 37B                                                     0.0s
 => [internal] load .dockerignore                                                       0.0s
 => => transferring context: 35B                                                        0.0s
 => [internal] load metadata for gcr.io/distroless/static:nonroot                       0.9s
 => [internal] load metadata for docker.io/library/golang:1.16                          1.0s
 => [builder 1/9] FROM docker.io/library/golang:1.16@sha256:527d720ce3e2bc9b8900c9c165  0.0s
 => [internal] load build context                                                       0.0s
 => => transferring context: 643B                                                       0.0s
 => [stage-1 1/3] FROM gcr.io/distroless/static:nonroot@sha256:07869abb445859465749913  0.0s
 => CACHED [builder 2/9] WORKDIR /workspace                                             0.0s
 => CACHED [builder 3/9] COPY go.mod go.mod                                             0.0s
 => CACHED [builder 4/9] COPY go.sum go.sum                                             0.0s
 => CACHED [builder 5/9] RUN go mod download                                            0.0s
 => CACHED [builder 6/9] COPY main.go main.go                                           0.0s
 => CACHED [builder 7/9] COPY api/ api/                                                 0.0s
 => CACHED [builder 8/9] COPY controllers/ controllers/                                 0.0s
 => CACHED [builder 9/9] RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o mana  0.0s
 => CACHED [stage-1 2/3] COPY --from=builder /workspace/manager .                       0.0s
 => exporting to image                                                                  0.0s
 => => exporting layers                                                                 0.0s
 => => writing image sha256:abcaffd48dc037de5e3cce48890e720c6bdcf86f229be598aaaeb76cf8  0.0s
 => => naming to ghcr.io/nakamasato/mysql-operator:latest                               0.0s

Use 'docker scan' to run Snyk tests against images to find vulnerabilities and learn how to fix them
/Applications/Xcode.app/Contents/Developer/usr/bin/make kuttl
kubectl kuttl test
=== RUN   kuttl
    harness.go:457: starting setup
    harness.go:245: running tests with KIND.
    harness.go:174: temp folder created /var/folders/5g/vmdg2t1j2011ggd9p983ns6h0000gn/T/kuttl484640091
    harness.go:203: node mount point /var/lib/docker/volumes/kind-0/_data
    harness.go:156: Starting KIND cluster
    kind.go:67: Adding Containers to KIND...
    kind.go:76: Add image mysql-operator:latest to node kind-control-plane
    harness.go:285: Successful connection to cluster at: https://127.0.0.1:57498
    logger.go:42: 22:09:26 |  | running command: [make install deploy IMG=mysql-operator VERSION=latest]
    logger.go:42: 22:09:26 |  | /Users/masato-naka/repos/nakamasato/mysql-operator/bin/controller-gen "crd:trivialVersions=true,preserveUnknownFields=false" rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
    logger.go:42: 22:09:28 |  | /Users/masato-naka/repos/nakamasato/mysql-operator/bin/kustomize build config/crd | kubectl apply -f -
    logger.go:42: 22:09:31 |  | customresourcedefinition.apiextensions.k8s.io/mysqls.mysql.nakamasato.com created
    logger.go:42: 22:09:31 |  | customresourcedefinition.apiextensions.k8s.io/mysqlusers.mysql.nakamasato.com created
    logger.go:42: 22:09:31 |  | cd config/manager && /Users/masato-naka/repos/nakamasato/mysql-operator/bin/kustomize edit set image controller=mysql-operator
    logger.go:42: 22:09:31 |  | /Users/masato-naka/repos/nakamasato/mysql-operator/bin/kustomize build config/default | kubectl apply -f -
    logger.go:42: 22:09:32 |  | namespace/mysql-operator-system created
    logger.go:42: 22:09:32 |  | customresourcedefinition.apiextensions.k8s.io/mysqls.mysql.nakamasato.com configured
    logger.go:42: 22:09:32 |  | customresourcedefinition.apiextensions.k8s.io/mysqlusers.mysql.nakamasato.com configured
    logger.go:42: 22:09:32 |  | serviceaccount/mysql-operator-controller-manager created
    logger.go:42: 22:09:32 |  | role.rbac.authorization.k8s.io/mysql-operator-leader-election-role created
    logger.go:42: 22:09:32 |  | clusterrole.rbac.authorization.k8s.io/mysql-operator-manager-role created
    logger.go:42: 22:09:32 |  | clusterrole.rbac.authorization.k8s.io/mysql-operator-metrics-reader created
    logger.go:42: 22:09:32 |  | clusterrole.rbac.authorization.k8s.io/mysql-operator-proxy-role created
    logger.go:42: 22:09:32 |  | rolebinding.rbac.authorization.k8s.io/mysql-operator-leader-election-rolebinding created
    logger.go:42: 22:09:32 |  | clusterrolebinding.rbac.authorization.k8s.io/mysql-operator-manager-rolebinding created
    logger.go:42: 22:09:32 |  | clusterrolebinding.rbac.authorization.k8s.io/mysql-operator-proxy-rolebinding created
    logger.go:42: 22:09:32 |  | configmap/mysql-operator-manager-config created
    logger.go:42: 22:09:32 |  | service/mysql-operator-controller-manager-metrics-service created
    logger.go:42: 22:09:32 |  | deployment.apps/mysql-operator-controller-manager created
    harness.go:353: running tests
    harness.go:74: going to run test suite with timeout of 120 seconds for each step
    harness.go:365: testsuite: tests/e2e/ has 1 tests
=== RUN   kuttl/harness
=== RUN   kuttl/harness/with-valid-mysql
=== PAUSE kuttl/harness/with-valid-mysql
=== CONT  kuttl/harness/with-valid-mysql
    logger.go:42: 22:09:32 | with-valid-mysql | Skipping creation of user-supplied namespace: default
    logger.go:42: 22:09:32 | with-valid-mysql/0-mysql-deployment | starting test step 0-mysql-deployment
    logger.go:42: 22:09:32 | with-valid-mysql/0-mysql-deployment | Deployment:default/mysql created
    logger.go:42: 22:09:32 | with-valid-mysql/0-mysql-deployment | Service:default/mysql created
    logger.go:42: 22:09:43 | with-valid-mysql/0-mysql-deployment | test step completed 0-mysql-deployment
    logger.go:42: 22:09:43 | with-valid-mysql/1-create-mysql-user | starting test step 1-create-mysql-user
    logger.go:42: 22:09:43 | with-valid-mysql/1-create-mysql-user | running command: [kubectl apply -k ../../../config/samples-on-k8s --namespace default]
    logger.go:42: 22:09:46 | with-valid-mysql/1-create-mysql-user | Warning: resource services/mysql is missing the kubectl.kubernetes.io/last-applied-configuration annotation which is required by kubectl apply. kubectl apply should only be used on resources created declaratively by either kubectl create --save-config or kubectl apply. The missing annotation will be patched automatically.
    logger.go:42: 22:09:46 | with-valid-mysql/1-create-mysql-user | service/mysql configured
    logger.go:42: 22:09:46 | with-valid-mysql/1-create-mysql-user | Warning: resource deployments/mysql is missing the kubectl.kubernetes.io/last-applied-configuration annotation which is required by kubectl apply. kubectl apply should only be used on resources created declaratively by either kubectl create --save-config or kubectl apply. The missing annotation will be patched automatically.
    logger.go:42: 22:09:46 | with-valid-mysql/1-create-mysql-user | deployment.apps/mysql configured
    logger.go:42: 22:09:46 | with-valid-mysql/1-create-mysql-user | mysql.mysql.nakamasato.com/mysql-sample created
    logger.go:42: 22:09:46 | with-valid-mysql/1-create-mysql-user | mysqluser.mysql.nakamasato.com/nakamasato created
    logger.go:42: 22:09:52 | with-valid-mysql/1-create-mysql-user | test step completed 1-create-mysql-user
    logger.go:42: 22:09:52 | with-valid-mysql | with-valid-mysql events from ns default:
    logger.go:42: 22:09:52 | with-valid-mysql | 2021-10-02 22:09:16 +0900 JST   Normal  Node kind-control-plane              Starting        Starting kubelet.
    logger.go:42: 22:09:52 | with-valid-mysql | 2021-10-02 22:09:16 +0900 JST   Normal  Node kind-control-plane              NodeHasSufficientMemory Node kind-control-plane status is now: NodeHasSufficientMemory
    logger.go:42: 22:09:52 | with-valid-mysql | 2021-10-02 22:09:16 +0900 JST   Normal  Node kind-control-plane              NodeHasNoDiskPressure   Node kind-control-plane status is now: NodeHasNoDiskPressure
    logger.go:42: 22:09:52 | with-valid-mysql | 2021-10-02 22:09:16 +0900 JST   Normal  Node kind-control-plane              NodeHasSufficientPID    Node kind-control-plane status is now: NodeHasSufficientPID
    logger.go:42: 22:09:52 | with-valid-mysql | 2021-10-02 22:09:16 +0900 JST   Normal  Node kind-control-plane              NodeAllocatableEnforced Updated Node Allocatable limit across pods
    logger.go:42: 22:09:52 | with-valid-mysql | 2021-10-02 22:09:26 +0900 JST   Normal  Node kind-control-plane              RegisteredNode  Node kind-control-plane event: Registered Node kind-control-plane in Controller
    logger.go:42: 22:09:52 | with-valid-mysql | 2021-10-02 22:09:27 +0900 JST   Normal  Node kind-control-plane              Starting        Starting kube-proxy.
    logger.go:42: 22:09:52 | with-valid-mysql | 2021-10-02 22:09:32 +0900 JST   Warning Pod mysql-5fd4b796b6-jhx52           FailedScheduling        0/1 nodes are available: 1 node(s) had taint {node.kubernetes.io/not-ready: }, that the pod didn't tolerate.
    logger.go:42: 22:09:52 | with-valid-mysql | 2021-10-02 22:09:32 +0900 JST   Normal  ReplicaSet.apps mysql-5fd4b796b6             SuccessfulCreate        Created pod: mysql-5fd4b796b6-jhx52
    logger.go:42: 22:09:52 | with-valid-mysql | 2021-10-02 22:09:32 +0900 JST   Normal  Deployment.apps mysql                ScalingReplicaSet       Scaled up replica set mysql-5fd4b796b6 to 1
    logger.go:42: 22:09:52 | with-valid-mysql | 2021-10-02 22:09:36 +0900 JST   Normal  Node kind-control-plane              NodeReady       Node kind-control-plane status is now: NodeReady
    logger.go:42: 22:09:52 | with-valid-mysql | 2021-10-02 22:09:41 +0900 JST   Normal  Pod mysql-5fd4b796b6-jhx52           Scheduled       Successfully assigned default/mysql-5fd4b796b6-jhx52 to kind-control-plane
    logger.go:42: 22:09:52 | with-valid-mysql | 2021-10-02 22:09:42 +0900 JST   Normal  Pod mysql-5fd4b796b6-jhx52.spec.containers{mysql}            Pulled  Container image "mysql:5.7" already present on machine
    logger.go:42: 22:09:52 | with-valid-mysql | 2021-10-02 22:09:42 +0900 JST   Normal  Pod mysql-5fd4b796b6-jhx52.spec.containers{mysql}            Created Created container mysql
    logger.go:42: 22:09:52 | with-valid-mysql | 2021-10-02 22:09:42 +0900 JST   Normal  Pod mysql-5fd4b796b6-jhx52.spec.containers{mysql}            Started Started container mysql
    logger.go:42: 22:09:52 | with-valid-mysql | Skipping deletion of user-supplied namespace: default
=== CONT  kuttl
    harness.go:399: run tests finished
    harness.go:508: cleaning up
    harness.go:517: collecting cluster logs to kind-logs-1633180192
    harness.go:563: removing temp folder: "/var/folders/5g/vmdg2t1j2011ggd9p983ns6h0000gn/T/kuttl484640091"
    harness.go:569: tearing down kind cluster
--- PASS: kuttl (295.94s)
    --- PASS: kuttl/harness (0.00s)
        --- PASS: kuttl/harness/with-valid-mysql (20.45s)
PASS
```

</details>

## 3. Errors

Please read [debug](debug.md)
