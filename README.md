# Starnet is galactic-scale orchestrator for distributed cosmos network.

## What is Starnet?

Starnet is a powerful and scalable testnet deployment system.

## How does it work?

injective-starnet is a Pulumi-based automation tool designed to provision and manage a full Injective validator network. It programmatically spins up a distributed network of validator nodes across GCP infrastructure, handles key distribution, configuration templating, genesis coordination, and ensures all nodes are launched with the correct Injective binary and parameters.

✨ Features

	*	Deploys N number of Injective validators and sentry nodes.
	*	Fully automates node provisioning (GCP cloud)

🛠️ Use cases

	*	Internal testing and CI environments
	*	Reproducing edge-case network conditions
	*	Rapid prototyping of new Injective modules or forks
	*	Chaos testing and performance profiling



## Requirements

* [Installed Pulumi CLI](https://www.pulumi.com/docs/iac/download-install/)
* [Installed gcloud cli](https://cloud.google.com/sdk/docs/install) (could be also installed with [brew](https://formulae.brew.sh/cask/google-cloud-sdk))
* GCP account with compute permissions.
* Pulumi account in injective organization.
* Local injective-core repo (binaries are built from your local source/branch)
* [chain-stresser installed](https://github.com/InjectiveLabs/chain-stresser?tab=readme-ov-file#installation)

## How do I use it?

### Use Github Actions to deploy Starnet

Currently only [Nuke The Network](https://github.com/InjectiveLabs/injective-starnet/actions/workflows/destroy.yaml) job works, you can use it nuke the running network.

### Use Local Machine to deploy Starnet

Clone repository

```
git clone org-44571224@github.com:InjectiveLabs/injective-starnet.git
cd injective/injective-starnet
```

Login to your GCP and Pulumi accounts
```
gcloud auth login
pulumi login
```

#### Generate network artifacts with chain stresser
Note: For testing purpose use smaller network 2-15 validators max. Otherwise GCP will compain about on-demand resource availabilities in zones (e.g you will get errors from GCP that requested HW is anavailable in zone X).

```
chain-stresser generate --instances <instance_num> --validators <validators_num> --sentries <sentries_num> --evm <evm_bool> --prod <prod_bool>
```


#### Align validators/sentries with the number of VM's we need to create

    vi Pulumi.starnet.yaml # change the nodePoolSize to fit the value you set when you where generering network artifacts.
    
#### Spin up the network

    pulumi up -y

#### When pulumi finishes, you will get validators IP's in output
    
Check the network status with `gex -h <validator_ip>:26657` 

#### To nuke the network

    pulumi destroy -y

**NOTE:** if you are not sure if you nuked network properly , go to injective-starnet repo and manually run [Nuke The Network](https://github.com/InjectiveLabs/injective-starnet/actions/workflows/destroy.yaml) job (top right corner have "run the workflow" buttin").
All ports are opened to public, p2p, RPC , gRPC
  
  
  
