package config

type Aws struct {
	AccessKey       string
	SecretAccessKey string
	S3              struct {
		Buckets struct {
			Media string
		}
	}
}
