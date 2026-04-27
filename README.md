# vaultswap

> CLI tool for syncing and rotating secrets across multiple secret managers (AWS, Vault, GCP)

---

## Installation

```bash
go install github.com/yourusername/vaultswap@latest
```

Or download a pre-built binary from the [releases page](https://github.com/yourusername/vaultswap/releases).

---

## Usage

```bash
# Sync a secret from AWS Secrets Manager to HashiCorp Vault
vaultswap sync \
  --src aws://us-east-1/my-secret \
  --dst vault://secret/my-secret

# Rotate a secret across all configured providers
vaultswap rotate --secret my-db-password --providers aws,vault,gcp

# Dry-run to preview changes without applying them
vaultswap sync --src aws://us-east-1/my-secret --dst gcp://my-project/my-secret --dry-run
```

### Supported Providers

| Provider | Identifier |
|---|---|
| AWS Secrets Manager | `aws` |
| HashiCorp Vault | `vault` |
| GCP Secret Manager | `gcp` |

### Configuration

vaultswap reads credentials from your environment or a config file at `~/.vaultswap/config.yaml`.

```yaml
providers:
  aws:
    region: us-east-1
  vault:
    address: https://vault.example.com
  gcp:
    project: my-gcp-project
```

---

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

---

## License

[MIT](LICENSE)