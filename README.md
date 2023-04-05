# MySQL Operator

[![codecov](https://codecov.io/gh/nakamasato/mysql-operator/branch/master/graph/badge.svg?token=AWM1SBTI19)](https://codecov.io/gh/nakamasato/mysql-operator)

This is a go-based Kubernetes operator built with [operator-sdk](https://sdk.operatorframework.io/docs/building-operators/golang/), which manages MySQL databases, schema, users, permissions for existing MySQL clusters. This operator DOES NOT manage MySQL cluster like other MySQL operator.

## Versions

- Go: 1.19
## Components

![](diagram.drawio.svg)

1. Custom Resource
    1. `MySQL`: MySQL cluster (holds credentials to connect to MySQL)
    1. `MySQLUser`: MySQL user (`mysqlName` and `host`)
    1. `MySQLDB`: MySQL database (`mysqlName` and `dbName`)
1. Reconciler
    1. `MySQLReconciler` is responsible for updating `MySQLClients` based on `MySQL` resource
    1. `MySQLUserReconciler` is responsible for managing `MySQLUser` using `MySQLClients`
    1. `MySQLDBReconciler` is responsible for managing `MySQLDB` using `MySQLClients`

## Getting Started

1. Install CRD
    ```
    kubectl apply -k https://github.com/nakamasato/mysql-operator/config/install
    ```
1. (Optional) prepare MySQL.
    ```
    kubectl apply -k https://github.com/nakamasato/mysql-operator/config/mysql
    ```
1. Apply custom resources (`MySQL`, `MySQLUser`, `MySQLDB`).

    `mysql.yaml` credentials to connect to the MySQL:

    ```yaml
    apiVersion: mysql.nakamasato.com/v1alpha1
    kind: MySQL
    metadata:
      name: mysql-sample
    spec:
      host: mysql.default # need to include namespace if you use Kubernetes Service as an endpoint.
      admin_user:
        name: root
        type: raw
      admin_password:
        name: password
        type: raw
    ```

    `mysqluser.yaml`: MySQL user

    ```yaml
    apiVersion: mysql.nakamasato.com/v1alpha1
    kind: MySQLUser
    metadata:
      name: nakamasato
    spec:
      mysqlName: mysql-sample
      host: '%'
    ```

    `mysqldb.yaml`: MySQL database

    ```yaml
    apiVersion: mysql.nakamasato.com/v1alpha1
    kind: MySQLDB
    metadata:
      name: sample-db # this is not a name for MySQL database but just a Kubernetes object name
    spec:
      dbName: sample_db # this is MySQL database name
      mysqlName: mysql-sample
    ```

    ```
    kubectl apply -k https://github.com/nakamasato/mysql-operator/config/samples-on-k8s
    ```
1. Check `MySQLUser` and `Secret` for the MySQL user

    ```
    kubectl get mysqluser
    NAME         PHASE   REASON
    nakamasato   Ready   Both secret and mysql user are successfully created.
    ```

    ```
    kubectl get secret
    NAME                            TYPE     DATA   AGE
    mysql-mysql-sample-nakamasato   Opaque   1      10s
    ```
1. Connect to MySQL with the secret
    ```
    kubectl exec -it $(kubectl get po | grep mysql | head -1 | awk '{print $1}') -- mysql -unakamasato -p$(kubectl get secret mysql-mysql-sample-nakamasato -o jsonpath='{.data.password}' | base64 --decode)
    ```
1. Delete custom resources (`MySQL`, `MySQLUser`, `MySQLDB`).
    Example:
    ```
    kubectl delete -k https://github.com/nakamasato/mysql-operator/config/samples-on-k8s
    ```

    <details><summary>NOTICE</summary>

    custom resources might get stuck if MySQL is deleted before (to be improved). → Remove finalizers to forcifully delete the stuck objects:
    ```
    kubectl patch mysqluser <resource_name> -p '{"metadata":{"finalizers": []}}' --type=merge
    ```
    ```
    kubectl patch mysql <resource_name> -p '{"metadata":{"finalizers": []}}' --type=merge
    ```

    ```
    kubectl patch mysqldb <resource_name> -p '{"metadata":{"finalizers": []}}' --type=merge
    ```

    </details>

1. (Optional) Delete MySQL
    ```
    kubectl delete -k https://github.com/nakamasato/mysql-operator/config/mysql
    ```
1. Uninstall `mysql-operator`
    ```
    kubectl delete -k https://github.com/nakamasato/mysql-operator/config/install
    ```

## With GCP Secret Manager

Instead of writing raw password in `MySQL.Spec.AdminPassword`, you can get the password for root user from an external secret manager (e.g. GCP) (ref: [Authenticate to Google Cloud using a service account](https://cloud.google.com/kubernetes-engine/docs/tutorials/authenticating-to-cloud-platform))
1. Set var PROJECT_ID
    ```
    PROJECT_ID=<your_project_id>
    gcloud config set project $PROJECT_ID
    ```
1. Create Secret for password
    ```
    echo -n "password" | gcloud secrets create mysql-password --data-file=-
    ```
1. Create service account
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
    gcloud iam service-accounts keys create config/default/sa-private-key.json --iam-account=mysql-operator@${PROJECT_ID}.iam.gserviceaccount.com
    ```
1. Update the following in `config/default/kustomization.yaml`
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
1. Run
    ```
    skaffold dev
    ```
1. Create custom resources

    Update `config/samples-with-k8s/mysql_v1alpha1_mysql.yaml` with `gcp_secret_name`:

    ```yaml
    apiVersion: mysql.nakamasato.com/v1alpha1
    kind: MySQL
    metadata:
      name: mysql-sample
    spec:
      host: mysql.default # need to include namespace if you use Kubernetes Service as an endpoint.
      admin_user:
        name: root
        type: raw
      admin_password:
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

## Exposed Metrics

- `mysql_user_created_total`
- `mysql_user_deleted_total`
## Contributing

[CONTRIBUTING](CONTRIBUTING.md)
