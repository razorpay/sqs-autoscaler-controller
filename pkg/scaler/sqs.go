package scaler

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

const (
	attrVisible  = "ApproximateNumberOfMessages"
	attrInFlight = "ApproximateNumberOfMessagesNotVisible"
)

// QueueStats holds the SQS queue depth signals used for scaling decisions.
type QueueStats struct {
	// Visible is ApproximateNumberOfMessages: messages waiting to be picked up.
	Visible int64
	// InFlight is ApproximateNumberOfMessagesNotVisible: messages currently
	// being processed by a worker (within their visibility timeout).
	InFlight int64
}

func CurrentQueueStats(sess *session.Session, queueUrl string) (QueueStats, error) {
	svc := sqs.New(sess)
	output, err := svc.GetQueueAttributes(&sqs.GetQueueAttributesInput{
		QueueUrl: &queueUrl,
		AttributeNames: []*string{
			aws.String(attrVisible),
			aws.String(attrInFlight),
		},
	})
	if err != nil {
		return QueueStats{}, err
	}

	visible, err := parseAttr(output.Attributes, attrVisible)
	if err != nil {
		return QueueStats{}, err
	}
	inFlight, err := parseAttr(output.Attributes, attrInFlight)
	if err != nil {
		return QueueStats{}, err
	}
	return QueueStats{Visible: visible, InFlight: inFlight}, nil
}

func parseAttr(attrs map[string]*string, name string) (int64, error) {
	s, ok := attrs[name]
	if !ok {
		return 0, fmt.Errorf("%s did not exist: %+v", name, attrs)
	}
	return strconv.ParseInt(*s, 10, 64)
}
