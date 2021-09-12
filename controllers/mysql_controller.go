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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	mysqlv1alpha1 "github.com/nakamasato/mysql-user-operator/api/v1alpha1"
)

const mysqlFinalizer = "mysql.nakamasato.com/finalizer"

// MySQLReconciler reconciles a MySQL object
type MySQLReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=mysql.nakamasato.com,resources=mysqls,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mysql.nakamasato.com,resources=mysqls/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mysql.nakamasato.com,resources=mysqls/finalizers,verbs=update

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
	log := log.FromContext(ctx)

	// Fetch MySQL
	mysql := &mysqlv1alpha1.MySQL{}
	err := r.Get(ctx, req.NamespacedName, mysql)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Fetch MySQL instance. MySQL not found.", "mysql.Name", mysql.Name, "mysql.Namespace", mysql.Namespace)
			return ctrl.Result{}, nil
		}

		log.Error(err, "Fetch MySQL instance. Failed to get MySQL.")
		return ctrl.Result{}, err
	}

	// Finalize if DeletionTimestamp exists
	isMysqlUserMarkedToBeDeleted := mysql.GetDeletionTimestamp() != nil
	if isMysqlUserMarkedToBeDeleted {
		log.Info("MySQL instance is marked to be deleted.")
		if controllerutil.ContainsFinalizer(mysql, mysqlFinalizer) {
			// Run finalization logic for mysqlFinalizer. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			referencedNum, err := r.countReferencesByMySQLUser(ctx, log, mysql);
			if err != nil {
				return ctrl.Result{}, err
			}

			// if referencedNum is greater than zero, requeue it.
			if referencedNum > 0 {
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: 0,
				}, nil // https://github.com/operator-framework/operator-sdk/issues/4209#issuecomment-729916367
			}
			// Remove mysqlFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			controllerutil.RemoveFinalizer(mysql, mysqlFinalizer)
			err = r.Update(ctx, mysql)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, err
	}

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(mysql, mysqlFinalizer) {
		controllerutil.AddFinalizer(mysql, mysqlFinalizer)
		err = r.Update(ctx, mysql)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MySQLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mysqlv1alpha1.MySQL{}).
		Complete(r)
}

func (r *MySQLReconciler) countReferencesByMySQLUser(ctx context.Context, log logr.Logger, mysql *mysqlv1alpha1.MySQL) (int, error) {
	// 1. Get the referenced MySQLUser instances.
	// 2. Return the number of referencing MySQLUser.
	mysqlUserList := &mysqlv1alpha1.MySQLUserList{}
	err := r.List(ctx, mysqlUserList)
	mysqlUserCount := 0
	if err != nil {
		return mysqlUserCount, err
	}

	for _, mysqlUser := range mysqlUserList.Items {
		if mysqlUser.Spec.MysqlName == mysql.Name {
			mysqlUserCount++
		}
	}
	if mysqlUserCount == 0 {
		return mysqlUserCount, nil
	} else {
		log.Info("Cannot remove mysql '%s' finalizer as is referenced by %d mysqlUsers", mysql.Name, mysqlUserCount)
		return mysqlUserCount, nil
	}
}
