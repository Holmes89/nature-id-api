package main

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
	"image/png"
	"nature-id-api/internal/predictor"
)

func main() {
	// set to use a video capture device 0
	deviceID := 0
	// open webcam
	webcam, err := gocv.OpenVideoCapture(deviceID)
	if err != nil {
		logrus.WithError(err).Fatal("unable to open webcam")
	}
	defer webcam.Close()

	// open display window
	window := gocv.NewWindow("Animal Detect")
	defer window.Close()

	// prepare image matrix
	img := gocv.NewMat()
	defer img.Close()

	pred, err := predictor.NewTensorflowService()
	if err != nil {
		logrus.WithField("err", err).Fatal("unable to start service")
	}

	for {
		if ok := webcam.Read(&img); !ok {
			logrus.WithField("device", deviceID).Fatal("unable to read device")
			return
		}
		if img.Empty() {
			continue
		}

		// detect thing
		i, err := img.ToImage()
		if err != nil {
			logrus.WithError(err).Error("unable to convert to image")
		}
		// create buffer
		buf := new(bytes.Buffer)

		// encode image to buffer
		err = png.Encode(buf, i)
		if err != nil {
			logrus.WithError(err).Error("unable to create buffer")
		}

		labels, err := pred.GetLabels(buf)
		if err != nil {
			logrus.WithError(err).Error("unable to predict")
		}

		for _, l := range labels {
			logrus.WithField("probability", l.Probability).Info(l.DisplayName)
		}

		// draw a rectangle around each face on the original image
		//for _, r := range rects {
		//	gocv.Rectangle(&img, r, blue, 3)
		//}

		// show the image in the window, and wait 1 millisecond
		window.IMShow(img)
		window.WaitKey(1)
	}
}
