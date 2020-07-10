# statsy

## Backend requirements

* [docker](https://www.docker.com/) - Build, Manage and Secure Your Apps Anywhere. Your Way.
* [docker-compose](https://docs.docker.com/compose/) - Compose is a tool for defining and running multi-container Docker applications. 
* [golang](https://golang.org/) - The Go Programming Language
* [golang mod](https://github.com/golang/go/wiki/Modules) - Go dependency management tool 

## Versions requirements
* golang **>=1.13.10**
* Docker **>=18.09.2**
* Docker-compose **>=1.21.0**

### Setup Linux

```bash
git clone git@github.com:donutloop/statsy.git
cd ./statsy
sudo docker-compose up
go build -o statsy ./cmd/statsy/main.go
SERVICE_ENV_FILE={{replace}}  ./statsy
```

#### example call for bin

```bash
SERVICE_ENV_FILE=/home/donutloop/workspace/statsy/services.env  ./statsy
```

#### example http call

```bash
curl -i --header "Content-Type: application/json" \
  --request POST \
  --data '{"customerID":1,"tagID":2,"userID":"aaaaaaaa-bbbb-cccc-1111-222222222222","remoteIP":"219.070.64.33","timestamp":1500000000}' \
  http://localhost:8080/customer/stats
```

### Docker loging 

```bash
 mysql -h localhost -P 3306 --protocol=tcp -u root -p
```

### Suggestions

* Change hourly_stats.request_count name to hourly_stats.valid_request_count
* Change hourly_stats.invalid_count name to hourly_stats.invalid_request_count
* Change `time` timestamp NOT NULL to `time` INT(4) unsigned NOT NULL