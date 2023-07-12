#!/bin/bash

cd ..
echo "Building Terraform plugin for Shieldoo"
env GOOS=darwin GOARCH=arm64 go build -o bin/terraform-provider-shieldoo
if [ $? -ne 0 ]; then
    echo "Failed to build Terraform plugin for Shieldoo"
    exit 1
fi

PLUGIN_ARCH=darwin_arm64
#PLUGIN_ARCH=linux_amd64

# Create the directory holding the newly built Terraform plugins
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/shieldoo-io/shieldoo/1.0.1/${PLUGIN_ARCH}
cp bin/terraform-provider-shieldoo ~/.terraform.d/plugins/registry.terraform.io/shieldoo-io/shieldoo/1.0.1/${PLUGIN_ARCH}/terraform-provider-shieldoo

echo "Testing Terraform plugin for Shieldoo"

export SHIELDOO_API_KEY="AAABBBCCCDDD"

cd ./local-build
rm -f .terraform.lock.hcl
terraform init
terraform plan

if [ "$1" == "-a" ]; then
    terraform apply -auto-approve
fi

if [ "$1" == "-r" ]; then
    rm -f terraform.tfstate
    terraform apply -auto-approve
fi

if [ "$1" == "-d" ]; then
    terraform destroy -auto-approve
fi