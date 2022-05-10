package main

import (
	"fmt"

	"github.com/pulumi/pulumi-azure/sdk/v4/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v4/go/azure/compute"
	"github.com/pulumi/pulumi-azure/sdk/v4/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type AzureVirtualMachineBlueprint struct {
	Subnet            AzureSubnetBlueprint
	NetworkInterfaces []*network.NetworkInterface
	Provider *azure.Provider
	AdminUsername        pulumi.String
	AdminSshKeys         *compute.LinuxVirtualMachineAdminSshKeyArray
	SourceImageReference *compute.LinuxVirtualMachineSourceImageReferenceArgs
	OsDisk               *compute.LinuxVirtualMachineOsDiskArgs
	Size pulumi.String
	CustomData *pulumi.String
	Tags pulumi.StringMap
}

var defaultImageRefArgs = compute.LinuxVirtualMachineSourceImageReferenceArgs{
	Publisher: pulumi.String("Canonical"),
	Offer:     pulumi.String("UbuntuServer"),
	Sku:       pulumi.String("18.04-LTS"),
	Version:   pulumi.String("latest"),
}

var defaultOsDiskArgs = compute.LinuxVirtualMachineOsDiskArgs{
	Caching:            pulumi.String("ReadWrite"),
	StorageAccountType: pulumi.String("Standard_LRS"),
}

func (obj *AzureVirtualMachineBlueprint) UpdateSshKeys(ctx *pulumi.Context, sshKeys map[string]pulumi.StringInput){

	keyArrays := compute.LinuxVirtualMachineAdminSshKeyArray{}
	for key, sshKey := range sshKeys{
		keyArrays = append(keyArrays, &compute.LinuxVirtualMachineAdminSshKeyArgs{
			Username:  pulumi.String(key),
			PublicKey: sshKey,
		})
	}

	obj.AdminSshKeys = &keyArrays
}

func (obj *AzureVirtualMachineBlueprint) CreateLinuxVM(ctx *pulumi.Context, name string) error {

	var networkInterfacesIds pulumi.StringArray
	random := RandomString(6)

	for i := 0; i < len(obj.NetworkInterfaces); i++ {
		networkInterfacesIds = append(networkInterfacesIds, obj.NetworkInterfaces[i].ID())
	}

	_, err := compute.NewLinuxVirtualMachine(ctx, fmt.Sprintf("%s-%s", name, random), &compute.LinuxVirtualMachineArgs{
		Location:            obj.Subnet.ResourceGroup.Location,
		ResourceGroupName:   obj.Subnet.ResourceGroup.Name,
		NetworkInterfaceIds: networkInterfacesIds,

		AdminUsername:        obj.AdminUsername,
		AdminSshKeys:         obj.AdminSshKeys,
		Size:                 obj.Size,
		SourceImageReference: obj.SourceImageReference,
		OsDisk:               obj.OsDisk,
		CustomData:           obj.CustomData,
		Tags:                 obj.Tags,
	}, pulumi.Provider(obj.Provider), pulumi.DeleteBeforeReplace(true))

	return err
}
