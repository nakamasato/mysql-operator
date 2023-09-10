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

	. "github.com/go-sql-driver/mysql"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mysqlv1alpha1 "github.com/nakamasato/mysql-operator/api/v1alpha1"
	mysqlinternal "github.com/nakamasato/mysql-operator/internal/mysql"
	secret "github.com/nakamasato/mysql-operator/internal/secret"
)

const mysqlFinalizer = "mysql.nakamasato.com/finalizer"

// MySQLReconciler reconciles a MySQL object
type MySQLReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	MySQLClients    mysqlinternal.MySQLClients
	MySQLDriverName string
	SecretManagers  map[string]secret.SecretManager
}

//+kubebuilder:rbac:groups=mysql.nakamasato.com,resources=mysqls,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mysql.nakamasato.com,resources=mysqls/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mysql.nakamasato.com,resources=mysqls/finalizers,verbs=update
//+kubebuilder:rbac:groups=mysql.nakamasato.com,resources=mysqlusers,verbs=list;
//+kubebuilder:rbac:groups=mysql.nakamasato.com,resources=mysqldbs,verbs=list;

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MySQL object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *MySQLReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithName("MySQLReconciler")

	// Fetch MySQL
	mysql := &mysqlv1alpha1.MySQL{}
	err := r.Get(ctx, req.NamespacedName, mysql)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("[FetchMySQL] Not found", "mysql.Name", mysql.Name, "mysql.Namespace", mysql.Namespace)
			return ctrl.Result{}, nil
		}

		log.Error(err, "[FetchMySQL] Failed to get MySQL")
		return ctrl.Result{}, err
	}

	// Add a finalizer if not exists
	if controllerutil.AddFinalizer(mysql, mysqlFinalizer) {
		if err := r.Update(ctx, mysql); err != nil {

			log.Error(err, "Failed to update MySQL after adding finalizer")
			return ctrl.Result{}, err
		}
	}

	// Get referenced number
	referencedUserNum, err := r.countReferencesByMySQLUser(ctx, mysql)
	if err != nil {
		log.Error(err, "Failed get referencedUserNum")
		return ctrl.Result{}, err
	}
	referencedDbNum, err := r.countReferencesByMySQLDB(ctx, mysql)
	if err != nil {
		log.Error(err, "Failed get referencedDbNum")
		return ctrl.Result{}, err
	}
	log.Info("Successfully got referenced num", "referencedUserNum", referencedUserNum, "referencedDbNum", referencedDbNum)

	// Update Status
	if mysql.Status.UserCount != int32(referencedUserNum) || mysql.Status.DBCount != int32(referencedDbNum) {
		mysql.Status.UserCount = int32(referencedUserNum)
		mysql.Status.DBCount = int32(referencedDbNum)
		err = r.Status().Update(ctx, mysql)
		if err != nil {
			log.Error(err, "[Status] Failed to update staus (UserCount and DBCount)",
				"UserCount", referencedUserNum, "DBCount", referencedDbNum)
			return ctrl.Result{RequeueAfter: time.Second}, nil
		}
		log.Info("[Status] updated", "UserCount", referencedUserNum, "DBCount", referencedDbNum)
	}

	// Update MySQLClients
	retry, err := r.UpdateMySQLClients(ctx, mysql)
	if err != nil {
		mysql.Status.Connected = false
		mysql.Status.Reason = err.Error()
		if err := r.Status().Update(ctx, mysql); err != nil {
			log.Error(err, "failed to update status (Connected & Reason)", "status", mysql.Status)
			return ctrl.Result{RequeueAfter: time.Second}, nil
		}
		return ctrl.Result{}, err
	} else if retry {
		return ctrl.Result{RequeueAfter: time.Second}, nil
	}

	connected, reason := true, "Ping succeded and updated MySQLClients"
	if mysql.Status.Connected != connected || mysql.Status.Reason != reason {
		mysql.Status.Connected = connected
		mysql.Status.Reason = reason
		if err := r.Status().Update(ctx, mysql); err != nil {
			log.Error(err, "failed to update status (Connected & Reason)", "status", mysql.Status)
			return ctrl.Result{RequeueAfter: time.Second}, nil
		}
	}

	if !mysql.GetDeletionTimestamp().IsZero() && controllerutil.ContainsFinalizer(mysql, mysqlFinalizer) {
		if r.finalizeMySQL(ctx, mysql) {
			if controllerutil.RemoveFinalizer(mysql, mysqlFinalizer) {
				if err := r.Update(ctx, mysql); err != nil {
					return ctrl.Result{}, err
				}
			}
		} else {
			log.Info("Could not complete finalizer. waiting another second")
			return ctrl.Result{RequeueAfter: time.Second}, nil
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MySQLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mysqlv1alpha1.MySQL{}).
		Owns(&mysqlv1alpha1.MySQLUser{}).
		Owns(&mysqlv1alpha1.MySQLDB{}).
		Complete(r)
}

func (r *MySQLReconciler) UpdateMySQLClients(ctx context.Context, mysql *mysqlv1alpha1.MySQL) (retry bool, err error) {
	log := log.FromContext(ctx).WithName("MySQLReconciler")
	// Get MySQL config from raw username and password or GCP secret manager
	cfg, err := r.getMySQLConfig(ctx, mysql)
	if err != nil {
		return true, err
	}
	if db, _ := r.MySQLClients.GetClient(mysql.GetKey()); db == nil {
		log.Info("MySQLClients doesn't have client", "key", mysql.GetKey())

		db, err := sql.Open(r.MySQLDriverName, cfg.FormatDSN())
		if err != nil {
			log.Error(err, "Failed to open MySQL database", "mysql.Name", mysql.Name)
			return true, err
		}
		err = db.PingContext(ctx)
		if err != nil {
			log.Error(err, "Ping failed", "mysql.Name", mysql.Name)
			return true, err
		}

		// key: mysql.Namespace-mysql.Name
		r.MySQLClients[mysql.GetKey()] = db
		log.Info("Successfully added MySQL client", "mysql.Name", mysql.Name)
	}

	// open connection for each MySQLDB
	mysqlDBList := &mysqlv1alpha1.MySQLDBList{}
	err = r.List(ctx, mysqlDBList, client.MatchingFields{"spec.mysqlName": mysql.Name})
	if err != nil {
		return true, err
	}
	for _, mysqlDB := range mysqlDBList.Items {
		if mysqlDB.Status.Phase != "Ready" {
			log.Info("mysqlDB is not ready", "mysqlDB", mysqlDB.Name, "mysqlDB.Status", mysqlDB.Status)
			return true, nil
		}
		if _, err := r.MySQLClients.GetClient(mysqlDB.GetKey()); err != nil {
			cfg.DBName = mysqlDB.Spec.DBName
			db, err := sql.Open(r.MySQLDriverName, cfg.FormatDSN())
			if err != nil {
				return true, err
			}
			err = db.PingContext(ctx)
			if err != nil {
				return true, err
			}
			r.MySQLClients[mysqlDB.GetKey()] = db
			log.Info("Successfully added MySQL client", "mysqlDB.Name", mysqlDB.Name)
		}
	}
	return false, nil
}

// If GcpSecretName is set, get password from GCP secret manager
// Otherwise user MySQL.Spec.AdminPassword
func (r *MySQLReconciler) getMySQLConfig(ctx context.Context, mysql *mysqlv1alpha1.MySQL) (Config, error) {
	log := log.FromContext(ctx)
	secretManager, ok := r.SecretManagers[mysql.Spec.AdminPassword.Type]
	if !ok {
		return Config{}, fmt.Errorf("the specified SecretManager type (%s) doesn't exist", mysql.Spec.AdminPassword.Type)
	}
	password, err := secretManager.GetSecret(ctx, mysql.Spec.AdminPassword.Name)
	if err != nil {
		log.Error(err, "failed to get secret from secret manager", "secret", mysql.Spec.AdminPassword.Name)
		return Config{}, err
	}
	secretManager, ok = r.SecretManagers[mysql.Spec.AdminUser.Type]
	if !ok {
		return Config{}, fmt.Errorf("the specified SecretManager type (%s) doesn't exist", mysql.Spec.AdminUser.Type)
	}
	user, err := secretManager.GetSecret(ctx, mysql.Spec.AdminUser.Name)
	if err != nil {
		return Config{}, err
	}

	return Config{
		User:                 user,
		Passwd:               password,
		Addr:                 fmt.Sprintf("%s:%d", mysql.Spec.Host, mysql.Spec.Port),
		Net:                  "tcp",
		AllowNativePasswords: true,
	}, nil
}

func (r *MySQLReconciler) countReferencesByMySQLUser(ctx context.Context, mysql *mysqlv1alpha1.MySQL) (int, error) {
	// 1. Get the referenced MySQLUser instances.
	// 2. Return the number of referencing MySQLUser.
	mysqlUserList := &mysqlv1alpha1.MySQLUserList{}
	err := r.List(ctx, mysqlUserList, client.MatchingFields{"spec.mysqlName": mysql.Name})

	if err != nil {
		return 0, err
	}
	return len(mysqlUserList.Items), nil
}

func (r *MySQLReconciler) countReferencesByMySQLDB(ctx context.Context, mysql *mysqlv1alpha1.MySQL) (int, error) {
	mysqlDBList := &mysqlv1alpha1.MySQLDBList{}
	err := r.List(ctx, mysqlDBList, client.MatchingFields{"spec.mysqlName": mysql.Name})

	if err != nil {
		return 0, err
	}
	return len(mysqlDBList.Items), nil
}

// finalizeMySQL return true if no user and no db is referencing the given MySQL
func (r *MySQLReconciler) finalizeMySQL(ctx context.Context, mysql *mysqlv1alpha1.MySQL) bool {
	log := log.FromContext(ctx).WithName("MySQLReconciler")
	if mysql.Status.UserCount > 0 || mysql.Status.DBCount > 0 {
		log.Info("there's referencing user or database", "UserCount", mysql.Status.UserCount, "DBCount", mysql.Status.DBCount)
		return false
	}
	if db, ok := r.MySQLClients[mysql.GetKey()]; ok {
		if err := db.Close(); err != nil {
			return false
		}
		delete(r.MySQLClients, mysql.GetKey())
		log.Info("Closed and removed MySQL client", "mysql.Key", mysql.GetKey())
	} else {
		log.Info("MySQL client doesn't exist", "mysql.Key", mysql.GetKey())
	}
	return true
}
