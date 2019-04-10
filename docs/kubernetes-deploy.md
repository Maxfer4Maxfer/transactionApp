# Deployment on a kubernetes cluster

In this deployment [Docker Machine](https://docs.docker.com/machine) and [Google Cloud Platform](https://cloud.google.com) will be used as a platform for Transaction App deployment.

## Prerequsites
You need to have a kubernetes cluster. The easiest way is to create a cluster using [Google Cloud Kubernetes Engine](https://cloud.google.com/kubernetes-engine/)


### Clone the git repository 
```bash
git clone https://github.com/Maxfer4Maxfer/TransactionApp.git
cd ./TransactionApp
```

### Docker images
Build docker images and upload them two the [Docker Hub](https://hub.docker.com):
```bash
DOCKER_HUB_USER=<YOUR_DOCKER_HUB_ACCOUNT>
APP_VERSION=<APP_VERSION>
### Build user interface
cd ui
npm run build
cd ..

docker build -t $DOCKER_HUB_USER/ui:$APP_VERSION ./ui
docker build -t $DOCKER_HUB_USER/apiserver:$APP_VERSION ./apiserver
docker build -t $DOCKER_HUB_USER/repository:$APP_VERSION ./repository
docker build -t $DOCKER_HUB_USER/worker:$APP_VERSION ./worker

docker push $DOCKER_HUB_USER/apiserver:$APP_VERSION
docker push $DOCKER_HUB_USER/ui:$APP_VERSION
docker push $DOCKER_HUB_USER/repository:$APP_VERSION
docker push $DOCKER_HUB_USER/worker:$APP_VERSION
```

### Docker images
Make changes in kubernetes yaml files:
```bash
DOCKER_HUB_USER=<YOUR_DOCKER_HUB_ACCOUNT>
APP_VERSION=<APP_VERSION>
cd ./kubernetes
sed -i.bu 's/docker_hub_user/'$DOCKER_HUB_USER'/g' ./apiserver-deployment.yml
sed -i.bu 's/app_version/'$APP_VERSION'/g' ./apiserver-deployment.yml
sed -i.bu 's/docker_hub_user/'$DOCKER_HUB_USER'/g' ./repo-deployment.yml
sed -i.bu 's/app_version/'$APP_VERSION'/g' ./repo-deployment.yml
sed -i.bu 's/docker_hub_user/'$DOCKER_HUB_USER'/g' ./worker-deployment.yml
sed -i.bu 's/app_version/'$APP_VERSION'/g' ./worker-deployment.yml
```

### Create configmaps
```bash
cd ./kubernetes
kubectl create configmap logstash-cfgmap --from-file ../elk/logstash/pipeline
kubectl create configmap filebeat-cfgmap --from-file ../elk/filebeat/filebeat.yml
kubectl create configmap prometheus-cfgmap --from-file ../prometheus
kubectl create configmap alertmanager-cfgmap --from-file ../prometheus/alertmanager.yml
kubectl create configmap grafana-cfgmap --from-file ../prometheus/grafana
kubectl create configmap grafana-dashboards-cfgmap --from-file ../prometheus/grafana/dashboards
```

### Deploy application
```bash
cd ./kubernetes
kubectl create -f .
```

## Access application
You need to know external IPs of ui and apiserver components.
```bash
kubectl get service ui-lb apiserver-lb
```

Go to http://<<UI_EXTERNAL_IP>>:80

Go to the **Setting Tab** and put there <<EXTERNAL_IP_OF_APISERVER>>
