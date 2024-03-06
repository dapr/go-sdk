# Makefile for building Go service and Docker image

# Go service details
APP_NAME := mock-client2
APP_SRC := app.go
APP_OUTPUT := build/$(APP_NAME)

# Docker details
DOCKERFILE_PATH := deployments
# CHANGE THIS to your Repo Name & in the corresponding kubernetes.yaml
DOCKER_ID ?= seahope
DOCKER_REPO := $(DOCKER_ID)/mock-client2-go-image
DOCKER_TAG := latest

# Kubernetes details
K8S_DEPLOYMENT_FILE := deployments/kubernetes.yaml

# Default target
all: build
# Build the Go service for amd64 architecture
build-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(APP_OUTPUT) $(APP_SRC)

# Build the Go service for arm64 architecture
build-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o $(APP_OUTPUT) $(APP_SRC)

# Build the Docker image
docker-build-arm:
	docker build --platform linux/arm64  -t $(DOCKER_REPO):$(DOCKER_TAG) -f $(DOCKERFILE_PATH)/Dockerfile .

docker-build-amd:
	docker build --platform linux/amd64  -t $(DOCKER_REPO):$(DOCKER_TAG) -f $(DOCKERFILE_PATH)/Dockerfile .

# Push the Docker image to DockerHub
docker-push:
	docker push $(DOCKER_REPO):$(DOCKER_TAG)

# Deploy the application to Kubernetes
deploy:
	-sed s/DOCKER_ID/$(DOCKER_ID)/ $(K8S_DEPLOYMENT_FILE) > deployments/t 
	-mv  deployments/t  $(K8S_DEPLOYMENT_FILE)
	kubectl create -f $(K8S_DEPLOYMENT_FILE)

# Undeploy the application from Kubernetes
undeploy:
	kubectl delete -f $(K8S_DEPLOYMENT_FILE)

# Build for amd64 and push to DockerHub
build-amd64-and-push: build-amd64 docker-build-amd docker-push

# Build for arm64 and push to DockerHub
build-arm64-and-push: build-arm64 docker-build-arm docker-push

# Build, push to DockerHub, and deploy to Kubernetes
deploy-all: build-amd64-and-push deploy

# Clean the built binary
clean:
	rm -f $(APP_OUTPUT)

.PHONY: all build-amd64 build-arm64 docker-build docker-push deploy build-amd64-and-push build-arm64-and-push deploy-all clean

