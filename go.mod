module github.com/nakamasato/mysql-operator

go 1.16

require (
	github.com/go-logr/logr v0.4.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.17.0
	github.com/prometheus/client_golang v1.11.0
	github.com/redhat-cop/operator-utils v1.3.1
	k8s.io/api v0.22.3
	k8s.io/apimachinery v0.22.3
	k8s.io/client-go v0.22.3
	k8s.io/kubectl v0.22.3 // indirect
	k8s.io/utils v0.0.0-20210819203725-bdf08cb9a70a // indirect
	sigs.k8s.io/controller-runtime v0.10.0
)
