FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git

COPY . .
# Using go get to fetch /go/src/golang.org/x/net/context
RUN go get -d -v
COPY . .
RUN go build -o /app/queue_test_task

FROM scratch

COPY --from=builder /app/queue_test_task /app/queue_test_task
ENTRYPOINT [ "/app/queue_test_task" ]