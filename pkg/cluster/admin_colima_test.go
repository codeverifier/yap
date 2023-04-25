package cluster

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/pseudonator/yap/pkg/api"
	"github.com/pseudonator/yap/pkg/internal/exec"
)

const (
	MetalLbCidr = "192.168.108.230/29"
)

func TestColimaStartFlags(t *testing.T) {
	f := newColimaFixture()
	ctx := context.Background()
	err := f.a.Create(ctx, &api.Cluster{
		Name: "test-cluster",
		Colima: &api.ColimaCluster{
			ContainerRuntime: "containerd",
			StartFlags:       []string{"--foo"},
			MetalLbCidr:      MetalLbCidr,
		},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{
		"colima", "start",
		"--foo",
		"--profile=test-cluster",
		"--kubernetes",
		"--runtime=containerd",
		"--kubernetes-disable=servicelb,traefik",
		"--install-metallb",
		"--metallb-address-pool=" + MetalLbCidr,
	}, f.runner.LastArgs)
}

type colimaFixture struct {
	runner *exec.FakeCmdRunner
	a      *colimaAdmin
}

func newColimaFixture() *colimaFixture {
	iostreams := genericclioptions.IOStreams{Out: os.Stdout, ErrOut: os.Stderr}
	runner := exec.NewFakeCmdRunner(func(argv []string) string {
		return ""
	})
	return &colimaFixture{
		runner: runner,
		a:      newColimaAdmin(iostreams, runner),
	}
}
