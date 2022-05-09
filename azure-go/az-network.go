package main

import (
	"fmt"

	"github.com/pulumi/pulumi-azure-native/sdk/go/azure/resources"
	"github.com/pulumi/pulumi-azure/sdk/v4/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"errors"
	"strconv"
	"strings"
)

type AzureNetworkBlueprint struct {
	VirtualNetwork *network.VirtualNetwork
	PublicSubnets  []*AzureSubnetBlueprint
	PrivateSubnets []*AzureSubnetBlueprint
	PublicIps      []*network.PublicIp
	ResourceGroup  resources.ResourceGroup
}

type AzureSubnetBlueprint struct {
	Subnet         *network.Subnet
	NetInterfaces  []*network.NetworkInterface
	SecurityGroups []*network.NetworkSecurityGroup
	ResourceGroup  resources.ResourceGroup
}

func (obj *AzureNetworkBlueprint) CreateVirtualNetwork(ctx *pulumi.Context, random string, addressSpace pulumi.StringArray) error {

	tmpNet, err := network.NewVirtualNetwork(ctx, fmt.Sprintf("vnet-%s", random), &network.VirtualNetworkArgs{
		AddressSpaces:     addressSpace,
		Location:          obj.ResourceGroup.Location,
		ResourceGroupName: obj.ResourceGroup.Name,
	})

	obj.VirtualNetwork = tmpNet
	return err
}

func (obj *AzureNetworkBlueprint) CreateSubnet(ctx *pulumi.Context, name string, addressPrefixes pulumi.StringArray) (AzureSubnetBlueprint, error) {
	random := RandomString(6)
	
	subnet, err := network.NewSubnet(ctx, fmt.Sprintf("%s-%s", name, random), &network.SubnetArgs{
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

func (obj *AzureNetworkBlueprint) CreatePublicIP(ctx *pulumi.Context, name string, domainName string) (network.PublicIp, error) {

	publicIp, err := network.NewPublicIp(ctx, fmt.Sprintf("%s", name), &network.PublicIpArgs{
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
	random := RandomString(6)

	nic, err := network.NewNetworkInterface(ctx, fmt.Sprintf("nic-%s", random), &network.NetworkInterfaceArgs{
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
	random := RandomString(6)

	nsg, err := network.NewNetworkSecurityGroup(ctx, fmt.Sprintf("nsg-%s", random), &network.NetworkSecurityGroupArgs{
		Location:          obj.ResourceGroup.Location,
		ResourceGroupName: obj.ResourceGroup.Location,
		SecurityRules:     rules,
	})

	if err != nil {
		return err
	}

	_, err = network.NewSubnetNetworkSecurityGroupAssociation(ctx, fmt.Sprintf("asoc-%s", random), &network.SubnetNetworkSecurityGroupAssociationArgs{
		SubnetId:               obj.Subnet.ID(),
		NetworkSecurityGroupId: nsg.ID(),
	}, pulumi.DependsOn([]pulumi.Resource{nsg}))

	obj.SecurityGroups = append(obj.SecurityGroups, nsg)

	return err
}

func (obj AzureNetworkBlueprint) CreateSubnetCollection(ctx *pulumi.Context, key string, len int, startCidr string) []*AzureSubnetBlueprint{

	if len > SUBNET_SIZE_LIMIT[key]{
		panic(errors.New(fmt.Sprintf("LimitTresholdExceeded: The current treshold limit is %d and your input was %d", SUBNET_SIZE_LIMIT[key], len)))
	}

	subnets := []*AzureSubnetBlueprint{}

	for i := 0; i < len; i++ {
		cidr := strings.Split(startCidr, ".")
		intVal, err := strconv.Atoi(cidr[2])
		errorHandle(err, true)
		cidr[2] = strconv.Itoa(intVal + i)
		
		subnet, err := obj.CreateSubnet(ctx, key, pulumi.StringArray{pulumi.String(strings.Join(cidr, "."))})
		errorHandle(err, true)
	
		subnets = append(subnets, &subnet)

		if key == SUBNET_PUBLIC_KEY_NAME{
			obj.PublicSubnets = append(obj.PublicSubnets, &subnet)
		} else {
			obj.PrivateSubnets = append(obj.PrivateSubnets, &subnet)
		}
	}

	// return object as we want to make changes on it
	return subnets
}
