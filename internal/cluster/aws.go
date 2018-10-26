// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cluster

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kubernauts/tk8/pkg/common"
	"github.com/kubernauts/tk8/pkg/installer"
	"github.com/kubernauts/tk8/pkg/provisioner"
	"github.com/kubernauts/tk8/pkg/templates"
)

var ec2IP string

func distSelect() (string, string) {
	//Read Configuration File
	awsAmiID, awsInstanceOS, sshUser := GetDistConfig()

	if awsAmiID != "" && sshUser == "" {
		log.Fatal("SSH Username is required when using custom AMI")
		return "", ""
	}
	if awsAmiID == "" && awsInstanceOS == "" {
		log.Fatal("Provide either of AMI ID or OS in the config file.")
		return "", ""
	}

	if awsAmiID != "" && sshUser != "" {
		awsInstanceOS = "custom"
		DistOSMap["custom"] = DistOS{
			User:     sshUser,
			AmiOwner: awsAmiID,
			OS:       "custom",
		}
	}

	return DistOSMap[awsInstanceOS].User, awsInstanceOS
}

func prepareConfigFiles(awsInstanceOS string) {
	if awsInstanceOS == "custom" {
		templates.ParseTemplate(templates.CustomInfrastructure, "./inventory/"+common.Name+"/provisioner/create-infrastructure.tf", DistOSMap[awsInstanceOS])
	} else {
		templates.ParseTemplate(templates.Infrastructure, "./inventory/"+common.Name+"/provisioner/create-infrastructure.tf", DistOSMap[awsInstanceOS])
	}

	templates.ParseTemplate(templates.Credentials, "./inventory/"+common.Name+"/provisioner/credentials.tfvars", GetCredentials())
	templates.ParseTemplate(templates.Variables, "./inventory/"+common.Name+"/provisioner/variables.tf", DistOSMap[awsInstanceOS])
	templates.ParseTemplate(templates.Terraform, "./inventory/"+common.Name+"/provisioner/terraform.tfvars", GetClusterConfig())
}

func prepareInventoryGroupAllFile(fileName string) *os.File {
	groupVars, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0600)
	common.ErrorCheck("Error while trying to open "+fileName+": %v.", err)
	return groupVars
}

func prepareInventoryClusterFile(fileName string) *os.File {
	k8sClusterFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0600)
	defer k8sClusterFile.Close()
	common.ErrorCheck("Error while trying to open "+fileName+": %v.", err)
	fmt.Fprintf(k8sClusterFile, "kubeconfig_localhost: true\n")
	return k8sClusterFile
}

// AWSCreate is used to create a infrastructure on AWS.
func AWSCreate() {
	SetClusterName()
	if _, err := os.Stat("./inventory/" + common.Name + "/provisioner/.terraform"); err == nil {
		fmt.Println("Configuration folder already exists")
	} else {
		sshUser, osLabel := distSelect()
		fmt.Printf("Prepairing Setup for user %s on %s\n", sshUser, osLabel)
		os.MkdirAll("./inventory/"+common.Name+"/provisioner", 0755)
		err := exec.Command("cp", "-rfp", "./kubespray/contrib/terraform/aws/.", "./inventory/"+common.Name+"/provisioner").Run()
		common.ErrorCheck("provisioner could not provided: %v", err)
		prepareConfigFiles(osLabel)
		provisioner.ExecuteTerraform("init", "./inventory/"+common.Name+"/provisioner/")
	}

	provisioner.ExecuteTerraform("apply", "./inventory/"+common.Name+"/provisioner/")

	// waiting for Loadbalancer and other not completed stuff
	fmt.Println("Infrastructure is upcoming.")
	time.Sleep(15 * time.Second)
	return

}

// AWSInstall is used for installing Kubernetes on the available infrastructure.
func AWSInstall() {
	// check if ansible is installed
	common.DependencyCheck("ansible")
	SetClusterName()
	// Copy the configuraton files as indicated in the kubespray docs
	if _, err := os.Stat("./inventory/" + common.Name + "/installer"); err == nil {
		fmt.Println("Configuration folder already exists")
	} else {
		os.MkdirAll("./inventory/"+common.Name+"/installer", 0755)
		mvHost := exec.Command("mv", "./inventory/hosts", "./inventory/"+common.Name+"/hosts")
		mvHost.Run()
		mvHost.Wait()
		mvShhBastion := exec.Command("cp", "./kubespray/ssh-bastion.conf", "./inventory/"+common.Name+"/ssh-bastion.conf")
		mvShhBastion.Run()
		mvShhBastion.Wait()
		//os.MkdirAll("./inventory/"+common.Name+"/installer/group_vars", 0755)
		cpSample := exec.Command("cp", "-rfp", "./kubespray/inventory/sample/.", "./inventory/"+common.Name+"/installer/")
		cpSample.Run()
		cpSample.Wait()

		cpKube := exec.Command("cp", "-rfp", "./kubespray/.", "./inventory/"+common.Name+"/installer/")
		cpKube.Run()
		cpKube.Wait()

		mvInstallerHosts := exec.Command("cp", "./inventory/"+common.Name+"/hosts", "./inventory/"+common.Name+"/installer/hosts")
		mvInstallerHosts.Run()
		mvInstallerHosts.Wait()
		mvProvisionerHosts := exec.Command("cp", "./inventory/"+common.Name+"/hosts", "./inventory/"+common.Name+"/installer/hosts")
		mvProvisionerHosts.Run()
		mvProvisionerHosts.Wait()

		// Check if Kubeadm is enabled
		EnableKubeadm()

		//Start Kubernetes Installation
		//Enable load balancer api access and copy the kubeconfig file locally
		loadBalancerName, err := exec.Command("sh", "-c", "grep apiserver_loadbalancer_domain_name= ./inventory/"+common.Name+"/installer/hosts | cut -d'=' -f2").CombinedOutput()

		if err != nil {
			fmt.Println("Problem getting the load balancer domain name", err)
		} else {
			var groupVars *os.File
			//Make a copy of kubeconfig on Ansible host
			if kubesprayVersion == "develop" {
				// Set Kube Network Proxy
				SetNetworkPlugin("./inventory/" + common.Name + "/installer/group_vars/k8s-cluster")
				prepareInventoryClusterFile("./inventory/" + common.Name + "/installer/group_vars/k8s-cluster/k8s-cluster.yml")
				groupVars = prepareInventoryGroupAllFile("./inventory/" + common.Name + "/installer/group_vars/all/all.yml")
			} else {
				// Set Kube Network Proxy
				SetNetworkPlugin("./inventory/" + common.Name + "/installer/group_vars")
				prepareInventoryClusterFile("./inventory/" + common.Name + "/installer/group_vars/k8s-cluster.yml")
				groupVars = prepareInventoryGroupAllFile("./inventory/" + common.Name + "/installer/group_vars/all.yml")
			}
			defer groupVars.Close()
			// Resolve Load Balancer Domain Name and pick the first IP

			elbNameRaw, _ := exec.Command("sh", "-c", "grep apiserver_loadbalancer_domain_name= ./inventory/"+common.Name+"/installer/hosts | cut -d'=' -f2 | sed 's/\"//g'").CombinedOutput()

			// Convert the Domain name to string, strip all spaces so that Lookup does not return errors
			elbName := strings.TrimSpace(string(elbNameRaw))
			fmt.Println(elbName)
			node, err := net.LookupHost(elbName)
			common.ErrorCheck("Error resolving ELB name: %v", err)
			elbIP := node[0]
			fmt.Println(node)

			DomainName := strings.TrimSpace(string(loadBalancerName))
			loadBalancerDomainName := "apiserver_loadbalancer_domain_name: " + DomainName

			fmt.Fprintf(groupVars, "#Set cloud provider to AWS\n")
			fmt.Fprintf(groupVars, "cloud_provider: 'aws'\n")
			fmt.Fprintf(groupVars, "#Load Balancer Configuration\n")
			fmt.Fprintf(groupVars, "loadbalancer_apiserver_localhost: false\n")
			fmt.Fprintf(groupVars, "%s\n", loadBalancerDomainName)
			fmt.Fprintf(groupVars, "loadbalancer_apiserver:\n")
			fmt.Fprintf(groupVars, "  address: %s\n", elbIP)
			fmt.Fprintf(groupVars, "  port: 6443\n")
		}
	}

	sshUser, osLabel := distSelect()
	installer.RunPlaybook("./inventory/"+common.Name+"/installer/", "cluster.yml", sshUser, osLabel)

	return
}

// AWSDestroy is used to destroy the infrastructure created.
func AWSDestroy() {
	SetClusterName()
	// Check if credentials file exist, if it exists skip asking to input the AWS values
	if _, err := os.Stat("./inventory/" + common.Name + "/provisioner/credentials.tfvars"); err == nil {
		fmt.Println("Credentials file already exists, creation skipped")
	} else {

		templates.ParseTemplate(templates.Credentials, "./inventory/"+common.Name+"/provisioner/credentials.tfvars", GetCredentials())
	}
	cpHost := exec.Command("cp", "./inventory/"+common.Name+"/hosts", "./inventory/hosts")
	cpHost.Run()
	cpHost.Wait()

	provisioner.ExecuteTerraform("destroy", "./inventory/"+common.Name+"/provisioner/")

	exec.Command("rm", "./inventory/hosts").Run()
	exec.Command("rm", "-rf", "./inventory/"+common.Name).Run()

	return
}

// AWSScale is used to scale the AWS infrastructure and Kubernetes
func AWSScale() {
	SetClusterName()
	// Scale the AWS infrastructure
	fmt.Printf("\t\t===============Starting AWS Scaling====================\n\n")
	sshUser, osLabel := distSelect()
	prepareConfigFiles(osLabel)
	provisioner.ExecuteTerraform("apply", "./inventory/"+common.Name+"/provisioner/")
	mvHost := exec.Command("mv", "./inventory/hosts", "./inventory/"+common.Name+"/provisioner/hosts")
	mvHost.Run()
	mvHost.Wait()
	// waiting for Infrastructure
	time.Sleep(30)
	// Scale the Kubernetes cluster
	fmt.Printf("\n\n\t\t===============Starting Kubernetes Scaling====================\n\n")
	_, err := os.Stat("./inventory/" + common.Name + "/provisioner/hosts")
	common.ErrorCheck("No host file found.", err)
	cpHost := exec.Command("cp", "./inventory/"+common.Name+"/provisioner/hosts", "./inventory/"+common.Name+"/installer/hosts")
	cpHost.Run()
	cpHost.Wait()
	installer.RunPlaybook("./inventory/"+common.Name+"/installer/", "scale.yml", sshUser, osLabel)

	return
}

// AWSReset is used to reset the Kubernetes on your AWS infrastructure.
func AWSReset() {
	SetClusterName()
	sshUser, osLabel := distSelect()
	installer.RunPlaybook("./inventory/"+common.Name+"/installer/", "reset.yml", sshUser, osLabel)
	// waiting for Infrastructure
	time.Sleep(30)
	AWSInstall()
	return
}

// AWSRemove is used to remove Kubernetes from your AWS infrastructure
func AWSRemove() {
	SetClusterName()
	sshUser, osLabel := distSelect()
	installer.RunPlaybook("./inventory/"+common.Name+"/installer/", "reset.yml", sshUser, osLabel)
}
