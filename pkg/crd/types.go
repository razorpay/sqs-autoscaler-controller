package crd

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SqsAutoScaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              AutoScalerSpec `json:"spec"`
}

type SqsAutoScalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []SqsAutoScaler `json:"items"`
}

type AutoScalerSpec struct {
	Queue      string    `json:"queue"`
	Deployment string    `json:"deployment"`
	MinPods    int32     `json:"minPods"`
	MaxPods    int32     `json:"maxPods"`
	ScaleUp    ScaleSpec `json:"scaleUp"`
	ScaleDown  ScaleSpec `json:"scaleDown"`
	// IncludeInFlightInScaleDown makes the scale-down decision consider
	// in-flight messages (ApproximateNumberOfMessagesNotVisible) in addition
	// to visible ones. This prevents premature scale-down of workers that are
	// still processing long-running tasks. Scale-up always uses visible
	// messages only: in-flight messages already have a worker.
	// Optional, defaults to false (legacy behaviour).
	IncludeInFlightInScaleDown bool `json:"includeInFlightInScaleDown,omitempty"`
}

type ScaleSpec struct {
	Threshold int64 `json:"threshold"`
	Amount    int32 `json:"amount"`
}
