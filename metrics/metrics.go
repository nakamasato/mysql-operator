package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)


const MetricsNamespace = "mysqloperator"


var (
	mysqlCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: MetricsNamespace,
			Name: "mysql_total",
			Help: "Number of mysql proccessed",
		},
	)
	mysqlUserCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: MetricsNamespace,
			Name: "mysql_user_total",
			Help: "Number of mysqlUser processed",
		},
	)
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(mysqlCounter, mysqlUserCounter)
}
