package wolframalpha

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"nature-id-api/internal"
	"nature-id-api/internal/speciesfinder"
	"net/http"
	"os"
	"strings"
)

const baseURL = "https://api.wolframalpha.com/v2/query?appid=%s&output=json&input=%s"
var (
	apiKey = os.Getenv("WOLFRAM_KEY")
)

type response struct {
	QueryResult struct {
		Pods []struct {
			Title string `json:"title"`
			Scanner string `json:"scanner"`
			Subpods []struct{
				Title string `json:"title"`
				Image struct {
					Source string `json:"src"`
				}
				PlainText string `json:"plaintext"`
			} `json:"subpods"`
		} `json:"pods"`
	} `json:"queryresult"`
}

type client struct {}

func NewClient() speciesfinder.Client {
	return &client{}
}

func (c *client) FetchMetaData(name string) (r internal.SpeciesMetaData, err error) {
	logrus.Info("calling wolframalpha")
	queryName := strings.Replace(name, " ", "+", -1)
	queryUrl := fmt.Sprintf(baseURL, apiKey, queryName)
	res, err := http.Get(queryUrl)
	if err != nil {
		logrus.WithError(err).Error("unable to fetch from wolframalpha client")
		return r, errors.New("call to wolframalpha failed")
	}
	defer res.Body.Close()
	content, err  := ioutil.ReadAll(res.Body)

	if err != nil {
		logrus.WithError(err).Error("unable to read content body")
		return r, errors.New("call to wolframalpha failed")
	}

	var resp response
	if err := json.Unmarshal(content, &resp); err != nil {
		logrus.WithError(err).Error("unable to unmarshal body")
		return r, errors.New("call to wolframalpha failed")
	}
	//
	//if resp.QueryResult.Error == "true"{
	//	return r, errors.New("call to wolframalpha failed")
	//}

	r = internal.SpeciesMetaData{
		Species:   name,
		Source:    "wolframalpha",
		Link:      fmt.Sprintf("https://www5a.wolframalpha.com/input/?i=%s", queryName),
		Name:      extractName(resp),
		ImagePath: "",
		Summary:   extractSummary(resp),
	}
	logrus.Info("wolframalpha called")
	return r, nil
}


func extractName(r response) string {
	for _, p := range r.QueryResult.Pods {
		if p.Scanner == "Identity" {
			if len(p.Subpods) > 0 {
				return p.Subpods[0].PlainText
			}
		}
	}
	logrus.Warn("missing name for species")
	return ""
}

func extractSummary(r response) string {
	content := ""
	for _, p := range r.QueryResult.Pods {
		if p.Title == "Taxonomy" || p.Title == "Biological properties"{
			if len(p.Subpods) > 0 {
				content += p.Subpods[0].PlainText +"\n"
			}
		}
	}
	return content
}