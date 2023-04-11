/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	mysqlv1alpha1 "github.com/nakamasato/mysql-operator/api/v1alpha1"
	mysqlinternal "github.com/nakamasato/mysql-operator/internal/mysql"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	mysqlDBFinalizer                   = "mysqldb.nakamasato.com/finalizer"
	mysqlDBPhaseNotReady               = "NotReady"
	mysqlDBReasonMySQLFetchFailed      = "Failed to fetch MySQL"
	mysqlDBReasonMySQLConnectionFailed = "Failed to connect to mysql"
	mysqlDBPhaseReady                  = "Ready"
	mysqlDBReasonCompleted             = "Database successfully created"
)

// MySQLDBReconciler reconciles a MySQLDB object
type MySQLDBReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	MySQLClients mysqlinternal.MySQLClients
}

//+kubebuilder:rbac:groups=mysql.nakamasato.com,resources=mysqlmysqlDBs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mysql.nakamasato.com,resources=mysqlmysqlDBs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mysql.nakamasato.com,resources=mysqlmysqlDBs/finalizers,verbs=update

// Reconcile function is responsible for managing MySQL database.
// Create database if not exists in the target MySQL and drop it if
// the corresponding object is deleted.
func (r *MySQLDBReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithName("MySQLDBReconciler")

	// 1. Fetch MySQLDB
	mysqlDB := &mysqlv1alpha1.MySQLDB{}
	err := r.Get(ctx, req.NamespacedName, mysqlDB)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("MySQLDB not found", "req.NamespacedName", req.NamespacedName)
			return ctrl.Result{}, nil
		}

		log.Error(err, "Failed to get MySQLDB")
		return ctrl.Result{}, err
	}

	// 2. Fetch MySQL
	mysql := &mysqlv1alpha1.MySQL{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: mysqlDB.Spec.MysqlName}, mysql); err != nil {
		log.Error(err, "[FetchMySQL] Failed")
		mysqlDB.Status.Phase = mysqlDBPhaseNotReady
		mysqlDB.Status.Reason = mysqlDBReasonMySQLFetchFailed
		if serr := r.Status().Update(ctx, mysqlDB); serr != nil {
			log.Error(serr, "Failed to update MySQLDB status", "Name", mysqlDB.Name)
		}
		return ctrl.Result{}, err
	}

	// 3. Add finalizer
	if controllerutil.AddFinalizer(mysqlDB, mysqlDBFinalizer) {
		err = r.Update(ctx, mysqlDB)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// 4. Get mysqlClient without specifying database
	mysqlClient, err := r.MySQLClients.GetClient(mysql.GetKey())
	if err != nil {
		log.Error(err, "Failed to get MySQL client", "key", mysqlDB.GetKey())
		return ctrl.Result{RequeueAfter: time.Second}, nil
	}

	// 5. finalize if marked as deleted
	if !mysqlDB.GetDeletionTimestamp().IsZero() {
		if controllerutil.ContainsFinalizer(mysqlDB, mysqlDBFinalizer) {
			if err := r.finalizeMySQLDB(ctx, mysqlClient, mysqlDB); err != nil {
				return ctrl.Result{}, err
			}
			if controllerutil.RemoveFinalizer(mysqlDB, mysqlDBFinalizer) {
				if err := r.Update(ctx, mysqlDB); err != nil {
					return ctrl.Result{}, err
				}
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, nil
	}

	// 6. Create database if not exists
	res, err := mysqlClient.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", mysqlDB.Spec.DBName))
	if err != nil {
		log.Error(err, "[MySQL] Failed to create MySQL database.", "mysql", mysql.Name, "database", mysqlDB.Spec.DBName)
		mysqlDB.Status.Phase = mysqlDBPhaseNotReady
		mysqlDB.Status.Reason = err.Error()
		if serr := r.Status().Update(ctx, mysqlDB); serr != nil {
			log.Error(serr, "Failed to update mysqlDB status", "mysqlDB", mysqlDB.Name)
			return ctrl.Result{RequeueAfter: time.Second}, nil
		}
		return ctrl.Result{}, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Error(err, "Failed to get res.RowsAffected")
		return ctrl.Result{}, err
	}
	if rows > 0 {
		mysqlDB.Status.Phase = mysqlDBPhaseReady
		mysqlDB.Status.Reason = mysqlDBReasonCompleted
		if serr := r.Status().Update(ctx, mysqlDB); serr != nil {
			log.Error(serr, "Failed to update MySQLDB status", "Name", mysqlDB.Spec.DBName)
			return ctrl.Result{RequeueAfter: time.Second}, nil
		}
	} else {
		log.Info("database already exists", "database", mysqlDB.Spec.DBName)
	}

	// 7. Get MySQL client for database
	mysqlClient, err = r.MySQLClients.GetClient(mysqlDB.GetKey())
	if err != nil {
		log.Error(err, "Failed to get MySQL Client", "key", mysqlDB.GetKey())
		return ctrl.Result{RequeueAfter: time.Second}, nil
	}

	// 6. Migrate database
	if mysqlDB.Spec.SchemaMigrationFromGitHub == nil {
		return ctrl.Result{}, nil
	}
	driver, err := migratemysql.WithInstance( // initialize db driver instance
		mysqlClient,
		&migratemysql.Config{DatabaseName: mysqlDB.Spec.DBName},
	)
	if err != nil {
		log.Error(err, "failed to create migratemysql.WithInstance")
		return ctrl.Result{}, err
	}

	m, err := migrate.NewWithDatabaseInstance( // initialize Migrate with db driver instance
		// "github://nakamasato/mysql-operator/config/sample-migrations#enable-to-migrate-schema-with-migrate", // Currently only support GitHub source
		mysqlDB.Spec.SchemaMigrationFromGitHub.GetSourceUrl(),
		mysqlDB.Spec.DBName,
		driver,
	)
	if err != nil {
		log.Error(err, "failed to initialize NewWithDatabaseInstance")
		return ctrl.Result{}, err
	}
	err = m.Up()

	if err != nil {
		if err.Error() == "no change" {
			log.Info("migrate no change")
		} else {
			log.Error(err, "failed to Up")
			return ctrl.Result{}, err
		}
	}
	log.Info("migrate completed")

	return ctrl.Result{}, nil
}

// finalizeMySQLDB drops MySQL database
func (r *MySQLDBReconciler) finalizeMySQLDB(ctx context.Context, mysqlClient *sql.DB, mysqlDB *mysqlv1alpha1.MySQLDB) error {
	_, err := mysqlClient.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", mysqlDB.Spec.DBName))
	return err
}

// SetupWithManager sets up the controller with the Manager.
func (r *MySQLDBReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mysqlv1alpha1.MySQLDB{}).
		Complete(r)
}
