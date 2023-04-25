package cluster

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/pkg/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"

	"github.com/pseudonator/yap/pkg/api"
	cexec "github.com/pseudonator/yap/pkg/internal/exec"
)

// minikubeAdmin uses the minikube CLI to manipulate a minikube cluster,
// once the underlying machine has been setup.
type minikubeAdmin struct {
	iostreams    genericclioptions.IOStreams
	runner       cexec.CmdRunner
	dockerClient dockerClient
}

func newMinikubeAdmin(iostreams genericclioptions.IOStreams, dockerClient dockerClient, runner cexec.CmdRunner) *minikubeAdmin {
	return &minikubeAdmin{
		iostreams:    iostreams,
		dockerClient: dockerClient,
		runner:       runner,
	}
}

func (a *minikubeAdmin) EnsureInstalled(ctx context.Context) error {
	_, err := exec.LookPath("minikube")
	if err != nil {
		return fmt.Errorf("minikube not installed. Please install minikube with these instructions: https://minikube.sigs.k8s.io/")
	}
	return nil
}

type minikubeVersionResponse struct {
	MinikubeVersion string `json:"minikubeVersion"`
}

func (a *minikubeAdmin) version(ctx context.Context) (semver.Version, error) {
	out := bytes.NewBuffer(nil)
	err := a.runner.RunIO(ctx,
		genericclioptions.IOStreams{Out: out, ErrOut: a.iostreams.ErrOut},
		"minikube", "version", "-o", "json")
	if err != nil {
		return semver.Version{}, fmt.Errorf("minikube version: %v", err)
	}

	decoder := json.NewDecoder(out)
	response := minikubeVersionResponse{}
	err = decoder.Decode(&response)
	if err != nil {
		return semver.Version{}, fmt.Errorf("minikube version: %v", err)
	}
	v := response.MinikubeVersion
	if v == "" {
		return semver.Version{}, fmt.Errorf("minikube version not found")
	}
	result, err := semver.ParseTolerant(v)
	if err != nil {
		return semver.Version{}, fmt.Errorf("minikube version: %v", err)
	}
	return result, nil
}

func (a *minikubeAdmin) Create(ctx context.Context, desired *api.Cluster) error {
	klog.V(3).Infof("Creating cluster with config:\n%+v\n---\n", desired)

	clusterName := desired.Name

	containerRuntime := "containerd"
	if desired.Minikube != nil && desired.Minikube.ContainerRuntime != "" {
		containerRuntime = desired.Minikube.ContainerRuntime
	}

	extraConfigs := []string{"kubelet.max-pods=500"}
	if desired.Minikube != nil && len(desired.Minikube.ExtraConfigs) > 0 {
		extraConfigs = desired.Minikube.ExtraConfigs
	}

	args := []string{
		"start",
	}

	if desired.Minikube != nil {
		args = append(args, desired.Minikube.StartFlags...)
	}

	args = append(args,
		"-p", clusterName,
		"--driver=docker",
		fmt.Sprintf("--container-runtime=%s", containerRuntime),
	)

	for _, c := range extraConfigs {
		args = append(args, fmt.Sprintf("--extra-config=%s", c))
	}

	if desired.MinCPUs != 0 {
		args = append(args, fmt.Sprintf("--cpus=%d", desired.MinCPUs))
	}
	if desired.KubernetesVersion != "" {
		args = append(args, "--kubernetes-version", desired.KubernetesVersion)
	}

	in := strings.NewReader("")

	err := a.runner.RunIO(ctx,
		genericclioptions.IOStreams{In: in, Out: a.iostreams.Out, ErrOut: a.iostreams.ErrOut},
		"minikube", args...)
	if err != nil {
		return errors.Wrap(err, "creating minikube cluster")
	}

	return nil
}

func (a *minikubeAdmin) Delete(ctx context.Context, config *api.Cluster) error {
	err := a.runner.RunIO(ctx, a.iostreams, "minikube", "delete", "-p", config.Name)
	if err != nil {
		return errors.Wrap(err, "deleting minikube cluster")
	}
	return nil
}
