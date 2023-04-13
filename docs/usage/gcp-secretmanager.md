# Read credentials from GCP SecretManager

mysql-operator can get the credentials of the MySQL user (which is used to access to the target MySQL cluster) from [GCP SecretManager](https://cloud.google.com/secret-manager)

## Prepare GCP resources

1. Set var PROJECT_ID
    ```
    PROJECT_ID=<your_project_id>
    gcloud config set project $PROJECT_ID
    ```
1. Create Secret `mysql-password` with value `password`
    ```
    echo -n "password" | gcloud secrets create mysql-password --data-file=-
    ```
1. Create service account `mysql-operator`
    ```
    gcloud iam service-accounts create mysql-operator --display-name=mysql-operator
    ```
1. Grant permission to the service account
    ```
    sa_email=$(gcloud iam service-accounts describe mysql-operator@${PROJECT_ID}.iam.gserviceaccount.com --format='value(email)')
    gcloud secrets add-iam-policy-binding mysql-password --role=roles/secretmanager.secretAccessor --member=serviceAccount:${sa_email}
    ```
1. Generate service account key json.
    ```
    gcloud iam service-accounts keys create sa-private-key.json --iam-account=mysql-operator@${PROJECT_ID}.iam.gserviceaccount.com
    ```

## Create Secret for service account key

```
kubectl create secret generic gcp-sa-private-key --from-file=sa-private-key.json
```

## Prepara mysql-operator yaml

1. `containers[].args`: Add `"--cloud-secret-manager=gcp"`
1. `containers[]`: Add the following codes
    ```yaml
          volumeMounts:
            - name: gcp-sa-private-key
              mountPath: /var/secrets/google
          env:
            - name: GOOGLE_APPLICATION_CREDENTIALS
              value: /var/secrets/google/sa-private-key.json
    ```
1. `volumes`:
    ```yaml
    volumes:
      - name: gcp-sa-private-key
        secret:
          secretName: gcp-sa-private-key
    ```
1.


## Prepare mysql-operator yaml

1. Uncomment the following piece of codes in `config/default/kustomization.yaml`
    ```yaml
    # [GCP SecretManager] Mount GCP service account key as secret
    secretGenerator:
    - name: gcp-sa-private-key
      files:
      - sa-private-key.json
    ```

    ```yaml
    # [GCP SecretManager] Mount GCP service account key as secret
    - manager_gcp_sa_secret_patch.yaml
    ```

    <details><summary>config/default/kustomization.yaml</summary>

    ```yaml
    namespace: mysql-operator-system
    namePrefix: mysql-operator-

    bases:
    - ../crd
    - ../rbac
    - ../manager

    # [GCP SecretManager] Mount GCP service account key as secret
    secretGenerator:
    - name: gcp-sa-private-key
      files:
      - sa-private-key.json

    patchesStrategicMerge:
    # [GCP SecretManager] Mount GCP service account key as secret
    - manager_gcp_sa_secret_patch.yaml
    ```

    </details>

## Run

1. Run
    ```
    skaffold dev
    ```
1. Create custom resources

    Update `adminPassword` with `type: gcp` in `config/samples-with-k8s/mysql_v1alpha1_mysql.yaml`:

    ```yaml
    apiVersion: mysql.nakamasato.com/v1alpha1
    kind: MySQL
    metadata:
      name: mysql-sample
    spec:
      host: mysql.default # need to include namespace if you use Kubernetes Service as an endpoint.
      adminUser:
        name: root
        type: raw
      adminPassword:
        name: mysql-password # echo -n "password" | gcloud secrets create mysql-password --data-file=-
        type: gcp
    ```

    ```
    kubectl apply -k config/samples-wtih-k8s
    ```

1. Check

    ```
    kubectl get -k config/samples-on-k8s
    NAME                                      HOST            ADMINUSER   USERCOUNT
    mysql.mysql.nakamasato.com/mysql-sample   mysql.default   root        1

    NAME                                        MYSQLUSER   SECRET   PHASE   REASON
    mysqluser.mysql.nakamasato.com/nakamasato   true        true     Ready   Both secret and mysql user are successfully created.
    ```

For more details, read [Workload Identity](https://cloud.google.com/kubernetes-engine/docs/concepts/workload-identity)
