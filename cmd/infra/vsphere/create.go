package vsphere

import (
	"context"
	"github.com/go-logr/logr"
)

type CreateInfraOptions struct {
	Region           string
	InfraID          string
	Name             string
	BaseDomain       string
	BaseDomainPrefix string
	Zones            []string
	OutputFile       string
	SSHKeyFile       string
}

type CreateInfraOutput struct {
	Name              string `json:"Name"`
	BaseDomain        string `json:"baseDomain"`
	BaseDomainPrefix  string `json:"baseDomainPrefix"`
	VCenter           string `json:"vCenter"`
	Username          string `json:"username"`
	Password          string `json:"password"`
	Datacenter        string `json:"datacenter"`
	DefaultDatastore  string `json:"defaultDatastore"`
	Folder            string `json:"folder"`
	Cluster           string `json:"cluster"`
	ResourcePool      string `json:"resourcePool"`
	TemplateVM        string `json:"templateVM"`
	NumCPUs           int32  `json:"numCpus"`
	NumCoresPerSocket int32  `json:"numCoresPerSocket"`
	MemoryMiB         int64  `json:"memoryMiB"`
	DiskSizeGB        int32  `json:"diskSizeGiB"`
	MachineCIDR       string `json:"machineCIDR"`
}

const (
	DefaultCIDRBlock = "10.0.0.0/16"
)

func (o *CreateInfraOptions) CreateInfra(ctx context.Context, l logr.Logger) (*CreateInfraOutput, error) {
	l.Info("Creating infrastructure", "id", o.InfraID)

	result := &CreateInfraOutput{
		MachineCIDR:      DefaultCIDRBlock,
		Name:             o.Name,
		BaseDomain:       o.BaseDomain,
		BaseDomainPrefix: o.BaseDomainPrefix,
	}

	return result, nil
}
