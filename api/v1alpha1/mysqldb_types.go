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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MySQLDBSpec defines the desired state of MySQLDB
type MySQLDBSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// MySQL Database name
	DBName string `json:"dbName"`

	// MySQL name
	MysqlName string `json:"mysqlName"`
}

// MySQLDBStatus defines the observed state of MySQLDB
type MySQLDBStatus struct {
	// database is created or not
	Phase  string `json:"phase,omitempty"`
	Reason string `json:"reason,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="The phase of MySQLDB"
//+kubebuilder:printcolumn:name="Reason",type="string",JSONPath=".status.reason",description="The reason for the current phase of this MySQLDB"

// MySQLDB is the Schema for the mysqldbs API
type MySQLDB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MySQLDBSpec   `json:"spec,omitempty"`
	Status MySQLDBStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MySQLDBList contains a list of MySQLDB
type MySQLDBList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MySQLDB `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MySQLDB{}, &MySQLDBList{})
}
