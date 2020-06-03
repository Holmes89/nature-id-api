package speciesfinder

import (
	"errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"nature-id-api/internal"
)

type speciesFinderService struct {
	clients []Client
	cache Cache
}

func NewSpeciesFinderService(cache Cache, clients []Client) internal.SpeciesFinder {
	return &speciesFinderService{
		clients: clients,
		cache:   cache,
	}
}

type Client interface {
	FetchMetaData(name string) (internal.SpeciesMetaData, error)
}

type Cache interface {
	Get(name string) []internal.SpeciesMetaData
	Put(name string, speciesData []internal.SpeciesMetaData)
}

func (s *speciesFinderService) FindMetaData(scientificName string) ([]internal.SpeciesMetaData, error) {
	// Check cache
	res := s.cache.Get(scientificName)
	if res != nil {
		return res, nil
	}

	// Cache miss
	res, err := s.callClients(scientificName)
	if err != nil {
		logrus.WithError(err).Error("failed to fetch data from client")
		return nil, errors.New("failed to find species information")
	}

	s.cache.Put(scientificName, res)
	return res, nil
}

func (s *speciesFinderService) callClients(scientificName string)  ([]internal.SpeciesMetaData, error) {

		var res []internal.SpeciesMetaData
		eg := &errgroup.Group{}
		for _, c := range s.clients {
			func(client Client) {
				eg.Go(func() error {
					r, err := client.FetchMetaData(scientificName)
					if err != nil {
						return err
					}
					res = append(res, r)
					return nil
				})
			}(c)
		}
		if err := eg.Wait(); err != nil {
			logrus.WithError(err).Error("client failed")
		}
		if len(res) == 0 {
			// No results came back from our clients so we will call the whole thing an error
			return nil, errors.New("client calls failed")
		}

		return res, nil
}