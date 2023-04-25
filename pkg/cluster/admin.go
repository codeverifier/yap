package cluster

import (
	"context"

	"github.com/pseudonator/yap/pkg/api"
)

// A cluster admin provides the basic start/stop functionality of a cluster,
// independent of the configuration of the machine it's running on.
type Admin interface {
	EnsureInstalled(ctx context.Context) error

	// Create a new cluster.
	//
	// Make a best effort attempt to delete any resources that might block creation
	// of the cluster.
	Create(ctx context.Context, desired *api.Cluster) error

	Delete(ctx context.Context, config *api.Cluster) error
}

// An extension of cluster admin that indicates the cluster configuration can be
// modified for use from inside containers.
type AdminInContainer interface {
	ModifyConfigInContainer(ctx context.Context, cluster *api.Cluster, containerID string, dockerClient dockerClient, configWriter configWriter) error
}
