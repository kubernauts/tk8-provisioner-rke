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
	"os"
	"os/exec"
	"time"

	"github.com/kubernauts/tk8-provisioner-rke/internal/templates"
)

var ec2IP string

func distSelect() (string, string) {
	//Read Configuration File
	AmiID, InstanceOS, sshUser := GetDistConfig()

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
		DistOSMap["custom"] = DistOS{
			User:     sshUser,
			AmiOwner: AmiID,
			OS:       "custom",
		}
	}

	return DistOSMap[InstanceOS].User, InstanceOS
}

func prepareConfigFiles(InstanceOS string) {
	if InstanceOS == "custom" {
		ParseTemplate(templates.CustomInfrastructure, "./inventory/"+Name+"/provisioner/create-infrastructure.tf", DistOSMap[InstanceOS])
	} else {
		ParseTemplate(templates.Infrastructure, "./inventory/"+Name+"/provisioner/create-infrastructure.tf", DistOSMap[InstanceOS])
	}

	ParseTemplate(templates.Credentials, "./inventory/"+Name+"/provisioner/credentials.tfvars", GetCredentials())
	ParseTemplate(templates.Variables, "./inventory/"+Name+"/provisioner/variables.tf", DistOSMap[InstanceOS])
	ParseTemplate(templates.Terraform, "./inventory/"+Name+"/provisioner/terraform.tfvars", GetClusterConfig())
}

func prepareInventoryGroupAllFile(fileName string) *os.File {
	groupVars, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0600)
	ErrorCheck("Error while trying to open "+fileName+": %v.", err)
	return groupVars
}

func prepareInventoryClusterFile(fileName string) *os.File {
	k8sClusterFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0600)
	defer k8sClusterFile.Close()
	ErrorCheck("Error while trying to open "+fileName+": %v.", err)
	fmt.Fprintf(k8sClusterFile, "kubeconfig_localhost: true\n")
	return k8sClusterFile
}

// Create is used to create a infrastructure on .
func Create() {

	if _, err := os.Stat("./inventory/" + Name + "/provisioner/.terraform"); err == nil {
		fmt.Println("Configuration folder already exists")
	} else {
		sshUser, osLabel := distSelect()
		fmt.Printf("Prepairing Setup for user %s on %s\n", sshUser, osLabel)
		os.MkdirAll("./inventory/"+Name+"/provisioner", 0755)
		err := exec.Command("cp", "-rfp", "./kubespray/contrib/terraform//.", "./inventory/"+Name+"/provisioner").Run()
		ErrorCheck("provisioner could not provided: %v", err)
		prepareConfigFiles(osLabel)
		ExecuteTerraform("init", "./inventory/"+Name+"/provisioner/")
	}

	ExecuteTerraform("apply", "./inventory/"+Name+"/provisioner/")

	// waiting for Loadbalancer and other not completed stuff
	fmt.Println("Infrastructure is upcoming.")
	time.Sleep(15 * time.Second)
	return

}

// Destroy is used to destroy the infrastructure created.
func Destroy() {
	// Check if credentials file exist, if it exists skip asking to input the  values
	if _, err := os.Stat("./inventory/" + Name + "/provisioner/credentials.tfvars"); err == nil {
		fmt.Println("Credentials file already exists, creation skipped")
	} else {

		ParseTemplate(templates.Credentials, "./inventory/"+Name+"/provisioner/credentials.tfvars", GetCredentials())
	}
	cpHost := exec.Command("cp", "./inventory/"+Name+"/hosts", "./inventory/hosts")
	cpHost.Run()
	cpHost.Wait()

	ExecuteTerraform("destroy", "./inventory/"+Name+"/provisioner/")

	exec.Command("rm", "./inventory/hosts").Run()
	exec.Command("rm", "-rf", "./inventory/"+Name).Run()

	return
}

// Scale is used to scale the  infrastructure and Kubernetes
func Scale() {
	var confirmation string
	// Scale the  infrastructure
	fmt.Printf("\t\t===============Starting  Scaling====================\n\n")
	_, osLabel := distSelect()
	prepareConfigFiles(osLabel)
	ExecuteTerraform("apply", "./inventory/"+Name+"/provisioner/")
	mvHost := exec.Command("mv", "./inventory/hosts", "./inventory/"+Name+"/provisioner/hosts")
	mvHost.Run()
	mvHost.Wait()

	// Scale the Kubernetes cluster

	return
}

// Reset is used to reset the  infrastructure and removing Kubernetes from it.
func Reset() {

	return
}

func Remove() {
	NotImplemented()
}
