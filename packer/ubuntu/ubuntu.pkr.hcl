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
  default = "ami-0c7217cdde317cfec" # Ubuntu 22.04 LTS in us-east-1
}

variable "instance_type" {
  type    = string
  default = "t3.micro"
}

variable "ami_name_prefix" {
  type    = string
  default = "ubuntu-custom"
}

variable "ssh_username" {
  type    = string
  default = "ubuntu"
}

locals {
  timestamp = formatdate("YYYYMMDDhhmmss", timestamp())
}

source "amazon-ebs" "ubuntu" {
  ami_name      = "${var.ami_name_prefix}-${local.timestamp}"
  instance_type = var.instance_type
  region        = var.aws_region
  source_ami    = var.source_ami

  ssh_username = var.ssh_username

  tags = {
    Name        = "${var.ami_name_prefix}-${local.timestamp}"
    OS          = "Ubuntu"
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
  name = "ubuntu-custom"
  sources = ["source.amazon-ebs.ubuntu"]

  provisioner "shell" {
    scripts = [
      "${path.root}/../scripts/base.sh",
      "${path.root}/../scripts/ubuntu.sh"
    ]
  }

  post-processor "manifest" {
    output = "manifest.json"
    strip_path = true
  }
}
