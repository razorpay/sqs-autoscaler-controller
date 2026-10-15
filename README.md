# SQS Autoscaler Controller

This provides a controller to autoscale Deployments according to the queue size of an SQS queue.

## Creating

```yaml
apiVersion: aws.uswitch.com/v1
kind: SqsAutoScaler
metadata:
  name: testing-scaler
  namespace: myns
spec:
  deployment: testapp
  maxPods: 20
  minPods: 1
  queue: https://sqs.eu-west-1.amazonaws.com/1234567890/some-queue
  scaleDown:
    amount: 2
    threshold: 10
  scaleUp:
    amount: 5
    threshold: 100
  # Optional (default false): also count in-flight messages
  # (ApproximateNumberOfMessagesNotVisible) for the scale-down decision, so
  # workers busy with long-running tasks aren't scaled away while the visible
  # queue looks drained. Scale-up always uses visible messages only.
  includeInFlightInScaleDown: true
```

```
$ kubectl apply -f scaler.yaml
$ kubectl get sqsautoscalers
```
