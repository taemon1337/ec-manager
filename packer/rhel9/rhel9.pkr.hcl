packer {
  required_plugins {
    amazon = {
      version = ">= 1.2.8"
      source  = "github.com/hashicorp/amazon"
    }
  }
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

locals {
  timestamp = formatdate("YYYYMMDDhhmmss", timestamp())
}

source "amazon-ebs" "rhel9" {
  ami_name      = "${var.ami_name_prefix}-${local.timestamp}"
  instance_type = var.instance_type
  region        = var.aws_region
  source_ami    = var.source_ami

  ssh_username = var.ssh_username

  tags = {
    Name        = "${var.ami_name_prefix}-${local.timestamp}"
    OS          = "RHEL9"
    Status      = "latest"
    BuildDate   = formatdate("YYYY-MM-DD", timestamp())
    BuildTime   = formatdate("hh:mm:ss", timestamp())
    "ami-migrate" = "latest"
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

  post-processor "manifest" {
    output = "manifest.json"
    strip_path = true
  }
}
