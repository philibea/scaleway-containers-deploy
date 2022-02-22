FROM golang:1.17

WORKDIR /scw-container-deploy
COPY . /scw-container-deploy

RUN go get -d -v

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -v -o scw-container-deploy .

ENTRYPOINT ["/scw-container-deploy"]