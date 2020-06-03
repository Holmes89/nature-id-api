package predictor

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/tensorflow/tensorflow/tensorflow/go"
	"github.com/tensorflow/tensorflow/tensorflow/go/op"
	"gocloud.dev/blob"
	"io"
	"io/ioutil"
	"log"
	"nature-id-api/internal"
	"os"
	"time"
)

const (
	graphFile = "./models/nature.pb"
	tagsFile  = "./models/tags.json"
)


type tfService struct {
	bucket *blob.Bucket
	labelMap map[int]*internal.Prediction
	graph    *tensorflow.Graph
	session *tensorflow.Session
}


func NewTensorflowPredictor(bucket *blob.Bucket) (internal.Predictor, error) {
	s := &tfService{
		bucket: bucket,
	}
	if err := s.loadLabelMap(); err != nil {
		logrus.WithField("err", err).Error("unable to load map")
		return nil, err
	}

	err := s.loadGraphAndSession("models/faster_rcnn_resnet50_fgvc_2018_07_19.pb")
	if err != nil {
		logrus.Error("unable to load model")
		return nil, err
	}
	logrus.Info("service created")
	return s, nil
}

func (s *tfService) Predict(img io.Reader) (internal.Predictions, error) {

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

	var labels internal.Predictions

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
	var labels []*internal.Prediction
	if err := json.Unmarshal(byteValue, &labels); err != nil {
		log.Fatal("unable to parse tags")
		return err
	}
	s.labelMap = make(map[int]*internal.Prediction)

	for _, l := range labels {
		s.labelMap[l.ID] = l
	}
	logrus.Info("label map created")
	return nil
}


func (s *tfService) normalizeImage(body io.Reader) (*tensorflow.Tensor, error) {
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

// Creates a graph to decode, resize and normalize an image
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



func (s *tfService) loadGraphAndSession(path string) error {
	// Load Model from bucket
	// TODO load labels and models together
	logrus.WithField("path", path).Info("downloading model")
	model, err := s.bucket.ReadAll(context.Background(), path)
	if err != nil {
		return err
	}
	logrus.Info("downloaded model")
	s.graph = tensorflow.NewGraph()
	if err := s.graph.Import(model, ""); err != nil {
		return err
	}
	s.session, err = tensorflow.NewSession(s.graph, nil)
	if err != nil {
		return err
	}
	return nil
}