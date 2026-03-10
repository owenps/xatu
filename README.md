# xatu ![xatu](xatu.gif)

AWS Logs TUI written in GO

## Getting Started

1. Download the latest binary for your platform from [Releases](../../releases)
1. Ensure your AWS credentials are configured (`aws configure` or SSO)
    - You'll need `logs:DescribeLogGroups`, `logs:FilterLogEvents` IAM permissions.
1. Run `./xatu` — the setup wizard will walk you through selecting a region, log groups, and naming your first context

> [!TIP]
> Move the binary to your PATH so you can run `xatu` from anywhere:
>```bash
>mv xatu /usr/local/bin/
>```

You can re-run the setup wizard anytime with
```bash
./xatu -setup
```

## Usage

> [!IMPORTANT]
> Use <kbd>?</kbd> from xatu to open the shortcuts reminders

### Contexts

xatu contexts are a set of log groups. They may extend in the future to a wider set of configurations.
Contexts encapsulate different scopes so that you can toggle quickly between them based on how you plan to use xatu.

<details>
<summary>Example context setups</summary>

1. `beta`, `prod` - two contexts, where all logs are for each environment
1. `service A`, `service B`, `service C` -  three contexts, divide by logs by service
1. `lambda`, `ecs-prod`, `ecs-test` - three contexts, divided by environment and service
    
</details>

Use <kbd>shift</kbd> + <kbd>tab</kbd> to toggle between contexts.

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

## Data

- Log data is aggregated while xatu is running and cleaned once stopped.
- All settings and preferences can be found at `~/.config/xatu/config.yaml`.

## Development

```bash
git clone <repo-url> && cd xatu
make run        # run from source
make build      # compile to bin/xatu
make test       # run tests
make lint       # run golangci-lint
```

Requires Go 1.25+.

## Thank you!

If you want to support the project, please consider leaving a ⭐︎ on the repository!
