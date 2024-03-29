# Versions

## operator-sdk

Originally created with [v1.10.1](https://github.com/operator-framework/operator-sdk/releases/tag/v1.10.1)

`Makefile` was updated with [v1.28.0](https://github.com/operator-framework/operator-sdk/releases/tag/v1.28.0)

Steps:

1. create temporary dir
1. create operator-sdk project
    ```
    operator-sdk init --domain nakamasato.com --repo github.com/nakamasato/mysql-operator
    ```
1. copy `Makefile` to this repo
1. Update a few points
    1. IMAGE_TAG_BASE
        ```
        IMAGE_TAG_BASE ?= nakamasato/mysql-operator
        ```
    1. IMG
        ```
        IMG ?= ghcr.io/nakamasato/mysql-operator
        ```
    1. test
        ```
        test: manifests generate fmt vet envtest ## Run tests.
                KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" $(GINKGO) -cover -coverprofile cover.out -covermode=atomic -sk
        ip-package=e2e ./...
        ```
    1. gingko
        ```
        GINKGO_VERSION ?= v2.9.2
        ```

        ```
        GINKGO = $(LOCALBIN)/ginkgo
        ginkgo:
            test -s $(LOCALBIN)/ginkgo && $(LOCALBIN)/ginkgo version | grep -q $(GINKGO_VERSION) || \
            GOBIN=$(LOCALBIN) go install github.com/onsi/ginkgo/v2/ginkgo@$(GINKGO_VERSION)
        ```
    1. helmify

        ```
        HELMIFY ?= $(LOCALBIN)/helmify

        .PHONY: helmify
        helmify: $(HELMIFY) ## Download helmify locally if necessary.
        $(HELMIFY): $(LOCALBIN)
        	test -s $(LOCALBIN)/helmify || GOBIN=$(LOCALBIN) go install github.com/arttor/helmify/cmd/helmify@latest

        helm: manifests kustomize helmify
        	$(KUSTOMIZE) build config/default | $(HELMIFY)
        ```

## kubebuilder

### Migration from go/v3 to go/v4

- https://book.kubebuilder.io/migration/manually_migration_guide_gov3_to_gov4
- https://github.com/kubernetes-sigs/kubebuilder/blob/master/testdata/project-v4/Makefile
- https://book.kubebuilder.io/migration/migration_guide_gov3_to_gov4


```
kubebuilder version
Version: main.version{KubeBuilderVersion:"3.11.0", KubernetesVendor:"1.27.1", GitCommit:"3a3d1d9573f5b8fe7252bf49cec6e67ba87c88e7", BuildDate:"2023-06-20T19:20:03Z", GoOs:"darwin", GoArch:"arm64"}
```

```
go mod init github.com/nakamasato/mysql-operator
```

```
kubebuilder init --domain nakamasato.com --plugins=go/v4
```

```
kubebuilder create api --group mysql --version v1alpha1 --kind MySQL --controller --api
kubebuilder create api --group mysql --version v1alpha1 --kind MySQLDB --controller --api
kubebuilder create api --group mysql --version v1alpha1 --kind MySQLUser --controller --api
```

Copy apis

Copy internal packages
```
cp -r ../mysql-operator/internal/metrics internal
cp -r ../mysql-operator/internal/mysql internal/mysql
cp -r ../mysql-operator/internal/secret internal/
cp -r ../mysql-operator/internal/utils internal/
```
