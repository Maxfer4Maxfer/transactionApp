# Simple docker deploy

## Prerequsites
A machine where **Transaction App** will be dployed should already has [Docker](https://www.docker.com) installed on it.

## Infrastructure services

### Clone the git repository 
```bash
git clone https://github.com/Maxfer4Maxfer/TransactionApp.git
cd ./TransactionApp
```

### NATS
```bash
docker run -d --name nats --network=app_net --network-alias=nats -p 4222:4222 -p 6222:6222 -p 8222:8222 nats
```

### MySQL
```bash
docker run --name mysql -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=repo -p 3306:3306 -d mysql:8 mysqld --sql_mode="" --default-authentication-plugin=mysql_native_password
```

### ELK
```bash
docker run -d --name elasticsearch --net app_net -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:6.6.2
docker run -d --name kibana --net app_net -p 5601:5601 docker.elastic.co/kibana/kibana:6.6.2

docker run -d \
  --name logstash \
  --net app_net \
  --volume "$(pwd)/elk/logstash/pipeline/:/usr/share/logstash/pipeline/" \
  docker.elastic.co/logstash/logstash:6.6.2

docker run -d \
  --name=filebeat \
  --net app_net \
  --user=root \
  --volume="$(pwd)/elk/filebeat/filebeat.yml:/usr/share/filebeat/filebeat.yml:ro" \
  --volume="/var/lib/docker/containers:/var/lib/docker/containers:ro" \
  --volume="/var/run/docker.sock:/var/run/docker.sock:ro" \
  docker.elastic.co/beats/filebeat:6.6.2 filebeat -e -strict.perms=false 
```

### Jaeger
```bash
docker run -d --name jaeger --net app_net -p 5775:5775/udp -p 16686:16686 jaegertracing/all-in-one:latest
```


### Prometheus and Kibana
```bash
docker run -d \
  --name prometheus \
  --net app_net \
  -p 9090:9090 \
  --volume="$(pwd)/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro" \
  --volume="$(pwd)/prometheus/alert.rules:/etc/prometheus/alert.rules:ro" \
  prom/prometheus:latest

docker run -d \
  --name grafana \
  --net app_net \
  -p 3000:3000 \
  --volume="$(pwd)/prometheus/grafana/datasources.yml:/etc/grafana/provisioning/datasources/datasources.yml:ro" \
  --volume="$(pwd)/prometheus/grafana/dashboards.yml:/etc/grafana/provisioning/dashboards/dashboards.yml:ro" \
  --volume="$(pwd)/prometheus/grafana/dashboards:/var/lib/grafana/dashboards:ro" \
  grafana/grafana:latest

docker run -d \
  --name alertmanager \
  --net app_net \
  -p 9093:9093 \
  --volume="$(pwd)/prometheus/alertmanager.yml:/etc/alertmanager/alertmanager.yml:ro" \
  prom/alertmanager:latest
```

## Application 
### apiserver
```bash
cd apiserver
docker build -t apiserver -f Dockerfile .
docker run -d --name apiserver --network=app_net --network-alias=apiserver -p 8081:8081 apiserver
```

### repository
```bash
cd repository
docker build -t repo -f Dockerfile .
docker run -d --name repo --network=app_net --network-alias=repo repo
```

#### worker
```bash
cd worker
docker build -t worker -f Dockerfile .
docker run -d --name worker1 --network=app_net worker
docker run -d --name worker2 --network=app_net worker
docker run -d --name worker3 --network=app_net worker
```

### ui
```bash
cd ui
npm run build
docker build -t ui -f Dockerfile .
docker run -d --name ui --network=app_net --network-alias=ui -p 8080:80 ui
```

## Access application
Go to http://localhost:8080
Directly do REST API requests to apiserver
```bash
curl -d "{}" -X POST http://localhost:8081/findfree
curl -d "{}" -X POST http://localhost:8081/getallnodes
curl -d "{}" -X POST http://localhost:8081/newjob
```


## Clean up installation
```bash
docker stop nats && docker rm nats
docker stop mysql && docker rm mysql
docker stop elasticsearch && docker rm elasticsearch
docker stop kibana && docker rm kibana
docker stop logstash && docker rm logstash
docker stop filebeat && docker rm filebeat
docker stop jaeger && docker rm jaeger
docker stop prometheus && docker rm prometheus
docker stop grafana && docker rm grafana
docker stop alertmanager && docker rm alertmanager
docker stop apiserver && docker rm apiserver
docker stop repo && docker rm repo
docker stop worker1 && docker rm worker1
docker stop worker2 && docker rm worker2
docker stop worker3 && docker rm worker3
```