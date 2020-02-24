# desktop-customisation-service
Self service image customisation for creation of pre-populated "desktop" environments, ideal for workshops, technical seminars where a tested, specific set of tools will need to be installed/consumed.

# Contents
## /containerd-snapshot-service
Contains a PoC agent to be run as a Daemonset on K8s worker nodes. 
The service exposes a HTTP GET endpoint to list the ContainerD containers on the node (by default in the CNI "k8s.io" ContainerD namespace), and then a POST endpoint to trigger a ContainerD checkpoint of a running container (given it's UUID) to a new image path (Given a fully qualified registry/user/imagename:version string) and upload the new image to the registry.

This functionality isn't exposed through CNI or Kubernetes, the idea is this agent can be used to automate the 'snapshotting' of changes to a desktop image our users may make on a VNC-exposed pod, whithout needing to build a Dockerfile for the new image, reducing the complexity to build workshop/kiosk custom images for a number of our speakers.

There is a sample daemonset yaml for kubernetes for deploying the agent, it's currently unauthenticated and listens on the node's IP (host networking) on TCP 8080. 
In the daemonset yaml you'll see an ENV variable for passing your dockerhub/registry credentials through to the service, from a secret/env etc.

### Useage
```
GET http://worker-node:8080

ContainerD Snapshotter Daemon v0.12
```
```
GET http://worker-node:8080/containers

ContainerID: 0968e1808e5bb6360fa6533cd5f343feaa262be244ad432991086f06d4b6e05a. 
ContainerID: 1309cafc19397d7deafe4593ee823d5eda6ca9d9b7981b95f85f9efe089d57f1. 
ContainerID: 370b9fcc085d39cf88a05624b9f1d5f9db45c90c2ba97642e2018365fa0a17bc. 
ContainerID: 38de077d07d0f31da97b619ce66e09b851ddacccf25850ae21127e55a0abb46a. 
ContainerID: 488a6d4b4d25a8661650132e571d9bb62537131b5e43e09a51ac5d60cada743b. 
ContainerID: 4d0494e34ffaebea96762d4fa6ea0e7518fd163605faadb208afc58cbf34d595. 
ContainerID: 61687e7a7139188c7ffe91924c90578ab5d4c699ac92faf089f6712251b7fb00. 
ContainerID: 6757872c32e162013542c48e30b94d81f44fe1a3b87f1fd95d41d2df7b279d21. 
ContainerID: 738cd16172f7e6b6d5524a142d0234a6098142d9add317665943a64c6d332b34. 
ContainerID: 7946abbb63fc810fe466de36ad02a6439b0f4084f7f50f7015b5f1411db6238f. 
ContainerID: b389c4c801cc8462daa13a51d50a165935fec31e3f61670182f42b979b9bddcf. 
ContainerID: f4ef6308e5651d7412104e8c00e3e2c1dc8aa589a4036860964f901a686a47c2. 
```

```
POST http://worker-node:8080/snapshot?containerid=UUID-from-kubectl-describe-po-xyz&imagetag=registry.hub.docker.com/your/destination-container:version

Received snapshot request for ContainerID: 738cd16172f7e6b6d5524a142d0234a6098142d9add317665943a64c6d332b34 
To destination: registry.hub.docker.com/blah/test-snapshot:v1 
ContainerID: 738cd16172f7e6b6d5524a142d0234a6098142d9add317665943a64c6d332b34 Found! Attempting Checkpoint...
```

### To Build
```
cd ./containerd-snapshot-service
docker build .
```

### directory contents
`ca-certificates.crt` - Needed in the container for the ctr binary to connect to public registries over HTTPS

`daemonset-containerd-snapshot.yaml` - Example DaemonSet deployment with ENV

`Dockerfile` - Compile and build the agent container

`main.go` - Very hacky POC snapshot agent. Currently calls out to `ctr` binary for the snapshot and push, as the ContainerD documentation leaves a lot to be desired.

