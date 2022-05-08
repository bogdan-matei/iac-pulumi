package main

import (
	"fmt"

	"github.com/pulumi/pulumi-azure-native/sdk/go/azure/resources"
	"github.com/pulumi/pulumi-azure/sdk/v4/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/google/uuid"
)

type AzureNetworkBlueprint struct {
	VirtualNetwork network.VirtualNetwork
	Subnets        []*AzureSubnetBlueprint
	PublicIps      []*network.PublicIp
	ResourceGroup  resources.ResourceGroup
}

type AzureSubnetBlueprint struct {
	Subnet         *network.Subnet
	NetInterfaces  []*network.NetworkInterface
	SecurityGroups []*network.NetworkSecurityGroup
	ResourceGroup  resources.ResourceGroup
}

func (obj *AzureNetworkBlueprint) CreateVirtualNetwork(ctx *pulumi.Context, addressSpace pulumi.StringArray) (network.VirtualNetwork, error) {
	tmpNet, err := network.NewVirtualNetwork(ctx, fmt.Sprintf("vnet-%s", uuid.New()), &network.VirtualNetworkArgs{
		AddressSpaces:     addressSpace,
		Location:          obj.ResourceGroup.Location,
		ResourceGroupName: obj.ResourceGroup.Name,
	})

	return *tmpNet, err
}

func (obj *AzureNetworkBlueprint) CreateSubnet(ctx *pulumi.Context, addressPrefixes pulumi.StringArray) (AzureSubnetBlueprint, error) {

	subnet, err := network.NewSubnet(ctx, fmt.Sprintf("subnet-%s", uuid.New()), &network.SubnetArgs{
		ResourceGroupName:  obj.ResourceGroup.Name,
		VirtualNetworkName: obj.VirtualNetwork.Name,
		AddressPrefixes:    addressPrefixes,
	})

	newSubnet := AzureSubnetBlueprint{
		Subnet:        subnet,
		ResourceGroup: obj.ResourceGroup,
	}

	return newSubnet, err
}

func (obj *AzureNetworkBlueprint) CreatePublicIP(ctx *pulumi.Context, domainName string) (network.PublicIp, error) {

	publicIp, err := network.NewPublicIp(ctx, fmt.Sprintf("pubip-%s", uuid.New()), &network.PublicIpArgs{
		AllocationMethod:     pulumi.String("Static"),
		DomainNameLabel:      pulumi.StringPtr(domainName),
		IdleTimeoutInMinutes: pulumi.IntPtr(5),
		ResourceGroupName:    obj.ResourceGroup.Name,
		Sku:                  pulumi.String("Standard"),
	})

	obj.PublicIps = append(obj.PublicIps, publicIp)

	return *publicIp, err
}

func (obj *AzureSubnetBlueprint) CreateNetInterface(ctx *pulumi.Context, publicIpId pulumi.String) (network.NetworkInterface, error) {
	nic, err := network.NewNetworkInterface(ctx, fmt.Sprintf("netinterf-%s", uuid.New()), &network.NetworkInterfaceArgs{
		Location:          obj.ResourceGroup.Location,
		ResourceGroupName: obj.ResourceGroup.Name,
		IpConfigurations: network.NetworkInterfaceIpConfigurationArray{
			&network.NetworkInterfaceIpConfigurationArgs{
				SubnetId:                   obj.Subnet.ID(),
				PrivateIpAddressAllocation: pulumi.String("Dynamic"),
				// omitting this will create a privat NIC
				PublicIpAddressId: publicIpId,
			},
		},
	})

	obj.NetInterfaces = append(obj.NetInterfaces, nic)

	return *nic, err
}

func (obj *AzureSubnetBlueprint) CreateSecurityGroup(ctx *pulumi.Context, rules network.NetworkSecurityGroupSecurityRuleArray) error {

	nsg, err := network.NewNetworkSecurityGroup(ctx, fmt.Sprintf("nsg-%s", uuid.New()), &network.NetworkSecurityGroupArgs{
		Location:          obj.ResourceGroup.Location,
		ResourceGroupName: obj.ResourceGroup.Location,
		SecurityRules:     rules,
	})

	if err != nil {
		return err
	}

	_, err = network.NewSubnetNetworkSecurityGroupAssociation(ctx, fmt.Sprintf("asoc-%s", uuid.New()), &network.SubnetNetworkSecurityGroupAssociationArgs{
		SubnetId:               obj.Subnet.ID(),
		NetworkSecurityGroupId: nsg.ID(),
	}, pulumi.DependsOn([]pulumi.Resource{nsg}))

	obj.SecurityGroups = append(obj.SecurityGroups, nsg)

	return err
}
