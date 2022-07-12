package nodepool

import (
	hyperv1 "github.com/openshift/hypershift/api/v1alpha1"
	capivsphere "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
)

func vSphereMachineTemplateSpec(hcluster *hyperv1.HostedCluster, nodePool *hyperv1.NodePool) *capivsphere.VSphereMachineTemplate {
	nodePoolPlatform := nodePool.Spec.Platform.VSphere
	cloneSpec := capivsphere.VirtualMachineCloneSpec{
		Template:          nodePoolPlatform.TemplateVM,
		CloneMode:         capivsphere.FullClone,
		Server:            nodePoolPlatform.VCenter,
		Datacenter:        nodePoolPlatform.Datacenter,
		Folder:            nodePoolPlatform.Folder,
		Datastore:         nodePoolPlatform.DefaultDatastore,
		ResourcePool:      nodePoolPlatform.ResourcePool,
		Network:           capivsphere.NetworkSpec{},
		NumCPUs:           nodePoolPlatform.NumCPUs,
		NumCoresPerSocket: nodePoolPlatform.NumCoresPerSocket,
		MemoryMiB:         nodePoolPlatform.MemoryMiB,
		DiskGiB:           nodePoolPlatform.DiskSizeGiB,
	}
	template := capivsphere.VSphereMachineTemplate{
		Spec: capivsphere.VSphereMachineTemplateSpec{
			Template: capivsphere.VSphereMachineTemplateResource{
				Spec: capivsphere.VSphereMachineSpec{
					VirtualMachineCloneSpec: cloneSpec,
				},
			},
		},
	}
	return &template
}
