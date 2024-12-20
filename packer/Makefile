.PHONY: all rhel9 ubuntu docker-build init validate

DOCKER_IMAGE := ami-migrate-packer:latest
AWS_REGION ?= us-east-1

all: docker-build rhel9 ubuntu

docker-build:
	docker build -t $(DOCKER_IMAGE) .

rhel9:
	@echo "Building RHEL 9 AMI..."
	docker run --rm -it \
		-v $(PWD):/workspace \
		-v $(HOME)/.aws:/root/.aws:ro \
		-e AWS_REGION=$(AWS_REGION) \
		$(DOCKER_IMAGE) \
		build \
		-var="aws_region=$(AWS_REGION)" \
		rhel9/rhel9.pkr.hcl

ubuntu:
	@echo "Building Ubuntu AMI..."
	docker run --rm -it \
		-v $(PWD):/workspace \
		-v $(HOME)/.aws:/root/.aws:ro \
		-e AWS_REGION=$(AWS_REGION) \
		$(DOCKER_IMAGE) \
		build \
		-var="aws_region=$(AWS_REGION)" \
		ubuntu/ubuntu.pkr.hcl

validate:
	@echo "Validating RHEL 9 template..."
	docker run --rm \
		-v $(PWD):/workspace \
		$(DOCKER_IMAGE) \
		validate rhel9/rhel9.pkr.hcl
	@echo "Validating Ubuntu template..."
	docker run --rm \
		-v $(PWD):/workspace \
		$(DOCKER_IMAGE) \
		validate ubuntu/ubuntu.pkr.hcl

init:
	@echo "Initializing RHEL 9 template..."
	docker run --rm \
		-v $(PWD):/workspace \
		$(DOCKER_IMAGE) \
		init -upgrade rhel9/rhel9.pkr.hcl
	@echo "Initializing Ubuntu template..."
	docker run --rm \
		-v $(PWD):/workspace \
		$(DOCKER_IMAGE) \
		init -upgrade ubuntu/ubuntu.pkr.hcl
