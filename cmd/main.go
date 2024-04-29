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

package main

import (
	"context"
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	"go.uber.org/zap/zapcore"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	mysqlv1alpha1 "github.com/nakamasato/mysql-operator/api/v1alpha1"
	controllers "github.com/nakamasato/mysql-operator/internal/controller"
	"github.com/nakamasato/mysql-operator/internal/mysql"
	"github.com/nakamasato/mysql-operator/internal/secret"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(mysqlv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var adminUserSecretType string
	var projectId string
	var secretNamespace string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&adminUserSecretType, "cloud-secret-manager", "",
		"The cloud secret manager to get credentials from. "+
			"Currently, only support gcp")
	flag.StringVar(&projectId, "gcp-project-id", "",
		"GCP project id. Set this value to use adminUserSecretType=gcp. "+
			"Also can be set by environment variable PROJECT_ID."+
			"If both are set, the flag is used.")
	flag.StringVar(&secretNamespace, "k8s-secret-namespace", "",
		"Kubernetes namespace where MYSQL credentials secrets is located. Set this value to use adminUserSecretType=k8s. "+
			"Also can be set by environment variable SECRET_NAMESPACE."+
			"If both are set, the flag is used.")
	opts := zap.Options{
		Development: true,
		TimeEncoder: zapcore.ISO8601TimeEncoder,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "dfc6d3c2.nakamasato.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	mysqlClients := mysql.MySQLClients{}

	if err = (&controllers.MySQLUserReconciler{
		Client:       mgr.GetClient(),
		Scheme:       mgr.GetScheme(),
		MySQLClients: mysqlClients,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "MySQLUser")
		os.Exit(1)
	}

	ctx := context.Background()
	secretManagers := map[string]secret.SecretManager{
		"raw": secret.RawSecretManager{},
	}
	switch adminUserSecretType {
	case "gcp":
		if projectId == "" {
			projectId = os.Getenv("PROJECT_ID")
		}
		gcpSecretManager, err := secret.NewGCPSecretManager(ctx, projectId)
		if err != nil {
			setupLog.Error(err, "failed to initialize GCPSecretManager")
			os.Exit(1)
		}
		defer gcpSecretManager.Close()
		setupLog.Info("Initialized gcpSecretManager", "projectId", projectId)
		secretManagers["gcp"] = gcpSecretManager
	case "k8s":
		if secretNamespace == "" {
			secretNamespace = os.Getenv("SECRET_NAMESPACE")
		}
		k8sSecretManager, err := secret.Newk8sSecretManager(ctx, secretNamespace, mgr.GetClient())
		if err != nil {
			setupLog.Error(err, "failed to initialize k8sSecretManager")
			os.Exit(1)
		}
		setupLog.Info("Initialized k8sSecretManager", "namespace", secretNamespace)
		secretManagers["k8s"] = k8sSecretManager
	}
	if err = (&controllers.MySQLReconciler{
		Client:          mgr.GetClient(),
		Scheme:          mgr.GetScheme(),
		MySQLClients:    mysqlClients,
		MySQLDriverName: "mysql",
		SecretManagers:  secretManagers,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "MySQL")
		os.Exit(1)
	}
	if err = (&controllers.MySQLDBReconciler{
		Client:       mgr.GetClient(),
		Scheme:       mgr.GetScheme(),
		MySQLClients: mysqlClients,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "MySQLDB")
		os.Exit(1)
	}

	// Set index for mysqluser with spec.mysqlName
	// this is necessary to get MySQLUser/MySQLDB that references a MySQL
	cache := mgr.GetCache()
	indexFunc := func(obj client.Object) []string {
		return []string{obj.(*mysqlv1alpha1.MySQLUser).Spec.MysqlName}
	}
	if err := cache.IndexField(context.TODO(), &mysqlv1alpha1.MySQLUser{}, "spec.mysqlName", indexFunc); err != nil {
		panic(err)
	}
	indexFunc = func(obj client.Object) []string {
		return []string{obj.(*mysqlv1alpha1.MySQLDB).Spec.MysqlName}
	}
	if err := cache.IndexField(context.TODO(), &mysqlv1alpha1.MySQLDB{}, "spec.mysqlName", indexFunc); err != nil {
		panic(err)
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
