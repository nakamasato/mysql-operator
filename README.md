# MySQL Operator

This is a go-based Kubernetes operator built with [operator-sdk](https://sdk.operatorframework.io/docs/building-operators/golang/), which manages MySQL databases, schema, users, permissions for existing MySQL clusters. This operator DOES NOT manage MySQL cluster like other MySQL operator.

## Versions

- Go: 1.16+
## Components

- `MySQL`: MySQL cluster
- `MySQLUser`: MySQL user (`mysqlName` and `host`)
- `MySQLDB`: MySQL database including schema repository (ToDo)

## Getting Started
TBD

## Contributing

[CONTRIBUTING](CONTRIBUTING.md)
