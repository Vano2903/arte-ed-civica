FROM golang:1.18.0-alpine3.15

EXPOSE 8080
WORKDIR /go/src/mostra
COPY go.mod go.sum /go/src/mostra/
RUN go mod download

COPY .env /go/src/mostra/
COPY responser/ /go/src/mostra/responser
COPY *.go /go/src/mostra/
COPY ./pages/ /go/src/mostra/pages/
RUN go build -o mostra

CMD ./mostra