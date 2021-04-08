#!/usr/bin/env bash
docker run -v $(pwd)/database/migrations:/migrations --network host migrate/migrate -path=/migrations/ -database "postgres://golangblog:123@localhost:7557/golangblog?sslmode=disable" up

