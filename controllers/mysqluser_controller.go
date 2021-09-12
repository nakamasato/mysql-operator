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

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"database/sql"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	_ "github.com/go-sql-driver/mysql"

	mysqlv1alpha1 "github.com/nakamasato/mysql-user-operator/api/v1alpha1"
)

const mysqlUserFinalizer = "mysqluser.nakamasato.com/finalizer"

// MySQLUserReconciler reconciles a MySQLUser object
type MySQLUserReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=mysql.nakamasato.com,resources=mysqlusers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mysql.nakamasato.com,resources=mysqlusers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mysql.nakamasato.com,resources=mysqlusers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MySQLUser object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *MySQLUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch MySQLUser
	mysqlUser := &mysqlv1alpha1.MySQLUser{}
	err := r.Get(ctx, req.NamespacedName, mysqlUser)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Fetch MySQLUser instance. MySQLUser not found.", "mysqlUser.Name", mysqlUser.Name, "mysqlUser.Namespace", mysqlUser.Namespace)
			return ctrl.Result{}, nil
		}

		log.Error(err, "Fetch MySQLUser instance. Failed to get MySQLUser.")
		return ctrl.Result{}, err
	}
	log.Info("Fetch MySQLUser instance. MySQLUser resource found.", "mysqlUser.Name", mysqlUser.Name, "mysqlUser.Namespace", mysqlUser.Namespace)

	mysqlName := mysqlUser.Spec.MysqlName

	// Fetch MySQL
	mysql := &mysqlv1alpha1.MySQL{}
	var mysqlNamespacedName = client.ObjectKey{Namespace: req.Namespace, Name: mysqlUser.Spec.MysqlName}
	if err := r.Get(ctx, mysqlNamespacedName, mysql); err != nil {
		log.Error(err, "unable to fetch MySQL")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Info("Fetched MySQL instance.")

	// Connect to MySQL
	db, err := r.getMySQLDB(log, mysql)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	// Finalize if DeletionTimestamp exists
	isMysqlUserMarkedToBeDeleted := mysqlUser.GetDeletionTimestamp() != nil
	if isMysqlUserMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(mysqlUser, mysqlUserFinalizer) {
			// Run finalization logic for mysqlUserFinalizer. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if err := r.finalizeMySQLUser(log, mysqlUser, mysql); err != nil {
				return ctrl.Result{}, err
			}
			// Remove mysqlUserFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			controllerutil.RemoveFinalizer(mysqlUser, mysqlUserFinalizer)
			err := r.Update(ctx, mysqlUser)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, err
	}

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(mysqlUser, mysqlUserFinalizer) {
		controllerutil.AddFinalizer(mysqlUser, mysqlUserFinalizer)
		err = r.Update(ctx, mysqlUser)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// Create MySQL user if not exists with password `password`
	log.Info("Create MySQL user if not.", "mysqlUser.Name", mysqlUser.Name, "mysqlUser.Namespace", mysqlUser.Namespace)
	password := "password" // TODO: #5 generate a random password
	_, err = db.Exec("CREATE USER IF NOT EXISTS '" + mysqlUser.Name + "'@'%' IDENTIFIED BY '" + password + "';")
	if err != nil {
		panic(err.Error())
	}

	// Create Secret with prefix `mysql-` + `<mysqlName>`
	data := make(map[string][]byte)
	data["password"] = []byte(password)
	secretName := "mysql-" + mysqlName + "-" + mysqlUser.Name
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: req.Namespace,
		},
	}
	err = ctrl.SetControllerReference(mysqlUser, secret, r.Scheme) // Set owner of this Secret
	if err != nil {
		log.Error(err, "Failed to SetControllerReference for Secret.")
		return ctrl.Result{}, err
	}
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, secret, func() error {
		secret.Data = data
		log.Info("Successfully created Secret.")
		return nil
	}); err != nil {
		log.Error(err, "Error creating or updating Secret.")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MySQLUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mysqlv1alpha1.MySQLUser{}).
		Complete(r)
}

func (r *MySQLUserReconciler) finalizeMySQLUser(log logr.Logger, mysqlUser *mysqlv1alpha1.MySQLUser, mysql *mysqlv1alpha1.MySQL) error {
	// 1. Get the referenced MySQL instance.
	// 2. Connect to MySQL.
	// 3. Delete the MySQL user.

	db, err := r.getMySQLDB(log, mysql)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	_, err = db.Exec("DROP USER IF EXISTS '" + mysqlUser.Name + "'@'%';")
	if err != nil {
		panic(err.Error())
	}

	log.Info("Successfully finalized mysqlUser")
	return nil
}

func (r *MySQLUserReconciler) getMySQLDB(log logr.Logger, mysql *mysqlv1alpha1.MySQL) (*sql.DB, error) {
	return sql.Open("mysql", mysql.Spec.AdminUser+":"+mysql.Spec.AdminPassword+"@tcp("+mysql.Spec.Host+":3306)/")
}
