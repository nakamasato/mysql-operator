# Run on GKE and manage Cloud SQL (MySQL) with GCP SecretManager

mysql-operator can get the credentials of the `MySQL` user (which is used to access to the target MySQL cluster) from [GCP SecretManager](https://cloud.google.com/secret-manager)

In this example, we'll use Cloud SQL for MySQL, and run mysql-operator on GKE.

## 1. Prepare GCP resources

### 1.1. Prepare env var and gcloud

1. Set environment variables
    ```
    INSTANCE_NAME=mysql-test
    ZONE=asia-northeast1-b
    REGION=asia-northeast1
    SECRET_NAME=mysql-password
    SA_NAME=mysql-operator
    GKE_CLUSTER_NAME=hello-cluster
    NAMESPACE=mysql-operator
    KSA_NAME=mysql-operator-controller-manager
    ```

1. Configure gcloud project

    ```
    PROJECT=<your_project_id>
    gcloud config set project $PROJECT
    ```


### 1.2. Create GKE cluster

1. Create GKE cluster

    ```
    gcloud container clusters create-auto $GKE_CLUSTER_NAME --location=$REGION
    ```
1. Set up kubeconfig
    ```
    gcloud container clusters get-credentials $GKE_CLUSTER_NAME --location=$REGION
    ```

### 1.3. Create Cloud SQL instance.

1. Generate random password for root user.
    ```
    ROOT_PASSWORD=$(openssl rand -base64 32)
    ```
1. Create Cloud SQL instance.
    ```
    gcloud sql instances create $INSTANCE_NAME \
    --cpu=1 \
    --memory=3840MiB \
    --zone=${ZONE} \
    --root-password=$ROOT_PASSWORD \
    --project ${PROJECT}
    ```

    For existing instance, you can reset the root password with the following command:

    ```
    gcloud sql users set-password root \
        --host=% \
        --instance=$INSTANCE_NAME \
        --password=$ROOT_PASSWORD
    ```

### 1.4. Create SecretManager secret for root password

1. Create Secret `mysql-password` with value `password`, which will be used for the credentials of custom resource `MySQL`.
    ```
    gcloud secrets create $SECRET_NAME --replication-policy="automatic" --project ${PROJECT}
    ```

    ```
    echo -n "${ROOT_PASSWORD}" | gcloud secrets versions add $SECRET_NAME --data-file=- --project ${PROJECT}
    ```
### 1.5. Create Service Account for `mysql-operator`

1. Create service account `mysql-operator`

    ```
    gcloud iam service-accounts create $SA_NAME --display-name=$SA_NAME
    ```

1. Grant necessary permission for the created `Secret` to the service account

    1. `roles/secretmanager.secretAccessor`: To allow mysql-operator to get root password from SecretManager

        ```
        gcloud secrets add-iam-policy-binding $SECRET_NAME \
            --member="serviceAccount:${SA_NAME}@${PROJECT}.iam.gserviceaccount.com" \
            --role="roles/secretmanager.secretAccessor" --project ${PROJECT}
        ```

    1. `roles/cloudsql.client`: To allow mysql-operator can connect to Cloud SQL
        ```
        gcloud projects add-iam-policy-binding $PROJECT \
            --member="serviceAccount:${SA_NAME}@${PROJECT}.iam.gserviceaccount.com" \
            --role="roles/cloudsql.client"
        ```

    1. `roles/iam.workloadIdentityUser`: To allow to Kubernete Pod to impersonate the Service Account

        ```
        gcloud iam service-accounts add-iam-policy-binding ${SA_NAME}@${PROJECT}.iam.gserviceaccount.com \
            --role roles/iam.workloadIdentityUser \
            --member "serviceAccount:${PROJECT}.svc.id.goog[${NAMESPACE}/${KSA_NAME}]"
        ```

        For more details, read [Workload Identity](https://cloud.google.com/kubernetes-engine/docs/concepts/workload-identity)

## 2. Install `mysql-operator` with Helm

1. Create a Namespace.

    ```
    kubectl create ns $NAMESPACE
    ```

1. Deploy with Helm.

    ```
    helm repo add nakamasato https://nakamasato.github.io/helm-charts
    helm repo update
    ```

    ```
    helm install mysql-operator nakamasato/mysql-operator \
        --set adminUserSecretType=gcp \
        --set gcpServiceAccount=${SA_NAME}@${PROJECT}.iam.gserviceaccount.com \
        --set gcpProjectId=$PROJECT \
        --set cloudSQL.instanceConnectionName=$PROJECT:$REGION:$INSTANCE_NAME \
        -n $NAMESPACE
    ```

1. Check Helm release.

    ```
    helm list -n $NAMESPACE
    NAME            NAMESPACE       REVISION        UPDATED                                 STATUS          CHART                   APP VERSION
    mysql-operator  mysql-operator  1               2023-09-09 12:03:54.220046 +0900 JST    deployed        mysql-operator-v0.3.0   v0.3.0
    ```
1. Check `mysql-operator` Pod.
    ```
    kubectl get pod -n $NAMESPACE
    NAME                                                 READY   STATUS    RESTARTS   AGE
    mysql-operator-controller-manager-77649f6bb9-xbt9l   2/2     Running   0          2m59s
    ```

## 3. Create custom resources (Manage MySQL users, databases, schemas, etc.)

1. Create sample `MySQL`, `MySQLUser`, `MySQLDB`.

    If you want to create sample MySQL, MySQLUser, and `MySQLDB` at once, you can use the following command:

    ```
    kubectl apply -k https://github.com/nakamasato/mysql-operator/config/samples-on-k8s-with-gcp-secretmanager
    ```

1. Create `MySQL`

    ```
    kubectl apply -f - <<EOF
    apiVersion: mysql.nakamasato.com/v1alpha1
    kind: MySQL
    metadata:
      name: $INSTANCE_NAME
    spec:
      host: "127.0.0.1" # auth SQL proxy
      adminUser:
        name: root
        type: raw
      adminPassword:
        name: $SECRET_NAME
        type: gcp
    EOF
    ```

    Check the status
    ```
    NAME         HOST        ADMINUSER   CONNECTED   USERCOUNT   DBCOUNT   REASON
    mysql-test   127.0.0.1   root        true        0           0         Ping succeded and updated MySQLClients
    ```

1. Create MySQL user `sample-user`.

    ```
    kubectl apply -f - <<EOF
    apiVersion: mysql.nakamasato.com/v1alpha1
    kind: MySQLUser
    metadata:
      name: sample-user
    spec:
      mysqlName: $INSTANCE_NAME
    EOF
    ```

    ```
    kubectl get secret
    NAME                           TYPE     DATA   AGE
    mysql-mysql-test-sample-user   Opaque   1      99s
    ```

1. Create MySQL DB `sample_db`

    ```
    kubectl apply -f - <<EOF
    apiVersion: mysql.nakamasato.com/v1alpha1
    kind: MySQLDB
    metadata:
      name: sample-db # this is a Kubernetes object name, not the name for MySQL database
    spec:
      dbName: sample_db # this is MySQL database name
      mysqlName: $INSTANCE_NAME
    EOF
    ```

    ```
    kubectl get mysqldb
    NAME        PHASE   REASON                          SCHEMAMIGRATION
    sample-db   Ready   Database successfully created   {"dirty":false,"version":0}
    ```

1. Check

    ```
    cloud-sql-proxy ${PROJECT}:${REGION}:${INSTANCE_NAME}
    ```

    ```
    mysql -uroot -p${ROOT_PASSWORD} --host 127.0.0.1
    ```

    ```sql
    mysql> show databases;
    +--------------------+
    | Database           |
    +--------------------+
    | information_schema |
    | mysql              |
    | performance_schema |
    | sample_db          |
    | sys                |
    +--------------------+
    5 rows in set (0.01 sec)
    ```

    ```sql
    mysql> select User, Host from mysql.user where User = 'sample-user';
    +-------------+------+
    | User        | Host |
    +-------------+------+
    | sample-user | %    |
    +-------------+------+
    1 row in set (0.01 sec)
    ```

## 4. Clean up

### 4.1. Kubernetes resources

Custom resources
```
kubectl delete mysqldb sample-db
kubectl delete mysqluser sample-user
kubectl delete mysql $INSTANCE_NAME
```

Uninstall `mysql-operator`

```
helm uninstall mysql-operator -n $NAMESPACE
```

### 4.2. GCP resources

```
gcloud container clusters delete $GKE_CLUSTER_NAME --location $REGION
gcloud sql instances delete ${INSTANCE_NAME} --project ${PROJECT}
gcloud iam service-accounts delete ${SA_NAME}@${PROJECT}.iam.gserviceaccount.com --project ${PROJECT}
gcloud secrets delete $SECRET_NAME --project ${PROJECT}
```
