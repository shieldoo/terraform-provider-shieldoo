terraform {
  required_providers {
    shieldoo = {
      source  = "shieldoo-io/shieldoo"
      version = ">=1.0.0"
    }
  }
  required_version = ">= 0.14"
}

provider "shieldoo" {
  endpoint = "http://localhost:9000"
}

data "shieldoo_firewall" "default" {
  name = "default"
}

output "firewall-id" {
  value = data.shieldoo_firewall.default.id
}

data "shieldoo_server" "server1" {
  name = "server1"
}

output "server-configuration" {
  value = nonsensitive(data.shieldoo_server.server1.configuration)
}
output "server-ip" {
  value = data.shieldoo_server.server1.ip_address
}
output "server-id" {
  value = data.shieldoo_server.server1.id
}

resource "shieldoo_firewall" "example1" {
  name = "furewall1"
  rules_inbound = [
    {
      port     = "any"
      protocol = "any"
    }
  ]
  rules_outbound = [
    {
      port     = "any"
      protocol = "icmp"
    },
    {
      port        = "22"
      protocol    = "tcp"
      group_names = ["Shieldoo_Admin-69"]
    },
    {
      port             = "80"
      protocol         = "tcp"
      group_object_ids = ["8ebe0ff5-d358-4787-bf6d-0f44d9b1129f"]
    }
  ]
}

resource "shieldoo_firewall" "example2" {
  name = "furewall2"
  rules_inbound = [
    {
      port     = "any"
      protocol = "icmp"
    },
    {
      port     = "11"
      protocol = "tcp"
    },
    {
      port      = "22"
      protocol  = "tcp"
      group_ids = ["localhost:groups:69"]
    },
    {
      port             = "80"
      protocol         = "tcp"
      group_object_ids = ["8ebe0ff5-d358-4787-bf6d-0f44d9b1129f"]
    }
  ]
}

output "create-firewall-example1" {
  value = shieldoo_firewall.example1.id
}

output "create-firewall-example2" {
  value = shieldoo_firewall.example2.id
}

resource "shieldoo_server" "example1" {
  name        = "example1"
  firewall_id = shieldoo_firewall.example1.id
  description = "example1 description"
}

output "name-server-example1-id" {
  value = shieldoo_server.example1.id
}
output "name-server-example1-configuration" {
  value = nonsensitive(shieldoo_server.example1.configuration)
}

resource "shieldoo_server" "example2" {
  name        = "example2"
  firewall_id = shieldoo_firewall.example1.id
  //description = shieldoo_server.example1.configuration
  group_ids = ["localhost:groups:69"]
  listeners = [
    {
      listen_port  = 80
      protocol     = "tcp"
      forward_port = 8080
      forward_host = "localhost"
      description  = "example2 description"
    },
    {
      listen_port  = 81
      protocol     = "tcp"
      forward_port = 8081
      forward_host = "localhost"
    }
  ]
}
output "name-server-example2-id" {
  value = shieldoo_server.example2.id
}

resource "shieldoo_server" "example3" {
  name        = "example3"
  firewall_id = shieldoo_firewall.example1.id
  description = "example3 description"
}

output "name-server-example3-id" {
  value = shieldoo_server.example3.id
}
