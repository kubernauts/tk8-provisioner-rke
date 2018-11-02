package cluster

import (
	"github.com/kubernauts/tk8/pkg/common"
	"github.com/spf13/viper"
)

var (
	kubesprayVersion = "version-0-4"
)

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
	common.ReadViperConfigFile("config")
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
