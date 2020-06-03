package internal

import "os"

func GetEnv(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}

type BucketConfig struct {
	ConnectionString string
	AccessID         string
	AccessKey        string
}

func LoadBucketConfig() BucketConfig {
	host := GetEnv("BUCKET_HOST", "gs://holmes89-projects")
	accessID := os.Getenv("ACCESS_ID")
	key := os.Getenv("ACCESS_KEY")
	return BucketConfig{
		ConnectionString: host,
		AccessID:         accessID,
		AccessKey:        key,
	}
}