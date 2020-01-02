package predictor

import (
	"bytes"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/tensorflow/tensorflow/tensorflow/go"
	"github.com/tensorflow/tensorflow/tensorflow/go/op"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"
)

const (
	graphFile = "./models/nature.pb"
	tagsFile  = "./models/tags.json"
)

type Service interface {
	GetLabels(img io.ReadCloser) (Labels, error)
}

type tfService struct {
	labelMap map[int]*Label
	graph    *tensorflow.Graph
	session *tensorflow.Session
}

func NewTensorflowService() (Service, error) {
	s := &tfService{}
	if err := s.loadLabelMap(); err != nil {
		logrus.WithField("err", err).Error("unable to load map")
		return nil, err
	}

	model, err := tensorflow.LoadSavedModel("./nature-model/saved_model", []string{"serve"}, nil)
	if err != nil {
		logrus.Error("unable to load model")
		return nil, err
	}
	s.session = model.Session
	s.graph = model.Graph

	logrus.Info("service created")
	return s, nil
}

func (s *tfService) GetLabels(img io.ReadCloser) (Labels, error) {

	// Get normalized tensor
	tensor, err := s.normalizeImage(img)
	if err != nil {
		log.Fatalf("unable to make a tensor from image: %v", err)
	}

	now := time.Now()
	logrus.Info("predicting")
	output, err := s.session.Run(
		map[tensorflow.Output]*tensorflow.Tensor{
			s.graph.Operation("image_tensor").Output(0): tensor,
		},
		[]tensorflow.Output{
			s.graph.Operation("detection_scores").Output(0),
			s.graph.Operation("detection_classes").Output(0),
			s.graph.Operation("num_detections").Output(0),
			s.graph.Operation("detection_boxes").Output(0),
		},
		nil)
	if err != nil {
		return nil, err
	}
	processedTime := time.Now().Sub(now)
	logrus.WithField("time", processedTime.String()).Info("predicting complete")
	scores := output[0].Value().([][]float32)[0] //Maps to above tensorflow output detection_scores
	ids := output[1].Value().([][]float32)[0]    //Maps to above tensorflow output detection_classes

	var labels Labels

	for i, sc := range scores {
		//todo len check
		id := ids[i]

		label, ok := s.labelMap[int(id)]
		if !ok {
			logrus.WithField("id", id).Warn("id does not exist")
			continue
		}
		label.Probability = sc * 100
		labels = append(labels, label)
	}

	return labels, nil
}

func (s *tfService) loadLabelMap() error {
	tagsFile, err := os.Open(tagsFile)
	if err != nil {
		return err
	}
	defer tagsFile.Close()

	byteValue, _ := ioutil.ReadAll(tagsFile)
	var labels []*Label
	if err := json.Unmarshal(byteValue, &labels); err != nil {
		log.Fatal("unable to parse tags")
		return err
	}
	s.labelMap = make(map[int]*Label)

	for _, l := range labels {
		s.labelMap[l.ID] = l
	}
	logrus.Info("label map created")
	return nil
}

func (s *tfService) loadGraph() error {
	// Load inception model
	model, err := ioutil.ReadFile(graphFile)
	if err != nil {
		return err
	}
	s.graph = tensorflow.NewGraph()
	if err := s.graph.Import(model, ""); err != nil {
		return err
	}
	logrus.Info("graph loaded")
	return nil
}

func (s *tfService) createSession() error {
	session, err := tensorflow.NewSession(s.graph, nil)
	if err != nil {
		return err
	}
	s.session = session
	logrus.Info("session created")
	return nil
}


func (s *tfService) normalizeImage(body io.ReadCloser) (*tensorflow.Tensor, error) {
	var buf bytes.Buffer
	io.Copy(&buf, body)
	logrus.Info("normalizing image")
	tensor, err := tensorflow.NewTensor(buf.String())
	if err != nil {
		return nil, err
	}

	graph, input, output, err := s.getNormalizedGraph()
	if err != nil {
		return nil, err
	}

	session, err := tensorflow.NewSession(graph, nil)
	if err != nil {
		return nil, err
	}

	normalized, err := session.Run(
		map[tensorflow.Output]*tensorflow.Tensor{
			input: tensor,
		},
		[]tensorflow.Output{
			output,
		},
		nil)
	if err != nil {
		log.Fatal(err.Error())
		return nil, err
	}

	logrus.Info("image normalized")
	return normalized[0], nil
}

// Creates a graph to decode, rezise and normalize an image
func (s *tfService) getNormalizedGraph() (graph *tensorflow.Graph, input, output tensorflow.Output, err error) {
	scope := op.NewScope()
	input = op.Placeholder(scope, tensorflow.String)
	decode := op.DecodeJpeg(scope, input, op.DecodeJpegChannels(3))
	output = op.ExpandDims(scope,
		// cast image to uint8
		op.Cast(scope, decode, tensorflow.Uint8),
		op.Const(scope.SubScope("make_batch"), int32(0)))

	graph, err = scope.Finalize()

	return graph, input, output, err
}

