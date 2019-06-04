FROM ubuntu AS nature-cnn
WORKDIR tmp
RUN apt-get update \
    && apt-get install -y wget \
    && wget http://download.tensorflow.org/models/object_detection/faster_rcnn_resnet50_fgvc_2018_07_19.tar.gz \
    && tar -xzf faster_rcnn_resnet50_fgvc_2018_07_19.tar.gz

FROM golang:stretch AS build-base
COPY go.mod /api/go.mod
COPY go.sum /api/go.sum
WORKDIR /tmp
RUN apt-get update && apt-get install -y wget gcc g++ libc-dev \
    && wget https://storage.googleapis.com/tensorflow/libtensorflow/libtensorflow-cpu-linux-x86_64-1.12.0.tar.gz \
    && tar -C /usr/local -xzf libtensorflow-cpu-linux-x86_64-1.12.0.tar.gz \
    && wget http://download.tensorflow.org/models/object_detection/faster_rcnn_resnet50_fgvc_2018_07_19.tar.gz \
RUN ldconfig
WORKDIR /api
RUN go mod download

FROM build-base AS build-env
COPY . /api
COPY --from=nature-cnn /tmp/faster_rcnn_resnet50_fgvc_2018_07_19/frozen_inference_graph.pb ./models/nature.pb
RUN CGO_ENABLED=1 GOOS=linux go build cmd/predict-api/main.go
CMD ["./main"]