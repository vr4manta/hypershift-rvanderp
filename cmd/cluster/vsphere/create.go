package vsphere

import (
	"context"
	"encoding/json"
	"fmt"
	apifixtures "github.com/openshift/hypershift/api/fixtures"
	"github.com/openshift/hypershift/cmd/cluster/core"
	vsphereinfra "github.com/openshift/hypershift/cmd/infra/vsphere"
	"github.com/openshift/hypershift/cmd/util"
	"github.com/openshift/hypershift/support/infraid"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

func NewCreateCommand(opts *core.CreateOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "vsphere",
		Short:        "Creates basic functional HostedCluster resources on vSphere",
		SilenceUsage: true,
	}

	opts.VSpherePlatform = core.VSpherePlatformOptions{
		TemplateVM:        "",
		NumCPUs:           4,
		NumCoresPerSocket: 1,
		MemoryMiB:         16384,
		DiskSizeGB:        120,
	}

	cmd.Flags().StringVar(&opts.VSpherePlatform.Username, "vsphere-user", opts.VSpherePlatform.Username, "vSphere username")
	cmd.Flags().StringVar(&opts.VSpherePlatform.Password, "vsphere-password", opts.VSpherePlatform.Password, "vSphere password")
	cmd.Flags().StringVar(&opts.VSpherePlatform.VCenter, "vcenter", opts.VSpherePlatform.VCenter, "vCenter hostname")
	cmd.Flags().StringVar(&opts.VSpherePlatform.Datacenter, "datacenter", opts.VSpherePlatform.Datacenter, "datacenter where compute VMs will be created")
	cmd.Flags().StringVar(&opts.VSpherePlatform.Folder, "folder", opts.VSpherePlatform.Folder, "folder where compute VMs will be created")
	cmd.Flags().StringVar(&opts.VSpherePlatform.ResourcePool, "resource-pool", opts.VSpherePlatform.ResourcePool, "resource pool where compute VMs will be created")
	cmd.Flags().StringVar(&opts.VSpherePlatform.Cluster, "cluster", opts.VSpherePlatform.Cluster, "cluster where compute VMs will be created")
	cmd.Flags().StringVar(&opts.VSpherePlatform.DefaultDatastore, "default-datastore", opts.VSpherePlatform.DefaultDatastore, "datastore where compute VMs will be created")
	cmd.Flags().StringVar(&opts.VSpherePlatform.TemplateVM, "template", opts.VSpherePlatform.TemplateVM, "name of VM which will be used as a template for compute VMs")
	cmd.Flags().StringVar(&opts.VSpherePlatform.Network, "network", opts.VSpherePlatform.Network, "name of network which will be used by VMs")
	cmd.Flags().Int32("disk-size", opts.VSpherePlatform.DiskSizeGB, "disk size in GB of assigned to compute VMs")
	cmd.Flags().Int32("vCPUs", opts.VSpherePlatform.NumCPUs, "number of vCPUs assigned to compute VMs")
	cmd.Flags().Int32("cores-per-socket", opts.VSpherePlatform.NumCoresPerSocket, "number of cores per socket assigned to compute VMs")
	cmd.Flags().Int64("memory", opts.VSpherePlatform.MemoryMiB, "amount of memory in MiB assigned to compute VMs")
	cmd.Flags().StringVar(&opts.CredentialSecretName, "secret-creds", opts.CredentialSecretName, "A Kubernetes secret with a platform credential. The secret must exist in the supplied \"--namespace\"")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		opts.Log.Info("Running command for create cluster vsphere")
		ctx := cmd.Context()
		if opts.Timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
			defer cancel()
		}

		if len(opts.CredentialSecretName) == 0 {
			if err := isRequiredOption("vsphere-user", opts.VSpherePlatform.Username); err != nil {
				return err
			}
			if err := isRequiredOption("vsphere-password", opts.VSpherePlatform.Password); err != nil {
				return err
			}
			if err := isRequiredOption("pull-secret", opts.PullSecretFile); err != nil {
				return err
			}
		} else {
			//Check the secret exists now, otherwise stop.
			opts.Log.Info("Retrieving credentials secret", "namespace", opts.Namespace, "name", opts.CredentialSecretName)
			if _, err := util.GetSecret(opts.CredentialSecretName, opts.Namespace); err != nil {
				return err
			}
		}

		opts.Log.Info("Calling create cluster")
		if err := CreateCluster(ctx, opts); err != nil {
			opts.Log.Error(err, "Failed to create cluster")
			os.Exit(1)
		}
		return nil
	}

	return cmd
}

func CreateCluster(ctx context.Context, opts *core.CreateOptions) error {
	if err := core.Validate(ctx, opts); err != nil {
		opts.Log.Error(err, "Validation failed.")
		return err
	}
	return core.CreateCluster(ctx, opts, applyPlatformSpecificsValues)
}

func applyPlatformSpecificsValues(ctx context.Context, exampleOptions *apifixtures.ExampleOptions, opts *core.CreateOptions) (err error) {
	opts.Log.Info("Applying platform specifics")
	client, err := util.GetClient()
	if err != nil {
		return err
	}
	infraID := opts.InfraID

	// Load or create infrastructure for the cluster
	var infra *vsphereinfra.CreateInfraOutput
	if len(opts.InfrastructureJSON) > 0 {
		rawInfra, err := ioutil.ReadFile(opts.InfrastructureJSON)
		if err != nil {
			return fmt.Errorf("failed to read infra json file: %w", err)
		}
		infra = &vsphereinfra.CreateInfraOutput{}
		if err = json.Unmarshal(rawInfra, infra); err != nil {
			return fmt.Errorf("failed to load infra json: %w", err)
		}
	}

	var VSphereUser, VSpherePass string
	if len(opts.CredentialSecretName) > 0 {
		secret, err := util.GetSecretWithClient(client, opts.CredentialSecretName, opts.Namespace)
		if err != nil {
			opts.Log.Error(err, "Unable to load credentials secret")
		} else {
			VSphereUser = string(secret.Data["username"])
			VSpherePass = string(secret.Data["password"])
		}
	}

	// Set the user/pass only if not passed in on cli
	if opts.VSpherePlatform.Username == "" {
		opts.VSpherePlatform.Username = VSphereUser
	}
	if opts.VSpherePlatform.Password == "" {
		opts.VSpherePlatform.Password = VSpherePass
	}

	if opts.BaseDomain == "" {
		if infra != nil {
			opts.BaseDomain = infra.BaseDomain
		} else {
			return fmt.Errorf("base-domain flag is required if infra-json is not provided")
		}
	}

	if infra == nil {
		if len(infraID) == 0 {
			infraID = infraid.New(opts.Name)
		}
		opt := vsphereinfra.CreateInfraOptions{
			InfraID:          infraID,
			Name:             opts.Name,
			BaseDomain:       opts.BaseDomain,
			BaseDomainPrefix: opts.BaseDomainPrefix,
		}
		infra, err = opt.CreateInfra(ctx, opts.Log)
		if err != nil {
			return fmt.Errorf("failed to create infra: %w", err)
		}
	}

	exampleOptions.BaseDomain = opts.BaseDomain
	exampleOptions.MachineCIDR = infra.MachineCIDR
	exampleOptions.InfraID = infraID

	exampleOptions.VSphere = &apifixtures.ExampleVSphereOptions{
		VCenter:           opts.VSpherePlatform.VCenter,
		Datacenter:        opts.VSpherePlatform.Datacenter,
		DefaultDatastore:  opts.VSpherePlatform.DefaultDatastore,
		Folder:            opts.VSpherePlatform.Folder,
		Cluster:           opts.VSpherePlatform.Cluster,
		ResourcePool:      opts.VSpherePlatform.ResourcePool,
		TemplateVM:        opts.VSpherePlatform.TemplateVM,
		NumCPUs:           opts.VSpherePlatform.NumCPUs,
		NumCoresPerSocket: opts.VSpherePlatform.NumCoresPerSocket,
		MemoryMiB:         opts.VSpherePlatform.MemoryMiB,
		DiskSizeGB:        opts.VSpherePlatform.DiskSizeGB,
		Network:           opts.VSpherePlatform.Network,
	}

	exampleOptions.VSphere.Creds = apifixtures.VSphereCreds{
		Username: opts.VSpherePlatform.Username,
		Password: opts.VSpherePlatform.Password,
	}

	opts.Log.Info("Finished applying platform specific")
	return nil
}

// IsRequiredOption returns a cobra style error message when the flag value is empty
func isRequiredOption(flag string, value string) error {
	if len(value) == 0 {
		return fmt.Errorf("required flag(s) \"%s\" not set", flag)
	}
	return nil
}
