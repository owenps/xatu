# xatu ![xatu](xatu.gif)

AWS Logs TUI written in GO

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

