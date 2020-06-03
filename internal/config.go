package internal

import (
	"fmt"
	"os"
)

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

type ModelConfig struct {
	Path string
	Name string
	LabelFile string
}
func LoadModelConfig() ModelConfig {
	return ModelConfig{
		Path: GetEnv("MODEL_PATH", "models/faster_rcnn_resnet50_fgvc_2018_07_19/"),
		Name: GetEnv("MODEL_NAME", "model.pb"),
		LabelFile: GetEnv("MODEL_PATH", "labels.json"),
	}
}

func (m ModelConfig) GetModelPath() string {
	return fmt.Sprintf("%s%s", m.Path, m.Name)
}

func (m ModelConfig) GetLabelFilePath() string {
	return fmt.Sprintf("%s%s", m.Path, m.LabelFile)
}