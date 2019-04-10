# Simple run

### Prerequsite
A machine where **Transaction App** will be deployed should already has infrastructure servises runned.
For simplisity you can run all of these services in canteiner using [Docker deploy](https://github.com/Maxfer4Maxfer/transactionApp/blob/master/docs/docker-deploy.md) instruction.

### Clone the git repository 
```bash
git clone https://github.com/Maxfer4Maxfer/TransactionApp.git
cd ./TransactionApp
```

### apiserver
```bash
cd ./apiserver
go run main.go --debug-addr=:8080 --http-addr=:8081 --jaeger-addr=localhost:5775 --repoIP=127.0.0.1 --repoPort=:8182
```

### repository
```bash
cd ./repository
go run main.go --debug-addr=:8180 --grpc-addr=:8182 --jaeger-addr=localhost:5775 --dsn="root:root@tcp(localhost:3306)/repo?charset=utf8&parseTime=True&loc=Local"
```

### worker
```bash
cd ./worker
go run main.go --debug-addr=:8280 --extIP=127.0.0.1 --extPort=:8282 --grpc-addr=:8282 --jaeger-addr=localhost:5775
go run main.go --debug-addr=:8380 --extIP=127.0.0.1 --extPort=:8382 --grpc-addr=:8382 --jaeger-addr=localhost:5775
go run main.go --debug-addr=:8480 --extIP=127.0.0.1 --extPort=:8482 --grpc-addr=:8482 --jaeger-addr=localhost:5775
```

### ui
```bash
cd ./ui
npm start
```

