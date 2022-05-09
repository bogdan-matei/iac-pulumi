package main

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"os"
)

var SUBNET_PUBLIC_KEY_NAME = "public"
var SUBNET_PRIVATE_KEY_NAME = "private"

var SUBNET_SIZE_LIMIT = map[string]int{
	SUBNET_PUBLIC_KEY_NAME: 10,
	SUBNET_PRIVATE_KEY_NAME: 30,
}

func errorHandle(err error, crit bool) {
	if err != nil {
		fmt.Println(err)

		if crit {
			os.Exit(1)
		}
	}
}


func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// config := config.New(ctx,"")

		core_randomId := RandomString(6)
		rg, err := CreateResourceGroup(ctx, core_randomId)

		errorHandle(err, true)

		net := AzureNetworkBlueprint{ResourceGroup: *rg}

		vnetPrefixes := pulumi.StringArray{ pulumi.String("10.0.0.0/16") }
		err = net.CreateVirtualNetwork(ctx, core_randomId,vnetPrefixes)

		errorHandle(err,true)

		publicSubnets := net.CreateSubnetCollection(ctx, SUBNET_PUBLIC_KEY_NAME, 2, "10.0.0.0/24")

		if publicSubnets != nil {

		}

		return nil
	})
}
