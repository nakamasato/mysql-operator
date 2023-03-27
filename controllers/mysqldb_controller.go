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
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mysqlv1alpha1 "github.com/nakamasato/mysql-operator/api/v1alpha1"
	mysqlinternal "github.com/nakamasato/mysql-operator/internal/mysql"
)

const (
	mysqlDBPhaseNotReady               = "NotReady"
	mysqlDBReasonMySQLFetchFailed      = "Failed to fetch MySQL"
	mysqlDBReasonMySQLConnectionFailed = "Failed to connect to mysql"
	mysqlDBPhaseReady                  = "Ready"
	mysqlDBReasonCompleted             = "Database successfully created"
)

// MySQLDBReconciler reconciles a MySQLDB object
type MySQLDBReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	MySQLClientFactory mysqlinternal.MySQLClientFactory
}

//+kubebuilder:rbac:groups=mysql.nakamasato.com,resources=mysqldbs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mysql.nakamasato.com,resources=mysqldbs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mysql.nakamasato.com,resources=mysqldbs/finalizers,verbs=update

// Reconcile function is responsible for managing MySQL database.
// Create database if not exists in the target MySQL and drop it if
// the corresponding object is deleted.
func (r *MySQLDBReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithName("MySQLDBReconciler")

	// 1. Fetch MySQL DB
	db := &mysqlv1alpha1.MySQLDB{}
	err := r.Get(ctx, req.NamespacedName, db)
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
	if err := r.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: db.Spec.MysqlName}, mysql); err != nil {
		log.Error(err, "[FetchMySQL] Failed")
		db.Status.Phase = mysqlDBPhaseNotReady
		db.Status.Reason = mysqlDBReasonMySQLFetchFailed
		if serr := r.Status().Update(ctx, db); serr != nil {
			log.Error(serr, "Failed to update MySQLDB status", "Name", db.Name)
		}
		return ctrl.Result{}, err
	}

	// 3. Connect to MySQL
	cfg := mysqlinternal.MySQLConfig{
		AdminUser:     mysql.Spec.AdminUser,
		AdminPassword: mysql.Spec.AdminPassword,
		Host:          mysql.Spec.Host,
	}
	mysqlClient, err := r.MySQLClientFactory(cfg)
	if err != nil {
		db.Status.Phase = mysqlDBPhaseNotReady
		db.Status.Reason = mysqlDBReasonMySQLConnectionFailed
		if serr := r.Status().Update(ctx, db); serr != nil {
			log.Error(serr, "Failed to update db status", "db", db.Name)
		}
		log.Error(err, "[MySQLClient] Failed to create")
		return ctrl.Result{}, err // requeue
	}
	log.Info("[MySQLClient] Ping")
	ctxPing, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	err = mysqlClient.PingContext(ctxPing)
	if err != nil {
		db.Status.Phase = mysqlDBPhaseNotReady
		db.Status.Reason = mysqlDBReasonMySQLConnectionFailed
		log.Error(err, "[MySQLClient] Failed to connect to MySQL", "mysqlName", mysql.Name)
		if serr := r.Status().Update(ctx, db); serr != nil {
			log.Error(serr, "Failed to update db status", "db", db.Name)
			return ctrl.Result{RequeueAfter: time.Second}, nil
		}
		return ctrl.Result{RequeueAfter: time.Second}, nil // requeue after 1 second
	}
	log.Info("[MySQLClient] Successfully connected")
	defer mysqlClient.Close()

	// 4. Delete if NotFound with finalizer

	// 5. Create if not exists
	err = mysqlClient.Exec("CREATE DATABASE IF NOT EXISTS " + db.Spec.DBName + ";")
	if err != nil {
		log.Error(err, "[MySQL] Failed to create MySQL database.", "mysql", mysql.Name, "database", db.Spec.DBName)
		db.Status.Phase = mysqlDBPhaseNotReady
		db.Status.Reason = err.Error()
		if serr := r.Status().Update(ctx, db); serr != nil {
			log.Error(serr, "Failed to update db status", "db", db.Name)
			return ctrl.Result{RequeueAfter: time.Second}, nil
		}
		return ctrl.Result{}, err
	}
	db.Status.Phase = mysqlDBPhaseReady
	db.Status.Reason = mysqlDBReasonCompleted
	if serr := r.Status().Update(ctx, db); serr != nil {
		log.Error(serr, "Failed to update MySQLDB status", "Name", db.Spec.DBName)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MySQLDBReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mysqlv1alpha1.MySQLDB{}).
		Complete(r)
}
