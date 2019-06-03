package main

import "github.com/prometheus/client_golang/prometheus"

const (
	namespace = "aws"
	subsystem = "sqs"
)

type queueMetric struct {
	name      string
	help      string
	attribute string
	Type      prometheus.ValueType
}

var metrics = []*queueMetric{
	{
		name:      "approximate_number_of_messages",
		help:      "The approximate number of messages available for retrieval from the queue.",
		attribute: "ApproximateNumberOfMessages",
		Type:      prometheus.GaugeValue,
	},
	{
		name:      "approximate_number_of_messages_delayed",
		help:      "The approximate number of messages in the queue that are delayed and not available for reading immediately.",
		attribute: "ApproximateNumberOfMessagesDelayed",
		Type:      prometheus.GaugeValue,
	},
	{
		name:      "approximate_number_of_messages_not_visible",
		help:      "The approximate number of messages that are in flight.",
		attribute: "ApproximateNumberOfMessagesNotVisible",
		Type:      prometheus.GaugeValue,
	},
}

func buildAttributes() []*string {
	var n []*string
	for _, m := range metrics {
		n = append(n, &m.attribute)
	}

	return n
}

func generateDesc(fqName, help string, variableLabels []string, constLabels prometheus.Labels) *prometheus.Desc {
	desc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, fqName), help,
		variableLabels, constLabels,
	)
	return desc
}
