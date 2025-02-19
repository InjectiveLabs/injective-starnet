# :star2: Starnet is galactic-scale orchestrator for distributed cosmos network.

## What is Starnet?

Starnet is a powerful and scalable testnet deployment system.

## How does it work?

Starnet uses Pulumi to create and manage network of nodes.
TBA

## Requirements
TBA

## How do I use it?
TBA

Starnet is designed to be used in a cloud environment, but can also be used on a local machine.



# TODO

- [x] Boot disk size should be smaller , 100gb but current image is created from 500gb disk.
- [x] Add ssh keys to the config and attach them to the instances
- [x] Fix IP ordering
- [x] Add RAID script to image
- [x] Create new image with RAID script and smaller boot disk
- [x] This image should have injectived sudo user, lets remove nikola user
- [x] Figureout how to add injectived binary , build from specific branch - We build from codebase
- [ ] Refactor logic to provision sentry nodes
- [ ] Write docs/README
- [x] Add GH Actions to deploy Starnet
  - [ ] Ensure only one running workflow at a time (Pulumi stacks are statefull, we can't spinup infinite number of nodes by mistake, but its good to have this check to prevent builds)
  - [ ] Add option to schedule nuking on creation (run the nuke workflow as cron job, handy if we forget to nuke)
  - [ ] Action input should be passed to pulumi also.
- [x] Add GH Actions to destroy Starnet
- [x] Impl Pulumi rollback logic, so we can destroy Starnet if something with binary goes wrong (the flow is reverse of deploy)
- [ ] Impl ssh wait logic. VM's are not ready to accept connections after creation, we need to retry.
