FROM golang:1.20.2-alpine3.17


RUN CGO_ENABLED=0 go build -o /bin/sqs-autoscaler-controller main.go && \
    chmod +x /bin/sqs-autoscaler-controller

ENTRYPOINT ["/bin/sqs-autoscaler-controller"]
