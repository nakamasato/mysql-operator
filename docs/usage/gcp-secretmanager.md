# Read credentials from GCP SecretManager

mysql-operator can get the credentials of the `MySQL` user (which is used to access to the target MySQL cluster) from [GCP SecretManager](https://cloud.google.com/secret-manager)

In this example, we'll use Cloud SQL for MySQL, and run mysql-operator on GKE.

## Prepare GCP resources

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

1. Create GKE cluster

    ```
    gcloud container clusters create-auto $GKE_CLUSTER_NAME --location=$REGION
    ```

    ```
    gcloud container clusters get-credentials $GKE_CLUSTER_NAME --location=$REGION
    ```

1. Create Cloud SQL instance.

    ```
    ROOT_PASSWORD=$(openssl rand -base64 32)
    ```

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

1. Create Secret `mysql-password` with value `password`, which will be used for the credentials of custom resource `MySQL`.
    ```
    gcloud secrets create $SECRET_NAME --replication-policy="automatic" --project ${PROJECT}
    ```

    ```
    echo -n "${ROOT_PASSWORD}" | gcloud secrets versions add $SECRET_NAME --data-file=- --project ${PROJECT}
    ```

1. Create service account `mysql-operator`
    ```
    gcloud iam service-accounts create $SA_NAME --display-name=$SA_NAME
    ```

1. Grant necessary permission for the created `Secret` to the service account

    `roles/secretmanager.secretAccessor`:
    ```
    gcloud secrets add-iam-policy-binding $SECRET_NAME \
        --member="serviceAccount:${SA_NAME}@${PROJECT}.iam.gserviceaccount.com" \
        --role="roles/secretmanager.secretAccessor" --project ${PROJECT}
    ```

    ```
    gcloud projects add-iam-policy-binding $PROJECT \
        --member="serviceAccount:${SA_NAME}@${PROJECT}.iam.gserviceaccount.com" \
        --role="roles/cloudsql.client"
    ```

1. Allow to Kubernete Pod to impersonate the Service Account

    ```
    gcloud iam service-accounts add-iam-policy-binding ${SA_NAME}@${PROJECT}.iam.gserviceaccount.com \
    --role roles/iam.workloadIdentityUser \
    --member "serviceAccount:${PROJECT}.svc.id.goog[${NAMESPACE}/${KSA_NAME}]"
    ```

1. Deploy with helm

    ```
    kubectl create ns $NAMESPACE
    ```

    ```
    helm install mysql-operator ./charts/mysql-operator \
        --set cloudSecretManagerType=gcp,gcpServiceAccount=${SA_NAME}@${PROJECT}.iam.gserviceaccount.com,gcpProjectId=$PROJECT,cloudSQL.instanceConnectionName=$PROJECT:$REGION:$INSTANCE_NAME \
        -n $NAMESPACE
    ```

    ```
    helm list -n $NAMESPACE
    NAME            NAMESPACE       REVISION        UPDATED                                 STATUS          CHART                   APP VERSION
    mysql-operator  mysql-operator  1               2023-09-09 12:03:54.220046 +0900 JST    deployed        mysql-operator-v0.3.0   v0.3.0
    ```

    ```
    kubectl get pod -n $NAMESPACE
    NAME                                                 READY   STATUS    RESTARTS   AGE
    mysql-operator-controller-manager-5d9bb58bcc-ngjrc   1/1     Running   0          4m3s
    ```

1. Create MySQL

    ```
    kubectl apply -f - <<EOF
    apiVersion: mysql.nakamasato.com/v1alpha1
    kind: MySQL
    metadata:
      name: mysql-sample
    spec:
      host: "127.0.0.1" # auth SQL
      adminUser:
        name: root
        type: raw
      adminPassword:
        name: $SECRET_NAME
        type: gcp
    EOF
    ```

For more details, read [Workload Identity](https://cloud.google.com/kubernetes-engine/docs/concepts/workload-identity)


## Clean up

```
gcloud container clusters delete $GKE_CLUSTER_NAME --location $REGION
gcloud sql instances delete ${INSTANCE_NAME} --project ${PROJECT}
gcloud iam service-accounts delete ${SA_NAME}@${PROJECT}.iam.gserviceaccount.com --project ${PROJECT}
gcloud secrets delete $SECRET_NAME --project ${PROJECT}
```
