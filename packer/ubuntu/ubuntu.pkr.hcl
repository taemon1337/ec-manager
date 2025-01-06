packer {
  required_plugins {
    amazon = {
      version = ">= 1.2.8"
      source  = "github.com/hashicorp/amazon"
    }
  }
}

locals {
  timestamp = formatdate("YYYYMMDDhhmmss", timestamp())
  version   = file("version.txt")
  clean_version = replace(trimspace(local.version), "[^a-zA-Z0-9._-]", "")  # Remove any invalid characters
}

variable "aws_region" {
  type    = string
  default = "us-east-1"
}

variable "source_ami" {
  type    = string
  default = "ami-0c7217cdde317cfec" # Ubuntu 22.04 LTS in us-east-1
}

variable "instance_type" {
  type    = string
  default = "t3.micro"
}

variable "ami_name_prefix" {
  type    = string
  default = "ubuntu-ami-base"
}

variable "ssh_username" {
  type    = string
  default = "ubuntu"
}

variable "ssh_keypair_name" {
  type    = string
  default = "packer-keypair"
}

variable "ssh_private_key_file" {
  type    = string
  default = "~/.ssh/packer-keypair.pem"
}

source "amazon-ebs" "ubuntu" {
  ami_name      = "${var.ami_name_prefix}-v${local.clean_version}-${local.timestamp}"
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
    OS          = "Ubuntu"
    Version     = local.version
    BuildDate   = formatdate("YYYY-MM-DD", timestamp())
    ami-migrate = "enabled"
  }

  launch_block_device_mappings {
    device_name = "/dev/sda1"
    volume_size = 20
    volume_type = "gp3"
    delete_on_termination = true
  }
}

build {
  name = var.ami_name_prefix
  sources = ["source.amazon-ebs.ubuntu"]

  provisioner "shell" {
    scripts = [
      "${path.root}/../scripts/base.sh",
      "${path.root}/../scripts/ubuntu.sh"
    ]
  }

  post-processor "shell-local" {
    inline = [
      "echo 'Updating AMI tags...'",
      "aws ec2 describe-images --filters Name=tag:OS,Values=Ubuntu Name=tag:ami-migrate,Values=latest --query 'Images[?ImageId!=`${build.ID}`].ImageId' --output text | tr '\t' '\n' | xargs -I {} aws ec2 create-tags --resources {} --tags Key=ami-migrate,Value=enabled || true",
      "aws ec2 create-tags --resources ${build.ID} --tags Key=ami-migrate,Value=latest",
      "echo 'AMI ${build.ID} is now tagged as latest'"
    ]
    environment_vars = [
      "AWS_REGION=${var.aws_region}"
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
