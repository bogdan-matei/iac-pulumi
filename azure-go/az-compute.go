package main

import (
	"fmt"

	"github.com/pulumi/pulumi-azure/sdk/v4/go/azure/compute"
	"github.com/pulumi/pulumi-azure/sdk/v4/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type AzureVirtualMachineBlueprint struct {
	Subent            AzureSubnetBlueprint
	NetworkInterfaces []*network.NetworkInterface
	Config            config.Config

	// wooferius
	AdminUsername        *pulumi.String
	AdminSshKeys         *compute.LinuxVirtualMachineAdminSshKeyArray
	SourceImageReference *compute.LinuxVirtualMachineSourceImageReferenceArgs
	OsDisk               *compute.LinuxVirtualMachineOsDiskArgs
	/*
			SourceImageReference: &compute.LinuxVirtualMachineSourceImageReferenceArgs{
			Publisher: pulumi.String("Canonical"),
			Offer:     pulumi.String("UbuntuServer"),
			Sku:       pulumi.String("18.04-LTS"),
			Version:   pulumi.String("latest"),
		}

		&compute.LinuxVirtualMachineOsDiskArgs{
			Caching:            pulumi.String("ReadWrite"),
			StorageAccountType: pulumi.String("Standard_LRS"),
		}
	*/

	//Standard_DS2_v2
	Size pulumi.String

	CustomData *pulumi.String
	random         string
	// 	"environment": pulumi.String(ctx.Stack())
	Tags pulumi.StringMap
}

func (obj *AzureVirtualMachineBlueprint) CreateLinuxVM(ctx *pulumi.Context, name string) error {

	var networkInterfacesIds pulumi.StringArray

	for i := 0; i < len(obj.NetworkInterfaces); i++ {
		networkInterfacesIds = append(networkInterfacesIds, obj.NetworkInterfaces[i].ID())
	}

	_, err := compute.NewLinuxVirtualMachine(ctx, fmt.Sprintf("%s-%s", name, obj.random), &compute.LinuxVirtualMachineArgs{
		Location:            obj.Subent.ResourceGroup.Location,
		ResourceGroupName:   obj.Subent.ResourceGroup.Name,
		NetworkInterfaceIds: networkInterfacesIds,

		AdminUsername:        obj.AdminUsername,
		AdminSshKeys:         obj.AdminSshKeys,
		Size:                 obj.Size,
		SourceImageReference: obj.SourceImageReference,
		OsDisk:               obj.OsDisk,
		CustomData:           obj.CustomData,
		Tags:                 obj.Tags,
	}, pulumi.DeleteBeforeReplace(true))

	return err
}
