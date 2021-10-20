package main

import (
	"fmt"

	"github.com/pulumi/pulumi-azure-native/sdk/go/azure/resources"
	"github.com/pulumi/pulumi-azure-native/sdk/go/azure/storage"
	"github.com/pulumi/pulumi-azure/sdk/v4/go/azure/compute"
	"github.com/pulumi/pulumi-azure/sdk/v4/go/azure/network"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

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
		})

		if err != nil {
			return err
		}

		mainNetworkInterface, err := network.NewNetworkInterface(ctx, "mainNetworkInterface", &network.NetworkInterfaceArgs{
			Location:          resourceGroup.Location,
			ResourceGroupName: resourceGroup.Name,
			IpConfigurations: network.NetworkInterfaceIpConfigurationArray{
				&network.NetworkInterfaceIpConfigurationArgs{
					Name:                       pulumi.String("testconfiguration1"),
					SubnetId:                   internal.ID(),
					PrivateIpAddressAllocation: pulumi.String("Dynamic"),
				},
			},
		})

		if err != nil {
			return err
		}

		password, err := random.NewRandomPassword(ctx, "password", &random.RandomPasswordArgs{
			Length:          pulumi.Int(16),
			Special:         pulumi.Bool(true),
			OverrideSpecial: pulumi.String(fmt.Sprintf("%v%v%v", "_", "%", "@")),
		})

		if err != nil {
			return err
		}

		ctx.Export("adminPassword", pulumi.All().ApplyT(func(args []interface{}) pulumi.Output {
			return password.Result
		},
		))

		_, err = compute.NewVirtualMachine(ctx, "mainVirtualMachine", &compute.VirtualMachineArgs{
			Location:          resourceGroup.Location,
			ResourceGroupName: resourceGroup.Name,
			NetworkInterfaceIds: pulumi.StringArray{
				mainNetworkInterface.ID(),
			},
			VmSize: pulumi.String("Standard_DS1_v2"),
			StorageImageReference: &compute.VirtualMachineStorageImageReferenceArgs{
				Publisher: pulumi.String("Canonical"),
				Offer:     pulumi.String("UbuntuServer"),
				Sku:       pulumi.String("18.04-LTS"),
				Version:   pulumi.String("latest"),
			},
			StorageOsDisk: &compute.VirtualMachineStorageOsDiskArgs{
				Name:            pulumi.String("myosdisk1"),
				Caching:         pulumi.String("ReadWrite"),
				CreateOption:    pulumi.String("FromImage"),
				ManagedDiskType: pulumi.String("Standard_LRS"),
			},
			OsProfile: &compute.VirtualMachineOsProfileArgs{
				ComputerName:  pulumi.String("hostname"),
				AdminUsername: pulumi.String("testadmin"),
				AdminPassword: password.Result,
			},
			OsProfileLinuxConfig: &compute.VirtualMachineOsProfileLinuxConfigArgs{
				DisablePasswordAuthentication: pulumi.Bool(false),
			},
			Tags: pulumi.StringMap{
				"environment": pulumi.String(ctx.Stack()),
			},
		})
		if err != nil {
			return err
		}

		return nil
	})
}
