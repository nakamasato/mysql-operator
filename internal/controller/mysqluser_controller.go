/*
Copyright 2023.

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
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	_ "github.com/go-sql-driver/mysql"

	mysqlv1alpha1 "github.com/nakamasato/mysql-operator/api/v1alpha1"
	"github.com/nakamasato/mysql-operator/internal/metrics"
	mysqlinternal "github.com/nakamasato/mysql-operator/internal/mysql"
	"github.com/nakamasato/mysql-operator/internal/utils"
)

const (
	mysqlUserFinalizer                     = "mysqluser.nakamasato.com/finalizer"
	mysqlUserReasonCompleted               = "Both secret and mysql user are successfully created."
	mysqlUserReasonMySQLConnectionFailed   = "Failed to connect to mysql"
	mysqlUserReasonMySQLFailedToCreateUser = "Failed to create MySQL user"
	mysqlUserReasonMySQLFetchFailed        = "Failed to fetch MySQL"
	mysqlUserPhaseReady                    = "Ready"
	mysqlUserPhaseNotReady                 = "NotReady"
)

// MySQLUserReconciler reconciles a MySQLUser object
type MySQLUserReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	MySQLClients mysqlinternal.MySQLClients
}

//+kubebuilder:rbac:groups=mysql.nakamasato.com,resources=mysqlusers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mysql.nakamasato.com,resources=mysqlusers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mysql.nakamasato.com,resources=mysqlusers/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;update;patch

// Reconcile function is responsible for managing MySQLUser.
// Create MySQL user if not exist in the target MySQL, and drop it
// if the corresponding object is deleted.
func (r *MySQLUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithName("MySQLUserReconciler")

	// Fetch MySQLUser
	mysqlUser := &mysqlv1alpha1.MySQLUser{}
	err := r.Get(ctx, req.NamespacedName, mysqlUser)
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
	mysqlUserGrants := mysqlUser.Spec.Grants

	// Fetch MySQL
	mysql := &mysqlv1alpha1.MySQL{}
	var mysqlNamespacedName = client.ObjectKey{Namespace: req.Namespace, Name: mysqlUser.Spec.MysqlName}
	if err := r.Get(ctx, mysqlNamespacedName, mysql); err != nil {
		log.Error(err, "[FetchMySQL] Failed")
		mysqlUser.Status.Phase = mysqlUserPhaseNotReady
		mysqlUser.Status.Reason = mysqlUserReasonMySQLFetchFailed
		if serr := r.Status().Update(ctx, mysqlUser); serr != nil {
			log.Error(serr, "Failed to update mysqluser status", "mysqlUser", mysqlUser.Name)
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Info("[FetchMySQL] Found")

	// SetOwnerReference if not exists
	if !r.ifOwnerReferencesContains(mysqlUser.ObjectMeta.OwnerReferences, mysql) {
		err := controllerutil.SetControllerReference(mysql, mysqlUser, r.Scheme)
		if err != nil {
			return ctrl.Result{}, err //requeue
		}
		err = r.Update(ctx, mysqlUser)
		if err != nil {
			return ctrl.Result{}, err //requeue
		}
	}

	// Get MySQL client
	mysqlClient, err := r.MySQLClients.GetClient(mysql.GetKey())
	if err != nil {
		log.Error(err, "Failed to get MySQL client", "key", mysql.GetKey())
		return ctrl.Result{}, err
	}

	if err != nil {
		mysqlUser.Status.Phase = mysqlUserPhaseNotReady
		mysqlUser.Status.Reason = mysqlUserReasonMySQLConnectionFailed
		log.Error(err, "[MySQLClient] Failed to connect to MySQL", "mysqlName", mysqlName)
		if serr := r.Status().Update(ctx, mysqlUser); serr != nil {
			log.Error(serr, "Failed to update mysqluser status", "mysqlUser", mysqlUser.Name)
			return ctrl.Result{RequeueAfter: time.Second}, nil
		}
		return ctrl.Result{RequeueAfter: time.Second}, nil // requeue after 1 second
	}
	log.Info("[MySQLClient] Successfully connected")

	// Finalize if DeletionTimestamp exists
	if !mysqlUser.GetDeletionTimestamp().IsZero() {
		log.Info("isMysqlUserMarkedToBeDeleted is true")
		if controllerutil.ContainsFinalizer(mysqlUser, mysqlUserFinalizer) {
			log.Info("ContainsFinalizer is true")
			// Run finalization logic for mysqlUserFinalizer. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if err := r.finalizeMySQLUser(ctx, mysqlClient, mysqlUser); err != nil {
				log.Error(err, "Failed to complete finalizeMySQLUser")
				return ctrl.Result{}, err
			}
			log.Info("finalizeMySQLUser completed")
			// Remove mysqlUserFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			log.Info("removing finalizer")
			if controllerutil.RemoveFinalizer(mysqlUser, mysqlUserFinalizer) {
				log.Info("RemoveFinalizer completed")
				err := r.Update(ctx, mysqlUser)
				log.Info("Update")
				if err != nil {
					log.Error(err, "Failed to update mysqlUser")
					return ctrl.Result{}, err
				}
				log.Info("Update completed")
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, nil // should return success when not having the finalizer
	}

	// Add finalizer for this CR
	log.Info("Add Finalizer for this CR")
	if controllerutil.AddFinalizer(mysqlUser, mysqlUserFinalizer) {
		log.Info("Added Finalizer")
		err = r.Update(ctx, mysqlUser)
		if err != nil {
			log.Info("Failed to update after adding finalizer")
			return ctrl.Result{}, err // requeue
		}
		log.Info("Updated successfully after adding finalizer")
	} else {
		log.Info("already has finalizer")
	}

	// Skip all the following steps if MySQL is being Deleted
	if !mysql.GetDeletionTimestamp().IsZero() {
		log.Info("MySQL is being deleted. MySQLUser cannot be created.", "mysql", mysql.Name, "mysqlUser", mysqlUser.Name)
		return ctrl.Result{}, err
	}

	// Get password from Secret if exists. Otherwise, generate new one.
	secretName := getSecretName(mysqlName, mysqlUserName)
	secret := &v1.Secret{}
	err = r.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: secretName}, secret)
	var password string
	if err != nil {
		if errors.IsNotFound(err) { // Secret doesn't exists -> generate password
			log.Info("[password] Generate new password for Secret", "secretName", secretName)
			password = utils.GenerateRandomString(16)
		} else {
			log.Error(err, "[password] Failed to get Secret", "secretName", secretName)
			return ctrl.Result{}, err // requeue
		}
	} else { // exists -> get password from Secret
		// password = string(secret.Data["password"])
		// TODO: check if the password is valid
		if err := r.updateGrants(ctx, mysqlClient, mysqlUserName, mysqlUser.Spec.Host, mysqlUserGrants); err != nil {
			log.Error(err, "Failed to update grants")
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	// Create MySQL user if not exists with the password set above.
	_, err = mysqlClient.ExecContext(ctx,
		fmt.Sprintf("CREATE USER IF NOT EXISTS '%s'@'%s' IDENTIFIED BY '%s'", mysqlUserName, mysqlUser.Spec.Host, password))
	if err != nil {
		log.Error(err, "[MySQL] Failed to create MySQL user.", "mysqlName", mysqlName, "mysqlUserName", mysqlUserName)
		mysqlUser.Status.Phase = mysqlUserPhaseNotReady
		mysqlUser.Status.Reason = mysqlUserReasonMySQLFailedToCreateUser
		mysqlUser.Status.MySQLUserCreated = false
		if serr := r.Status().Update(ctx, mysqlUser); serr != nil {
			log.Error(serr, "Failed to update mysqluser status", "mysqlUser", mysqlUser.Name)
			return ctrl.Result{RequeueAfter: time.Second}, nil
		}
		return ctrl.Result{RequeueAfter: time.Second}, nil // requeue after 1 second
	}

	if err := r.updateGrants(ctx, mysqlClient, mysqlUserName, mysqlUser.Spec.Host, mysqlUserGrants); err != nil {
		log.Error(err, "Failed to update grants")
		return ctrl.Result{}, err
	}

	log.Info("[MySQL] Created or updated", "name", mysqlUserName, "mysqlUser.Namespace", mysqlUser.Namespace)
	metrics.MysqlUserCreatedTotal.Increment() // TODO: increment only when a user is created
	mysqlUser.Status.Phase = mysqlUserPhaseNotReady
	mysqlUser.Status.Reason = "mysql user is successfully created. Secret is being created."
	mysqlUser.Status.MySQLUserCreated = true

	err = r.createSecret(ctx, password, secretName, mysqlUser.Namespace, mysqlUser)
	// TODO: #35 add test if mysql user is successfully created but secret is failed to create
	if err != nil {
		log.Error(err, "Failed to create secret", "secretName", secretName, "namespace", mysqlUser.Namespace, "mysqlUser", mysqlUser.Name)
		mysqlUser.Status.Reason = "Failed to create Secret"
		mysqlUser.Status.SecretCreated = false
		if serr := r.Status().Update(ctx, mysqlUser); serr != nil {
			log.Error(serr, "Failed to update mysqluser status", "mysqlUser", mysqlUser.Name)
			return ctrl.Result{RequeueAfter: time.Second}, nil
		}
		return ctrl.Result{}, err
	}
	mysqlUser.Status.Phase = mysqlUserPhaseReady
	mysqlUser.Status.Reason = mysqlUserReasonCompleted
	mysqlUser.Status.SecretCreated = true
	if serr := r.Status().Update(ctx, mysqlUser); serr != nil {
		log.Error(serr, "Failed to update mysqluser status", "mysqlUser", mysqlUser.Name)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MySQLUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mysqlv1alpha1.MySQLUser{}).
		Complete(r)
}

// finalizeMySQLUser drops MySQL user
func (r *MySQLUserReconciler) finalizeMySQLUser(ctx context.Context, mysqlClient *sql.DB, mysqlUser *mysqlv1alpha1.MySQLUser) error {
	if mysqlUser.Status.MySQLUserCreated {
		_, err := mysqlClient.ExecContext(ctx, fmt.Sprintf("DROP USER IF EXISTS '%s'@'%s'", mysqlUser.Name, mysqlUser.Spec.Host))
		if err != nil {
			return err
		}
		metrics.MysqlUserDeletedTotal.Increment()
	}

	return nil
}

func getSecretName(mysqlName string, mysqlUserName string) string {
	str := []string{"mysql", mysqlName, mysqlUserName}
	return strings.Join(str, "-")
}

func (r *MySQLUserReconciler) createSecret(ctx context.Context, password string, secretName string, namespace string, mysqlUser *mysqlv1alpha1.MySQLUser) error {
	log := log.FromContext(ctx)
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
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, secret, func() error {
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

func (r *MySQLUserReconciler) updateGrants(ctx context.Context, mysqlClient *sql.DB, mysqlUserName string, mysqlUserHost string, grants []mysqlv1alpha1.Grant) error {
	log := log.FromContext(ctx)
	_, err := mysqlClient.ExecContext(ctx, fmt.Sprintf("REVOKE ALL PRIVILEGES, GRANT OPTION FROM '%s'@'%s';", mysqlUserName, mysqlUserHost))
	if err != nil {
		log.Error(err, "[MySQLUserGrant] Revoke failed")
	}

	for _, grant := range grants {
		_, err := mysqlClient.ExecContext(ctx, fmt.Sprintf("GRANT %s ON %s TO '%s'@'%s'", grant.Privileges, grant.On, mysqlUserName, mysqlUserHost))
		if err != nil {
			return err
		}
		log.Info("[MySQLUserGrant] Grant", "name", mysqlUserName, "grant.Privileges", grant.Privileges, "on", grant.On)
	}
	return nil
}
