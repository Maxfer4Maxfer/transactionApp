# Docker compose and Docker machine

In this deployment [Docker Machine](https://docs.docker.com/machine) and [Google Cloud Platform](https://cloud.google.com) will be used as a platform for Transaction App deployment.

## Prerequsites

Set up environment variables:
```bash
GOOGLE_PROJECT=docker-28092018    
GOOGLE_ZONE=europe-west1-b
```
Create a virtual machine in GCP
```bash
docker-machine create --driver google \
--google-project $GOOGLE_PROJECT \
--google-machine-image https://www.googleapis.com/compute/v1/projects/ubuntu-os-cloud/global/images/family/ubuntu-1604-lts \
--google-machine-type n1-standard-4 \
--google-zone $GOOGLE_ZONE \
--google-open-port 80/tcp \
--google-open-port 3000/tcp \
--google-open-port 3306/tcp \
--google-open-port 8081/tcp \
--google-open-port 5601/tcp \
--google-open-port 9090/tcp \
--google-open-port 16686/tcp \
vm1

eval $(docker-machine env vm1)
```

|Port|Description|
|---|---|
|80|transaction-app| 
|3000|gragana|
|3306|mysql|
|8081|apiserver|
|5601|kibana|
|9090|prometheus|
|16686|jaeger|


## Deploy application 

### Clone the git repository 
```bash
git clone https://github.com/Maxfer4Maxfer/TransactionApp.git
cd ./TransactionApp
```

### Config application
```bash
docker-machine ls  # Shows you a IP of created virtual machine (vm1)
TRANSACTION_APP_IP=123.123.123.123 #Change to the vm1's IP address
sed -i.bu 's/localhost/'$TRANSACTION_APP_IP'/g' ./ui/src/App.js
```

### Build go vendor folders and ui output scripts
```bash
cd apiserver; go mod vendor; cd ..;
cd repository; go mod vendor; cd ..;
cd worker; go mod vendor; cd ..;
cd ui; npm install && npm run build; cd ..;
```

### Copy configs and source files to the docker machine
```bash
docker-machine ssh vm1 "sudo rm -fR ~/*"
docker-machine scp -r $(pwd)/prometheus vm1:~
docker-machine scp -r $(pwd)/elk vm1:~
rm -fR ui/node_modules
docker-machine scp -r $(pwd)/ui vm1:~
docker-machine scp -r $(pwd)/apiserver vm1:~
docker-machine scp -r $(pwd)/repository vm1:~
docker-machine scp -r $(pwd)/worker vm1:~
```

### Run docker-compose
```bash
docker-compose up -d
```

## Access application
Go to http://<<TRANSACTION_APP_IP>>:80

Directly do REST API requests to apiserver
```bash
curl -d "{}" -X POST http://<TRANSACTION_APP_IP>:8081/findfree
curl -d "{}" -X POST http://<TRANSACTION_APP_IP>:8081/getallnodes
curl -d "{}" -X POST http://<TRANSACTION_APP_IP>:8081/newjob
```

## Clean up installation
Stop application and delete a docker virtual machine
```bash
docker-compose down
docker-machine rm vm1
```


