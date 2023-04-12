terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.16"
    }
    shieldoo = {
      source  = "shieldoo-io/shieldoo"
      version = ">=1.0.0"
    }
  }
  required_version = ">= 0.14"
}

provider "aws" {
  region = "eu-west-1"
}

// create shieldoo firewall for server
resource "shieldoo_firewall" "fw_example1" {
  name = "fw-example1"
  rules_inbound = [
    {
      port     = "any"
      protocol = "icmp"
    },
    {
      port     = "22"
      protocol = "tcp"
    }
  ]
}

// create shieldoo server configuration
resource "shieldoo_server" "aws_server_1" {
  name        = "aws-server-1"
  firewall_id = shieldoo_firewall.fw_example1.id
  description = "example description"
}

// create cloud-init configuration for server
data "cloudinit_config" "server_config" {
  gzip          = true
  base64_encode = true
  part {
    content_type = "text/cloud-config"
    content = templatefile("${path.module}/cloud-config.yaml", {
      config : shieldoo_server.aws_server_1.configuration
    })
  }
}

// create aws server with cloud-init configuration
resource "aws_instance" "aws_server_1" {
  ami           = "ami-09dd5f12915cfb387" // AWS Linux default image by 2023-04-12
  instance_type = "t2.micro"
  key_name      = "valda"
  user_data     = data.cloudinit_config.server_config.rendered
  tags = {
    Name = "aws-test-1"
  }
}
