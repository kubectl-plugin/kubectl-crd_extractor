// Copyright 2023 The kubectl-plugin Authors
// SPDX-License-Identifier: Apache-2.0

package extractor

import (
	"context"
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// groupResource contains the APIGroup and APIResource
type groupResource struct {
	APIGroup        string
	APIGroupVersion string
	APIResource     metav1.APIResource
}

func (gr groupResource) ToRG() string {
	return fmt.Sprintf("%s.%s", gr.APIResource.Name, gr.APIGroup)
}

func GetCR(ctx context.Context, restClientGetter genericclioptions.RESTClientGetter, gvk schema.GroupVersionKind) (*apiextensionsv1.CustomResourceValidation, error) {
	discoveryClient, err := restClientGetter.ToDiscoveryClient()
	if err != nil {
		return nil, fmt.Errorf("ToDiscoveryClient: %w", err)
	}
	lists, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		return nil, fmt.Errorf("get ServerPreferredResources: %w", err)
	}

	resource := groupResource{}
	for _, list := range lists {
		if len(list.APIResources) == 0 {
			continue
		}
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			continue
		}
		for _, res := range list.APIResources {
			if len(res.Verbs) == 0 {
				continue
			}
			if gv.Group == gvk.Group && gv.Version == gvk.Version && res.Kind == gvk.Kind {
				resource = groupResource{
					APIGroup:        gv.Group,
					APIGroupVersion: gv.String(),
					APIResource:     res,
				}
			}
		}
	}

	restConfig, err := restClientGetter.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("ToRESTConfig: %w", err)
	}

	clientSet, err := apiextensionsclientset.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("apiextensionsclientset.NewForConfig: %w", err)
	}
	crd, err := clientSet.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, resource.ToRG(), metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get crd: %w", err)
	}

	var cr *apiextensionsv1.CustomResourceValidation
	for _, version := range crd.Spec.Versions {
		if version.Name == gvk.Version {
			cr = version.Schema
		}
	}

	return cr, nil
}
