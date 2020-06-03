package storage

import (
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"gocloud.dev/blob"
	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/gcp"
	"nature-id-api/internal"
	"net/url"
)

func NewGCPBucketStorage(config internal.BucketConfig) (*blob.Bucket, error) {
	ctx := context.Background()

	urlString := config.ConnectionString
	urlParts, _ := url.Parse(urlString)
	// Your GCP credentials.
	// See https://cloud.google.com/docs/authentication/production
	// for more info on alternatives.
	creds, err := gcp.DefaultCredentials(ctx)
	if err != nil {
		logrus.Fatal(err)
	}

	accessID := config.AccessID
	accessKey := config.AccessKey

	if accessID == "" || accessKey == "" {
		logrus.Warn("unable to find access information using default credentials")
		credsMap := make(map[string]string)
		json.Unmarshal(creds.JSON, &credsMap)
		accessID = credsMap["client_id"]
		accessKey = credsMap["private_key"]
	}

	opts := &gcsblob.Options{
		GoogleAccessID: accessID,
		PrivateKey:     []byte(accessKey),
	}
	// Create an HTTP client.
	// This example uses the default HTTP transport and the credentials
	// created above.
	client, err := gcp.NewHTTPClient(
		gcp.DefaultTransport(),
		gcp.CredentialsTokenSource(creds))
	if err != nil {
		logrus.Fatal(err)
	}

	// Create a *blob.Bucket.
	return gcsblob.OpenBucket(ctx, client, urlParts.Host, opts)
}