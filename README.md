# Terraform Provider Shieldoo (EXPERIMENTAL)

This repository is for a [Shieldoo](https://www.shieldoo.io) provider. It is containing:

- A resource and a data source (`internal/provider/`),
- Examples (`examples/`) and generated documentation (`docs/`),
- Miscellaneous meta files.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.19

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

### Configure provider

Configure in terraform:

```terraform
provider "shieldoo" {
    endpoint = "http://localhost:9000"
    apikey = "AAABBBCCCDDD"
}
```

Configure in OS environment variables:

```bash
export SHIELDOO_ENDPOINT="http://localhost:9000"
export SHIELDOO_API_KEY="AAABBBCCCDDD"
```

### Sample deployment AWS EC2 instance with shieldoo

[AWS EC2 terraform example](examples/aws)

### Sample complex deployment

Complex example:

```terraform
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
```

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```

### Test build/execute

```bash
env GOOS=darwin GOARCH=arm64 go build -o bin/terraform-provider-shieldoo

PLUGIN_ARCH=darwin_arm64
#PLUGIN_ARCH=linux_amd64

# Create the directory holding the newly built Terraform plugins
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/shieldoo-io/shieldoo/1.0.0/${PLUGIN_ARCH}
cp bin/terraform-provider-shieldoo ~/.terraform.d/plugins/registry.terraform.io/shieldoo-io/shieldoo/1.0.0/${PLUGIN_ARCH}/terraform-provider-shieldoo

```
