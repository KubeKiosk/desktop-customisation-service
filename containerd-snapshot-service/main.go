package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/gorilla/mux"
)

var version string = "v0.12"
var id int = 1
var containerdNamespace string = "k8s.io"
var containerdSocketPath string = "/run/k3s/containerd/containerd.sock"

//Not relevant for current bug/testing. Ignore this block comment.
/* type containerToSnapshot struct {
	ID                      string `json:"ID"`
	ContainerdUUID          string `json:"ContainerdUUID"`
	ContainerName           string `json:"ContainerName"`
	ContainerSnapshotStatus string `json:"ContainerSnapshotStatus"`
}

type allContainersToSnapshot []containerToSnapshot

var dummydb = allContainersToSnapshot{
	{
		ID:                      "1",
		ContainerdUUID:          "containerd://ac04ea8cc9a611f7a43342a4e7fbdf4f7cd486f55123c9947bcbe2d7da8608c5",
		ContainerName:           "docker.io/metahertz/kiosk-demo-custom-snapshot:1",
		ContainerSnapshotStatus: "false",
	},
} */

func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ContainerD Snapshotter Daemon %s", version)
}

//Not relevant for current bug/testing. Ignore this block comment.
/* func requestSnapshot(w http.ResponseWriter, r *http.Request) {
	var snapshotRequest containerToSnapshot
	err := json.NewDecoder(r.Body).Decode(&snapshotRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id = id + 1
	snapshotRequest.ID = strconv.Itoa(id)
	dummydb = append(dummydb, snapshotRequest)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(snapshotRequest)
	// checkpoint containerd task the task then push it to a registry
	//checkpoint, err := task.Checkpoint(context)
	//err := client.Push(context, "myregistry/checkpoints/redis:master", checkpoint)
}

func getSnapshotStatus(w http.ResponseWriter, r *http.Request) {
	snapshotRequestID := mux.Vars(r)["id"]

	for _, snapshotStatus := range dummydb {
		if snapshotStatus.ID == snapshotRequestID {
			json.NewEncoder(w).Encode(snapshotStatus)
		}
	}
}

func getAllSnapshotRequests(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(dummydb)
} */

func getContainerdContainers(w http.ResponseWriter, r *http.Request) {
	// Default CRI containerd namespace referenced here: https://github.com/containerd/cri/blob/master/docs/crictl.md
	ContainerdClient, err := containerd.New(containerdSocketPath, containerd.WithDefaultNamespace(containerdNamespace))
	if err != nil {
		log.Fatal(err)
	}
	defer ContainerdClient.Close()

	// https://containerd.io/docs/getting-started/ create a new context with an "example" namespace
	ContainerdCtx := namespaces.WithNamespace(context.Background(), containerdNamespace)
	containers, err := ContainerdClient.Containers(ContainerdCtx)
	for _, container := range containers {
		fmt.Fprintf(w, "ContainerID: %s. \n", container.ID())
	}
}

func snapshotContainerdContainer(w http.ResponseWriter, r *http.Request) {
	//snapshotRequestID := mux.Vars(r)["id"]
	snapshotRequestID := r.FormValue("containerid")
	snapshotImageDestination := r.FormValue("imagetag")
	// Default CRI containerd namespace referenced here: https://github.com/containerd/cri/blob/master/docs/crictl.md
	ContainerdClient, err := containerd.New(containerdSocketPath, containerd.WithDefaultNamespace(containerdNamespace))
	if err != nil {
		fmt.Fprintf(w, "An error occured connecting to ContainerD\n")
		log.Fatal(err)
	}
	defer ContainerdClient.Close()

	dockerRegistryCreds := os.Getenv("X_REGISTRY_CREDS")

	log.Println(" Received snapshot request for ContainerID: ", snapshotRequestID, "\n To destination: ", snapshotImageDestination)
	fmt.Fprintf(w, "Received snapshot request for ContainerID: %s \n To destination: %s", snapshotRequestID, snapshotImageDestination)
	//fmt.Fprintf(w, "Push with creds: %s \n", dockerRegistryCreds)
	// https://containerd.io/docs/getting-started/ create a new context with an "example" namespace
	ContainerdCtx := namespaces.WithNamespace(context.Background(), containerdNamespace)
	containers, err := ContainerdClient.Containers(ContainerdCtx)

	if err != nil {
		fmt.Fprintf(w, "An error occured requesting containerID %s.\n", snapshotRequestID)
		log.Fatal(err)
	}

	for _, container := range containers {
		if container.ID() == snapshotRequestID {
			fmt.Fprintf(w, "ContainerID: %s Found! Attempting Checkpoint...\n", container.ID())
			log.Println("ContainerID: %s Found! Attempting Checkpoint...\n", container.ID())

			binary, lookErr := exec.LookPath("/ctr")
			if lookErr != nil {
				panic(lookErr)
			} else {
				log.Println("Found ", binary, "binary on PATH")
			}
			cmdSnapshot, err := exec.Command("ctr", "-a", containerdSocketPath, "-namespace", containerdNamespace, "container", "checkpoint", snapshotRequestID, snapshotImageDestination).CombinedOutput()
			if err != nil {
				log.Printf("Command cmdSnapshot error in exec.Command")
				//log.Fatal(err)
			}

			log.Printf("Waiting for cmdSnapshot command to finish...")
			log.Printf(string(cmdSnapshot))

			cmdPush, err := exec.Command("ctr", "-a", containerdSocketPath, "-namespace", containerdNamespace, "image", "push", "--user", dockerRegistryCreds, snapshotImageDestination).CombinedOutput()
			if err != nil {
				log.Printf("Command cmdPush error in exec.Command")
				//log.Fatal(err)
			}
			log.Printf("Waiting for cmdPush command to finish...")
			log.Printf(string(cmdPush))
			//fmt.Fprintf(w, string(cmdPush))

			// WE ARE USING CTR BINARY VIA EXEC, TODO-Move to proper golang client.
			//TODO Docs say we should need a lot less params to the methods than we are being asked for see: https://github.com/containerd/containerd#checkpoint-and-restore,
			//TODO however this was with a container+task created in the very same session from the same client as seen in the instructions/link above.
			//TODO But instead, we're finding a Container object via the above code (because we'll already know the ContainerD UUID of the container we want to operate on)
			//TODO and then we're trying to call container.Task().Snapshot(), which is asking for a `cio.Attach` object, i'm struggling to create after reading the code here: https://github.com/containerd/containerd/blob/master/cio/io.go
			//TODO GoDocs for the ContainerD task/client API's here : https://godoc.org/github.com/containerd/containerd#Task
			//TODO Also, if cio.Attach issue is workable, i'm also not sure our params to .Checkpoint are correct either.

			//fifoSet, err := cio.NewFIFOSetInDir("/tmp/", "fifo1", false)
			//checkpointImage, err := container.Task(ContainerdCtx, cio.Attach(<BROKEN>)).Checkpoint(ContainerdCtx, "", withCheckpointImagePath(testCheckpointHardPath))
			//err := ContainerdClient.Push(ContainerdCtx, "", checkpointImage)

		} else {
			log.Println("ContainerID:", container.ID(), " Does not match. Ignoring")
		}

	}
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homeLink)
	//router.HandleFunc("/snapshot", requestSnapshot).Methods("POST")
	//router.HandleFunc("/snapshot", getAllSnapshotRequests).Methods("GET")
	//router.HandleFunc("/snapshot/{id}", getSnapshotStatus).Methods("GET")
	router.HandleFunc("/containers", getContainerdContainers).Methods("GET")
	//router.HandleFunc("/containers/{id}", getContainerdContainerObject).Methods("GET")
	router.HandleFunc("/snapshot", snapshotContainerdContainer).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
	containerdNamespace = os.Getenv("X_CONTAINERD_NAMESPACE")
	containerdSocketPath = os.Getenv("X_CONTAINERD_SOCKET_PATH")

}
