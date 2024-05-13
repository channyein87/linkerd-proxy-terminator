package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// get pod
	namespaceName := getNamespace()
	podName := os.Getenv("HOSTNAME")
	podInfo, err := getPodInfo(clientset, podName, namespaceName)
	if err != nil {
		panic(err.Error())
	}

	// check if there is container linkerd-proxy in the podInfo containers list.
	linkerdProxy := false
	for _, container := range podInfo.Spec.Containers {
		if container.Name == "linkerd-proxy" {
			fmt.Printf("Found linkerd-proxy container\n")
			linkerdProxy = true
			break
		}
	}

	if !linkerdProxy {
		fmt.Printf("No linkerd-proxy container found. Goodbye..\n")
		return
	}

	// create a new variable watchContainers array list and store the list of containers except linkerd-proxy and proxy-kill
	var watchContainers []string
	watchContainers = []string{}
	for _, container := range podInfo.Spec.Containers {
		if container.Name != "linkerd-proxy" && container.Name != "proxy-kill" {
			watchContainers = append(watchContainers, container.Name)
		}
	}
	fmt.Printf("Watching containers: %v\n", watchContainers)

	// for _, container := range podInfo.Status.ContainerStatuses {
	// 	if container.Name == "linkerd-proxy" {
	// 		// loop until container.Ready is true
	// 		for !container.Ready {
	// 			fmt.Printf("Waiting for linkerd-proxy container to be ready\n")
	// 		}
	// 		fmt.Printf("linkerd-proxy container is ready\n")
	// 	}
	// }

	runningWatchContainers := len(watchContainers)
	for runningWatchContainers > 0 {
		fmt.Printf("Running watch containers count: %v\n", runningWatchContainers)
		time.Sleep(5 * time.Second)
		// get pod info
		podInfo, err = getPodInfo(clientset, podName, namespaceName)
		if err != nil {
			panic(err.Error())
		}

		for _, watchContainer := range watchContainers {
			for _, container := range podInfo.Status.ContainerStatuses {
				if container.Name == watchContainer {
					fmt.Printf("Container: %s, Status: %s\n", container.Name, container.State.String())
					if container.State.Terminated != nil {
						fmt.Printf("Container %s has terminated as it is %s\n", watchContainer, container.State.Terminated.Reason)
						runningWatchContainers--
					} else {
						fmt.Printf("Container %s is still running. Terminated: %s\n", watchContainer, container.State.Terminated)
					}
				}
			}
		}
	}

	fmt.Println("All watching containers are terminated. Killing linkerd-proxy...")
	err = killLinkerProxy()
	if err != nil {
		panic(err.Error())
	}
}

// getNamespace function which reads the file /var/run/secrets/kubernetes.io/serviceaccount and return the namespace name as string
func getNamespace() string {
	// create the file
	file, err := os.Open("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		log.Fatal(err)
	}
	// read the file
	data, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	// return the namespace name
	return string(data)
}

// getPodInfo function which returns the pod info from clientset by passing podName and namespaceName
func getPodInfo(clientset *kubernetes.Clientset, podName string, namespaceName string) (*v1.Pod, error) {
	pod, err := clientset.CoreV1().Pods(namespaceName).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Printf("Pod %s not found in %s namespace\n", podName, namespaceName)
		} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
			fmt.Printf("Error getting pod %v\n", statusError.ErrStatus.Message)
		} else {
			return nil, err
		}
	} else {
		fmt.Printf("Found %s pod in %s namespace\n", podName, namespaceName)
	}
	return pod, err
}

// killLinkerProxy function by calling POST request to http://localhost:4191/shutdown
func killLinkerProxy() error {
	url := "http://localhost:4191/shutdown"

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}
	_, err = client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return err
	}

	fmt.Println("Killed linkerd-proxy!")
	return nil
}
