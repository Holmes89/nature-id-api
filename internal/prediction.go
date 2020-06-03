package internal

import "io"

// Prediction type
type Prediction struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
	Probability float32 `json:"probability"`
}

type Predictions []*Prediction

func (a Predictions) Len() int           { return len(a) }
func (a Predictions) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Predictions) Less(i, j int) bool { return a[i].Probability > a[j].Probability }


type Predictor interface {
	Predict(img io.Reader) (Predictions, error)
}
