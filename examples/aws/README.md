# example EC2 deployment

Example deployment of EC2 AWS instance with shieldoo configuration.

### expected configuration

Expected configuration environment variables

```bash
export AWS_ACCESS_KEY_ID="#####"
export AWS_SECRET_ACCESS_KEY="#####"

export SHIELDOO_API_KEY="#####"
export SHIELDOO_ENDPOINT="https://#####.shieldoo.net"
```

### deploy

```bash
# deployment
terraform apply

# destroy
terraform destroy
```