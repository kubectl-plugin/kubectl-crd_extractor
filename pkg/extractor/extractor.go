// Copyright 2023 The kubectl-plugin Authors
// SPDX-License-Identifier: Apache-2.0

package extractor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/yaml"
)

// groupResource contains the APIGroup and APIResource
type groupResource struct {
	APIGroup        string
	APIGroupVersion string
	APIResource     metav1.APIResource
}

func (gr groupResource) ToGVK() string {
	return fmt.Sprintf("%s.%s", gr.APIResource.Name, gr.APIGroup)
}

func GetCRDs(ctx context.Context, restClientGetter genericclioptions.RESTClientGetter, g, v, k string) error {
	discoveryClient, err := restClientGetter.ToDiscoveryClient()
	if err != nil {
		return fmt.Errorf("ToDiscoveryClient: %w", err)
	}
	lists, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		return fmt.Errorf("get ServerPreferredResources: %w", err)
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
			if gv.Group == g && gv.Version == v && k == res.Kind {
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
		return fmt.Errorf("ToRESTConfig: %w", err)
	}

	clientSet := apiextensionsclientset.NewForConfigOrDie(restConfig)
	crd, err := clientSet.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, resource.ToGVK(), metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get crd: %w", err)
	}

	var oapiSchema *apiextensionsv1.CustomResourceValidation
	for _, version := range crd.Spec.Versions {
		if version.Name == v {
			oapiSchema = version.Schema
		}
	}

	b, err := yaml.Marshal(oapiSchema.OpenAPIV3Schema)
	if err != nil {
		return fmt.Errorf("marshal yaml: %w", err)
	}

	data, err := yaml.YAMLToJSON(b)
	if err != nil {
		return fmt.Errorf("YAMLToJSON: %w", err)
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "  "); err != nil {
		return fmt.Errorf("indent json: %w", err)
	}

	f, err := os.Create(strings.ToLower(fmt.Sprintf("%s_%s.json", k, v)))
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if _, err := buf.WriteTo(f); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
