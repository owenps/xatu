package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
)

type LogGroup struct {
	Name string
	ARN  string
}

func (c *Client) DiscoverLogGroups(ctx context.Context) ([]LogGroup, error) {
	var groups []LogGroup
	var nextToken *string

	for {
		out, err := c.CWL.DescribeLogGroups(ctx, &cloudwatchlogs.DescribeLogGroupsInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, err
		}

		for _, lg := range out.LogGroups {
			name := ""
			arn := ""
			if lg.LogGroupName != nil {
				name = *lg.LogGroupName
			}
			if lg.Arn != nil {
				arn = *lg.Arn
			}
			groups = append(groups, LogGroup{Name: name, ARN: arn})
		}

		if out.NextToken == nil {
			break
		}
		nextToken = out.NextToken
	}

	return groups, nil
}
