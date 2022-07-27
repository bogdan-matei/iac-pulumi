package main

import (
	"fmt"

	"github.com/pulumi/pulumi-azure/sdk/v4/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v4/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

var SUBNET_PUBLIC_KEY_NAME = "public"
var SUBNET_PRIVATE_KEY_NAME = "private"

var SUBNET_SIZE_LIMIT = map[string]int{
	SUBNET_PUBLIC_KEY_NAME:  10,
	SUBNET_PRIVATE_KEY_NAME: 30,
}

type ClusterConfig struct {
	DnsName     string
	WorkerNodes int
}

func createVM(ctx *pulumi.Context, name string, subnet AzureSubnetBlueprint, provider *azure.Provider, publicIp *network.PublicIp) *AzureVirtualMachineBlueprint {
	config := config.New(ctx, "")

	AdminUsername := pulumi.String("wooferius")

	// First subnet will be used for the controlPlane
	nic, err := subnet.CreateNetInterface(ctx, publicIp)

	errorHandle(err, false)

	linuxVM := AzureVirtualMachineBlueprint{
		Subnet:               subnet,
		Provider:             provider,
		Size:                 pulumi.String(config.Require("defaultNodeSku")),
		SourceImageReference: &defaultImageRefArgs,
		OsDisk:               &defaultOsDiskArgs,
		AdminUsername:        AdminUsername,
		NetworkInterfaces:    []*network.NetworkInterface{},
		Tags: pulumi.StringMap{
			"environment": pulumi.String(ctx.Stack()),
		},
	}

	controlPlaneSshKeys := map[string]pulumi.StringInput{
		string(AdminUsername): config.RequireSecret("public-key"),
	}

	linuxVM.NetworkInterfaces = append(linuxVM.NetworkInterfaces, &nic)
	linuxVM.UpdateSshKeys(ctx, controlPlaneSshKeys)

	linuxVM.CreateLinuxVM(ctx, name)

	return &linuxVM
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		config := config.New(ctx, "")

		var clusterConfig ClusterConfig
		config.RequireObject("cluster", &clusterConfig)

		core_randomId := RandomString(6)
		rg, err := CreateResourceGroup(ctx, core_randomId)

		errorHandle(err, true)

		net := AzureNetworkBlueprint{ResourceGroup: *rg}

		vnetPrefixes := pulumi.StringArray{pulumi.String("10.0.0.0/16")}
		err = net.CreateVirtualNetwork(ctx, core_randomId, vnetPrefixes)

		errorHandle(err, true)

		publicSubnets := net.CreateSubnetCollection(ctx, SUBNET_PUBLIC_KEY_NAME, 2, "10.0.0.0/24")

		publicSubnetRules := network.NetworkSecurityGroupSecurityRuleArray{
			&network.NetworkSecurityGroupSecurityRuleArgs{
				Name:                     pulumi.String("externalK8sApiServer"),
				Priority:                 pulumi.Int(110),
				Direction:                pulumi.String("Inbound"),
				Access:                   pulumi.String("Allow"),
				Protocol:                 pulumi.String("Tcp"),
				SourcePortRange:          pulumi.String("*"),
				SourceAddressPrefix:      pulumi.String("*"),
				DestinationAddressPrefix: pulumi.String("*"),
				DestinationPortRange:     pulumi.String("6443"),
			},
		}

		privateSubnetRules := network.NetworkSecurityGroupSecurityRuleArray{
			&network.NetworkSecurityGroupSecurityRuleArgs{
				Name:                     pulumi.String("allowRemoteSSH"),
				Priority:                 pulumi.Int(100),
				Direction:                pulumi.String("Inbound"),
				Access:                   pulumi.String("Allow"),
				Protocol:                 pulumi.String("Tcp"),
				SourcePortRange:          pulumi.String("*"),
				// for the moment the controlPlane will be deployed in the first subnet. This needs to be changed to support multiple subents
				SourceAddressPrefixes:    publicSubnets[0].Subnet.AddressPrefixes,
				DestinationAddressPrefix: pulumi.String("*"),
				DestinationPortRange:     pulumi.String("22"),
			},
		}

		for i := 0 ; i < len(publicSubnets); i++ {
			publicSubnets[i].CreateSecurityGroup(ctx, publicSubnetRules)
		}

		privateSubnets := net.CreateSubnetCollection(ctx, SUBNET_PRIVATE_KEY_NAME, 3, "10.0.20.0/24")

		for i := 0 ; i < len(privateSubnets); i++ {
			privateSubnets[i].CreateSecurityGroup(ctx, privateSubnetRules)
		}

		controlPlanePublicIP, err := net.CreatePublicIP(ctx, "controlPlane", pulumi.String(clusterConfig.DnsName))

		errorHandle(err, false)

		computeProvider, err := azure.NewProvider(ctx, "computeProvider", &azure.ProviderArgs{
			Features: azure.ProviderFeaturesArgs{
				VirtualMachine: azure.ProviderFeaturesVirtualMachineArgs{
					DeleteOsDiskOnDeletion: pulumi.Bool(true),
				},
				ResourceGroup: azure.ProviderFeaturesResourceGroupArgs{
					PreventDeletionIfContainsResources: pulumi.Bool(true),
				},
			},
		})

		errorHandle(err, true)

		// this is the ControlPlane, it has assigned a publicIp
		createVM(ctx, "controlPlane",*publicSubnets[0], computeProvider, &controlPlanePublicIP)

		if publicSubnets != nil && privateSubnets != nil {
			fmt.Println(controlPlanePublicIP.ID())
			fmt.Println(computeProvider.ID())
		}

		// these are the VM's for the worker nodes
		for i := 0; i < clusterConfig.WorkerNodes; i++ {
			createVM(ctx, fmt.Sprintf("workerNode-%d", i), *privateSubnets[0], computeProvider, nil)
		}

		return nil
	})
}
