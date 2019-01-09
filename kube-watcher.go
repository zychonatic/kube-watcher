package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"bytes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"path/filepath"
	"time"
)

var (
	nSpaces    []string
	kubeconfig string
	url        string
)

func init() {
	url = "http://localhost:9200/kubeevents/_doc/"
	kubeconfig = filepath.Join("../../../../../MobaXterm/home/config.development")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		log.Fatal("failed to get namespaces:", err)
	}

	for _, namespace := range namespaces.Items {
		//if (event.Type != "Normal") {
		nSpaces = append(nSpaces, namespace.Name)
		//diff := time.Now().Sub(event.LastTimestamp)
		//fmt.Printf(diff)
		//fmt.Printf("[%d] %s\n", i, event.LastTimestamp)
		//}
	}

	for _, nSpace := range nSpaces {
		//if (event.Type != "Normal") {
		fmt.Printf("getting Namespace %s for checking\n", nSpace)
		//diff := time.Now().Sub(event.LastTimestamp)
		//fmt.Printf(diff)
		//fmt.Printf("[%d] %s\n", i, event.LastTimestamp)
		//}
	}

}
func main() {
	//kubeconfig := filepath.Join("../../../../../MobaXterm/home/config.development")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	for {
		fmt.Println("starting checking at ", time.Now())
		for _, nSpace := range nSpaces {
			events, err := clientset.CoreV1().Events(nSpace).List(metav1.ListOptions{})
			if err != nil {
				log.Fatal("failed to get events:", err)
			}

			// print pods
			currentTime := time.Now().Unix()
			for _, event := range events.Items {
				//if (event.Type != "Normal") {
				if ((currentTime - event.LastTimestamp.Unix()) < 180) && ((currentTime - event.LastTimestamp.Unix()) > 0) {
					message := map[string]interface{}{
						"timestamp": event.LastTimestamp,
						"namespace": nSpace,
						"reason":    event.Reason,
						"message":   event.Message,
						"firstSeen": event.FirstTimestamp,
						"name":      event.Name,
						"type":      event.Type,
						"source":    event.Source,
						"count":     event.Count,
					}
					jsonMessage, _ := json.Marshal(message)
					fmt.Println(string(jsonMessage))
					req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonMessage))
					req.Header.Set("Content-Type", "application/json")
					if err != nil {
						// Handle error
						panic(err)
					}
					client := &http.Client{}
					resp, err := client.Do(req)
					defer resp.Body.Close()
					//fmt.Println(nSpace, event.FirstTimestamp, event.LastTimestamp, event.Name, event.Reason, event.Message)
				}
				//diff := time.Now().Sub(event.LastTimestamp)
				//fmt.Printf(diff)
				//fmt.Printf("[%d] %s\n", i, event.LastTimestamp)
				//}
			}
		}
		time.Sleep(180 * time.Second)
	}

}
