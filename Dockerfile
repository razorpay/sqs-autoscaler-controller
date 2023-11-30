FROM c.rzp.io/razorpay/onggi-multi-arch:rzp-golden-image-base-golang-1.20


ENV SRC_DIR=/go/src/github.com/razorpay/sqs-autoscaler-controller

WORKDIR $SRC_DIR

ADD . $SRC_DIR/

RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -o /bin/sqs-autoscaler-controller main.go && \
    chmod +x /bin/sqs-autoscaler-controller

ENTRYPOINT ["/bin/sqs-autoscaler-controller"]
