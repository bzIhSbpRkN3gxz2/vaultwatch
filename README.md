# vaultwatch

A CLI tool that monitors HashiCorp Vault secret leases and alerts on expiring or orphaned credentials.

## Installation

```bash
go install github.com/yourusername/vaultwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/vaultwatch.git
cd vaultwatch && go build -o vaultwatch .
```

## Usage

Set your Vault address and token, then run:

```bash
export VAULT_ADDR="https://vault.example.com"
export VAULT_TOKEN="s.your-token-here"

# Monitor all leases and alert if expiring within 24 hours
vaultwatch monitor --threshold 24h

# List orphaned credentials
vaultwatch scan --orphaned

# Watch a specific secret path
vaultwatch watch secret/data/myapp --interval 5m
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--threshold` | Alert window before expiry | `12h` |
| `--interval` | Polling interval | `10m` |
| `--output` | Output format (`text`, `json`) | `text` |

## Configuration

vaultwatch can be configured via a `~/.vaultwatch.yaml` file:

```yaml
vault_addr: https://vault.example.com
threshold: 24h
interval: 5m
output: json
```

## Requirements

- Go 1.21+
- HashiCorp Vault 1.10+

## License

MIT © 2024 Your Name