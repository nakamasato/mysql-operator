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
	"strings"

	"github.com/redhat-cop/operator-utils/pkg/util"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	_ "github.com/go-sql-driver/mysql"

	mysqlv1alpha1 "github.com/nakamasato/mysql-operator/api/v1alpha1"
	"github.com/nakamasato/mysql-operator/internal/metrics"
	mysqlinternal "github.com/nakamasato/mysql-operator/internal/mysql"
	"github.com/nakamasato/mysql-operator/internal/utils"
)

const (
	mysqlUserFinalizer                   = "mysqluser.nakamasato.com/finalizer"
	mysqlUserReasonCompleted             = "Both secret and mysql user are successfully created."
	mysqlUserReasonMySQLConnectionFailed = "Failed to connect to mysql"
	mysqlUserReasonMySQLFetchFailed      = "Failed to fetch MySQL"
	mysqlUserPhaseReady                  = "Ready"
	mysqlUserPhaseNotReady               = "NotReady"
)

// MySQLUserReconciler reconciles a MySQLUser object
type MySQLUserReconciler struct {
	util.ReconcilerBase
	Log                logr.Logger
	Scheme             *runtime.Scheme
	MySQLClientFactory mysqlinternal.MySQLClientFactory
}

//+kubebuilder:rbac:groups=mysql.nakamasato.com,resources=mysqlusers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mysql.nakamasato.com,resources=mysqlusers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mysql.nakamasato.com,resources=mysqlusers/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;update;patch

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
	err := r.GetClient().Get(ctx, req.NamespacedName, mysqlUser)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("[FetchMySQLUser] Not found", "req.NamespacedName", req.NamespacedName)
			return ctrl.Result{}, nil
		}

		log.Error(err, "[FetchMySQLUser] Failed")
		return ctrl.Result{}, err
	}
	log.Info("[FetchMySQLUser] Found.", "name", mysqlUser.ObjectMeta.Name, "mysqlUser.Namespace", mysqlUser.Namespace)
	mysqlUserName := mysqlUser.ObjectMeta.Name
	mysqlName := mysqlUser.Spec.MysqlName

	// Fetch MySQL
	mysql := &mysqlv1alpha1.MySQL{}
	var mysqlNamespacedName = client.ObjectKey{Namespace: req.Namespace, Name: mysqlUser.Spec.MysqlName}
	if err := r.GetClient().Get(ctx, mysqlNamespacedName, mysql); err != nil {
		log.Error(err, "[FetchMySQL] Failed")
		mysqlUser.Status.Phase = mysqlUserPhaseNotReady
		mysqlUser.Status.Reason = mysqlUserReasonMySQLFetchFailed
		// return ctrl.Result{}, client.IgnoreNotFound(err)
		return r.ManageError(ctx, mysqlUser, err)
	}
	log.Info("[FetchMySQL] Found")

	// SetOwnerReference if not exists
	if !r.ifOwnerReferencesContains(mysqlUser.ObjectMeta.OwnerReferences, mysql) {
		err := controllerutil.SetControllerReference(mysql, mysqlUser, r.Scheme)
		if err != nil {
			return r.ManageError(ctx, mysqlUser, err) // requeue
		}
		err = r.GetClient().Update(ctx, mysqlUser)
		if err != nil {
			return r.ManageError(ctx, mysqlUser, err) // requeue
		}
	}

	// Connect to MySQL
	cfg := mysqlinternal.MySQLConfig{
		AdminUser:     mysql.Spec.AdminUser,
		AdminPassword: mysql.Spec.AdminPassword,
		Host:          mysql.Spec.Host,
	}
	mysqlClient, err := r.MySQLClientFactory(cfg)
	if err != nil {
		mysqlUser.Status.Phase = mysqlUserPhaseNotReady
		mysqlUser.Status.Reason = mysqlUserReasonMySQLConnectionFailed
		log.Error(err, "[MySQLClient] Failed to create")
		return r.ManageError(ctx, mysqlUser, err) // requeue
	}
	log.Info("[MySQLClient] Ping")
	err = mysqlClient.Ping()
	if err != nil {
		mysqlUser.Status.Phase = mysqlUserPhaseNotReady
		mysqlUser.Status.Reason = mysqlUserReasonMySQLConnectionFailed
		log.Error(err, "[MySQLClient] Failed to connect to MySQL", "mysqlName", mysqlName)
		return r.ManageError(ctx, mysqlUser, err) // requeue
	}
	log.Info("[MySQLClient] Successfully connected")
	defer mysqlClient.Close()

	// Finalize if DeletionTimestamp exists
	isMysqlUserMarkedToBeDeleted := mysqlUser.GetDeletionTimestamp() != nil
	if isMysqlUserMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(mysqlUser, mysqlUserFinalizer) {
			// Run finalization logic for mysqlUserFinalizer. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if err := r.finalizeMySQLUser(log, mysqlUser, mysql); err != nil {
				// return ctrl.Result{}, err
				return r.ManageError(ctx, mysqlUser, err) // requeue
			}
			// Remove mysqlUserFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			controllerutil.RemoveFinalizer(mysqlUser, mysqlUserFinalizer)
			err := r.GetClient().Update(ctx, mysqlUser)
			if err != nil {
				// return ctrl.Result{}, err
				return r.ManageError(ctx, mysqlUser, err) // requeue
			}
			return ctrl.Result{}, nil
		}
		// return ctrl.Result{}, err
		return r.ManageSuccess(ctx, mysqlUser) // should return success when not having the finalizer
	}

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(mysqlUser, mysqlUserFinalizer) {
		controllerutil.AddFinalizer(mysqlUser, mysqlUserFinalizer)
		err = r.GetClient().Update(ctx, mysqlUser)
		if err != nil {
			return r.ManageError(ctx, mysqlUser, err) // requeue
		}
	}

	// Get password from Secret if exists. Otherwise, generate new one.
	secretName := getSecretName(mysqlName, mysqlUserName)
	secret := &v1.Secret{}
	err = r.GetClient().Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: secretName}, secret)
	var password string
	if err != nil {
		if errors.IsNotFound(err) { // Secret doesn't exists -> generate password
			log.Info("[password] Generate new password for Secret", "secretName", secretName)
			password = utils.GenerateRandomString(16)
		} else {
			log.Error(err, "[password] Failed to get Secret", "secretName", secretName)
			return r.ManageError(ctx, mysqlUser, err) // requeue
		}
	} else { // exists -> get password from Secret
		password = string(secret.Data["password"])
	}

	// Create MySQL user if not exists with the password set above.
	err = mysqlClient.Exec("CREATE USER IF NOT EXISTS '" + mysqlUserName + "'@'" + mysqlUser.Spec.Host + "' IDENTIFIED BY '" + password + "';")
	if err != nil {
		log.Error(err, "[MySQL] Failed to create MySQL user.", "mysqlName", mysqlName, "mysqlUserName", mysqlUserName)
		return r.ManageError(ctx, mysqlUser, err) // requeue
	}
	log.Info("[MySQL] Created or updated", "name", mysqlUserName, "mysqlUser.Namespace", mysqlUser.Namespace)
	metrics.MysqlUserCreatedTotal.Increment()
	mysqlUser.Status.Phase = mysqlUserPhaseReady
	mysqlUser.Status.Reason = "mysql user are successfully created. Secret is being created."

	err = r.createSecret(ctx, log, password, secretName, mysqlUser.Namespace, mysqlUser)
	// TODO: #35 add test if mysql user is successfully created but secret is failed to create
	if err != nil {
		return r.ManageError(ctx, mysqlUser, err)
	}
	mysqlUser.Status.Phase = mysqlUserPhaseReady
	mysqlUser.Status.Reason = mysqlUserReasonCompleted

	return r.ManageSuccess(ctx, mysqlUser)
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

	cfg := mysqlinternal.MySQLConfig{
		AdminUser:     mysql.Spec.AdminUser,
		AdminPassword: mysql.Spec.AdminPassword,
		Host:          mysql.Spec.Host,
	}
	mysqlClient, err := r.MySQLClientFactory(cfg)
	if err != nil {
		return err
	}
	err = mysqlClient.Ping()
	if err != nil {
		return err
	}

	defer mysqlClient.Close()

	err = mysqlClient.Exec("DROP USER IF EXISTS '" + mysqlUser.ObjectMeta.Name + "'@'%';")
	if err != nil {
		log.Error(err, "Failed to drop MySQL user.", "mysqlUser", mysqlUser.ObjectMeta.Name)
		return err
	}
	metrics.MysqlUserDeletedTotal.Increment()

	log.Info("Successfully finalized mysqlUser")
	return nil
}

func getSecretName(mysqlName string, mysqlUserName string) string {
	str := []string{"mysql", mysqlName, mysqlUserName}
	return strings.Join(str, "-")
}

func (r *MySQLUserReconciler) createSecret(ctx context.Context, log logr.Logger, password string, secretName string, namespace string, mysqlUser *mysqlv1alpha1.MySQLUser) error {
	data := make(map[string][]byte)
	data["password"] = []byte(password)
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
	}
	err := ctrl.SetControllerReference(mysqlUser, secret, r.Scheme) // Set owner of this Secret
	if err != nil {
		log.Error(err, "Failed to SetControllerReference for Secret.")
		return err
	}
	if _, err := ctrl.CreateOrUpdate(ctx, r.GetClient(), secret, func() error {
		secret.Data = data
		log.Info("Successfully created Secret.")
		return nil
	}); err != nil {
		log.Error(err, "Error creating or updating Secret.")
		return err
	}
	return nil
}

func (r *MySQLUserReconciler) ifOwnerReferencesContains(ownerReferences []metav1.OwnerReference, mysql *mysqlv1alpha1.MySQL) bool {
	for _, ref := range ownerReferences {
		if ref.APIVersion == "mysql.nakamasato.com/v1alpha1" && ref.Kind == "MySQL" && ref.UID == mysql.UID {
			return true
		}
	}
	return false
}
