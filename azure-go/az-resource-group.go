package main

import (
	"fmt"

	// "github.com/google/uuid"
	"github.com/pulumi/pulumi-azure-native/sdk/go/azure/resources"
	"github.com/pulumi/pulumi-azure-native/sdk/go/azure/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)


func CreateResourceGroup(ctx *pulumi.Context,random string) (*resources.ResourceGroup, error) {
	// Create an Azure Resource Group
	resourceGroup, err := resources.NewResourceGroup(ctx, fmt.Sprintf("rg-%s", random), nil)
	if err != nil {
		return nil, err
	}

	// Create an Azure resource (Storage Account)
	account, err := storage.NewStorageAccount(ctx, fmt.Sprintf("sa%s", random), &storage.StorageAccountArgs{
		ResourceGroupName: resourceGroup.Name,
		Sku: &storage.SkuArgs{
			Name: storage.SkuName_Standard_LRS,
		},
		Kind: storage.KindStorageV2,
	})

	if err != nil {
		return nil, err
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

	return resourceGroup, err
}
