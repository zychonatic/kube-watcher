this is a simple tool that collects every 3 mins the latest kubernetes events from all namespaces in the cluster and sends them into an elasticsearch index

-) go get github.com/zychonatic/kube-watcher

-) go build kube-watcher

-) adapt config.yaml
this config must be in the WORKDIR, because kube-watcher is getting it from "./config.yaml"

-) ./kube-watcher &

