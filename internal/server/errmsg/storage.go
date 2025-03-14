package errmsg

const (
	IntervalMustBePositive      = "interval must be a positive int value"
	StorageIsNil                = "storage is nil"
	UnableToSaveMetricInStorage = "unable to save metrics to storage"
	UnableToPingDB              = "unable to ping database"
	URLCannotBeEmpty            = "url cannot be empty"
	UnableToOpenFile            = "unable to open file"
	UnableToRestoreMetric       = "unable to restore metric"
)
