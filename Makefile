main_package_path = ./cmd/main.go
binary_name = alessor
docker_usrname = zdev19
image_name = lessor-service
container_name = lessor-service-container

## development
## build:
.PHONY: build
build:
    # Include additional build steps, like TypeScript, SCSS or Tailwind compilation here...
    go build -o=/tmp/bin/${binary_name} ${main_package_path}

## tidy: tidy modfiles and format .go files
.PHONY: tidy
tidy:
    go mod tidy -v
    go fmt ./...

## run-bin: run the application binary
.PHONY: run-bin
run-bin: build
    /tmp/bin/${binary_name}

## run:
.PHONY: run
run: 
	go run main_package_path

.PHONY: run-live
run-live:
    go run github.com/cosmtrek/air@v1.43.0 \
        --build.cmd "make build" --build.bin "/tmp/bin/${binary_name}" --build.delay "100" \
        --build.exclude_dir "" \
        --build.include_ext "go, tpl, tmpl, html, css, scss, js, ts, sql, jpeg, jpg, gif, png, bmp, svg, webp, ico" \
        --misc.clean_on_exit "true"

## helpers

## help:
.PHONY: help
help:
    @echo 'Usage:'
    @sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## confirm:
.PHONY: confirm
confirm:
    @echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

## no-dirt:
.PHONY: no-dirty
no-dirty:
    @test -z "$(shell git status --porcelain)"

## audit: run quality control checks
.PHONY: audit
audit: test
    go mod tidy -diff
    go mod verify
    test -z "$(shell gofmt -l .)" 
    go vet ./...
    go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
    go run golang.org/x/vuln/cmd/govulncheck@latest ./...

## test: run all tests
.PHONY: test
test:
    go test -v -race -buildvcs ./...

## test-cover: run all tests and display coverage
.PHONY: test-cover
test/cover:
    go test -v -race -buildvcs -coverprofile=/tmp/coverage.out ./...
    go tool cover -html=/tmp/coverage.out

# deployment
# docker-build: builds docker image with dockerfile
.PHONY: docker-build
docker-build:
	docker build --platform=linux/amd64 -t ${docker_usrname}/${image_name}:latest .

# docker-push: pushes image to dockerhub
.PHONY: docker-push
docker-push:
	docker push ${docker_usrname}/${image_name}:latest

# docker-run: runs docker container
.PHONY: docker-run
docker-run: docker-build
	docker run -d -p 8087:8087 ${container_name}

# docker-interact: connect to running container
.PHONY: docker-iteract
docker-interact:
	docker exec -it ${container_name} bash



