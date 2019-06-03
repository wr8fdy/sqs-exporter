package main

import (
	"log"
	"net/url"
	"path"
	"strconv"

	"github.com/aws/aws-sdk-go/service/sqs"
)

type queue struct {
	name       string
	url        *string
	tags       map[string]*string
	attributes map[string]*string
}

func (q *queue) getQueueAttributes(client *sqs.SQS, attributes []*string) error {
	input := &sqs.GetQueueAttributesInput{
		QueueUrl:       q.url,
		AttributeNames: attributes,
	}
	output, err := client.GetQueueAttributes(input)
	if err != nil {
		return err
	}

	q.attributes = output.Attributes

	return nil
}

func (q *queue) getAttributeValue(attribute string) float64 {
	s, ok := q.attributes[attribute]
	if !ok {
		return 0
	}

	v, err := strconv.ParseFloat(*s, 64)
	if err != nil {
		log.Printf("Failed to parse value to float64: %v", err)
	}
	return v
}

func getQueueName(queueUrl string) string {
	u, err := url.Parse(queueUrl)
	if err != nil {
		log.Fatalf("Failed to parse queue URL: %v", err)
	}
	return path.Base(u.Path)
}
