.PHONY: all rhel9 ubuntu docker-build init validate bump-rhel9 bump-ubuntu

DOCKER_IMAGE := ec-manager-packer:latest
AWS_REGION ?= us-east-1
AWS_ACCESS_KEY_ID ?= $(shell aws configure get aws_access_key_id)
AWS_SECRET_ACCESS_KEY ?= $(shell aws configure get aws_secret_access_key)
BUMP_TYPE ?= patch

all: docker-build init rhel9 ubuntu

docker-build:
	docker build -t $(DOCKER_IMAGE) .

init:
	@echo "Initializing Packer plugins..."
	docker run --rm \
		-v $(PWD):/workspace \
		-v $(HOME)/.config/packer:/root/.config/packer \
		$(DOCKER_IMAGE) \
		init rhel9/rhel9.pkr.hcl
	docker run --rm \
		-v $(PWD):/workspace \
		-v $(HOME)/.config/packer:/root/.config/packer \
		$(DOCKER_IMAGE) \
		init ubuntu/ubuntu.pkr.hcl

bump-rhel9:
	@echo "Bumping RHEL 9 version ($(BUMP_TYPE))..."
	./scripts/bump-version.sh $(BUMP_TYPE) rhel9

bump-ubuntu:
	@echo "Bumping Ubuntu version ($(BUMP_TYPE))..."
	./scripts/bump-version.sh $(BUMP_TYPE) ubuntu

rhel9: init
	@echo "Building RHEL 9 AMI..."
	docker run --rm \
		-v $(PWD):/workspace \
		-v $(HOME)/.config/packer:/root/.config/packer \
		-v $(HOME)/.aws:/root/.aws:ro \
		-e AWS_REGION=$(AWS_REGION) \
		-e AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) \
		-e AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY) \
		$(DOCKER_IMAGE) \
		build \
		-var="aws_region=$(AWS_REGION)" \
		rhel9/rhel9.pkr.hcl

ubuntu: init
	@echo "Building Ubuntu AMI..."
	docker run --rm \
		-v $(PWD):/workspace \
		-v $(HOME)/.config/packer:/root/.config/packer \
		-v $(HOME)/.aws:/root/.aws:ro \
		-e AWS_REGION=$(AWS_REGION) \
		-e AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) \
		-e AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY) \
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
