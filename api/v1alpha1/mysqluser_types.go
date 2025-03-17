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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Grant defines the privileges and the resource for a MySQL user
type Grant struct {
	// Privileges for the MySQL user
	Privileges string `json:"privileges"`

	// Resource on which the privileges are applied
	On string `json:"on"`
}

// MySQLUserSpec defines the desired state of MySQLUser
type MySQLUserSpec struct {

	// MySQL (CRD) name to reference to, which decides the destination MySQL server
	MysqlName string `json:"mysqlName"`

	// +kubebuilder:default=%
	// +kubebuilder:validation:Optional

	// MySQL hostname for MySQL account
	Host string `json:"host"`

	// Grants for the MySQL user
	Grants []Grant `json:"grants,omitempty"`
}

// MySQLUserStatus defines the observed state of MySQLUser
type MySQLUserStatus struct {

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
	Phase      string             `json:"phase,omitempty"`
	Reason     string             `json:"reason,omitempty"`

	// +kubebuilder:default=false

	// true if MySQL user is created
	MySQLUserCreated bool `json:"mysql_user_created,omitempty"`

	// +kubebuilder:default=false

	// true if Secret is created
	SecretCreated bool `json:"secret_created,omitempty"`
}

func (m *MySQLUser) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *MySQLUser) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="MySQLUser",type="boolean",JSONPath=".status.mysql_user_created",description="true if MySQL user is created"
//+kubebuilder:printcolumn:name="Secret",type="boolean",JSONPath=".status.secret_created",description="true if Secret is created"
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="The phase of this MySQLUser"
//+kubebuilder:printcolumn:name="Reason",type="string",JSONPath=".status.reason",description="The reason for the current phase of this MySQLUser"

// MySQLUser is the Schema for the mysqlusers API
type MySQLUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MySQLUserSpec   `json:"spec,omitempty"`
	Status MySQLUserStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MySQLUserList contains a list of MySQLUser
type MySQLUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MySQLUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MySQLUser{}, &MySQLUserList{})
}
