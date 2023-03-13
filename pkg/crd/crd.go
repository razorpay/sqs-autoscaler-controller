package crd

import (
	"context"
	"reflect"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// metav1 "k8s.io/apiextensions-apiserver/vendor/k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
)

func EnsureResource(client apiextensionsclient.Interface) error {
	crd := &v1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "sqsautoscalers.aws.uswitch.com",
		},
		Spec: v1.CustomResourceDefinitionSpec{
			Group: "aws.uswitch.com",
			Scope: v1.NamespaceScoped,
			Versions: []v1.CustomResourceDefinitionVersion{
				{
					Name:       "v1",
					Served:     true,
					Storage:    true,
					Deprecated: false,
					Schema:     &v1.CustomResourceValidation{},
				},
				{
					Name:       "v1beta1",
					Served:     true,
					Storage:    false,
					Deprecated: true,
					Schema:     &v1.CustomResourceValidation{},
				},
			},
			Names: v1.CustomResourceDefinitionNames{
				Singular: "sqsautoscaler",
				Plural:   "sqsautoscalers",
				Kind:     reflect.TypeOf(SqsAutoScaler{}).Name(),
			},
		},
	}
	_, err := client.ApiextensionsV1().CustomResourceDefinitions().Create(context.TODO(), crd, metav1.CreateOptions{})
	if err != nil && apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func NewClient(cfg *rest.Config) (*rest.RESTClient, *runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	builder := runtime.NewSchemeBuilder(addKnownTypes)
	err := builder.AddToScheme(scheme)
	if err != nil {
		return nil, nil, err
	}

	config := *cfg
	config.GroupVersion = &SchemeGroupVersion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme)}

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, nil, err
	}

	return client, scheme, err
}

const (
	GroupName = "aws.uswitch.com"
	Version   = "v1"
)

var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: Version}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion, &SqsAutoScaler{}, &SqsAutoScalerList{})
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
