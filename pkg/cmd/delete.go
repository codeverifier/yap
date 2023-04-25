package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/pseudonator/yap/pkg/api"
	"github.com/pseudonator/yap/pkg/cluster"
	"github.com/pseudonator/yap/pkg/visitor"
)

type DeleteOptions struct {
	*genericclioptions.PrintFlags
	*genericclioptions.FileNameFlags
	genericclioptions.IOStreams

	IgnoreNotFound bool
	Filenames      []string

	// We currently only support two modes - "true" and "false".
	// But we expect that there may be more modes in the future
	// (like what happened with kubectl delete --cascade).
	Cascade string

	clusterController clusterController
}

func NewDeleteOptions() *DeleteOptions {
	o := &DeleteOptions{
		PrintFlags: genericclioptions.NewPrintFlags("deleted"),
		IOStreams:  genericclioptions.IOStreams{Out: os.Stdout, ErrOut: os.Stderr, In: os.Stdin},
	}
	o.FileNameFlags = &genericclioptions.FileNameFlags{Filenames: &o.Filenames}
	return o
}

func (o *DeleteOptions) Command() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "delete -f FILENAME",
		Short: "Delete a currently running cluster",
		Example: "  yap delete -f cluster.yaml\n" +
			"  yap delete cluster minikube",
		Run: o.Run,
	}

	cmd.SetOut(o.Out)
	cmd.SetErr(o.ErrOut)
	o.FileNameFlags.AddFlags(cmd.Flags())

	cmd.Flags().BoolVar(&o.IgnoreNotFound, "ignore-not-found", o.IgnoreNotFound, "If the requested object does not exist the command will return exit code 0.")
	cmd.Flags().StringVar(&o.Cascade, "cascade", "false",
		"If 'true', objects will be deleted recursively. "+
			"For example, deleting a cluster will delete any connected registries. Defaults to 'false'.")

	return cmd
}

func (o *DeleteOptions) Run(cmd *cobra.Command, args []string) {
	err := o.run(args)
	if err != nil {
		_, _ = fmt.Fprintf(o.ErrOut, "%v\n", err)
		os.Exit(1)
	}
}

type deleter interface {
	Delete(ctx context.Context, name string) error
}

type clusterController interface {
	deleter
	Get(ctx context.Context, name string) (*api.Cluster, error)
}

func (o *DeleteOptions) run(args []string) error {
	err := o.validateCascade()
	if err != nil {
		return err
	}

	ctx := context.TODO()
	resources, err := o.parseExplicitResources(args)
	if err != nil {
		return err
	}

	printer, err := toPrinter(o.PrintFlags)
	if err != nil {
		return err
	}

	for _, resource := range resources {
		switch resource := resource.(type) {
		case *api.Cluster:
			controller, err := o.getClusterController()
			if err != nil {
				return err
			}

			cluster.FillDefaults(resource)

			name := resource.Name

			// Normalize the name of the cluster so that
			// 'yap delete cluster kind' works.
			cluster, err := normalizedGet(ctx, controller, name)
			if err == nil {
				name = cluster.Name
			}

			err = controller.Delete(ctx, name)

			if err != nil {
				if o.IgnoreNotFound && errors.IsNotFound(err) {
					continue
				}
				return err
			}
			err = printer.PrintObj(resource, o.Out)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("cannot delete: %T", resource)
		}
	}
	return nil
}

func (o *DeleteOptions) parseExplicitResources(args []string) ([]runtime.Object, error) {
	hasFiles := len(o.Filenames) > 0
	hasNames := len(args) >= 2
	if !(hasFiles || hasNames) {
		return nil, fmt.Errorf("Expected resources, specified as files ('yap delete -f') or names ('yap delete cluster foo`)")
	}
	if hasFiles && hasNames {
		return nil, fmt.Errorf("Can only specify one of {files, resource names}")
	}

	if hasFiles {
		visitors, err := visitor.FromStrings(o.Filenames, o.In)
		if err != nil {
			return nil, err
		}

		return visitor.DecodeAll(visitors)
	}

	var resources []runtime.Object
	t := args[0]
	names := args[1:]
	switch t {
	case "cluster", "clusters":
		for _, name := range names {
			resources = append(resources, &api.Cluster{
				TypeMeta: cluster.TypeMeta(),
				Name:     name,
			})
		}
	default:
		return nil, fmt.Errorf("Unrecognized type: %s", t)
	}
	return resources, nil
}

func (o *DeleteOptions) getClusterController() (clusterController, error) {
	if o.clusterController == nil {
		controller, err := cluster.DefaultController(o.IOStreams)
		if err != nil {
			return nil, err
		}
		o.clusterController = controller
	}
	return o.clusterController, nil
}

func (o *DeleteOptions) validateCascade() error {
	if o.Cascade == "" || o.Cascade == "true" || o.Cascade == "false" {
		return nil
	}
	return fmt.Errorf("Invalid cascade: %s. Valid values: true, false.", o.Cascade)
}
