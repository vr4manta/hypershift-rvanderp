package nodepool

import (
	"encoding/base64"

	hyperv1 "github.com/openshift/hypershift/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	capivsphere "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
)

func vSphereMachineTemplateSpec(hcluster *hyperv1.HostedCluster, nodePool *hyperv1.NodePool, userDataSecret *corev1.Secret) *capivsphere.VSphereMachineTemplateSpec {
	nodePoolPlatform := nodePool.Spec.Platform.VSphere
	cloneSpec := capivsphere.VirtualMachineCloneSpec{
		Template:     nodePoolPlatform.TemplateVM,
		CloneMode:    capivsphere.FullClone,
		Server:       nodePoolPlatform.VCenter,
		Datacenter:   nodePoolPlatform.Datacenter,
		Folder:       nodePoolPlatform.Folder,
		Datastore:    nodePoolPlatform.DefaultDatastore,
		ResourcePool: nodePoolPlatform.ResourcePool,
		Network: capivsphere.NetworkSpec{
			Devices: []capivsphere.NetworkDeviceSpec{
				{
					NetworkName: nodePoolPlatform.Network,
					DHCP4:       true,
					DHCP6:       true,
					IPAddrs:     []string{},
				},
			},
		},
		NumCPUs:           nodePoolPlatform.NumCPUs,
		NumCoresPerSocket: nodePoolPlatform.NumCoresPerSocket,
		MemoryMiB:         nodePoolPlatform.MemoryMiB,
		DiskGiB:           nodePoolPlatform.DiskSizeGiB,
		CustomVMXKeys: map[string]string{
			"guestinfo.ignition.config.data":          base64.StdEncoding.EncodeToString(userDataSecret.Data["value"]),
			"guestinfo.ignition.config.data.encoding": "base64",
			"disk.EnableUUID":                         "TRUE",
		},
	}
	template := capivsphere.VSphereMachineTemplateSpec{
		Template: capivsphere.VSphereMachineTemplateResource{
			Spec: capivsphere.VSphereMachineSpec{
				VirtualMachineCloneSpec: cloneSpec,
			},
		},
	}
	return &template
}
