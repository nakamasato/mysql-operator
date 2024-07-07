# helm

## Create Helm chart

With [helmify](https://github.com/arttor/helmify), you can create a helm chart

1. Update Makefile
    ```
    HELMIFY ?= $(LOCALBIN)/helmify

    .PHONY: helmify
    helmify: $(HELMIFY) ## Download helmify locally if necessary.
    $(HELMIFY): $(LOCALBIN)
    	test -s $(LOCALBIN)/helmify || GOBIN=$(LOCALBIN) go install github.com/arttor/helmify/cmd/helmify@latest

    helm: manifests kustomize helmify
    	$(KUSTOMIZE) build config/install | $(HELMIFY)
    ```
1. Run
    ```
    make helm
    ```
1. Check generated files
    ```
    chart
    ├── Chart.yaml
    ├── templates
    │   ├── _helpers.tpl
    │   ├── deployment.yaml
    │   ├── leader-election-rbac.yaml
    │   ├── manager-config.yaml
    │   ├── manager-rbac.yaml
    │   ├── metrics-reader-rbac.yaml
    │   ├── metrics-service.yaml
    │   ├── mysql-crd.yaml
    │   ├── mysqldb-crd.yaml
    │   ├── mysqluser-crd.yaml
    │   └── proxy-rbac.yaml
    └── values.yaml

    1 directory, 13 files
    ```
1. Update name in `chart/Chart.yaml`
    ```yaml
    name: mysql-operator
    ```
1. Update `chart/templates/deployment.yaml` for your purpose
    What we do here is basically to enable to change `Deployment` from `Values`. (ref: [#199](https://github.com/nakamasato/mysql-operator/pull/199/commits/cc245343a9a24eee35425ef7d665c9d17996c7a8))
1. Package
    ```
    helm package chart --app-version v0.2.0
    ```
    The command will generate `mysql-operator-0.1.0.tgz`
## Publish package to Helm chart repo.

https://github.com/nakamasato/helm-charts is used for repo.
All we need to do is to update the chart source file under [charts/mysql-operator](https://github.com/nakamasato/helm-charts/tree/main/charts/mysql-operator) in the repo.

We use GitHub Actions to update the repo.

## Install mysql-operator with the Helm chart (from local source file)

1. Install mysql-operator with helm

    ```
    helm install mysql-operator-0.1.0.tgz --generate-name
    ```

    Optionally, you can add `--set adminUserSecretType=gcp --set gcpProjectId=$PROJECT_ID` to use GCP SecretManager to get AdminUser and/or AdminPassword.

    <details>

    ```
    NAME: mysql-operator-0-1680907162
    LAST DEPLOYED: Sat Apr  8 07:13:58 2023
    NAMESPACE: default
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None
    ```

    </details>

1. List

    ```
    helm list
    ```

    ```
    NAME                            NAMESPACE       REVISION        UPDATED                                 STATUS          CHART                   APP VERSION
    mysql-operator-0-1680907162     default         1               2023-04-08 07:39:22.416055 +0900 JST    deployed        mysql-operator-0.1.0    v0.2.0
    ```
1. Check operator is running
    ```
    kubectl get po
    NAME                                                             READY   STATUS    RESTARTS   AGE
    mysql-operator-0-1680907162-controller-manager-f9d855dc9-d4psm   0/1     Running   0          13s
    ```
1. (Optional) upgrade an existing release
    ```
    helm upgrade mysql-operator-0-1680913123 $HELM_PATH --set adminUserSecretType=gcp --set gcpProjectId=$PROJECT_ID
    ```
1. Uninstall
    ```
    helm uninstall mysql-operator-0-1680907162
    ```

## Usage

[Install with Helm](../usage/install-with-helm.md)

## Development Tips

1. Check resulting yaml file
    ```
    helm template chart
    ```
