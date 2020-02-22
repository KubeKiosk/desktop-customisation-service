# desktop-customisation-service
Self service image customisation for creation of pre-populated "desktop" environments, ideal for workshops, technical seminars where a tested, specific set of tools will need to be installed/consumed.

# Initial Roadmap
Two options for people to self-service their own 'desktops' to be consumable in a KubeKiosk environment.

## 1 Allow automated submission of Dockerfile 
Simple submission of Dockerfile, based on our tested images with `FROM`, CI/CD build. Expose via VNC allowing pre-onsite 'staging'.

## 2 Allow for "Edit and Snapshot" live environment via VNC.
Dynamically provide an instance of the base desktop Image, with a VNC/NoVNC sidecar, expose the HTTP UI of the noVNC connection via ingress to the user. 

Allow the user to edit the environment by running commands, installing software etc, then provide a way (API+UI) of signalling they are finished. At this point, create a new image of the containers state, and save as their new "desktop" environment image. Allowing users not familiar with Docker or container tooling to still benefit from environment customization.


# Techical Notes

## Current issues with "2 - Edit and Snapshot".
While `docker commit` is exactly the command we'd want here to snapshot the users' changes, this gets harder if we want this process running on Kubenetes (instead of custom orchestrating a node with docker on it for the self-servie image creation.)

K3OS and Considerable others use ContainerD, via CRI, under the kubelet.
While ContainerD does support a snpashot+image creation, which should suit our needs, these ContainerD API's arent surfaced through to k8s in the CRI spec.

ContainerD Checkpointing: https://github.com/containerd/containerd#checkpoint-and-restore
Kubernetes CRI Spec: https://github.com/kubernetes/cri-api/blob/master/pkg/apis/runtime/v1alpha2/api.proto

Therefore, in order to do this, I see a couple of options, neither of which are very nice.

### A. Agent on each K8S worker via daemonset to connect to ContainerD socket.
Within a POD describe output, we do get the ContainerD UUID of a given container, and the node it's on. 
We should be able to use this to send a request to a simple API running on our agent, on the specific worker (exposed via a nodeport), to call the ContainerD checkpointing API's via the go client.

```
k3os-24912 [~]$ kubectl describe po desktop-daemonset-j5hp8
Name:         desktop-daemonset-j5hp8
Namespace:    default
Priority:     0
Node:         k3os-24912/172.17.0.13
Start Time:   Tue, 11 Feb 2020 23:50:50 +0000
Labels:       controller-revision-hash=68d87d4d9
              name=desktop-kiosk
              pod-template-generation=1
Annotations:  <none>
Status:       Running
IP:           172.17.0.13
IPs:
  IP:           172.17.0.13
Controlled By:  DaemonSet/desktop-daemonset
Containers:
  desktop-kiosk:
    Container ID:   containerd://ac04ea8cc9a611f7a43342a4e7fbdf4f7cd486f55123c9947bcbe2d7da8608c5
    Image:          metahertz/kiosk-demo:v0.1
    Image ID:       docker.io/metahertz/kiosk-demo@sha256:71149efc607a453201ab01ac2a0970155ded2ea3e679d0faa4d713dab5bbecd9
    Port:           <none>
    Host Port:      <none>
    State:          Running
      Started:      Tue, 11 Feb 2020 23:50:50 +0000
```
### B. The alternative would be to try and use a priviledged dind (docker in docker) container, to temporarily host a docker daemon in a pod, then use this DockerD to

* Spin up the UI container+VNC combination through Docker, but scheduled on the k8s worker node through k8s.
* Use the `docker commit` command when the user is ready to finalize their customization.

This second option seems even worse and hackier than the first, and depends on more "stuff". 
Both would lead to the needed customized image being tagged and saved to a registry of our choosing.

I've also asked on the dockercommunity>containerd and kubernetes>general slack for pointers on where to ask, maybe i'm missing some advanced parameter passing in CRI, but i doubt it, based on this closed K8s issue asking for `docker commit` functionality: 

https://github.com/kubernetes/kubernetes/issues/80818
