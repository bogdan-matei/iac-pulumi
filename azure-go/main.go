package main

import (
	"github.com/pulumi/pulumi-azure-native/sdk/go/azure/resources"
	"github.com/pulumi/pulumi-azure-native/sdk/go/azure/storage"
	"github.com/pulumi/pulumi-azure/sdk/v4/go/azure/compute"
	"github.com/pulumi/pulumi-azure/sdk/v4/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		// Create an Azure Resource Group
		resourceGroup, err := resources.NewResourceGroup(ctx, "resourceGroup", nil)
		if err != nil {
			return err
		}

		// Create an Azure resource (Storage Account)
		account, err := storage.NewStorageAccount(ctx, "sa", &storage.StorageAccountArgs{
			ResourceGroupName: resourceGroup.Name,
			Sku: &storage.SkuArgs{
				Name: storage.SkuName_Standard_LRS,
			},
			Kind: storage.KindStorageV2,
		})
		if err != nil {
			return err
		}

		// Export the primary key of the Storage Account
		ctx.Export("primaryStorageKey", pulumi.All(resourceGroup.Name, account.Name).ApplyT(
			func(args []interface{}) (string, error) {
				resourceGroupName := args[0].(string)
				accountName := args[1].(string)
				accountKeys, err := storage.ListStorageAccountKeys(ctx, &storage.ListStorageAccountKeysArgs{
					ResourceGroupName: resourceGroupName,
					AccountName:       accountName,
				})
				if err != nil {
					return "", err
				}

				return accountKeys.Keys[0].Value, nil
			},
		))

		mainVirtualNetwork, err := network.NewVirtualNetwork(ctx, "mainVirtualNetwork", &network.VirtualNetworkArgs{
			AddressSpaces: pulumi.StringArray{
				pulumi.String("10.0.0.0/16"),
			},
			Location:          resourceGroup.Location,
			ResourceGroupName: resourceGroup.Name,
		})

		if err != nil {
			return err
		}

		internal, err := network.NewSubnet(ctx, "internal", &network.SubnetArgs{
			ResourceGroupName:  resourceGroup.Name,
			VirtualNetworkName: mainVirtualNetwork.Name,
			AddressPrefixes: pulumi.StringArray{
				pulumi.String("10.0.2.0/24"),
			},
			Name: pulumi.String("private-vnet-subnet"),
		})

		if err != nil {
			return err
		}

		publicIp, err := network.NewPublicIp(ctx, "publicControlPlaneIp", &network.PublicIpArgs{
			AllocationMethod:     pulumi.String("Static"),
			DomainNameLabel:      pulumi.StringPtr("wooferiuskubeadmtraining"),
			IdleTimeoutInMinutes: pulumi.IntPtr(5),
			ResourceGroupName:    resourceGroup.Name,
			Sku:                  pulumi.String("Standard"),
		})

		if err != nil {
			return err
		}

		mainNetworkInterface, err := network.NewNetworkInterface(ctx, "mainNetworkInterface", &network.NetworkInterfaceArgs{
			Location:          resourceGroup.Location,
			ResourceGroupName: resourceGroup.Name,
			IpConfigurations: network.NetworkInterfaceIpConfigurationArray{
				&network.NetworkInterfaceIpConfigurationArgs{
					Name:                       pulumi.String("ipconfig1"),
					SubnetId:                   internal.ID(),
					PrivateIpAddressAllocation: pulumi.String("Dynamic"),
					PublicIpAddressId:          publicIp.ID(),
				},
			},
		})

		if err != nil {
			return err
		}

		_, err = compute.NewLinuxVirtualMachine(ctx, "clusterControlPlane", &compute.LinuxVirtualMachineArgs{
			Location:          resourceGroup.Location,
			ResourceGroupName: resourceGroup.Name,
			NetworkInterfaceIds: pulumi.StringArray{
				mainNetworkInterface.ID(),
			},

			AdminUsername: pulumi.String("wooferius"),
			AdminSshKeys: compute.LinuxVirtualMachineAdminSshKeyArray{
				&compute.LinuxVirtualMachineAdminSshKeyArgs{
					Username:  pulumi.String("wooferius"),
					PublicKey: cfg.RequireSecret("public-key"),
				},
			},
			Size: pulumi.String("Standard_DS1_v2"),
			SourceImageReference: &compute.LinuxVirtualMachineSourceImageReferenceArgs{
				Publisher: pulumi.String("Canonical"),
				Offer:     pulumi.String("UbuntuServer"),
				Sku:       pulumi.String("18.04-LTS"),
				Version:   pulumi.String("latest"),
			},
			OsDisk: &compute.LinuxVirtualMachineOsDiskArgs{
				Caching:            pulumi.String("ReadWrite"),
				StorageAccountType: pulumi.String("Standard_LRS"),
			},
			Tags: pulumi.StringMap{
				"environment": pulumi.String(ctx.Stack()),
			},
		})

		if err != nil {
			return err
		}

		primaryNSG, err := network.NewNetworkSecurityGroup(ctx, "primaryNSG", &network.NetworkSecurityGroupArgs{
			Location:          resourceGroup.Location,
			ResourceGroupName: resourceGroup.Name,
			SecurityRules: network.NetworkSecurityGroupSecurityRuleArray{
				&network.NetworkSecurityGroupSecurityRuleArgs{
					Name:                     pulumi.String("allowRemote"),
					Priority:                 pulumi.Int(100),
					Direction:                pulumi.String("Inbound"),
					Access:                   pulumi.String("Allow"),
					Protocol:                 pulumi.String("Tcp"),
					SourcePortRange:          pulumi.String("*"),
					SourceAddressPrefix:      cfg.RequireSecret("sourceIp"),
					DestinationAddressPrefix: pulumi.String("*"),
					DestinationPortRange:     pulumi.String("22"),
				},
			},
		})

		if err != nil {
			return err
		}

		_, err = network.NewSubnetNetworkSecurityGroupAssociation(ctx, "primarySubnetSGAssoc", &network.SubnetNetworkSecurityGroupAssociationArgs{
			SubnetId:               internal.ID(),
			NetworkSecurityGroupId: primaryNSG.ID(),
		})

		if err != nil {
			return err
		}

		return nil
	})
}
