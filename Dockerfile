FROM golang:stretch AS build
WORKDIR /tmp
RUN apt-get update && apt-get install -y wget gcc g++ libc-dev \
    && wget https://storage.googleapis.com/tensorflow/libtensorflow/libtensorflow-cpu-linux-x86_64-1.15.0.tar.gz \
    && tar -C /usr/local -xzf libtensorflow-cpu-linux-x86_64-1.15.0.tar.gz
RUN ldconfig
COPY go.mod /api/go.mod
COPY go.sum /api/go.sum
WORKDIR /api
RUN go mod download
EXPOSE 8080
COPY . /api
RUN CGO_ENABLED=1 GOOS=linux go build cmd/predict-api/main.go

FROM debian:sid-slim AS prod
WORKDIR /tmp
RUN apt-get update && apt-get install -y wget  \
    && wget https://storage.googleapis.com/tensorflow/libtensorflow/libtensorflow-cpu-linux-x86_64-1.15.0.tar.gz \
    && apt-get purge -y wget  \
    && tar -C /usr/local -xzf libtensorflow-cpu-linux-x86_64-1.15.0.tar.gz \
    && rm -rf libtensorflow-cpu-linux-x86_64-1.15.0.tar.gz \
    && ldconfig
WORKDIR /
EXPOSE 8080
COPY --from=build /api/main /main
CMD ["./main"]