#!/bin/sh

rm -f api
rm -f job

go build -o api ./cmd/api/main.go
go build -o job ./cmd/job/main.go

chmod +x ./api
chmod +x ./job

docker-compose up -d

pm2 delete all
pm2 start ./api --name api
pm2 start ./job --name job




