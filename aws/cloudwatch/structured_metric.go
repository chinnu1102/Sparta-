package cloudwatch

// So a metric is the top level fields that map to the Metric
// info in the serialization layer. So we need a map of names to their
// info. And we can map the rest in the log/publish statement...
import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

var envMap map[string]string

func init() {
	// Get them all and turn it into a map...
	// Ref: https://docs.aws.amazon.com/lambda/latest/dg/lambda-environment-variables.html
	envMap = make(map[string]string)
	envVars := os.Environ()
	for _, eachValue := range envVars {
		parts := strings.Split(eachValue, "=")
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}
}

// MetricDirective represents an element in the array

// MetricUnit Represents a MetricUnit type
type MetricUnit string

const (
	// UnitSeconds Seconds
	UnitSeconds MetricUnit = "Seconds"
	// UnitMicroseconds Microseconds
	UnitMicroseconds MetricUnit = "Microseconds"
	// UnitMilliseconds Milliseconds
	UnitMilliseconds MetricUnit = "Milliseconds"
	// UnitBytes Bytes
	UnitBytes MetricUnit = "Bytes"
	//UnitKilobytes Kilobytes
	UnitKilobytes MetricUnit = "Kilobytes"
	//UnitMegabytes Megabytes
	UnitMegabytes MetricUnit = "Megabytes"
	//UnitGigabytes Gigabytes
	UnitGigabytes MetricUnit = "Gigabytes"
	//UnitTerabytes Terabytes
	UnitTerabytes MetricUnit = "Terabytes"
	//UnitBits Bits
	UnitBits MetricUnit = "Bits"
	//UnitKilobits Kilobits
	UnitKilobits MetricUnit = "Kilobits"
	//UnitMegabits Megabits
	UnitMegabits MetricUnit = "Megabits"
	//UnitGigabits Gigabits
	UnitGigabits MetricUnit = "Gigabits"
	//UnitTerabits Terabits
	UnitTerabits MetricUnit = "Terabits"
	//UnitPercent Percent
	UnitPercent MetricUnit = "Percent"
	//UnitCount Count
	UnitCount MetricUnit = "Count"
	//UnitBytesPerSecond BytesPerSecond
	UnitBytesPerSecond MetricUnit = "Bytes/Second"
	//UnitKilobytesPerSecond KilobytesPerSecond
	UnitKilobytesPerSecond MetricUnit = "Kilobytes/Second"
	//UnitMegabytesPerSecond MegabytesPerSecond
	UnitMegabytesPerSecond MetricUnit = "Megabytes/Second"
	//UnitGigabytesPerSecond GigabytesPerSecond
	UnitGigabytesPerSecond MetricUnit = "Gigabytes/Second"
	//UnitTerabytesPerSecond TerabytesPerSecond
	UnitTerabytesPerSecond MetricUnit = "Terabytes/Second"
	//UnitBitsPerSecond BitsPerSecond
	UnitBitsPerSecond MetricUnit = "Bits/Second"
	//UnitKilobitsPerSecond KilobitsPerSecond
	UnitKilobitsPerSecond MetricUnit = "Kilobits/Second"
	//UnitMegabitsPerSecond MegabitsPerSecond
	UnitMegabitsPerSecond MetricUnit = "Megabits/Second"
	//UnitGigabitsPerSecond GigabitsPerSecond
	UnitGigabitsPerSecond MetricUnit = "Gigabits/Second"
	//UnitTerabitsPerSecond TerabitsPerSecond
	UnitTerabitsPerSecond MetricUnit = "Terabits/Second"
	//UnitCountPerSecond CountPerSecond
	UnitCountPerSecond MetricUnit = "Count/Second"
	// UnitNone No units
	UnitNone MetricUnit = "None"
)

// MetricValue represents a metric value
type MetricValue struct {
	Value interface{}
	Unit  MetricUnit
}

// MetricDirective is the directive that encapsulates a metric
type MetricDirective struct {
	// Dimensions corresponds to the JSON schema field "Dimensions".
	Dimensions map[string]string

	// Metrics corresponds to the JSON schema field "Metrics".
	Metrics map[string]MetricValue

	// namespace corresponds to the JSON schema field "Namespace".
	namespace string
}

// EmbeddedMetric represents an embedded metric that should be published
type EmbeddedMetric struct {
	metrics    []*MetricDirective
	properties map[string]interface{}
}

// WithProperty is a fluent builder to add property to the EmbeddedMetric state.
// Properties should be used for high cardintality values that need to be
// searchable, but not treated as independent metrics
func (em *EmbeddedMetric) WithProperty(key string, value interface{}) *EmbeddedMetric {
	if em.properties == nil {
		em.properties = make(map[string]interface{})
	}
	em.properties[key] = value
	return em
}

// NewMetricDirective returns an initialized MetricDirective
// that's included in the EmbeddedMetric instance
func (em *EmbeddedMetric) NewMetricDirective(namespace string,
	dimensions map[string]string) *MetricDirective {
	md := &MetricDirective{
		namespace:  namespace,
		Dimensions: dimensions,
		Metrics:    make(map[string]MetricValue),
	}
	if md.Dimensions == nil {
		md.Dimensions = make(map[string]string)
	}
	em.metrics = append(em.metrics, md)
	return md
}

// PublishToSink writes the EmbeddedMetric info to the provided writer
func (em *EmbeddedMetric) PublishToSink(additionalProperties map[string]interface{},
	sink io.Writer) {
	// BEGIN - Preconditions
	for _, eachDirective := range em.metrics {
		// Precondition...
		if len(eachDirective.Dimensions) > 9 {
			fmt.Printf("DimensionSet for structured metric must not have more than 9 elements. Count: %d",
				len(eachDirective.Dimensions))
		}
	}
	// END - Preconditions
	for eachKey, eachValue := range additionalProperties {
		em = em.WithProperty(eachKey, eachValue)
	}
	rawJSON, rawJSONErr := json.Marshal(em)
	var writtenErr error
	if rawJSONErr == nil {
		_, writtenErr = io.WriteString(sink, (string)(rawJSON))
	} else {
		_, writtenErr = io.WriteString(sink, fmt.Sprintf("Error publishing metric: %v", rawJSONErr))
	}
	if writtenErr != nil {
		fmt.Printf("ERROR: %#v", writtenErr)
	}
}

// Publish the metric to the logfile
func (em *EmbeddedMetric) Publish(additionalProperties map[string]interface{}) {
	em.PublishToSink(additionalProperties, os.Stdout)
}

// MarshalJSON is a custom marshaller to ensure that the marshalled
// headers are always lowercase
func (em *EmbeddedMetric) MarshalJSON() ([]byte, error) {
	/* From: https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch_Embedded_Metric_Format_Generation_CloudWatch_Agent.html

	The logs must contain a log_group_name key that tells the agent which log group to use.

	Each log event must be on a single line. In other words, a log event cannot contain the newline (\n) character.
	*/
	jsonMap := map[string]interface{}{
		"log_group_name": envMap["AWS_LAMBDA_LOG_GROUP_NAME"],
		"log_steam_name": envMap["AWS_LAMBDA_LOG_STREAM_NAME"],
	}
	for eachKey, eachValue := range em.properties {
		jsonMap[eachKey] = eachValue
	}
	// Walk everything and create the references...
	cwMetrics := &emfAWS{
		Timestamp:         int((time.Now().UnixNano() / int64(time.Millisecond))),
		CloudWatchMetrics: []emfAWSCloudWatchMetricsElem{},
	}
	for _, eachDirective := range em.metrics {
		metricsElem := emfAWSCloudWatchMetricsElem{
			Dimensions: [][]string{},
			Namespace:  eachDirective.namespace,
			Metrics:    []emfAWSCloudWatchMetricsElemMetricsElem{},
		}

		// Create the references and update the metrics...
		for eachKey, eachMetric := range eachDirective.Metrics {
			jsonMap[eachKey] = eachMetric.Value
			metricsElem.Metrics = append(metricsElem.Metrics,
				emfAWSCloudWatchMetricsElemMetricsElem{
					Name: eachKey,
					Unit: string(eachMetric.Unit),
				})
		}
		for eachKey, eachValue := range eachDirective.Dimensions {
			jsonMap[eachKey] = eachValue
			metricsElem.Dimensions = append(metricsElem.Dimensions,
				[]string{eachKey})
		}
		cwMetrics.CloudWatchMetrics = append(cwMetrics.CloudWatchMetrics,
			metricsElem)
	}
	jsonMap["_aws"] = cwMetrics
	return json.Marshal(jsonMap)
}

// JSON encoding the fields gives us the top level keys, which we need
// to map to the Metrics...
// https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch_Embedded_Metric_Format_Specification.html#CloudWatch_Embedded_Metric_Format_Specification_structure_target

// NewEmbeddedMetric returns a new fully initialized embedded metric. Callers
// should populate the Fields
func NewEmbeddedMetric() (*EmbeddedMetric, error) {
	embeddedMetric := &EmbeddedMetric{
		metrics:    []*MetricDirective{},
		properties: make(map[string]interface{}),
	}
	return embeddedMetric, nil
}

// NewEmbeddedMetricWithProperties returns an EmbeddedMetric with the
// user supplied properties
func NewEmbeddedMetricWithProperties(props map[string]interface{}) (*EmbeddedMetric, error) {
	embeddedMetric := &EmbeddedMetric{
		metrics:    []*MetricDirective{},
		properties: props,
	}
	if embeddedMetric.properties == nil {
		embeddedMetric.properties = make(map[string]interface{})
	}
	return embeddedMetric, nil
}
