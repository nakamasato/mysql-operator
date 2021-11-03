module github.com/nakamasato/mysql-operator

go 1.16

require (
	github.com/go-logr/logr v0.4.0 // indirect
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/redhat-cop/operator-utils v1.1.4
	k8s.io/api v0.21.2 // indirect
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
	sigs.k8s.io/controller-runtime v0.9.2
)

replace github.com/nakamasato/mysql-operator/internal/mysql => ./internal/mysql
