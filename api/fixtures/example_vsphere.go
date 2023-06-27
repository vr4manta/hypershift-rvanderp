package fixtures

type ExampleVSphereOptions struct {
	Creds             VSphereCreds
	VCenter           string
	Datacenter        string
	DefaultDatastore  string
	Folder            string
	Cluster           string
	ResourcePool      string
	TemplateVM        string
	Network           string
	NumCPUs           int32
	NumCoresPerSocket int32
	MemoryMiB         int64
	DiskSizeGB        int32
}

type VSphereCreds struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}
