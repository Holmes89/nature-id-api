package cache

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"nature-id-api/internal"
	"nature-id-api/internal/speciesfinder"
)

type memoryCache struct {
	cache map[string][]byte
}

func NewMemoryCache() speciesfinder.Cache {
	m := make(map[string][]byte)
	return &memoryCache{cache: m}
}

func(c *memoryCache) Get(name string) (m []internal.SpeciesMetaData) {
	key := cleanName(name)
	b, ok := c.cache[key]
	if !ok {
		logrus.WithField("key", key).Warn("cache miss")
		return nil
	}
	if err := json.Unmarshal(b, &m); err != nil {
		logrus.WithError(err).Error("failed unmarshal cache")
		return nil
	}
	return m
}

func(c *memoryCache) Put(name string, speciesData []internal.SpeciesMetaData) {
	key := cleanName(name)
	b, err := json.Marshal(&speciesData)
	if err != nil {
		logrus.WithField("key", key).WithError(err).Error("unable to marshall data")
		return
	}
	c.cache[key] = b
}