ifdef RACE
	OPTS = -race -v
	else
	OPTS = -v
endif

all: clean vet test build

cleanall: clean dockerclean

ci: clean vet bench build

cover: clean vet build nogogo coveralls

vet:
	go vet ./redisearch/...

nogogo:
	! grep -rnw 'redisearch' --include=\*.go -e 'github.com/gogo/protobuf/proto'

lint:
	golint ./redisearch/...

generatemock:
	go generate ./...

test:
	go test -cover ./redisearch/...

build:
	go build $(OPTS) ./...

build-linux:
	GOOS=linux GOARCH=amd64 go build $(OPTS) ./...

bench:
	go test -cover --bench ./redisearch/... ./redisearch/...

benchmark: bench

clean:
	go clean ./...

race:
	RACE=true make all

doc:
	godoc -http=:6060

update:
	govendor fetch +vendor

dockerclean:
	echo "remove exited containers"
	docker ps --filter status=dead --filter status=exited -aq | xargs  docker rm -v
	docker images --no-trunc | grep "<none>" | awk '{print $3}' | xargs  docker rmi
	echo "^ above errors are ok"

coveralls:
	go test -v -covermode=count -coverprofile=coverage.out ./redisearch/...
