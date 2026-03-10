# xatu ![xatu](xatu.gif)

AWS Logs TUI written in GO

## Getting Started

1. Download the latest binary for your platform from [Releases](../../releases)
2. Ensure your AWS credentials are configured (`aws configure` or SSO)
3. Run `./xatu` — the setup wizard will walk you through selecting a region, log groups, and naming your first context

Optionally, move the binary to your PATH so you can run `xatu` from anywhere:

```bash
mv xatu /usr/local/bin/
```

Your configuration is saved to `~/.config/xatu/config.yaml`. To re-run the setup wizard at any time:

```bash
./xatu -setup
```

### Required IAM Permissions

```
logs:DescribeLogGroups
logs:FilterLogEvents
```

## Costs

xatu makes outbound API calls to AWS to fetch log data. Costs scale with the volume of data processed.

> [!TIP]
> xatu can help you reduce AWS costs compared to using the console directly. To keep costs low:
> - Minimize the number of log groups in your xatu contexts
> - Keep insight queries targeted at a short time window
> - Switch to manual polling instead of automatic intervals.

The following estimates are approximations for reference only.

| Scenario | Assumptions | Cost/hour | Cost/day (8hr) |
| :--- | :--- | :--- | :--- |
| Light polling | 5s interval, ~1 MB logs/min | ~$0.0015/hr | ~$0.01 |
| Medium polling | 5s interval, ~10 MB logs/min | ~$0.015/hr | ~$0.12 |
| Heavy polling | 5s interval, ~100 MB logs/min | ~$0.15/hr | ~$1.20 |
| Insights query | 1 GB scanned per query | $0.0057/query | varies |
| LiveTail (not supported in xatu) | Per session | $0.60/hr | $4.80 |

## Development

```bash
git clone <repo-url> && cd xatu
make run        # run from source
make build      # compile to bin/xatu
make test       # run tests
make lint       # run golangci-lint
```

Requires Go 1.25+.
