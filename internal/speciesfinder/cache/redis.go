package cache

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"nature-id-api/internal"
	"nature-id-api/internal/speciesfinder"
	"strings"
	"time"
)

type redisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) speciesfinder.Cache {
	return &redisCache{client: client}
}

func(c *redisCache) Get(name string) (m []internal.SpeciesMetaData) {

	key := cleanName(name)
	r := c.client.Get(context.Background(), key)
	if err := r.Err(); err != nil {
		if err == redis.Nil {
			logrus.WithField("key", key).Warn("cache miss")
			return nil
		}
		logrus.WithError(err).Error("failed fetching data from cache")
		return nil
	}
	b, err := r.Bytes()
	if err != nil {
		logrus.WithError(err).Error("failed fetching data from cache")
		return nil
	}

	if err := json.Unmarshal(b, &m); err != nil {
		logrus.WithError(err).Error("failed unmarshal cache")
		return nil
	}

	return m
}

func(c *redisCache) Put(name string, speciesData []internal.SpeciesMetaData) {
	key := cleanName(name)
	b, err := json.Marshal(&speciesData)
	if err != nil {
		logrus.WithField("key", key).WithError(err).Error("unable to marshall data")
		return
	}
	s := c.client.Set(context.Background(), key, b, time.Hour*24*14) //evict after two weeks
	if err := s.Err(); err != nil {
		logrus.WithError(err).WithField("key", key).Error("unable to put in cache")
	}
}

func cleanName(name string) string {
	key := strings.ToLower(name)
	key = strings.Replace(key, " ", "", -1)
	key = strings.Replace(key, "+", "", -1)
	key = strings.Replace(key, "-", "", -1)
	key = strings.Replace(key, "_", "", -1)
	return key
}