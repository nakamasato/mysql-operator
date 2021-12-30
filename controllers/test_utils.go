package controllers

import (
	"context"

	mysqlv1alpha1 "github.com/nakamasato/mysql-operator/api/v1alpha1"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func cleanUpMySQL(ctx context.Context, k8sClient client.Client, namespace string) {
	err := k8sClient.DeleteAllOf(ctx, &mysqlv1alpha1.MySQL{}, client.InNamespace(namespace))
	Expect(err).NotTo(HaveOccurred())
	mysqlList := &mysqlv1alpha1.MySQLList{}
	Eventually(func() int {
		err := k8sClient.List(ctx, mysqlList, &client.ListOptions{})
		if err != nil {
			return -1
		}
		return len(mysqlList.Items)
	}).Should(Equal(0))
}

func cleanUpMySQLUser(ctx context.Context, k8sClient client.Client, namespace string) {
	err := k8sClient.DeleteAllOf(ctx, &mysqlv1alpha1.MySQLUser{}, client.InNamespace(namespace))
	Expect(err).NotTo(HaveOccurred())
	mysqlUserList := &mysqlv1alpha1.MySQLUserList{}
	Eventually(func() int {
		err := k8sClient.List(ctx, mysqlUserList, &client.ListOptions{})
		if err != nil {
			return -1
		}
		return len(mysqlUserList.Items)
	}).Should(Equal(0))
}

func cleanUpSecret(ctx context.Context, k8sClient client.Client, namespace string) {
	err := k8sClient.DeleteAllOf(ctx, &v1.Secret{}, client.InNamespace(namespace))
	Expect(err).NotTo(HaveOccurred())
}

func newMySQLUser(apiVersion, namespace, name, mysqlName string) *mysqlv1alpha1.MySQLUser {
	return &mysqlv1alpha1.MySQLUser{
		TypeMeta: metav1.TypeMeta{APIVersion: apiVersion, Kind: "MySQLUser"},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec:   mysqlv1alpha1.MySQLUserSpec{MysqlName: mysqlName},
		Status: mysqlv1alpha1.MySQLUserStatus{},
	}
}

func addOwnerReferenceToMySQL(mysqlUser *mysqlv1alpha1.MySQLUser, mysql *mysqlv1alpha1.MySQL) *mysqlv1alpha1.MySQLUser {
	mysqlUser.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion:         "mysql.nakamasato.com/v1alpha1",
			Kind:               "MySQL",
			Name:               mysql.Name,
			UID:                mysql.UID,
			BlockOwnerDeletion: pointer.BoolPtr(true),
			Controller:         pointer.BoolPtr(true),
		},
	}
	return mysqlUser
}
