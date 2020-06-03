package wiki

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"nature-id-api/internal"
	"nature-id-api/internal/speciesfinder"
	"net/http"
	"strings"
)

const baseURL = "https://en.wikipedia.org/api/rest_v1/page/summary/"

type response struct {
	Title string `json:"title"`
	Thumbnail struct {
		Source string `json:"source"`
	} `json:"thumbnail"`
	OriginalImage struct {
		Source string `json:"source"`
	} `json:"originalimage"`
	ContentUrls struct {
		Desktop struct {
			Page string `json:"page"`
		} `json:"desktop"`
	} `json:"content_urls"`
	Extract string `json:"extract"`
}

type client struct {}

func NewClient() speciesfinder.Client {
	return &client{}
}

func (c *client) FetchMetaData(name string) (r internal.SpeciesMetaData, err error) {
	logrus.Info("calling wikipedia")
	underscoredName := strings.Replace(name, " ", "_", -1)
	queryUrl := fmt.Sprintf("%s%s", baseURL, underscoredName)
	res, err := http.Get(queryUrl)
	if err != nil {
		logrus.WithError(err).Error("unable to fetch from wiki client")
		return r, errors.New("call to wikipedia failed")
	}
	defer res.Body.Close()
	content, err  := ioutil.ReadAll(res.Body)

	if err != nil {
		logrus.WithError(err).Error("unable to read content body")
		return r, errors.New("call to wikipedia failed")
	}
	var resp response
	if err := json.Unmarshal(content, &resp); err != nil {
		logrus.WithError(err).Error("unable to unmarshal body")
		return r, errors.New("call to wikipedia failed")
	}

	r = internal.SpeciesMetaData{
		Species:   name,
		Source:    "wikipedia",
		Link:      resp.ContentUrls.Desktop.Page,
		Name:      resp.Title,
		ImagePath: resp.OriginalImage.Source,
		Summary:   resp.Extract,
	}
	logrus.Info("wikipedia call complete")
	return r, nil
}
