package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/prometheus/client_golang/prometheus"
)

type collector struct {
	client *sqs.SQS
	mutex  *sync.Mutex

	queues     []*queue
	attributes []*string

	metrics      []*queueMetric
	totalScrapes prometheus.Counter
}

func newCollector(ctx context.Context, updateInterval time.Duration, prefix *string, regex *regexp.Regexp, tagsAsLables bool) *collector {
	sess := session.Must(session.NewSession())

	c := &collector{
		client: sqs.New(sess),
		mutex:  &sync.Mutex{},

		metrics:    metrics,
		attributes: buildAttributes(),

		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, subsystem, "total_scrapes"),
			Help: "Current total AWS SQS scrapes.",
		}),
	}

	go c.queueListUpdater(ctx, updateInterval, queuePrefix, regex, tagsAsLables)

	return c
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- generateDesc(metric.name, metric.help, nil, nil)
	}
	ch <- c.totalScrapes.Desc()
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	c.totalScrapes.Inc()
	defer func() {
		ch <- c.totalScrapes
	}()

	log.Printf("Collecting metrics for %d queues", len(c.queues))

	for _, q := range c.queues {
		lables := []string{"name"}
		values := []string{q.name}

		for key, value := range q.tags {
			lables = append(lables, fmt.Sprintf("tag_%s", key))
			values = append(values, *value)
		}

		if err := q.getQueueAttributes(c.client, c.attributes); err != nil {
			log.Printf("Failed to get queue attributes: %v", err)
			continue
		}

		for _, m := range c.metrics {
			ch <- prometheus.MustNewConstMetric(
				generateDesc(m.name, m.help, lables, nil),
				m.Type,
				q.getAttributeValue(m.attribute),
				values...,
			)
		}
	}
}

func (c *collector) queueListUpdater(ctx context.Context, updateInverval time.Duration, prefix *string, regex *regexp.Regexp, tagsAsLables bool) {
	log.Printf("Start queue list updater")

	queueListInput := &sqs.ListQueuesInput{}

	if len(*prefix) > 0 {
		queueListInput.QueueNamePrefix = prefix
	}

	f := func() error {
		result, err := c.client.ListQueues(queueListInput)
		if err != nil {
			return err
		}

		var tmpQueues []*queue

		for _, q := range result.QueueUrls {
			queueName := getQueueName(*q)
			if !regex.MatchString(queueName) {
				continue
			}

			tmpQ := &queue{name: queueName, url: q}

			if tagsAsLables {
				tagsInput := &sqs.ListQueueTagsInput{
					QueueUrl: q,
				}

				tagsOutput, err := c.client.ListQueueTags(tagsInput)
				if err != nil {
					return err
				}
				tmpQ.tags = tagsOutput.Tags
			}
			tmpQueues = append(tmpQueues, tmpQ)
		}

		log.Printf("Found %d queues", len(tmpQueues))

		c.mutex.Lock()
		c.queues = tmpQueues
		c.mutex.Unlock()

		return nil
	}

	if err := f(); err != nil {
		log.Fatalf("Failed to get queue list: %v", err)
	}

	innerTicker := time.NewTicker(updateInverval)
	defer innerTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-innerTicker.C:
			if err := f(); err != nil {
				log.Fatalf("Failed to get queue list: %v", err)
			}
		}
	}
}
