package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)


const MetricsNamespace = "mysqloperator"

type MysqlUserTotalAdaptor struct {
	metric prometheus.Counter
}

func (m MysqlUserTotalAdaptor) Increment() {
	m.metric.Inc()
}


var (
	mysqlUserTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: MetricsNamespace,
			Name: "mysql_user_total",
			Help: "Number of mysqlUser processed",
		},
	)

	MysqlUserTotal *MysqlUserTotalAdaptor = &MysqlUserTotalAdaptor{metric: mysqlUserTotal}
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(mysqlUserTotal)
}
