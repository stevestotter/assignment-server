FROM golang:1.14

WORKDIR /go/src/assignment-server
COPY . .

RUN make deps
RUN go install -v ./...

ENV KAFKA_URL="kafka:9093"

CMD ["assignment-server"]