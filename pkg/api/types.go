package api

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"

	"github.com/pseudonator/yap/pkg/api/k3dv1alpha4"
)

// TypeMeta partially copies apimachinery/pkg/apis/meta/v1.TypeMeta
// No need for a direct dependence; the fields are stable.
type TypeMeta struct {
	Kind       string `json:"kind,omitempty" yaml:"kind,omitempty"`
	APIVersion string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
}

// Cluster contains cluster configuration.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Cluster struct {
	TypeMeta `yaml:",inline"`

	// The cluster name. Pulled from .kube/config.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// The name of the tool used to create this cluster.
	Product string `json:"product,omitempty" yaml:"product,omitempty"`

	// Make sure that the cluster has access to at least this many
	// CPUs. This is mostly helpful for ensuring that your Docker/Lima
	// VM has enough CPU. If yap can't guarantee this many
	// CPU, it will return an error.
	MinCPUs int `json:"minCPUs,omitempty" yaml:"minCPUs,omitempty"`

	// The desired version of Kubernetes to run.
	//
	// Examples:
	// v1.19.1
	// v1.14.0
	// Must start with 'v' and contain a major, minor, and patch version.
	//
	// Not all cluster products allow you to customize this.
	KubernetesVersion string `json:"kubernetesVersion,omitempty" yaml:"kubernetesVersion,omitempty"`

	// The Kind cluster config. Only applicable for clusters with product: kind.
	//
	// Full documentation at:
	// https://pkg.go.dev/sigs.k8s.io/kind/pkg/apis/config/v1alpha4#Cluster
	//
	// Properties of this config may be overridden by properties of the yap
	// Cluster config. For example, the name field of the top-level Cluster object
	// wins over one specified in the Kind config.
	KindV1Alpha4Cluster *v1alpha4.Cluster `json:"kindV1Alpha4Cluster,omitempty" yaml:"kindV1Alpha4Cluster,omitempty"`

	// The Minikube cluster config. Only applicable for clusters with product: minikube.
	Minikube *MinikubeCluster `json:"minikube,omitempty" yaml:"minikube,omitempty"`

	// The K3D cluster config. Only applicable for clusters with product: k3d.
	K3D *K3DCluster `json:"k3d,omitempty" yaml:"k3d,omitempty"`

	// The Colima cluster config. Only applicable for clusters with product: colima.
	Colima *ColimaCluster `json:"colima,omitempty" yaml:"colima,omitempty"`

	// Most recently observed status of the cluster.
	// Populated by the system.
	// Read-only.
	Status ClusterStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

type ClusterStatus struct {
	// When the cluster was first created.
	CreationTimestamp metav1.Time `json:"creationTimestamp,omitempty" yaml:"creationTimestamp,omitempty"`

	// The number of CPU. Only applicable to local clusters.
	CPUs int `json:"cpus,omitempty" yaml:"cpus,omitempty"`

	// Whether this is the current cluster in `kubectl`
	Current bool `json:"current,omitempty" yaml:"current,omitempty"`

	// The version of Kubernetes currently running.
	//
	// Reported by the Kubernetes API. May contain a build tag.
	//
	// Examples:
	// v1.19.1
	// v1.18.10-gke.601
	// v1.19.3-34+fa32ff1c160058
	KubernetesVersion string `json:"kubernetesVersion,omitempty" yaml:"kubernetesVersion,omitempty"`

	// Populated when we encounter an error reading the cluster status.
	Error string `json:"error,omitempty"`
}

// MinikubeCluster describes minikube-specific options for starting a cluster.
//
// Options in this struct, when possible, should match the flags
// to `minikube start`.
//
// Prefer setting features on the ClusterSpec rather than on the MinikubeCluster
// object when possible. For example, this object doesn't have a `kubernetesVersion`
// field, because it's supported by ClusterSpec.
//
// yap's logic for diffing clusters and applying changes is less robust
// for cluster-specific config flags.
type MinikubeCluster struct {
	// The container runtime of the cluster. Defaults to containerd.
	ContainerRuntime string `json:"containerRuntime,omitempty" yaml:"containerRuntime,omitempty"`

	// Extra config options passed directly to Minikube's --extra-config flags.
	// When not set, we will default to starting minikube with these configs:
	//
	// kubelet.max-pods=500
	ExtraConfigs []string `json:"extraConfigs,omitempty" yaml:"extraConfigs,omitempty"`

	// Unstructured flags to pass to minikube on `minikube start`.
	// These flags will be passed before default flags.
	StartFlags []string `json:"startFlags,omitempty" yaml:"startFlags,omitempty"`
}

// K3DCluster describes k3d-specific options for starting a cluster.
//
// Prefer setting features on the ClusterSpec rather than on the K3dCluster
// object when possible.
//
// yap's logic for diffing clusters and applying changes is less robust
// for cluster-specific configs.
type K3DCluster struct {
	// K3D's own cluster config format.
	//
	// Documentation: https://k3d.io/v5.4.6/usage/configfile/
	//
	// Uses this schema: https://github.com/k3d-io/k3d/blob/v5.4.6/pkg/config/v1alpha4/types.go
	V1Alpha4Simple *k3dv1alpha4.SimpleConfig `json:"v1alpha4Simple,omitempty" yaml:"v1alpha4Simple,omitempty"`
}

type ColimaCluster struct {
	// The container runtime of the cluster. Defaults to containerd.
	ContainerRuntime string `json:"containerRuntime,omitempty" yaml:"containerRuntime,omitempty"`

	// Unstructured flags to pass to colima on `colima start`.
	StartFlags []string `json:"startFlags,omitempty" yaml:"startFlags,omitempty"`

	// MetalLB address pool
	MetalLbCidr string `json:"metallbCidr,omitempty" yaml:"metallbCidr,omitempty"`
}

// ClusterList is a list of Clusters.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ClusterList struct {
	TypeMeta `json:",inline"`

	// List of clusters.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md
	Items []Cluster `json:"items" protobuf:"bytes,2,rep,name=items"`
}
