package errmsg

const (
	InvalidMetricType             = "invalid metric type"
	InvalidMetricValue            = "invalid metric value"
	MetricCounterDeltaCannotBeNil = "metric counter delta cannot be nil"
	MetricGaugeValueCannotBeNil   = "metric gauge value cannot be nil"
	MetricNameRequired            = "metric name is required"
	MetricNotFound                = "metric not found"
	UnableToDecodeJSON            = "cannot decode JSON body"
	UnableToEncodeJSON            = "cannot encode JSON body"
	UnableToParseInt              = "unable to parse int"
	UnableToParseFloat            = "unable to parse float"
	UnableToParseTemplate         = "cannot parse template"
	UnableToWriteTemplate         = "cannot write template"
	UnableToWriteResponse         = "cannot write response body"
)
