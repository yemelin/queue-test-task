FROM golang:alpine AS builder

COPY . .
RUN go build -o /app/queue_test_task

FROM scratch

COPY --from=builder /app/queue_test_task /app/queue_test_task
ENTRYPOINT [ "/app/queue_test_task" ]