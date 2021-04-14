# TODO: Make this a multistage image build
# TODO: Include Redis/Postgres args as inputs
FROM golang:1.15
ADD . /go/src/url-shortener
WORKDIR /go/src/url-shortener
RUN go build -o url-shortener main.go
CMD ["url-shortener"]