# :star2: Starnet is galactic-scale orchestrator for distributed cosmos network.

  

## What is Starnet?

  

Starnet is a powerful and scalable testnet deployment system.

  

## How does it work?

  

Starnet uses Pulumi to create and manage network of nodes.

TBA

  

## Requirements

TBA

  

## How do I use it?

  

### Use Github Actions to deploy Starnet

Currently only [Nuke The Network](https://github.com/InjectiveLabs/injective-starnet/actions/workflows/destroy.yaml) job works, you can use it nuke the running network.

  

### Use Local Machine to deploy Starnet

  

Currently we have set up a local machine to deploy Starnet on mainne-devlab node (playground machine).

  

ssh config

  

    Host mainne-devlab
    
      
    
     HostName 37.59.23.104
    
      
    
     User root
    
      
    
     Port 22
    
      
    
    I dentityFile ~/.ssh/id_rsa

  
  

When you ssh, run the following commands:

  

    cd injective/injective-starnet
    git pull  
    gcloud auth login --cred-file=gcloud.json
    export GOOGLE_APPLICATION_CREDENTIALS=$(pwd)/gcloud.json

  
  

#### Tweak local script that wraps chain-stresser, adjust the number of validators
Note: For testing purpose use smaller network 2-10 validators max.

    vi scripts/local.sh
    # generate artifacts
    ./scripts/local.sh # This should output you chain-stresser-deploy folder in same path

#### Align validators with the number of VM's we need to create

    vi Pulumi.starnet.yaml # change the nodePoolSize to fit the value you set in scripts/local.sh
    
#### Spin up the network

    pulumi up -y

#### When pulumi finishes, you will get validators IP's in output

Copy any validator IP and  ssh (from that devlab machine, but you should be able to ssh from your machine also)

    ssh injectived@<validator_ip_from_output>
    
Check the network status with `gex` 


#### To nuke the network

    pulumi destroy -y

**NOTE:** if you are not sure if you nuked network properly , go to injective-starnet repo and manually run [Nuke The Network](https://github.com/InjectiveLabs/injective-starnet/actions/workflows/destroy.yaml) job (top right corner have "run the workflow" buttin").
All ports are opened to public, p2p, RPC , gRPC 
  
  
  
