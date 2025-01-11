packer {
  required_plugins {
    amazon = {
      version = ">= 1.2.8"
      source  = "github.com/hashicorp/amazon"
    }
  }
}

variable "project" {
  type    = string
  default = "ec-manager"
}

variable "aws_region" {
  type    = string
  default = "us-east-1"
}

variable "source_ami" {
  type    = string
  default = "ami-0fe630eb857a6ec83" # RHEL 9 in us-east-1
}

variable "instance_type" {
  type    = string
  default = "t3.micro"
}

variable "ami_name_prefix" {
  type    = string
  default = "rhel9-custom"
}

variable "ssh_username" {
  type    = string
  default = "ec2-user"
}

variable "ssh_keypair_name" {
  type    = string
  default = "packer-keypair"
}

variable "ssh_private_key_file" {
  type    = string
  default = "~/.ssh/packer-keypair.pem"
}

locals {
  timestamp = formatdate("YYYYMMDDhhmmss", timestamp())
  version   = file("version.txt")
  clean_version = replace(trimspace(local.version), "[^a-zA-Z0-9._-]", "")  # Remove any invalid characters
}

source "amazon-ebs" "rhel9" {
  ami_name      = "${var.ami_name_prefix}-v${local.clean_version}"
  instance_type = var.instance_type
  region        = var.aws_region
  source_ami    = var.source_ami

  # create the instance with the specified key
  ssh_keypair_name = var.ssh_keypair_name

  # use the specified SSH private key (must correspond to the private key in the keypair)
  ssh_private_key_file = var.ssh_private_key_file

  ssh_username = var.ssh_username

  tags = {
    Name        = "${var.ami_name_prefix}-v${local.version}-${local.timestamp}"
    OS          = "RHEL9"
    Version     = local.version
    BuildDate   = formatdate("YYYY-MM-DD", timestamp())
    ami-migrate = "latest"
    Project     = var.project
  }

  launch_block_device_mappings {
    device_name = "/dev/sda1"
    volume_size = 20
    volume_type = "gp3"
    delete_on_termination = true
  }
}

build {
  name = "rhel9-custom"
  sources = ["source.amazon-ebs.rhel9"]

  provisioner "shell" {
    scripts = [
      "${path.root}/../scripts/base.sh",
      "${path.root}/../scripts/rhel9.sh"
    ]
  }

  post-processor "shell-local" {
    inline = [
      "echo 'Updating AMI tags...'",
      "aws ec2 describe-images --filters Name=tag:OS,Values=RHEL9 Name=tag:ami-migrate,Values=latest --query 'Images[?ImageId!=`${build.ID}`].ImageId' --output text | tr '\t' '\n' | xargs -I {} aws ec2 create-tags --resources {} --tags Key=ami-migrate,Value=enabled || true",
      "aws ec2 create-tags --resources ${build.ID} --tags Key=ami-migrate,Value=latest",
      "echo 'AMI ${build.ID} is now tagged as latest'"
    ]
  }

  post-processor "manifest" {
    output = "manifest.json"
    strip_path = true
    custom_data = {
      version = local.version
    }
  }
}
