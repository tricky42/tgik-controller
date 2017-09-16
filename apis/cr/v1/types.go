package v1

import (
	"fmt"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	EnvironmentsPlural  string = "environments"
	EnvironmentGroup    string = "sf.yaas.io"
	EnvironmentVersion  string = "v1"
	FullEnvironmentName string = EnvironmentsPlural + "." + EnvironmentGroup
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Environment struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata"`
	Spec               EnvironmentSpec   `json:"spec"`
	Status             EnvironmentStatus `json:"status,omitempty"`
}

func (env Environment) String() string {
	return fmt.Sprintf(" {\n   Name:         %s\n   Namespace:    %s\n   Guid:         %s\n   Scpid:        %s\n   State:        %s\n   Message:      %s\n}", env.ObjectMeta.Name, env.ObjectMeta.Namespace, env.Spec.Guid, env.Spec.Scpid, env.Status.State, env.Status.Message)
}

type EnvironmentSpec struct {
	Guid  string `json:"guid"`
	Scpid string `json:"scpid"`
}

type EnvironmentStatus struct {
	State   EnvironmentState `json:"state,omitempty"`
	Message string           `json:"message,omitempty"`
}

type EnvironmentState string

const (
	EnvironmentStateCreated   EnvironmentState = "Created"
	EnvironmentStateProcessed EnvironmentState = "Processed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type EnvironmentList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`
	Items            []Environment `json:"items"`
}

func (envList EnvironmentList) String() string {
	var out string = "{ "
	for _, env := range envList.Items {
		out += env.ObjectMeta.Name + ", "
	}
	out += " }\n"
	return fmt.Sprintf(out)
}
