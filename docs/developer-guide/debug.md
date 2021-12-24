# Debug

## Errors

### Server rejected event

```
[manager] E1224 00:09:17.005902       1 event.go:264] Server rejected event '&v1.Event{TypeMeta:v1.TypeMeta{Kind:"", APIVersion:""}, ObjectMeta:v1.ObjectMeta{Name:"nakamasato.16c38846ea2599bc", GenerateName:"", Namespace:"default", SelfLink:"", UID:"", ResourceVersion:"", Generation:0, CreationTimestamp:v1.Time{Time:time.Time{wall:0x0, ext:0, loc:(*time.Location)(nil)}}, DeletionTimestamp:(*v1.Time)(nil), DeletionGracePeriodSeconds:(*int64)(nil), Labels:map[string]string(nil), Annotations:map[string]string(nil), OwnerReferences:[]v1.OwnerReference(nil), Finalizers:[]string(nil), ClusterName:"", ManagedFields:[]v1.ManagedFieldsEntry(nil)}, InvolvedObject:v1.ObjectReference{Kind:"MySQLUser", Namespace:"default", Name:"nakamasato", UID:"7a35c237-4cb0-4611-af47-87c60cf77842", APIVersion:"mysql.nakamasato.com/v1alpha1", ResourceVersion:"485016", FieldPath:""}, Reason:"ProcessingError", Message:"dial tcp: lookup mysql.default on 10.96.0.10:53: no such host", Source:v1.EventSource{Component:"mysqluser_controller", Host:""}, FirstTimestamp:v1.Time{Time:time.Time{wall:0xc06961c3dad8c3bc, ext:242576599201, loc:(*time.Location)(0x25472e0)}}, LastTimestamp:v1.Time{Time:time.Time{wall:0xc06961cb3b4f70f0, ext:272155074101, loc:(*time.Location)(0x25472e0)}}, Count:16, Type:"Warning", EventTime:v1.MicroTime{Time:time.Time{wall:0x0, ext:0, loc:(*time.Location)(nil)}}, Series:(*v1.EventSeries)(nil), Action:"", Related:(*v1.ObjectReference)(nil), ReportingController:"", ReportingInstance:""}': 'events "nakamasato.16c38846ea2599bc" is forbidden: User "system:serviceaccount:mysql-operator-system:mysql-operator-controller-manager" cannot patch resource "events" in API group "" in the namespace "default"' (will not retry!)
```
