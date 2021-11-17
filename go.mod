module github.com/nakamasato/mysql-operator

go 1.16

require (
	github.com/go-logr/logr v0.4.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/prometheus/client_golang v1.11.0
	github.com/redhat-cop/operator-utils v1.2.0
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
	sigs.k8s.io/controller-runtime v0.9.2
)
