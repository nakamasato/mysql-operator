# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Core Commands

### Building and Testing

```bash
# Build the operator binary
make build

# Run the operator locally
make run

# Run unit tests
make test

# Run tests with specific focus
bin/ginkgo -skip-package=e2e --focus "Should have finalizer" --fail-fast ./...

# Run e2e tests with Ginkgo
make e2e-with-ginkgo

# Run e2e tests with KUTTL
make e2e-with-kuttl

# Run linting
make lint

# Generate CRD manifests and code
make manifests generate
```

### Deployment

```bash
# Install CRDs into the K8s cluster
make install

# Uninstall CRDs from the K8s cluster
make uninstall

# Deploy controller to the K8s cluster
make deploy

# Undeploy controller from the K8s cluster
make undeploy

# Build Docker image with the manager
make docker-build

# Push Docker image with the manager
make docker-push

# Deploy using Skaffold (local development)
skaffold dev
```

### Working with Helm Charts

```bash
# Generate Helm chart from kustomize configs
make helm
```

## Architecture Overview

This is a Kubernetes operator built with [operator-sdk](https://sdk.operatorframework.io/) that manages MySQL databases, schema, users, and permissions in existing MySQL servers. It does not manage MySQL clusters themselves.

### Custom Resources

1. **MySQL**: Defines the connection to a MySQL server (host, port, admin credentials)
2. **MySQLUser**: Defines a MySQL user for a specific MySQL instance
3. **MySQLDB**: Defines a MySQL database for a specific MySQL instance, with optional schema migration

### Controllers

1. **MySQLReconciler**: Manages `MySQLClients` based on `MySQL` and `MySQLDB` resources
2. **MySQLUserReconciler**: Creates/deletes MySQL users defined in `MySQLUser` and creates Kubernetes Secrets to store passwords
3. **MySQLDBReconciler**: Creates/deletes databases and handles schema migrations defined in `MySQLDB`

### Key Components

- **MySQLClients**: Manages database connections to MySQL servers
- **SecretManagers**: Interface for retrieving secrets (raw, GCP Secret Manager, Kubernetes Secrets)
- **Metrics**: Exposes Prometheus metrics for MySQL user operations

## Secret Management

The operator supports three methods for handling admin credentials:

1. **Raw**: Plaintext credentials in the CR spec
2. **GCP Secret Manager**: Credentials stored in Google Cloud Secret Manager
3. **Kubernetes Secrets**: Credentials stored in Kubernetes Secrets

Set the secret type when starting the operator:

```bash
# For GCP Secret Manager
go run ./cmd/main.go --admin-user-secret-type=gcp --gcp-project-id=$PROJECT_ID

# For Kubernetes Secrets
go run ./cmd/main.go --admin-user-secret-type=k8s --k8s-secret-namespace=default
```

## Database Migrations

The operator supports schema migrations through the `MySQLDB` custom resource using the golang-migrate library. Migrations can be sourced from GitHub repositories.

Example configuration:

```yaml
apiVersion: mysql.nakamasato.com/v1alpha1
kind: MySQLDB
metadata:
  name: sample-db
spec:
  dbName: sample_db
  mysqlName: mysql-sample
  schemaMigrationFromGitHub:
    owner: myorg
    repo: myrepo
    path: migrations
    ref: main
```

## Development Workflow

1. Start a local Kubernetes cluster (kind, minikube, etc.)
2. Run a MySQL server (either locally with Docker or in the cluster)
3. Install CRDs and run the operator
4. Apply sample custom resources
5. Validate the resources were created in MySQL

## Monitoring

The operator exports Prometheus metrics:
- `mysql_user_created_total`
- `mysql_user_deleted_total`

To view metrics, you can set up Prometheus and ServiceMonitor.