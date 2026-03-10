package aws

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cwltypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

// FetchLogs retrieves log events from the given log groups within the time range.
func (c *Client) FetchLogs(ctx context.Context, logGroups []string, startTime, endTime time.Time) ([]LogEntry, error) {
	var allEntries []LogEntry

	startMs := startTime.UnixMilli()
	endMs := endTime.UnixMilli()

	for _, group := range logGroups {
		var nextToken *string

		for {
			input := &cloudwatchlogs.FilterLogEventsInput{
				LogGroupName: &group,
				StartTime:    &startMs,
				EndTime:      &endMs,
				NextToken:    nextToken,
			}

			out, err := c.CWL.FilterLogEvents(ctx, input)
			if err != nil {
				var rnf *cwltypes.ResourceNotFoundException
				if errors.As(err, &rnf) {
					break // skip this log group, continue with others
				}
				return allEntries, err
			}

			for _, event := range out.Events {
				entry := LogEntry{
					LogGroup: group,
				}
				if event.Timestamp != nil {
					entry.Timestamp = time.UnixMilli(*event.Timestamp)
				}
				if event.Message != nil {
					entry.Message = *event.Message
				}
				if event.LogStreamName != nil {
					entry.LogStream = *event.LogStreamName
				}
				if event.IngestionTime != nil {
					entry.IngestionTime = time.UnixMilli(*event.IngestionTime)
				}
				allEntries = append(allEntries, entry)
			}

			if out.NextToken == nil {
				break
			}
			nextToken = out.NextToken
		}
	}

	sort.Slice(allEntries, func(i, j int) bool {
		return allEntries[i].Timestamp.Before(allEntries[j].Timestamp)
	})

	return allEntries, nil
}
