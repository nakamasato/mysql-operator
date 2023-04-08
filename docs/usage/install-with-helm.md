# Install with Helm

## 1. Preparation

```
HELM_PATH=mysql-operator-0.1.0.tgz
```

## 2. Install

### 2.1. Install without GCP SecretManager

```
helm install $HELM_PATH --generate-name
```

### 2.2. Install with GCP credentials json

1. Check the yaml to be applied with the `template` command

    ```
    helm template $HELM_PATH --set cloudSecretManagerType=gcp --set gcpProjectId=$PROJECT_ID
    ```

    Check point:
    - [ ] secret is mounted
    - [ ] env var `GOOGLE_APPLICATION_CREDENTIALS` is set
    - [ ] env var `PROJECT_ID` is set

1. Install

    ```
    helm install $HELM_PATH --set cloudSecretManagerType=gcp --set gcpProjectId=$PROJECT_ID --generate-name
    ```

### 2.3. Install without GCP credentials json (e.g. Run on GCP resource)

TODO

## 3. Upgrade

When you want to modify helm release (start operator with new settings or arguments), you can upgrade an existing release.

1. Get target release
    ```
    helm list
    ```
1. Upgrade
    ```
    helm upgrade <target release to upgrade> $HELM_PATH --set cloudSecretManagerType=gcp --set gcpProjectId=$PROJECT_ID
    ```

## 4. Uninstall

1. Check helm release to uninstall
    ```
    helm list
    NAME                            NAMESPACE       REVISION        UPDATED                                 STATUS          CHART                   APP VERSION
    mysql-operator-0-1680913123     default         2               2023-04-08 09:28:31.937086 +0900 JST    deployed        mysql-operator-0.1.0    v0.2.0
    ```
1. Uninstall
    ```
    helm uninstall mysql-operator-0-1680913123
    ```
