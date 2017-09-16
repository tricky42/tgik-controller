package client

import (
	"fmt"
	"reflect"
	"time"

	crv1 "github.com/tricky42/sf_envs/apis/cr/v1"
	apiv1 "k8s.io/api/core/v1"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
)

//CreateCustomResourceDefinition tries to create the CustomResource Definition for environments and waits for max 10 seconds for it to be visible in the system.
func CreateCustomResourceDefinition(clientset apiextensionsclient.Interface) (*apiextensionsv1beta1.CustomResourceDefinition, error) {
	crd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: crv1.FullEnvironmentName,
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   crv1.EnvironmentGroup,
			Version: crv1.SchemeGroupVersion.Version,
			Scope:   apiextensionsv1beta1.NamespaceScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural: crv1.EnvironmentsPlural,
				Kind:   reflect.TypeOf(crv1.Environment{}).Name(),
			},
		},
	}
	_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	if err != nil {
		return nil, err
	}

	// wait for CRD being established
	err = wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
		crd, err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crv1.FullEnvironmentName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		for _, cond := range crd.Status.Conditions {
			switch cond.Type {
			case apiextensionsv1beta1.Established:
				if cond.Status == apiextensionsv1beta1.ConditionTrue {
					return true, err
				}
			case apiextensionsv1beta1.NamesAccepted:
				if cond.Status == apiextensionsv1beta1.ConditionFalse {
					fmt.Printf("Name conflict: %v\n", cond.Reason)
				}
			}
		}
		return false, err
	})
	if err != nil {
		deleteErr := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(crv1.FullEnvironmentName, nil)
		if deleteErr != nil {
			return nil, errors.NewAggregate([]error{err, deleteErr})
		}
		return nil, err
	}
	return crd, nil
}

//WaitForEnvironmentInstanceProcessed waits until the Environment has been processed.
func WaitForEnvironmentInstanceProcessed(environmentClient *rest.RESTClient, name string) error {
	return wait.Poll(100*time.Millisecond, 10*time.Second, func() (bool, error) {
		var environment crv1.Environment
		err := environmentClient.Get().
			Resource(crv1.EnvironmentsPlural).
			Namespace(apiv1.NamespaceDefault).
			Name(name).
			Do().Into(&environment)

		if err == nil && environment.Status.State == crv1.EnvironmentStateProcessed {
			return true, nil
		}

		return false, err
	})
}

func NewClient(cfg *rest.Config) (*rest.RESTClient, *runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	if err := crv1.AddToScheme(scheme); err != nil {
		return nil, nil, err
	}

	config := *cfg
	config.GroupVersion = &crv1.SchemeGroupVersion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme)}

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, nil, err
	}

	return client, scheme, nil
}

