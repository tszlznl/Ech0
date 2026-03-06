package storagex

// ObjectStorageConfig holds the connection and behavior settings
// for an S3-compatible object storage backend.
// This lives in pkg/storagex so that internal/storage never needs
// to import business model packages.
type ObjectStorageConfig struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	BucketName string
	Region     string
	Provider   string
	UseSSL     bool
	CDNURL     string
	PathPrefix string
}
