FROM golang:1.20

WORKDIR /app


COPY go.mod ./
COPY go.sum ./
COPY *.go ./

RUN go get -d -v

ENV GCO_ENABLED=0

RUN go build -ldflags="-w -s" -o /scw-container-deploy .


ENTRYPOINT ["/scw-container-deploy"]
