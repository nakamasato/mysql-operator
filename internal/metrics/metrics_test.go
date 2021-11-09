package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestMySQLUserCreatedMetrics(t *testing.T) {
	MysqlUserCreatedTotal.Increment()
	actual := testutil.ToFloat64(mysqlUserCreatedTotal)
	assertFloat64(t, float64(1), actual)

	MysqlUserCreatedTotal.Increment()
	actual = testutil.ToFloat64(mysqlUserCreatedTotal)
	assertFloat64(t, float64(2), actual)
}

func TestMySQLUserDeletedMetrics(t *testing.T) {
	MysqlUserDeletedTotal.Increment()
	actual := testutil.ToFloat64(mysqlUserDeletedTotal)
	assertFloat64(t, float64(1), actual)

	MysqlUserDeletedTotal.Increment()
	actual = testutil.ToFloat64(mysqlUserDeletedTotal)
	assertFloat64(t, float64(2), actual)
}

func assertFloat64(t *testing.T, expected, actual float64) {
	if actual != expected {
		t.Errorf("value is not %f", expected)
	}
}
