# ✨ Starnet: The Galactic-Scale Orchestrator for Distributed Cosmos Networks

## What is Starnet?

injective-starnet is a Pulumi-based automation tool designed to provision and manage a full Injective validator network. It programmatically spins up a distributed network of validator nodes across GCP infrastructure, handles key distribution, configuration templating, genesis coordination, and ensures all nodes are launched with the correct Injective binary and parameters built from provided git branch on remote machines.

## How does it work?

✨ Features

* Deploys N number of Injective validators and sentry nodes
* Fully automates node provisioning (GCP cloud)
* Builds and deploys injective-core binaries from any branch to remote nodes
* Secure SSH key forwarding for repository access

🛠️ Use cases

* Internal testing and CI environments
* Reproducing edge-case network conditions
* Rapid prototyping of new Injective modules or forks
* Chaos testing and performance profiling

## Requirements

* [Pulumi CLI](https://www.pulumi.com/docs/iac/download-install/)
* [Google Cloud SDK](https://cloud.google.com/sdk/docs/install) (or install via [brew](https://formulae.brew.sh/cask/google-cloud-sdk))
* GCP account with compute permissions
* Pulumi account in the Injective organization (if using your own account, you'll need to manually create the injective-starnet stack)
* [chain-stresser](https://github.com/InjectiveLabs/chain-stresser?tab=readme-ov-file#installation) tool installed
* Access to injective-core GitHub repository (we use SSH agent forwarding for secure repository access - [your keys never leave your machine](https://docs.github.com/en/authentication/connecting-to-github-with-ssh/using-ssh-agent-forwarding))

## Quick Start

### Installation

```bash
git clone org-44571224@github.com:InjectiveLabs/injective-starnet.git
cd injective/injective-starnet
make install
```

### Configuration

#### 1. Generate Network Artifacts

Use chain-stresser to generate your network configuration. For testing, we recommend starting with 2-15 validators to avoid GCP resource constraints.

```bash
chain-stresser generate \
  --instances <instance_num> \
  --validators <validators_num> \
  --sentries <sentries_num> \
  --evm <evm_bool> \
  --prod <prod_bool>
```

#### 2. Authenticate Services

Obtain credentials by logging into your GCP and Pulumi accounts:

```bash
# Login to Google Cloud
gcloud auth login

# Login to Pulumi
pulumi login
```

### Deployment

#### Deploy Network

Deploy the network with the following command:

```bash
injective-starnet network up \
  --validators <num_of_validators> \
  --sentries <num_of_sentries> \
  [--artifacts-path <absolute_path_to_chain-stresser-deploy>] \
  [--build-branch <branch_name>]
```

**Parameters:**
* `--validators`: Number of validator nodes to deploy
* `--sentries`: Number of sentry nodes to deploy
* `--artifacts-path`: (Optional) Path to chain-stresser-deploy directory (absolute path to chain-stresser-deploy folder, the output of chain-stresser generate command).
* `--build-branch`: (Optional) Override the injective-core branch to build from

> **Note:** The number of validators and sentries must match the values used in the chain-stresser generate command.

#### Verify Deployment

After deployment, you'll receive the IP addresses of your validators and sentries. Check the network status:

```bash
gex -h <validator_ip>:26657
```

### Network Management

#### Preview Changes (Optional)

To check the current state of your network and see what changes would be applied:

```bash
injective-starnet network preview
```

#### Destroy Network

To tear down the network:

```bash
injective-starnet network destroy
```

> **Note:** If you're unsure about the network state, you can use the [Nuke The Network](https://github.com/InjectiveLabs/injective-starnet/actions/workflows/destroy.yaml) GitHub Action as a fallback option.

### Network Access

All nodes expose the following ports publicly:
* P2P: For node-to-node communication
* RPC: For JSON-RPC API access
* gRPC: For gRPC API access

You can use network up outputed IP's to use API's or ssh to any of nodes for debugging purpose (use key under pkg/pulumi/keys).
  