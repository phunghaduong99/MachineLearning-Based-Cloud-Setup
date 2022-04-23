## If already setup Prometheus, skip 1 and 2 step
## 1. Install Helm chart

```
curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
chmod 700 get_helm.sh
./get_helm.sh
```


## 2. Install Prometheus by Helm chart

Add the latest helm repository in Kubernetes

```
helm repo add stable https://charts.helm.sh/stable
```

Add the Prometheus community helm chart in Kubernetes

```
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
```

Lets pull  prometheus-community/kube-prometheus-stack

```
helm pull prometheus-community/kube-prometheus-stack --untar
```

Lets install prometheus 

```
helm install prometheus kube-prometheus-stack/ --values kube-prometheus-stack/values.yaml
```

Then prometheus and grafana service start, you will see service prometheus-grafana and prometheus-kube-prometheus-prometheus

```
kubectl get svc
```

In order to publish port service , execute command and change type:LoadBalancer

```
kubectl edit svc prometheus-kube-prometheus-prometheus
```

If you want uninstall prometheus 

```
helm uninstall prometheus
```

## 3. Config Prometheus for pushing metrics


Lets config  prometheus values.yaml

```
sudo vi kube-prometheus-stack/values.yaml
```

Change remoteWrite by your http://your-ip:8080/receive
```
    ## The remote_write spec configuration for Prometheus.
    ## ref: https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api.md#remotewritespec
    remoteWrite:
     - url: http://192.168.11.20:8080/receive
```

## 4.Install Kafka
Install kafka server like docs
https://tecadmin.net/install-apache-kafka-ubuntu/


Then go to Kafka directory 


```
cd /usr/local/kafka
sudo vi config/server.properties
```
Change properties like:
```
listeners= PLAINTEXT://192.168.11.20:9092
```

Restart kafka
```
sudo systemctl restart kafka
```

## 5.Install Prometheus kafka adapter 

Add the latest helm repository in Kubernetes

```
git clone https://github.com/Telefonica/prometheus-kafka-adapter
cd prometheus-kafka-adapter/
sudo vi config.go

```

Let change 2 things in config.go. 192.168.11.20 = your-ip for Prometheus push metrics on and Kafka server listening to. RulesText is filter of what metrics you want.

```
kafkaBrokerList        = "192.168.11.20:9092"
kafkaTopic             = "metrics"
```

```
        if value := os.Getenv("MATCH"); value != "" {
                matchList, err := parseMatchList(value)
                if err != nil {
                        logrus.WithError(err).Fatalln("couldn't parse the match rules")
                }
                fmt.Println(matchList)
                match = matchList
        }
        rulesText := `['container_tasks_state{container="middleend"}', 'container_spec_cpu_shares{container="kube-prometheus-stack", metrics_path="/metrics/cadvisor"}']`

```
Then build images telefonica/prometheus-kafka-adapter:latest

```
sudo make
```

Then start container by docker 

```
sudo docker run -p 8080:8080 -it telefonica/prometheus-kafka-adapter:latest
```