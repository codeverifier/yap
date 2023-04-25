package cmd

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/pseudonator/yap/pkg/api"
	clusterid "github.com/pseudonator/yap/pkg/internal/cluster"
)

type clusterGetter interface {
	Get(ctx context.Context, name string) (*api.Cluster, error)
}

// We create clusters like:
// yap create cluster kind
// For most clusters, the name of the cluster will match the name of the product.
// But for cases where they don't match, we want
// `yap delete cluster kind` to automatically map to `yap delete cluster kind-kind`
func normalizedGet(ctx context.Context, controller clusterGetter, name string) (*api.Cluster, error) {
	cluster, err := controller.Get(ctx, name)
	if err == nil {
		return cluster, nil
	}

	if !errors.IsNotFound(err) {
		return nil, err
	}

	origErr := err
	retryName := ""
	if name == string(clusterid.ProductKIND) {
		retryName = clusterid.ProductKIND.DefaultClusterName()
	} else if name == string(clusterid.ProductK3D) {
		retryName = clusterid.ProductK3D.DefaultClusterName()
	}

	if retryName == "" {
		return nil, origErr
	}

	cluster, err = controller.Get(ctx, retryName)
	if err == nil {
		return cluster, nil
	}
	return nil, origErr
}
