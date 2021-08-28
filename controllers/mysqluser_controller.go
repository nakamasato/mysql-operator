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
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	cachev1alpha1 "github.com/nakamasato/mysql-user-operator/api/v1alpha1"
)

// MySQLUserReconciler reconciles a MySQLUser object
type MySQLUserReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cache.nakamasato.com,resources=mysqlusers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cache.nakamasato.com,resources=mysqlusers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cache.nakamasato.com,resources=mysqlusers/finalizers,verbs=update

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
	mysqlUser := &cachev1alpha1.MySQLUser{}
	err := r.Get(ctx, req.NamespacedName, mysqlUser)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Fetch MySQLUser instance. MySQLUser not found.")
			return ctrl.Result{}, err
		}

		log.Error(err, "Fetch MySQLUser instance. Failed to get MySQLUser.")
		return ctrl.Result{}, err
	}
	log.Info("Fetch MySQLUser instance. MySQLUser resource found.", "mysqlUser.Name", mysqlUser.Name, "mysqlUser.Namespace", mysqlUser.Namespace)

	// Fetch MySQL
	var mysql cachev1alpha1.MySQL
	mysqlName := "mysql-sample" // TODO: extract from mysqlUser.mysqlName
	var mysqlNamespacedName = client.ObjectKey{Namespace: req.Namespace, Name: mysqlName}
	if err := r.Get(ctx, mysqlNamespacedName, &mysql); err != nil {
		log.Error(err, "unable to fetch MySQL")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Info("Fetched MySQL instance.")

	// Connect to MySQL
	db, err := sql.Open("mysql", "root:password@tcp(localhost:3306)/")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	// Create MySQL user if not exists with password `password`
	err = db.QueryRow("SELECT * FROM mysql.user WHERE User = ?", 1).Scan(&mysqlUser.Name)
	switch {
	case err == sql.ErrNoRows:
		fmt.Println("レコードが存在しません")
		log.Info("mysql.user doesn't exist. will be created.", "mysqlUser.Name", mysqlUser.Name, "mysqlUser.Namespace", mysqlUser.Namespace)
		password := "password" // TODO: generate a random password
		_, err = db.Exec("CREATE USER IF NOT EXISTS '" + mysqlUser.Name + "'@'%' IDENTIFIED BY '" + password + "';")
		if err != nil {
			panic(err.Error())
		}
		// create Secret with prefix `mysql-` + `<mysqlName>`
		secretName := "mysql-" + mysqlName + "-" + mysqlUser.Name
		data := make(map[string][]byte)
		data["password"] = []byte(password)
		secret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: req.Namespace,
			},
		}
		if _, err := ctrl.CreateOrUpdate(ctx, r.Client, secret, func() error {
			secret.Data = data
			log.Info("Successfully created Secret.")
			return nil
		}); err != nil {
			log.Error(err, "Error creating or updating Secret.")
			return ctrl.Result{}, err
		}

	case err != nil:
		panic(err.Error())
	default:
		log.Info("default.", "mysqlUser.Name", mysqlUser.Name, "mysqlUser.Namespace", mysqlUser.Namespace)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MySQLUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.MySQLUser{}).
		Complete(r)
}
