package scaler

import (
	"testing"

	"github.com/razorpay/sqs-autoscaler-controller/pkg/crd"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// helpers

func int32Ptr(i int32) *int32 { return &i }

func makeScaler(minPods, maxPods int32, upThreshold, downThreshold int64, upAmount, downAmount int32) *crd.SqsAutoScaler {
	return &crd.SqsAutoScaler{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: crd.AutoScalerSpec{
			MinPods: minPods,
			MaxPods: maxPods,
			ScaleUp: crd.ScaleSpec{
				Threshold: upThreshold,
				Amount:    upAmount,
			},
			ScaleDown: crd.ScaleSpec{
				Threshold: downThreshold,
				Amount:    downAmount,
			},
		},
	}
}

func makeDeployment(replicas int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		Status: appsv1.DeploymentStatus{Replicas: replicas},
		Spec:   appsv1.DeploymentSpec{Replicas: int32Ptr(replicas)},
	}
}

// min / max

func TestMin(t *testing.T) {
	tests := []struct {
		a, b, want int32
	}{
		{1, 2, 1},
		{2, 1, 1},
		{5, 5, 5},
		{0, 10, 0},
		{-1, 0, -1},
	}
	for _, tt := range tests {
		if got := min(tt.a, tt.b); got != tt.want {
			t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		a, b, want int32
	}{
		{1, 2, 2},
		{2, 1, 2},
		{5, 5, 5},
		{0, 10, 10},
		{-1, 0, 0},
	}
	for _, tt := range tests {
		if got := max(tt.a, tt.b); got != tt.want {
			t.Errorf("max(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

// targetReplicas

func TestTargetReplicas(t *testing.T) {
	s := Scaler{}

	tests := []struct {
		name     string
		queueLen int64
		replicas int32
		want     int32
	}{
		// scale up: queue >= upThreshold(100), add 5, cap at maxPods(20)
		{
			name: "scale up - normal",
			queueLen: 100, replicas: 5, want: 10,
		},
		{
			name: "scale up - capped at maxPods",
			queueLen: 200, replicas: 18, want: 20,
		},
		{
			name: "scale up - already at max",
			queueLen: 500, replicas: 20, want: 20,
		},
		// scale down: queue <= downThreshold(10), subtract 2, floor at minPods(1)
		{
			name: "scale down - normal",
			queueLen: 5, replicas: 10, want: 8,
		},
		{
			name: "scale down - floored at minPods",
			queueLen: 0, replicas: 2, want: 1,
		},
		{
			name: "scale down - already at min",
			queueLen: 0, replicas: 1, want: 1,
		},
		// steady state: 10 < queue < 100
		{
			name: "steady state - no change",
			queueLen: 50, replicas: 7, want: 7,
		},
		{
			name: "steady state - at upper boundary (99)",
			queueLen: 99, replicas: 3, want: 3,
		},
		{
			name: "steady state - at lower boundary (11)",
			queueLen: 11, replicas: 3, want: 3,
		},
		// exact threshold boundaries
		{
			name: "exact scale up threshold",
			queueLen: 100, replicas: 1, want: 6,
		},
		{
			name: "exact scale down threshold",
			queueLen: 10, replicas: 5, want: 3,
		},
	}

	// scaler: minPods=1, maxPods=20, upThreshold=100, downThreshold=10, upAmount=5, downAmount=2
	scale := makeScaler(1, 20, 100, 10, 5, 2)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := makeDeployment(tt.replicas)
			got, err := s.targetReplicas(QueueStats{Visible: tt.queueLen}, scale, d)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("targetReplicas(queueLen=%d, replicas=%d) = %d, want %d",
					tt.queueLen, tt.replicas, got, tt.want)
			}
		})
	}
}

func TestTargetReplicas_ZeroReplicas(t *testing.T) {
	s := Scaler{}
	scale := makeScaler(0, 10, 100, 10, 3, 1)
	d := makeDeployment(0)

	// queue in steady state — should stay at 0
	got, err := s.targetReplicas(QueueStats{Visible: 50}, scale, d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 0 {
		t.Errorf("expected 0 replicas in steady state, got %d", got)
	}
}

func TestTargetReplicas_LargeScaleUp(t *testing.T) {
	s := Scaler{}
	// upAmount=100 but maxPods=5 — result must be capped
	scale := makeScaler(1, 5, 50, 5, 100, 1)
	d := makeDeployment(3)

	got, err := s.targetReplicas(QueueStats{Visible: 1000}, scale, d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 5 {
		t.Errorf("expected capped at maxPods=5, got %d", got)
	}
}

func TestTargetReplicas_LargeScaleDown(t *testing.T) {
	s := Scaler{}
	// downAmount=100 but minPods=2 — result must be floored
	scale := makeScaler(2, 20, 100, 50, 1, 100)
	d := makeDeployment(10)

	got, err := s.targetReplicas(QueueStats{Visible: 0}, scale, d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 2 {
		t.Errorf("expected floored at minPods=2, got %d", got)
	}
}

// in-flight aware scale-down (includeInFlightInScaleDown)

func TestTargetReplicas_InFlightAwareScaleDown(t *testing.T) {
	s := Scaler{}

	// scaler: minPods=1, maxPods=20, upThreshold=100, downThreshold=10, upAmount=5, downAmount=2
	newScaler := func(includeInFlight bool) *crd.SqsAutoScaler {
		scale := makeScaler(1, 20, 100, 10, 5, 2)
		scale.Spec.IncludeInFlightInScaleDown = includeInFlight
		return scale
	}

	tests := []struct {
		name            string
		includeInFlight bool
		visible         int64
		inFlight        int64
		replicas        int32
		want            int32
	}{
		{
			name:            "disabled - in-flight ignored, scales down",
			includeInFlight: false,
			visible:         0, inFlight: 30, replicas: 10, want: 8,
		},
		{
			name:            "enabled - in-flight holds replicas (visible drained, workers busy)",
			includeInFlight: true,
			visible:         0, inFlight: 30, replicas: 10, want: 10,
		},
		{
			name:            "enabled - scales down when visible+inFlight within threshold",
			includeInFlight: true,
			visible:         4, inFlight: 6, replicas: 10, want: 8,
		},
		{
			name:            "enabled - visible+inFlight exactly at threshold scales down",
			includeInFlight: true,
			visible:         5, inFlight: 5, replicas: 10, want: 8,
		},
		{
			name:            "enabled - visible+inFlight just above threshold holds",
			includeInFlight: true,
			visible:         5, inFlight: 6, replicas: 10, want: 10,
		},
		{
			name:            "enabled - floored at minPods when queue fully idle",
			includeInFlight: true,
			visible:         0, inFlight: 0, replicas: 2, want: 1,
		},
		// scale-up must stay visible-only even when the flag is on:
		// in-flight messages already have a worker and must not add pods.
		{
			name:            "enabled - scale up ignores in-flight",
			includeInFlight: true,
			visible:         50, inFlight: 500, replicas: 7, want: 7,
		},
		{
			name:            "enabled - scale up on visible alone",
			includeInFlight: true,
			visible:         100, inFlight: 0, replicas: 5, want: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scale := newScaler(tt.includeInFlight)
			d := makeDeployment(tt.replicas)
			got, err := s.targetReplicas(QueueStats{Visible: tt.visible, InFlight: tt.inFlight}, scale, d)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("targetReplicas(visible=%d, inFlight=%d, includeInFlight=%v, replicas=%d) = %d, want %d",
					tt.visible, tt.inFlight, tt.includeInFlight, tt.replicas, got, tt.want)
			}
		})
	}
}
