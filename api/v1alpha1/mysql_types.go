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

package v1alpha1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MySQLSpec holds the connection information for the target MySQL cluster.
type MySQLSpec struct {

	// Host is MySQL host of target MySQL cluster.
	Host string `json:"host"`

	//+kubebuilder:default=3306

	// Port is MySQL port of target MySQL cluster.
	Port int16 `json:"port,omitempty"`

	// AdminUser is MySQL user to connect target MySQL cluster.
	AdminUser string `json:"admin_user"`

	// AdminPassword is MySQL password to connect target MySQL cluster.
	AdminPassword string `json:"admin_password,omitempty"`

	// Secret name for admin password in GCP Secret Manager.
	GcpSecretName string `json:"gcp_secret_name,omitempty"`
}

// MySQLStatus defines the observed state of MySQL
type MySQLStatus struct {

	//+kubebuilder:default=0

	// The number of users in this MySQL
	UserCount int32 `json:"userCount"`

	//+kubebuilder:default=0

	// The number of database in this MySQL
	DBCount int32 `json:"dbCount"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Host",type=string,JSONPath=`.spec.host`
//+kubebuilder:printcolumn:name="AdminUser",type=string,JSONPath=`.spec.admin_user`
//+kubebuilder:printcolumn:name="UserCount",type="integer",JSONPath=".status.userCount",description="The number of MySQLUsers that belongs to the MySQL"

// MySQL is the Schema for the mysqls API
type MySQL struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MySQLSpec   `json:"spec,omitempty"`
	Status MySQLStatus `json:"status,omitempty"`
}

func (m MySQL) GetKey() string {
	return fmt.Sprintf("%s-%s", m.Namespace, m.Name)
}

//+kubebuilder:object:root=true

// MySQLList contains a list of MySQL
type MySQLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MySQL `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MySQL{}, &MySQLList{})
}
