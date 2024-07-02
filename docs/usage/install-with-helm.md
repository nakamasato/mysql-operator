# Install with Helm

## 1. Preparation

```
helm repo add nakamasato https://nakamasato.github.io/helm-charts
helm repo update
```

## 2. Install

### 2.1. Install without GCP SecretManager

```
helm install mysql-operator nakamasato/mysql-operator
```

### 2.2. Install with GCP credentials json

1. Check the yaml to be applied with the `template` command

    ```
    helm template mysql-operator nakamasato/mysql-operator --set adminUserSecretType=gcp --set gcpProjectId=$PROJECT_ID
    ```

    Check point:
    - [ ] secret is mounted
    - [ ] env var `GOOGLE_APPLICATION_CREDENTIALS` is set
    - [ ] env var `PROJECT_ID` is set

1. Install

    ```
    helm install mysql-operator nakamasato/mysql-operator --set adminUserSecretType=gcp --set gcpProjectId=$PROJECT_ID --generate-name
    ```

### 2.3. Install without GCP credentials json (e.g. Run on GCP resource)

```
helm install mysql-operator ./charts/mysql-operator \
    --dry-run \
    --set adminUserSecretType=gcp \
    --set gcpServiceAccount=${SA_NAME}@${PROJECT}.iam.gserviceaccount.com \
    --set gcpProjectId=$PROJECT \
    --namespace mysql-operator
```

For more details, [GCP SecretManager](gcp-secretmanager.md)

## 3. Upgrade

When you want to modify helm release (start operator with new settings or arguments), you can upgrade an existing release.

1. Get target release
    ```
    helm list
    ```
1. Upgrade
    ```
    helm upgrade mysql-operator nakamasato/mysql-operator --set adminUserSecretType=gcp --set gcpProjectId=$PROJECT_ID
    ```

## 4. Uninstall

1. Check helm release to uninstall
    ```
    helm list
    NAME            NAMESPACE       REVISION        UPDATED                                 STATUS          CHART                      APP VERSION
    mysql-operator  default         2               2023-04-08 12:38:58.65552 +0900 JST     deployed        mysql-operator-0.1.0       v0.2.0
    ```
1. Uninstall
    ```
    helm uninstall mysql-operator
    ```
