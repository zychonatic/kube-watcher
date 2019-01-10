package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

var (
	nSpaces    []string
	kubeconfig string
	baseurl    string
	kubeevents = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "kubeevents", Help: "Kubeevents"}, []string{"namespace", "type", "name"})
)

func init() {
	prometheus.MustRegister(kubeevents)
	viper.SetConfigName("config")
	viper.AddConfigPath("./")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}
	baseurl = "http://" + viper.GetString("eshost") + ":9200/kubeevents-"
	kubeconfig = filepath.Join(viper.GetString("kubeconfig"))
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
		nSpaces = append(nSpaces, namespace.Name)
	}

	for _, nSpace := range nSpaces {
		fmt.Printf("getting Namespace %s for checking\n", nSpace)
	}

}
func watcher() {
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
		today := time.Now()
		date := today.Format("2006.01.02")
		url := baseurl + date + "/_doc/"
		for _, nSpace := range nSpaces {
			events, err := clientset.CoreV1().Events(nSpace).List(metav1.ListOptions{})
			if err != nil {
				log.Fatal("failed to get events:", err)
			}

			currentTime := time.Now().Unix()
			for _, event := range events.Items {
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
					kubeevents.WithLabelValues(nSpace, string(event.Type), string(event.Name)).Inc()
					jsonMessage, _ := json.Marshal(message)
					fmt.Println(string(jsonMessage))
					req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonMessage))
					req.Header.Set("Content-Type", "application/json")
					if err != nil {
						panic(err)
					}
					client := &http.Client{}
					resp, err := client.Do(req)
					defer resp.Body.Close()
				}
			}
		}
		time.Sleep(180 * time.Second)
	}
}

func main() {
	go watcher()
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))

}
