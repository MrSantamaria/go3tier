FROM golang:1.22 as builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o email_queue

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/email_queue .
EXPOSE 8080
CMD ["./email_queue"]