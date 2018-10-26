package cluster

import (
	"github.com/kubernauts/tk8/pkg/common"
	"github.com/spf13/viper"
)

// AwsCredentials defines the structure to hold AWS auth credentials.
type AwsCredentials struct {
	AwsAccessKeyID   string
	AwsSecretKey     string
	AwsAccessSSHKey  string
	AwsDefaultRegion string
}

var (
	kubesprayVersion = "version-0-4"
)

// DistOS defines the structure to hold the dist OS informations.
// It is possible to easily extend the list of OS.
// Append new entry to cluster.DistOSMap and use the key(string) in the config.
type DistOS struct {
	User     string
	AmiOwner string
	OS       string
}

// DistOSMap holds the main OS distrubution mapping informations.
var DistOSMap = map[string]DistOS{
	"centos": DistOS{
		User:     "centos",
		AmiOwner: "688023202711",
		OS:       "dcos-centos7-*",
	},
	"ubuntu": DistOS{
		User:     "ubuntu",
		AmiOwner: "099720109477",
		OS:       "ubuntu/images/hvm-ssd/ubuntu-xenial-16.04-amd64-*",
	},
	"coreos": DistOS{
		User:     "core",
		AmiOwner: "595879546273",
		OS:       "CoreOS-stable-*",
	},
}

// ClusterConfig holds the info required to create a cluster.
// This value is read from the config.yaml file through viper.
type ClusterConfig struct {
	AwsClusterName               string
	AwsVpcCidrBlock              string
	AwsCidrSubnetsPrivate        string
	AwsCidrSubnetsPublic         string
	AwsBastionSize               string
	AwsKubeMasterNum             string
	AwsKubeMasterSize            string
	AwsEtcdNum                   string
	AwsEtcdSize                  string
	AwsKubeWorkerNum             string
	AwsKubeWorkerSize            string
	AwsElbAPIPort                string
	K8sSecureAPIPort             string
	KubeInsecureApiserverAddress string
}

// ReadViperConfigFile is define the config paths and read the configuration file.
func ReadViperConfigFile(configName string) {
	viper.SetConfigName(configName)
	viper.AddConfigPath(".")
	viper.AddConfigPath("/tk8")
	verr := viper.ReadInConfig() // Find and read the config file.
	if verr != nil {             // Handle errors reading the config file.
		CreateConfig()
	}
}

// GetDistConfig is used to get config details specific to a particular distribution.
// Used to determine various details such as the SSH user about the distribution.
func GetDistConfig() (string, string, string) {
	ReadViperConfigFile("config")
	awsAmiID := viper.GetString("aws.ami_id")
	awsInstanceOS := viper.GetString("aws.os")
	sshUser := viper.GetString("aws.ssh_user")
	return awsAmiID, awsInstanceOS, sshUser
}

// GetCredentials get the aws credentials from the config file.
func GetCredentials() AwsCredentials {
	ReadViperConfigFile("config")
	return AwsCredentials{
		AwsAccessKeyID:   viper.GetString("aws.aws_access_key_id"),
		AwsSecretKey:     viper.GetString("aws.aws_secret_access_key"),
		AwsAccessSSHKey:  viper.GetString("aws.aws_ssh_keypair"),
		AwsDefaultRegion: viper.GetString("aws.aws_default_region"),
	}
}

type RKEConfig struct {
	ClusterName         string
	AWSRegion           string
	RKENodeInstanceType string
	NodeCount           int
	SSHKeyPath          string
	CloudProvider       string
	NodeOS              string
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

func SetClusterName() {
	if len(common.Name) < 1 {
		config := GetRKEConfig()
		common.Name = config.ClusterName
	}
}

type EKSConfig struct {
	ClusterName         string
	AWSRegion           string
	NodeInstanceType    string
	DesiredCapacity     int
	AutoScallingMinSize int
	AutoScallingMaxSize int
	KeyPath             string
}

// GetCredentials get the aws credentials from the config file.
func GetEKSConfig() EKSConfig {
	ReadViperConfigFile("config")
	return EKSConfig{
		ClusterName:         viper.GetString("eks.cluster-name"),
		AWSRegion:           viper.GetString("eks.aws_region"),
		NodeInstanceType:    viper.GetString("eks.node-instance-type"),
		DesiredCapacity:     viper.GetInt("eks.desired-capacity"),
		AutoScallingMinSize: viper.GetInt("eks.autoscalling-min-size"),
		AutoScallingMaxSize: viper.GetInt("eks.autoscalling-max-size"),
		KeyPath:             viper.GetString("eks.key-file-path"),
	}
}

// GetClusterConfig get the configuration from the config file.
func GetClusterConfig() ClusterConfig {
	ReadViperConfigFile("config")
	return ClusterConfig{
		AwsClusterName:               viper.GetString("aws.clustername"),
		AwsVpcCidrBlock:              viper.GetString("aws.aws_vpc_cidr_block"),
		AwsCidrSubnetsPrivate:        viper.GetString("aws.aws_cidr_subnets_private"),
		AwsCidrSubnetsPublic:         viper.GetString("aws.aws_cidr_subnets_public"),
		AwsBastionSize:               viper.GetString("aws.aws_bastion_size"),
		AwsKubeMasterNum:             viper.GetString("aws.aws_kube_master_num"),
		AwsKubeMasterSize:            viper.GetString("aws.aws_kube_master_size"),
		AwsEtcdNum:                   viper.GetString("aws.aws_etcd_num"),
		AwsEtcdSize:                  viper.GetString("aws.aws_etcd_size"),
		AwsKubeWorkerNum:             viper.GetString("aws.aws_kube_worker_num"),
		AwsKubeWorkerSize:            viper.GetString("aws.aws_kube_worker_size"),
		AwsElbAPIPort:                viper.GetString("aws.aws_elb_api_port"),
		K8sSecureAPIPort:             viper.GetString("aws.k8s_secure_api_port"),
		KubeInsecureApiserverAddress: viper.GetString("aws."),
	}
}

// EnableKubeadm check for kubeadm_enable option and set the config respectively in playbook.
func EnableKubeadm() {
	ReadViperConfigFile("config")
	kubeadmEnabled := viper.GetString("aws.kubeadm_enabled")
	if kubeadmEnabled == "true" {
		viper.SetConfigName("main")
		viper.AddConfigPath("./kubespray/roles/kubespray-defaults/defaults")
		err := viper.ReadInConfig()
		common.ErrorCheck("Error reading the main.yaml config file", err)
		viper.Set("kubeadm_enabled", true)
		err = viper.WriteConfig()
		common.ErrorCheck("Error writing the main.yaml config file", err)
	}
}

func SetNetworkPlugin(clusterFolder string) {

	ReadViperConfigFile("config")
	kubeNetworkPlugin := viper.GetString("aws.kube_network_plugin")
	viper.SetConfigName("k8s-cluster")
	viper.AddConfigPath(clusterFolder)
	err := viper.ReadInConfig()
	common.ErrorCheck("Error reading the main.yaml config file", err)
	if len(kubeNetworkPlugin) > 3 {
		viper.Set("kube_network_plugin", kubeNetworkPlugin)
		err = viper.WriteConfig()
	}

}
