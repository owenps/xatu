package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
)

// QueryStatus represents the state of a running Insights query.
type QueryStatus string

const (
	QueryStatusScheduled QueryStatus = "Scheduled"
	QueryStatusRunning   QueryStatus = "Running"
	QueryStatusComplete  QueryStatus = "Complete"
	QueryStatusFailed    QueryStatus = "Failed"
	QueryStatusCancelled QueryStatus = "Cancelled"
	QueryStatusTimeout   QueryStatus = "Timeout"
)

// QueryResult holds the output from a completed Insights query.
type QueryResult struct {
	Status  QueryStatus
	Results []map[string]string // each row is field→value
	Stats   QueryStats
}

// QueryStats holds query execution statistics.
type QueryStats struct {
	RecordsMatched  float64
	RecordsScanned  float64
	BytesScanned    float64
}

// StartInsightsQuery starts a CloudWatch Logs Insights query and returns the query ID.
func (c *Client) StartInsightsQuery(ctx context.Context, logGroups []string, query string, start, end time.Time) (string, error) {
	input := &cloudwatchlogs.StartQueryInput{
		LogGroupNames: logGroups,
		QueryString:   aws.String(query),
		StartTime:     aws.Int64(start.Unix()),
		EndTime:       aws.Int64(end.Unix()),
	}

	out, err := c.CWL.StartQuery(ctx, input)
	if err != nil {
		return "", fmt.Errorf("start query: %w", err)
	}

	return aws.ToString(out.QueryId), nil
}

// GetQueryResults polls for the results of a running Insights query.
func (c *Client) GetQueryResults(ctx context.Context, queryID string) (*QueryResult, error) {
	out, err := c.CWL.GetQueryResults(ctx, &cloudwatchlogs.GetQueryResultsInput{
		QueryId: aws.String(queryID),
	})
	if err != nil {
		return nil, fmt.Errorf("get query results: %w", err)
	}

	status := QueryStatus(out.Status)

	var rows []map[string]string
	for _, row := range out.Results {
		m := make(map[string]string, len(row))
		for _, field := range row {
			m[aws.ToString(field.Field)] = aws.ToString(field.Value)
		}
		rows = append(rows, m)
	}

	result := &QueryResult{
		Status:  status,
		Results: rows,
	}

	if out.Statistics != nil {
		result.Stats = QueryStats{
			RecordsMatched: out.Statistics.RecordsMatched,
			RecordsScanned: out.Statistics.RecordsScanned,
			BytesScanned:   out.Statistics.BytesScanned,
		}
	}

	return result, nil
}

// StopInsightsQuery cancels a running query.
func (c *Client) StopInsightsQuery(ctx context.Context, queryID string) error {
	_, err := c.CWL.StopQuery(ctx, &cloudwatchlogs.StopQueryInput{
		QueryId: aws.String(queryID),
	})
	if err != nil {
		return fmt.Errorf("stop query: %w", err)
	}
	return nil
}

