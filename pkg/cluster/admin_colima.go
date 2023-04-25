package cluster

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"

	"github.com/pseudonator/yap/pkg/api"
	cexec "github.com/pseudonator/yap/pkg/internal/exec"
)

const (
	cmdName = "colima"
)

type colimaAdmin struct {
	iostreams genericclioptions.IOStreams
	runner    cexec.CmdRunner
}

func newColimaAdmin(iostreams genericclioptions.IOStreams, runner cexec.CmdRunner) *colimaAdmin {
	return &colimaAdmin{
		iostreams: iostreams,
		runner:    runner,
	}
}

func (a *colimaAdmin) EnsureInstalled(ctx context.Context) error {
	_, err := exec.LookPath("colima")
	if err != nil {
		return fmt.Errorf("colima not installed. Please install colima with these instructions: ")
	}
	return nil
}

func (a *colimaAdmin) Create(ctx context.Context, desired *api.Cluster) error {
	klog.V(3).Infof("Creating cluster with config:\n%+v\n---\n", desired)

	clusterName := desired.Name

	containerRuntime := "containerd"
	if desired.Colima != nil && desired.Colima.ContainerRuntime != "" {
		containerRuntime = desired.Colima.ContainerRuntime
	}

	args := []string{
		"start",
	}

	if desired.Colima != nil {
		args = append(args, desired.Colima.StartFlags...)
	}

	args = append(args,
		fmt.Sprintf("--profile=%s", clusterName),
		"--kubernetes",
		fmt.Sprintf("--runtime=%s", containerRuntime),
		// Disabled servicelb (https://docs.k3s.io/networking#service-load-balancer) in favor of metallb
		// Disabled traefik (https://docs.k3s.io/networking#service-load-balancer)
		fmt.Sprintf("--kubernetes-disable=servicelb,traefik"),
		"--install-metallb",
		fmt.Sprintf("--metallb-address-pool=%s", desired.Colima.MetalLbCidr),
	)

	if desired.MinCPUs != 0 {
		args = append(args, fmt.Sprintf("--cpu=%d", desired.MinCPUs))
	}

	if desired.KubernetesVersion != "" {
		args = append(args, "--kubernetes-version", desired.KubernetesVersion)
	}

	in := strings.NewReader("")

	err := a.runner.RunIO(ctx,
		genericclioptions.IOStreams{In: in, Out: a.iostreams.Out, ErrOut: a.iostreams.ErrOut},
		cmdName, args...)
	if err != nil {
		return errors.Wrap(err, "creating colima kubernetes cluster")
	}

	return nil
}

func (a *colimaAdmin) Delete(ctx context.Context, config *api.Cluster) error {
	err := a.runner.RunIO(ctx, a.iostreams, cmdName, "delete", "-p", config.Name)
	if err != nil {
		return errors.Wrap(err, "deleting colima cluster")
	}
	return nil
}
