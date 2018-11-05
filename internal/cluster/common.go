package cluster

import (
	"log"

	"github.com/kubernauts/tk8/pkg/common"
	"github.com/spf13/viper"
)

var (
	kubesprayVersion = "version-0-4"
)

type AwsCredentials struct {
	AwsAccessKeyID   string
	AwsSecretKey     string
	AwsAccessSSHKey  string
	AwsDefaultRegion string
}

type RKEConfig struct {
	ClusterName         string
	AWSRegion           string
	RKENodeInstanceType string
	NodeCount           int
	SSHKeyPath          string
	CloudProvider       string
	NodeOS              string
	Authorization       string
}

func GetRKEConfig() RKEConfig {
	ReadViperConfigFile("config")
	return RKEConfig{
		ClusterName:         viper.GetString("rke.cluster_name"),
		AWSRegion:           viper.GetString("rke.rke_aws_region"),
		RKENodeInstanceType: viper.GetString("rke.rke_node_instance_type"),
		NodeCount:           viper.GetInt("rke.node_count"),
		CloudProvider:       viper.GetString("rke.cloud_provider"),
		Authorization:       viper.GetString("rke.authorization"),
	}
}

func SetClusterName() {
	if len(common.Name) < 1 {
		config := GetRKEConfig()
		common.Name = config.ClusterName
	}
}

// ReadViperConfigFile is define the config paths and read the configuration file.
func ReadViperConfigFile(configName string) {
	viper.SetConfigName(configName)
	viper.AddConfigPath(".")
	viper.AddConfigPath("/tk8")
	verr := viper.ReadInConfig() // Find and read the config file.
	if verr != nil {             // Handle errors reading the config file.
		log.Fatalln(verr)
	}
}
