package provisioner

import "github.com/kubernauts/tk8-provisioner-rke/internal/cluster"

type RKE struct {
}

func (p RKE) Init(args []string) {
	cluster.KubesprayInit()
	cluster.Create()
}

func (p RKE) Setup(args []string) {
	cluster.Install()

}

func (p RKE) Scale(args []string) {
	cluster.Scale()

}

func (p RKE) Reset(args []string) {
	cluster.Reset()

}

func (p RKE) Remove(args []string) {
	cluster.Remove()

}

func (p RKE) Upgrade(args []string) {
	cluster.NotImplemented()
}

func (p RKE) Destroy(args []string) {
	cluster.Destroy()
}

func NewRKE() cluster.Provisioner {
	cluster.SetClusterName()
	provisioner := new(RKE)
	return provisioner
}
