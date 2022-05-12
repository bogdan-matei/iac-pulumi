# azure-go

This folder contains Pulumi code written in Go. The purpose of it is to be able to set-up a self-managed Kubernetes cluster using [kubeadm](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/).

## Structure

```bash
├── Pulumi.dev.yaml # configuration of the stack
├── Pulumi.yaml # Pulumi project information
├── README.md
├── az-compute.go # defines structure for Azure VMs + functions that can be used for that type
├── az-network.go # defines structures for Azure Network and Subnets + functions that can be used for these types
├── az-resource-group.go # defines the intial Azure resources needed to create resource (Resources Group, Storage Account)
├── az-utils.go # defines utility functions used across the project
├── go.mod # go modules specs
├── go.sum # go library file (similar to package-json.lock)
├── main.go # default path to create the infrastructure
├── node_setup.sh # bash functions used to bootstrap vm. !Needs to be broken and used inside the custom data of the VM's!
└── notes.md # notes about setting up a self-managed cluster
```

## How to use

Prequisites:
- Azure account
  - `az login` is required before you can run the pulumi code
- Pulumi Account
  - `pulumi.com` backend is used, visit https://app.pulumi.com/ to create and account
- Pulumi CLI, Go, Plugin (equivalent of terraform providers)
  - Below you have a partial output of the `pulumi about` command
```
CLI          
Version      3.32.1
Go Version   go1.18.1
Go Compiler  gc

Plugins
NAME          VERSION
azure         4.42.0
azure-native  1.64.0
go            unknown

Current Stack: dev

Backend        
Name           pulumi.com
URL            https://app.pulumi.com/bogdan-matei
User           bogdan-matei
Organizations  bogdan-matei

NAME                                       VERSION
github.com/pulumi/pulumi-azure-native/sdk  v1.64.0
github.com/pulumi/pulumi-azure/sdk/v4      v4.42.0
github.com/pulumi/pulumi/sdk/v3            v3.32.1
```

## Issues

- OsDisk's created during the creation of the LinxuVM don't get destroyed at the end
  - Disclaimer from https://www.pulumi.com/registry/packages/azure/api-docs/compute/linuxvirtualmachine/ mention that this should happen automatically
  - Currently trying to set up a custom provider to compare the behaviour
    - `warning: provider config warning: Argument is deprecated`

## To Do's

- Provide boostrap for ControlPlane node & worker nodes
- Provide easy to use scaling mechanism via pulumi
  - new nodes need to automatically join the cluster
- Refactor code if needed and make it more generic and usable
- Write documentation