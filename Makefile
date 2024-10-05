export PROJECT_ROOT=$(shell pwd)
export GO111MODULE=on
export GOPRIVATE="wwwin-github.cisco.com"
export GOPROXY="https://proxy.golang.org, https://${ARTIFACTORY_USER}:${ARTIFACTORY_PASSWORD}@engci-maven-master.cisco.com/artifactory/api/go/nyota-go, direct"
REPO_NAME = wwwin-github.cisco.com/eti/sre-go-helloworld

all: deps target

target:
	echo "Running build"
	go build

deps:
	echo "Get Depedendent Modules"
	go get ./...
	go mod download
	go mod tidy

clean:
	echo "Running build"
	@rm -rf coverage coverage.html
	@rm sre-go-helloworld

test:
	echo "Running tests"
	go test -v --cover ./...

sonar: test
	sonar-scanner -Dsonar.projectVersion="$(version)"

test_debug: debug_kill
	@cd $(PROJECT_ROOT)
	dlv test $(REPO_NAME) --headless --api-version=2 --listen "0.0.0.0:2345" --log=true

debug_kill:
	-kill -9 `ps -ef | grep 'dlv debug' | grep -v grep | awk '{print $$2}'`
	-kill -9 `ps -ef | grep '/debug' | grep -v grep | awk '{print $$2}'`
