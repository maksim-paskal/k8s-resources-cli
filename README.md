# Kubernetes resource advisor

This tool helps you to find the right resources for your Kubernetes pods.

## Install prometheus

To test this tool you need to install prometheus.

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

helm upgrade prometheus prometheus-community/prometheus \
--install \
--namespace=prometheus \
--create-namespace \
--set nodeExporter.enabled=false \
--set alertmanager.enabled=false \
--set pushgateway.enabled=false \
--set kube-state-metrics.resources.requests.cpu=10m \
--set kube-state-metrics.resources.requests.memory=50Mi \
--set configmapReload.prometheus.enabled=false \
--set server.strategy.type=Recreate \
--set server.resources.requests.cpu=100m \
--set server.resources.requests.memory=1Gi \
--set server.global.scrape_interval=15s \
--set server.retention=3d

# prometheus will be available on this ip
kubectl -n prometheus get svc prometheus-server -o go-template='{{ .spec.clusterIP }}'
```

## Install latest k8s-resources-cli

Go to <https://github.com/maksim-paskal/k8s-resources-cli/releases/latest> and install binnary to your operating system.

### MacOS

```bash
brew install maksim-paskal/tap/k8s-resources-cli
```

### Linux

```bash
# for example install v0.0.4 on linux amd64
version=0.0.4; sudo curl -L -o /usr/local/bin/k8s-resources-cli https://github.com/maksim-paskal/k8s-resources-cli/releases/download/v${version}/k8s-resources-cli_${version}_linux_amd64

# make it executable
sudo chmod +x /usr/local/bin/k8s-resources-cli
```

## Calculate pod resources to all pods in cluster

```bash
k8s-resources-cli \
-kubeconfig=$HOME/.kube/config \
-prometheus.retention=3d \
-strategy=aggressive \
-prometheus.url=http://$(kubectl -n prometheus get svc prometheus-server -o go-template='{{ .spec.clusterIP }}')
```

Strategy can be `aggressive` this strategy will try to find container resources limits with 99th percentile of resource usage and `conservative` strategy will try to find container resources limits with maximum resource usage.

Example output:

```text
PodName                                 |ContainerName          |MemoryRequest    |MemoryLimit      |CPURequest |CPULimit
---------------------------------------------------------------------------------------------------------------------------
coredns-5b6f598c6b-nht4q                |coredns                |70Mi / 30.90Mi   |170Mi / 32.21Mi  |100m / 5m  |0 / 8m
fluentd-gcw28                           |fluentd                |200Mi / 183.85Mi |200Mi / 190.61Mi |100m / 8m  |100m / 76m
local-path-provisioner-84bb864455-m7vjn |local-path-provisioner |0 / 14.30Mi      |0 / 14.90Mi      |0 / 1m     |0 / 1m
metrics-server-ff9dbcb6c-hfbrx          |metrics-server         |70Mi / 30.98Mi   |0 / 31.82Mi      |100m / 8m  |0 / 11m
```

columns show current container resources memory and cpu usage and `/` recommended values based on strategy.
