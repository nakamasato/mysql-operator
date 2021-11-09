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
	mysqlUserCreatedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: MetricsNamespace,
			Name:      "mysql_user_created_total",
			Help:      "Number of created MySQL User",
		},
	)

	mysqlUserDeletedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: MetricsNamespace,
			Name:      "mysql_user_deleted_total",
			Help:      "Number of deleted MySQL User",
		},
	)

	MysqlUserCreatedTotal *MysqlUserTotalAdaptor = &MysqlUserTotalAdaptor{metric: mysqlUserCreatedTotal}
	MysqlUserDeletedTotal *MysqlUserTotalAdaptor = &MysqlUserTotalAdaptor{metric: mysqlUserDeletedTotal}
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(
		mysqlUserCreatedTotal,
		mysqlUserDeletedTotal,
	)
}
