package crd

import "k8s.io/apimachinery/pkg/runtime"

func (in *ScaleSpec) DeepCopyInto(out *ScaleSpec) {
	*out = *in
	out.Threshold = in.Threshold
	out.Amount = in.Amount
}

func (in *AutoScalerSpec) DeepCopyInto(out *AutoScalerSpec) {
	*out = *in
	out.Queue = in.Queue
	out.Deployment = in.Deployment
	out.MinPods = in.MinPods
	out.MaxPods = in.MaxPods
	in.ScaleUp.DeepCopyInto(&out.ScaleUp)
	in.ScaleDown.DeepCopyInto(&out.ScaleDown)
}

func (in *SqsAutoScaler) DeepCopyInto(out *SqsAutoScaler) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

func (in *SqsAutoScaler) DeepCopy() *SqsAutoScaler {
	if in == nil {
		return nil
	}
	out := new(SqsAutoScaler)
	in.DeepCopyInto(out)
	return out
}

func (in *SqsAutoScaler) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *SqsAutoScalerList) DeepCopyInto(out *SqsAutoScalerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SqsAutoScaler, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

func (in *SqsAutoScalerList) DeepCopy() *SqsAutoScalerList {
	if in == nil {
		return nil
	}
	out := new(SqsAutoScalerList)
	in.DeepCopyInto(out)
	return out
}

func (in *SqsAutoScalerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
