/*
Copyright 2022.

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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ApplicationSpec defines the desired state of Application
type ApplicationSpec struct {
	//+kubebuilder:default:="1.0.0"
	Version string `json:"version,omitempty"`
	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:default:=1
	AmountPods int32 `json:"amountPods"`
	// +kubebuilder:default:="database"
	DatabaseName string `json:"databaseName,omitempty"`
	// +kubebuilder:default:="databaseNamespace"
	DatabaseNamespace string `json:"databaseNamespace,omitempty"`
	// +kubebuilder:default:="https://raw.githubusercontent.com/IBM/multi-tenancy/main/installapp/postgres-config/create-populate-tenant-a.sql"
	SchemaUrl string `json:"schemaUrl,omitempty"`
	Title     string `json:"title,omitempty"`
}

// ApplicationStatus defines the observed state of Application
type ApplicationStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions    []metav1.Condition `json:"conditions"`
	SchemaCreated bool               `json:"schemaCreated"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

// Application is the Schema for the applications API
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec,omitempty"`
	Status ApplicationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ApplicationList contains a list of Application
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Application{}, &ApplicationList{})
}

func (application *Application) GetConditions() []metav1.Condition {
	return application.Status.Conditions
}

func (application *Application) SetConditions(conditions []metav1.Condition) {
	application.Status.Conditions = conditions
}
