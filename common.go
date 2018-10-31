package provisioner

import (
	"fmt"
	"log"
	"os/exec"
	"os/user"
	"strings"

	"github.com/blang/semver"
	"github.com/kubernauts/tk8-provisioner-rke/internal/cluster"
	"github.com/kubernauts/tk8/pkg/provisioner"
	"path/filepath"
)

type RKE struct {
}

// var Name string

func (p RKE) Init(args []string) {
	// Name = cluster.Name
	// if len(Name) == 0 {
	// 	Name = "TK8RKE"
	// }
	// cluster.KubesprayInit()
	// cluster.Create()
}

func checkRKEProvisionerFile(name string) (bool, error) {
	matches, err := filepath.Glob(name + "*")
	if err != nil {
		return false, err
	}
	return len(matches) > 0, nil
}

func (p RKE) Setup(args []string) {
	kube, err := exec.LookPath("kubectl")
	if err != nil {
		log.Fatal("kubectl not found, kindly check")
	}
	fmt.Printf("Found kubectl at %s\n", kube)
	rr, err := exec.Command("kubectl", "version", "--client", "--short").Output()
	if err != nil {
		log.Fatal(err)
	}

	log.Println(string(rr))
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	rkeProvisionerPath := usr.HomeDir + "/.terraform.d/plugins/terraform-provider-rke"
	exists, _ := checkRKEProvisionerFile(rkeProvisionerPath)
	if !exists {
		fmt.Println("Terraform RKE provisioner plugin does not exists. Please check https://github.com/yamamoto-febc/terraform-provider-rke for installation instructions specific to your OS.")
	} else {
		fmt.Println("Terraform RKE provisioner exists. Good to go!")
	}

	//Check if kubectl version is greater or equal to 1.10

	parts := strings.Split(string(rr), " ")

	KubeCtlVer := strings.Replace((parts[2]), "v", "", -1)

	v1, err := semver.Make("1.10.0")
	v2, err := semver.Make(strings.TrimSpace(KubeCtlVer))

	if v2.LT(v1) {
		log.Fatalln("kubectl client version on this system is less than the required version 1.10.0")
	}

	// Check if rke is installed
	if _, err := exec.LookPath("rke"); err != nil {
		log.Fatalln("RKE binary not found. Please install it. While RKE binary is not required while setting up cluster, it is strongly recommended for further interactions with cluster")
	}
	if _, err := exec.LookPath("terraform"); err != nil {
		log.Fatalln("Terraform binary not found in the installation folder")
	}

	log.Println("Terraform binary exists in the installation folder, terraform version:")

	terr, err := exec.Command("terraform", "version").Output()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(terr))

	cluster.Install()

}

func (p RKE) Scale(args []string) {
	provisioner.NotImplemented()

}

func (p RKE) Reset(args []string) {
	cluster.RKEReset()

}

func (p RKE) Remove(args []string) {
	// remove rke cluster, not complete infra
	// equivalent to rke remove --config rancher-cluster.yml
	cluster.RKERemove()

}

func (p RKE) Upgrade(args []string) {
	provisioner.NotImplemented()
}

func (p RKE) Destroy(args []string) {
	// teardown complete infra
	cluster.RKEDestroy()
}

func NewRKE() provisioner.Provisioner {
	cluster.SetClusterName()
	provisioner := new(RKE)
	return provisioner
}
