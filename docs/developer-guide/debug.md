# Debug

## Errors

### Server rejected event

Fails:

```bash
[manager] 2021-12-25T04:16:33.293Z      INFO    controller-runtime.manager.controller.mysqluser Fetch MySQLUser instance. MySQLUser resource found.   {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default", "name": "nakamasato", "mysqlUser.Namespace": "default"}
[manager] 2021-12-25T04:16:33.293Z      INFO    controller-runtime.manager.controller.mysqluser Fetched MySQL instance.       {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-25T04:16:33.293Z      INFO    controller-runtime.manager.controller.mysqluser started mysqlClient.Ping()    {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-25T04:16:33.305Z      ERROR   controller-runtime.manager.controller.mysqluser Failed to connect to MySQL.   {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default", "mysqlName": "mysql-sample", "error": "dial tcp: lookup mysql.default on 10.96.0.10:53: no such host"}
[manager] sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).reconcileHandler
[manager]       /go/pkg/mod/sigs.k8s.io/controller-runtime@v0.9.2/pkg/internal/controller/controller.go:298
[manager] sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).processNextWorkItem
[manager]       /go/pkg/mod/sigs.k8s.io/controller-runtime@v0.9.2/pkg/internal/controller/controller.go:253
[manager] sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).Start.func2.2
[manager]       /go/pkg/mod/sigs.k8s.io/controller-runtime@v0.9.2/pkg/internal/controller/controller.go:214
[manager] 2021-12-25T04:16:33.306Z      DEBUG   controller-runtime.manager.events       Warning {"object": {"kind":"MySQLUser","namespace":"default","name":"nakamasato","uid":"947ff02f-1950-4f37-9dcd-0e2d0d11db84","apiVersion":"mysql.nakamasato.com/v1alpha1","resourceVersion":"572560"}, "reason": "ProcessingError", "message": "dial tcp: lookup mysql.default on 10.96.0.10:53: no such host"}
[manager] E1225 04:16:33.315173       1 event.go:264] Server rejected event '&v1.Event{TypeMeta:v1.TypeMeta{Kind:"", APIVersion:""}, ObjectMeta:v1.ObjectMeta{Name:"nakamasato.16c3e45a191f6bd4", GenerateName:"", Namespace:"default", SelfLink:"", UID:"", ResourceVersion:"", Generation:0, CreationTimestamp:v1.Time{Time:time.Time{wall:0x0, ext:0, loc:(*time.Location)(nil)}}, DeletionTimestamp:(*v1.Time)(nil), DeletionGracePeriodSeconds:(*int64)(nil), Labels:map[string]string(nil), Annotations:map[string]string(nil), OwnerReferences:[]v1.OwnerReference(nil), Finalizers:[]string(nil), ClusterName:"", ManagedFields:[]v1.ManagedFieldsEntry(nil)}, InvolvedObject:v1.ObjectReference{Kind:"MySQLUser", Namespace:"default", Name:"nakamasato", UID:"947ff02f-1950-4f37-9dcd-0e2d0d11db84", APIVersion:"mysql.nakamasato.com/v1alpha1", ResourceVersion:"572560", FieldPath:""}, Reason:"ProcessingError", Message:"dial tcp: lookup mysql.default on 10.96.0.10:53: no such host", Source:v1.EventSource{Component:"mysqluser_controller", Host:""}, FirstTimestamp:v1.Time{Time:time.Time{wall:0xc069c4a1366643d4, ext:64615291501, loc:(*time.Location)(0x25472e0)}}, LastTimestamp:v1.Time{Time:time.Time{wall:0xc069c4a8523b024c, ext:93043659801, loc:(*time.Location)(0x25472e0)}}, Count:15, Type:"Warning", EventTime:v1.MicroTime{Time:time.Time{wall:0x0, ext:0, loc:(*time.Location)(nil)}}, Series:(*v1.EventSeries)(nil), Action:"", Related:(*v1.ObjectReference)(nil), ReportingController:"", ReportingInstance:""}': 'events "nakamasato.16c3e45a191f6bd4" is forbidden: User "system:serviceaccount:mysql-operator-system:mysql-operator-controller-manager" cannot patch resource "events" in API group "" in the namespace "default"' (will not retry!)
[manager] 2021-12-25T04:16:33.329Z      ERROR   controller-runtime.manager.controller.mysqluser Reconciler error     {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default", "error": "dial tcp: lookup mysql.default on 10.96.0.10:53: no such host"}
[manager] sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).processNextWorkItem
[manager]       /go/pkg/mod/sigs.k8s.io/controller-runtime@v0.9.2/pkg/internal/controller/controller.go:253
[manager] sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).Start.func2.2
[manager]       /go/pkg/mod/sigs.k8s.io/controller-runtime@v0.9.2/pkg/internal/controller/controller.go:214
[manager] 2021-12-25T04:16:33.329Z      INFO    controller-runtime.manager.controller.mysqluser Fetch MySQLUser instance. MySQLUser resource found.   {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default", "name": "nakamasato", "mysqlUser.Namespace": "default"}
[manager] 2021-12-25T04:16:33.330Z      INFO    controller-runtime.manager.controller.mysqluser Fetched MySQL instance.       {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-25T04:16:33.330Z      INFO    controller-runtime.manager.controller.mysqluser started mysqlClient.Ping()    {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-25T04:16:33.360Z      ERROR   controller-runtime.manager.controller.mysqluser Failed to connect to MySQL.   {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default", "mysqlName": "mysql-sample", "error": "dial tcp: lookup mysql.default on 10.96.0.10:53: no such host"}
[manager] sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).reconcileHandler
[manager]       /go/pkg/mod/sigs.k8s.io/controller-runtime@v0.9.2/pkg/internal/controller/controller.go:298
[manager] sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).processNextWorkItem
[manager]       /go/pkg/mod/sigs.k8s.io/controller-runtime@v0.9.2/pkg/internal/controller/controller.go:253
[manager] sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).Start.func2.2
[manager]       /go/pkg/mod/sigs.k8s.io/controller-runtime@v0.9.2/pkg/internal/controller/controller.go:214
[manager] 2021-12-25T04:16:33.361Z      DEBUG   controller-runtime.manager.events       Warning {"object": {"kind":"MySQLUser","namespace":"default","name":"nakamasato","uid":"947ff02f-1950-4f37-9dcd-0e2d0d11db84","apiVersion":"mysql.nakamasato.com/v1alpha1","resourceVersion":"572612"}, "reason": "ProcessingError", "message": "dial tcp: lookup mysql.default on 10.96.0.10:53: no such host"}
[manager] E1225 04:16:33.366899       1 event.go:264] Server rejected event '&v1.Event{TypeMeta:v1.TypeMeta{Kind:"", APIVersion:""}, ObjectMeta:v1.ObjectMeta{Name:"nakamasato.16c3e45a191f6bd4", GenerateName:"", Namespace:"default", SelfLink:"", UID:"", ResourceVersion:"", Generation:0, CreationTimestamp:v1.Time{Time:time.Time{wall:0x0, ext:0, loc:(*time.Location)(nil)}}, DeletionTimestamp:(*v1.Time)(nil), DeletionGracePeriodSeconds:(*int64)(nil), Labels:map[string]string(nil), Annotations:map[string]string(nil), OwnerReferences:[]v1.OwnerReference(nil), Finalizers:[]string(nil), ClusterName:"", ManagedFields:[]v1.ManagedFieldsEntry(nil)}, InvolvedObject:v1.ObjectReference{Kind:"MySQLUser", Namespace:"default", Name:"nakamasato", UID:"947ff02f-1950-4f37-9dcd-0e2d0d11db84", APIVersion:"mysql.nakamasato.com/v1alpha1", ResourceVersion:"572612", FieldPath:""}, Reason:"ProcessingError", Message:"dial tcp: lookup mysql.default on 10.96.0.10:53: no such host", Source:v1.EventSource{Component:"mysqluser_controller", Host:""}, FirstTimestamp:v1.Time{Time:time.Time{wall:0xc069c4a1366643d4, ext:64615291501, loc:(*time.Location)(0x25472e0)}}, LastTimestamp:v1.Time{Time:time.Time{wall:0xc069c4a855820c0c, ext:93098644501, loc:(*time.Location)(0x25472e0)}}, Count:16, Type:"Warning", EventTime:v1.MicroTime{Time:time.Time{wall:0x0, ext:0, loc:(*time.Location)(nil)}}, Series:(*v1.EventSeries)(nil), Action:"", Related:(*v1.ObjectReference)(nil), ReportingController:"", ReportingInstance:""}': 'events "nakamasato.16c3e45a191f6bd4" is forbidden: User "system:serviceaccount:mysql-operator-system:mysql-operator-controller-manager" cannot patch resource "events" in API group "" in the namespace "default"' (will not retry!)
[manager] 2021-12-25T04:16:33.371Z      ERROR   controller-runtime.manager.controller.mysqluser Reconciler error     {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default", "error": "dial tcp: lookup mysql.default on 10.96.0.10:53: no such host"}
[manager] sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).processNextWorkItem
[manager]       /go/pkg/mod/sigs.k8s.io/controller-runtime@v0.9.2/pkg/internal/controller/controller.go:253
[manager] sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).Start.func2.2
[manager]       /go/pkg/mod/sigs.k8s.io/controller-runtime@v0.9.2/pkg/internal/controller/controller.go:214
```

Recovers next reconciliation loop (1min30seconds after the previous loop):

```bash
[manager] 2021-12-25T04:17:55.179Z      INFO    controller-runtime.manager.controller.mysqluser Fetch MySQLUser instance. MySQLUser resource found.   {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default", "name": "nakamasato", "mysqlUser.Namespace": "default"}
[manager] 2021-12-25T04:17:55.179Z      INFO    controller-runtime.manager.controller.mysqluser Fetched MySQL instance.       {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-25T04:17:55.179Z      INFO    controller-runtime.manager.controller.mysqluser started mysqlClient.Ping()    {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-25T04:17:55.183Z      INFO    controller-runtime.manager.controller.mysqluser Successfully created mysqlClient      {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-25T04:17:55.294Z      INFO    controller-runtime.manager.controller.mysqluser Generate new password for Secret      {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default", "secretName": "mysql-mysql-sample-nakamasato"}
[manager] 2021-12-25T04:17:55.294Z      INFO    controller-runtime.manager.controller.mysqluser Create MySQL user if not.     {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default", "name": "nakamasato", "mysqlUser.Namespace": "default"}
[manager] 2021-12-25T04:17:55.297Z      INFO    controller-runtime.manager.controller.mysqluser Successfully created Secret.  {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-25T04:17:55.320Z      INFO    controller-runtime.manager.controller.mysqluser Fetch MySQLUser instance. MySQLUser resource found.   {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default", "name": "nakamasato", "mysqlUser.Namespace": "default"}
[manager] 2021-12-25T04:17:55.320Z      INFO    controller-runtime.manager.controller.mysqluser Fetched MySQL instance.       {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-25T04:17:55.320Z      INFO    controller-runtime.manager.controller.mysqluser started mysqlClient.Ping()    {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-25T04:17:55.329Z      INFO    controller-runtime.manager.controller.mysqluser Successfully created mysqlClient      {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-25T04:17:55.329Z      INFO    controller-runtime.manager.controller.mysqluser Create MySQL user if not.     {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default", "name": "nakamasato", "mysqlUser.Namespace": "default"}
[manager] 2021-12-25T04:17:55.367Z      INFO    controller-runtime.manager.controller.mysqluser Successfully created Secret.  {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-25T04:17:55.376Z      ERROR   util.api        unable to update status {"error": "Operation cannot be fulfilled on mysqlusers.mysql.nakamasato.com \"nakamasato\": the object has been modified; please apply your changes to the latest version and try again"}
[manager] github.com/redhat-cop/operator-utils/pkg/util.(*ReconcilerBase).ManageSuccessWithRequeue
[manager]       /go/pkg/mod/github.com/redhat-cop/operator-utils@v1.2.2/pkg/util/reconciler.go:423
[manager] github.com/redhat-cop/operator-utils/pkg/util.(*ReconcilerBase).ManageSuccess
[manager]       /go/pkg/mod/github.com/redhat-cop/operator-utils@v1.2.2/pkg/util/reconciler.go:434
[manager] github.com/nakamasato/mysql-operator/controllers.(*MySQLUserReconciler).Reconcile
[manager]       /workspace/controllers/mysqluser_controller.go:184
[manager] sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).reconcileHandler
[manager]       /go/pkg/mod/sigs.k8s.io/controller-runtime@v0.9.2/pkg/internal/controller/controller.go:298
[manager] sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).processNextWorkItem
[manager]       /go/pkg/mod/sigs.k8s.io/controller-runtime@v0.9.2/pkg/internal/controller/controller.go:253
[manager] sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).Start.func2.2
[manager]       /go/pkg/mod/sigs.k8s.io/controller-runtime@v0.9.2/pkg/internal/controller/controller.go:214
[manager] 2021-12-25T04:17:55.379Z      ERROR   controller-runtime.manager.controller.mysqluser Reconciler error     {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default", "error": "Operation cannot be fulfilled on mysqlusers.mysql.nakamasato.com \"nakamasato\": the object has been modified; please apply your changes to the latest version and try again"}
[manager] sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).processNextWorkItem
[manager]       /go/pkg/mod/sigs.k8s.io/controller-runtime@v0.9.2/pkg/internal/controller/controller.go:253
[manager] sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).Start.func2.2
[manager]       /go/pkg/mod/sigs.k8s.io/controller-runtime@v0.9.2/pkg/internal/controller/controller.go:214
[manager] 2021-12-25T04:17:55.379Z      INFO    controller-runtime.manager.controller.mysqluser Fetch MySQLUser instance. MySQLUser resource found.   {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default", "name": "nakamasato", "mysqlUser.Namespace": "default"}
[manager] 2021-12-25T04:17:55.379Z      INFO    controller-runtime.manager.controller.mysqluser Fetched MySQL instance.       {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-25T04:17:55.379Z      INFO    controller-runtime.manager.controller.mysqluser started mysqlClient.Ping()    {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-25T04:17:55.385Z      INFO    controller-runtime.manager.controller.mysqluser Successfully created mysqlClient      {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-25T04:17:55.385Z      INFO    controller-runtime.manager.controller.mysqluser Create MySQL user if not.     {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default", "name": "nakamasato", "mysqlUser.Namespace": "default"}
[manager] 2021-12-25T04:17:55.389Z      INFO    controller-runtime.manager.controller.mysqluser Successfully created Secret.  {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-25T04:17:55.399Z      INFO    controller-runtime.manager.controller.mysqluser Fetch MySQLUser instance. MySQLUser resource found.   {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default", "name": "nakamasato", "mysqlUser.Namespace": "default"}
[manager] 2021-12-25T04:17:55.399Z      INFO    controller-runtime.manager.controller.mysqluser Fetched MySQL instance.       {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-25T04:17:55.399Z      INFO    controller-runtime.manager.controller.mysqluser started mysqlClient.Ping()    {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-25T04:17:55.475Z      INFO    controller-runtime.manager.controller.mysqluser Successfully created mysqlClient      {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-25T04:17:55.475Z      INFO    controller-runtime.manager.controller.mysqluser Create MySQL user if not.     {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default", "name": "nakamasato", "mysqlUser.Namespace": "default"}
[manager] 2021-12-25T04:17:55.477Z      INFO    controller-runtime.manager.controller.mysqluser Successfully created Secret.  {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
```

1. Reproduce steps:
    1. `skaffold dev`
    1. `kubectl delete -k config/mysql`
    1. `kubectl apply -k config/samples-on-k8s`
    1. `kubectl apply -k config/mysql`

    https://pkg.go.dev/github.com/redhat-cop/operator-utils@v1.1.4/pkg/util#ReconcilerBase.ManageSuccessWithRequeue

1. Cause:

    ```bash
    [manager] E1225 04:16:33.315173       1 event.go:264] Server rejected event '&v1.Event{TypeMeta:v1.TypeMeta{Kind:"", APIVersion:""}, ObjectMeta:v1.ObjectMeta{Name:"nakamasato.16c3e45a191f6bd4", GenerateName:"", Namespace:"default", SelfLink:"", UID:"", ResourceVersion:"", Generation:0, CreationTimestamp:v1.Time{Time:time.Time{wall:0x0, ext:0, loc:(*time.Location)(nil)}}, DeletionTimestamp:(*v1.Time)(nil), DeletionGracePeriodSeconds:(*int64)(nil), Labels:map[string]string(nil), Annotations:map[string]string(nil), OwnerReferences:[]v1.OwnerReference(nil), Finalizers:[]string(nil), ClusterName:"", ManagedFields:[]v1.ManagedFieldsEntry(nil)}, InvolvedObject:v1.ObjectReference{Kind:"MySQLUser", Namespace:"default", Name:"nakamasato", UID:"947ff02f-1950-4f37-9dcd-0e2d0d11db84", APIVersion:"mysql.nakamasato.com/v1alpha1", ResourceVersion:"572560", FieldPath:""}, Reason:"ProcessingError", Message:"dial tcp: lookup mysql.default on 10.96.0.10:53: no such host", Source:v1.EventSource{Component:"mysqluser_controller", Host:""}, FirstTimestamp:v1.Time{Time:time.Time{wall:0xc069c4a1366643d4, ext:64615291501, loc:(*time.Location)(0x25472e0)}}, LastTimestamp:v1.Time{Time:time.Time{wall:0xc069c4a8523b024c, ext:93043659801, loc:(*time.Location)(0x25472e0)}}, Count:15, Type:"Warning", EventTime:v1.MicroTime{Time:time.Time{wall:0x0, ext:0, loc:(*time.Location)(nil)}}, Series:(*v1.EventSeries)(nil), Action:"", Related:(*v1.ObjectReference)(nil), ReportingController:"", ReportingInstance:""}': 'events "nakamasato.16c3e45a191f6bd4" is forbidden: User "system:serviceaccount:mysql-operator-system:mysql-operator-controller-manager" cannot patch resource "events" in API group "" in the namespace "default"' (will not retry!)
    ```

1. Reason: `User "system:serviceaccount:mysql-operator-system:mysql-operator-controller-manager" cannot patch resource "events" in API group "" in the namespace "default"'`

1. Solution: Add `//+kubebuilder:rbac:groups=core,resources=events,verbs=create;update;patch` to the controller. https://github.com/nakamasato/mysql-operator/pull/87

1. After the fix:

    ```
    kubectl get event --field-selector involvedObject.kind=MySQLUser
    LAST SEEN   TYPE      REASON            OBJECT                 MESSAGE
    23s         Warning   ProcessingError   mysqluser/nakamasato   dial tcp: lookup mysql.default on 10.96.0.10:53: no such host
    ```

### dial tcp: lookup mysql.default on 10.96.0.10:53: no such host (it takes 1:30 to create Secret since MySQL recovers.)

```bash
[manager] 2021-12-26T23:16:21.480Z      INFO    controller-runtime.manager.controller.mysqluser Fetch MySQLUser instance. MySQLUser resource found. {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default", "name": "nakamasato", "mysqlUser.Namespace": "default"}
[manager] 2021-12-26T23:16:21.481Z      INFO    controller-runtime.manager.controller.mysqluser Fetched MySQL instance. {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-26T23:16:21.481Z      INFO    controller-runtime.manager.controller.mysqluser started mysqlClient.Ping()      {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-26T23:16:21.492Z      ERROR   controller-runtime.manager.controller.mysqluser Failed to connect to MySQL.     {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default", "mysqlName": "mysql-sample", "error": "dial tcp: lookup mysql.default on 10.96.0.10:53: no such host"}
[manager] sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).reconcileHandler
[manager]       /go/pkg/mod/sigs.k8s.io/controller-runtime@v0.9.2/pkg/internal/controller/controller.go:298
[manager] sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).processNextWorkItem
[manager]       /go/pkg/mod/sigs.k8s.io/controller-runtime@v0.9.2/pkg/internal/controller/controller.go:253
[manager] sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).Start.func2.2
[manager]       /go/pkg/mod/sigs.k8s.io/controller-runtime@v0.9.2/pkg/internal/controller/controller.go:214
[manager] 2021-12-26T23:16:21.492Z      DEBUG   controller-runtime.manager.events       Warning {"object": {"kind":"MySQLUser","namespace":"default","name":"nakamasato","uid":"2d7b8b4a-fa11-4558-aaf2-1d4eb9edfa95","apiVersion":"mysql.nakamasato.com/v1alpha1","resourceVersion":"1134"}, "reason": "ProcessingError", "message": "dial tcp: lookup mysql.default on 10.96.0.10:53: no such host"}
[manager] 2021-12-26T23:16:21.503Z      ERROR   controller-runtime.manager.controller.mysqluser Reconciler error        {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default", "error": "dial tcp: lookup mysql.default on 10.96.0.10:53: no such host"}
[manager] sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).processNextWorkItem
[manager]       /go/pkg/mod/sigs.k8s.io/controller-runtime@v0.9.2/pkg/internal/controller/controller.go:253
[manager] sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).Start.func2.2
[manager]       /go/pkg/mod/sigs.k8s.io/controller-runtime@v0.9.2/pkg/internal/controller/controller.go:214
[manager] 2021-12-26T23:16:21.503Z      INFO    controller-runtime.manager.controller.mysqluser Fetch MySQLUser instance. MySQLUser resource found. {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default", "name": "nakamasato", "mysqlUser.Namespace": "default"}
[manager] 2021-12-26T23:16:21.503Z      INFO    controller-runtime.manager.controller.mysqluser Fetched MySQL instance. {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-26T23:16:21.503Z      INFO    controller-runtime.manager.controller.mysqluser started mysqlClient.Ping()      {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-26T23:16:21.517Z      ERROR   controller-runtime.manager.controller.mysqluser Failed to connect to MySQL.     {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default", "mysqlName": "mysql-sample", "error": "dial tcp: lookup mysql.default on 10.96.0.10:53: no such host"}
[manager] sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).reconcileHandler
[manager]       /go/pkg/mod/sigs.k8s.io/controller-runtime@v0.9.2/pkg/internal/controller/controller.go:298
[manager] sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).processNextWorkItem
[manager]       /go/pkg/mod/sigs.k8s.io/controller-runtime@v0.9.2/pkg/internal/controller/controller.go:253
[manager] sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).Start.func2.2
[manager]       /go/pkg/mod/sigs.k8s.io/controller-runtime@v0.9.2/pkg/internal/controller/controller.go:214
[manager] 2021-12-26T23:16:21.518Z      DEBUG   controller-runtime.manager.events       Warning {"object": {"kind":"MySQLUser","namespace":"default","name":"nakamasato","uid":"2d7b8b4a-fa11-4558-aaf2-1d4eb9edfa95","apiVersion":"mysql.nakamasato.com/v1alpha1","resourceVersion":"1211"}, "reason": "ProcessingError", "message": "dial tcp: lookup mysql.default on 10.96.0.10:53: no such host"}
[manager] 2021-12-26T23:16:21.538Z      ERROR   controller-runtime.manager.controller.mysqluser Reconciler error        {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default", "error": "dial tcp: lookup mysql.default on 10.96.0.10:53: no such host"}
[manager] sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).processNextWorkItem
[manager]       /go/pkg/mod/sigs.k8s.io/controller-runtime@v0.9.2/pkg/internal/controller/controller.go:253
[manager] sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).Start.func2.2
[manager]       /go/pkg/mod/sigs.k8s.io/controller-runtime@v0.9.2/pkg/internal/controller/controller.go:214
[manager] 2021-12-26T23:17:43.284Z      INFO    controller-runtime.manager.controller.mysqluser Fetch MySQLUser instance. MySQLUser resource found. {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default", "name": "nakamasato", "mysqlUser.Namespace": "default"}
[manager] 2021-12-26T23:17:43.284Z      INFO    controller-runtime.manager.controller.mysqluser Fetched MySQL instance. {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-26T23:17:43.284Z      INFO    controller-runtime.manager.controller.mysqluser started mysqlClient.Ping()      {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-26T23:17:43.289Z      INFO    controller-runtime.manager.controller.mysqluser Successfully created mysqlClient        {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-26T23:17:43.402Z      INFO    controller-runtime.manager.controller.mysqluser Generate new password for Secret        {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default", "secretName": "mysql-mysql-sample-nakamasato"}
[manager] 2021-12-26T23:17:43.402Z      INFO    controller-runtime.manager.controller.mysqluser Create MySQL user if not.       {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default", "name": "nakamasato", "mysqlUser.Namespace": "default"}
[manager] 2021-12-26T23:17:43.404Z      INFO    controller-runtime.manager.controller.mysqluser Successfully created Secret.    {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-26T23:17:43.422Z      INFO    controller-runtime.manager.controller.mysqluser Fetch MySQLUser instance. MySQLUser resource found. {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default", "name": "nakamasato", "mysqlUser.Namespace": "default"}
[manager] 2021-12-26T23:17:43.422Z      INFO    controller-runtime.manager.controller.mysqluser Fetched MySQL instance. {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-26T23:17:43.422Z      INFO    controller-runtime.manager.controller.mysqluser started mysqlClient.Ping()      {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-26T23:17:43.427Z      INFO    controller-runtime.manager.controller.mysqluser Successfully created mysqlClient        {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
[manager] 2021-12-26T23:17:43.427Z      INFO    controller-runtime.manager.controller.mysqluser Create MySQL user if not.       {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default", "name": "nakamasato", "mysqlUser.Namespace": "default"}
[manager] 2021-12-26T23:17:43.431Z      INFO    controller-runtime.manager.controller.mysqluser Successfully created Secret.    {"reconciler group": "mysql.nakamasato.com", "reconciler kind": "MySQLUser", "name": "nakamasato", "namespace": "default"}
```

1. Reproduce steps:
    1. `skaffold dev`
    1. `kubectl delete -k config/mysql`
    1. `kubectl apply -k config/samples-on-k8s`
    1. `kubectl apply -k config/mysql`
