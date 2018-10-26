package cluster

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/kubernauts/tk8-provisioner-rke/internal/templates"
	"github.com/spf13/viper"
)

type RKEConfig struct {
	ClusterName         string
	AWSRegion           string
	RKENodeInstanceType string
	NodeCount           int
	SSHKeyPath          string
	CloudProvider       string
	NodeOS              string
}

type rkeDistOS struct {
	User     string
	AmiOwner string
	NodeOS   string
}

// DistOSMap holds the main OS distrubution mapping informations.
var rkeDistOSMap = map[string]rkeDistOS{
	"centos": rkeDistOS{
		User:     "centos",
		AmiOwner: "688023202711",
		NodeOS:   "dcos-centos7-*",
	},
	"ubuntu": rkeDistOS{
		User:     "ubuntu",
		AmiOwner: "099720109477",
		NodeOS:   "ubuntu/images/hvm-ssd/ubuntu-xenial-16.04-amd64-*",
	},
	"coreos": rkeDistOS{
		User:     "core",
		AmiOwner: "595879546273",
		NodeOS:   "CoreOS-stable-*",
	},
}

func rkeDistSelect() (string, string) {
	//Read Configuration File
	AmiID, InstanceOS, sshUser := rkeGetDistConfig()

	if AmiID != "" && sshUser == "" {
		log.Fatal("SSH Username is required when using custom AMI")
		return "", ""
	}
	if AmiID == "" && InstanceOS == "" {
		log.Fatal("Provide either of AMI ID or OS in the config file.")
		return "", ""
	}

	if AmiID != "" && sshUser != "" {
		InstanceOS = "custom"
		rkeDistOSMap["custom"] = rkeDistOS{
			User:     sshUser,
			AmiOwner: AmiID,
			NodeOS:   "custom",
		}
	}

	return rkeDistOSMap[InstanceOS].User, InstanceOS
}

// GetDistConfig is used to get config details specific to a particular distribution.
// Used to determine various details such as the SSH user about the distribution.
func rkeGetDistConfig() (string, string, string) {
	ReadViperConfigFile("config")
	awsAmiID := viper.GetString("rke.ami_id")
	awsInstanceOS := viper.GetString("rke.node_os")
	sshUser := viper.GetString("rke.ssh_user")
	return awsAmiID, awsInstanceOS, sshUser
}

func GetRKEConfig() RKEConfig {
	ReadViperConfigFile("config")
	return RKEConfig{
		ClusterName:         viper.GetString("rke.cluster_name"),
		AWSRegion:           viper.GetString("rke.rke_aws_region"),
		RKENodeInstanceType: viper.GetString("rke.rke_node_instance_type"),
		NodeCount:           viper.GetInt("rke.node_count"),
		CloudProvider:       viper.GetString("rke.cloud_provider"),
	}
}

func rkePrepareConfigFiles(InstanceOS string, Name string) {
	fmt.Println(InstanceOS)
	ParseTemplate(templates.VariablesRKE, "./inventory/"+Name+"/provisioner/variables.tf", GetRKEConfig())
	ParseTemplate(templates.DistVariablesRKE, "./inventory/"+Name+"/provisioner/modules/rke/distos.tf", rkeDistOSMap[InstanceOS])
	ParseTemplate(templates.Credentials, "./inventory/"+Name+"/provisioner/credentials.tfvars", GetCredentials())

}

// Install is used to setup the Kubernetes Cluster with RKE
func Install() {
	var Name string
	config := GetRKEConfig()
	Name = config.ClusterName
	os.MkdirAll("./inventory/"+Name+"/provisioner/modules/rke", 0755)
	exec.Command("cp", "-rfp", "./provisioner/rke/", "./inventory/"+Name+"/provisioner").Run()
	rkeSSHUser, rkeOSLabel := rkeDistSelect()
	fmt.Printf("Prepairing Setup for user %s on %s\n", rkeSSHUser, rkeOSLabel)
	rkePrepareConfigFiles(rkeOSLabel, Name)
	// Check if a terraform state file already exists
	if _, err := os.Stat("./inventory/" + Name + "/provisioner/terraform.tfstate"); err == nil {
		log.Println("There is an existing cluster, please remove terraform.tfstate file or delete the installation before proceeding")
	} else {
		log.Println("starting terraform init")
		log.Println("terrinit dir %s", Name)
		terrInit := exec.Command("terraform", "init")
		terrInit.Dir = "./inventory/" + Name + "/provisioner/"
		out, _ := terrInit.StdoutPipe()
		terrInit.Start()
		scanInit := bufio.NewScanner(out)
		for scanInit.Scan() {
			m := scanInit.Text()
			fmt.Println(m)
		}

		terrInit.Wait()
	}

	//	log.Println("starting terraform apply")
	//	terrSet := exec.Command("terraform", "apply", "-var-file=credentials.tfvars", "-auto-approve")
	//	terrSet.Dir = "./inventory/" + Name + "/provisioner/"
	//	stdout, err := terrSet.StdoutPipe()
	//	terrSet.Stderr = terrSet.Stdout
	//	terrSet.Start()
	//
	//	scanner := bufio.NewScanner(stdout)
	//	for scanner.Scan() {
	//		m := scanner.Text()
	//		fmt.Println(m)
	//	}
	//
	//	terrSet.Wait()
	//	if err != nil {
	//		panic(err)
	//	}
	log.Println("starting terraform plan")
	terrSet := exec.Command("terraform", "plan", "-var-file=credentials.tfvars")
	terrSet.Dir = "./inventory/" + Name + "/provisioner/"
	stdout, err := terrSet.StdoutPipe()
	terrSet.Stderr = terrSet.Stdout
	terrSet.Start()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		m := scanner.Text()
		fmt.Println(m)
	}

	terrSet.Wait()
	if err != nil {
		panic(err)
	}
	// Export KUBECONFIG file to the installation folder
	//log.Println("Moving kubeconfig and rke cluster config files to the installation folder")
	//kubeConfigold := "./inventory/" + Name + "/provisioner/kube_config_rancher-cluster.yml"
	//kubeConfignew := "./kube_config_rancher-cluster.yml"
	//rkeConfigold := "./inventory/" + Name + "/provisioner/rancher-cluster.yml"
	//rkeConfignew := "./rancher-cluster.yml"
	//errmvKubeconfig := os.Rename(kubeConfigold, kubeConfignew)

	//if errmvKubeconfig != nil {
	//	fmt.Println(errmvKubeconfig)
	//}

	//errmvRkeConfig := os.Rename(rkeConfigold, rkeConfignew)

	//if errmvRkeConfig != nil {
	//	fmt.Println(errmvRkeConfig)
	//}

	//log.Println("Voila! Kubernetes cluster created with RKE is up and running")
	//log.Println("Writing private_key to the file from terraform output")
	//writePrivKey := exec.Command("terraform", "output", "private_key", ">>", "./rke-ssh-key.pem")
	//writePrivKey.Dir = "./inventory/" + Name + "/provisioner/"
	//stdout, err = writePrivKey.StdoutPipe()
	//writePrivKey.Stderr = writePrivKey.Stdout
	//writePrivKey.Start()

	//scanner = bufio.NewScanner(stdout)
	//for scanner.Scan() {
	//	m := scanner.Text()
	//	fmt.Println(m)
	//}

	//writePrivKey.Wait()
	//if err != nil {
	//	panic(err)
	//}

	os.Exit(0)

}

// Reset is used to reset the  Kubernetes Cluster back to rollout on the infrastructure.
func RKEReset() {
	NotImplemented()
}

// Remove is used to remove the Kubernetes Cluster from the infrastructure
func RKERemove() {
	// NotImplemented()
	log.Println("Removing rke cluster")
	rkeRemove := exec.Command("rke", "remove", "--config", "./rancher-cluster.yml")
	stdout, err := rkeRemove.StdoutPipe()
	rkeRemove.Stderr = rkeRemove.Stdout
	rkeRemove.Start()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		m := scanner.Text()
		fmt.Println(m)
	}

	rkeRemove.Wait()
	if err != nil {
		panic(err)
	}
	log.Println("Successfully removed rke cluster")
}

func RKEDestroy() {
	config := GetRKEConfig()
	Name := config.ClusterName
	log.Println("starting terraform destroy")
	terrDes := exec.Command("terraform", "destroy", "-var-file=credentials.tfvars", "-auto-approve")
	terrDes.Dir = "./inventory/" + Name + "/provisioner/"
	stdout, err := terrDes.StdoutPipe()
	terrDes.Stderr = terrDes.Stdout
	terrDes.Start()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		m := scanner.Text()
		fmt.Println(m)
	}

	terrDes.Wait()
	if err != nil {
		panic(err)
	}

}
