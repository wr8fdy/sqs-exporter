# builder
FROM golang:1.12.5-alpine3.9 AS builder

ENV GO111MODULE=on

WORKDIR /go/src/sqs_exporter

COPY ./* /go/src/sqs_exporter/

RUN apk --no-cache add git
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o sqs_prom

# final
FROM alpine:3.9

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /go/src/sqs_exporter/sqs_prom .

EXPOSE 9108

CMD ["./sqs_prom"]
