/*
Copyright 2018 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package serviceaccount

import (
	"context"
	"encoding/json"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/mutate-v1-service-account,mutating=true,failurePolicy=fail,groups="",resources=serviceaccounts,verbs=create;update,versions=v1,name=mserviceaccount.kb.io

// ServiceAccountMutator mutates ServiceAccounts
type ServiceAccountMutator struct {
	Client  client.Client
	decoder *admission.Decoder
}

// ServiceAccountMutator adds an annotation to every incoming service accounts.
func (a *ServiceAccountMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	serviceAccount := &corev1.ServiceAccount{}

	err := a.decoder.Decode(req, serviceAccount)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// TODO: implement service account augmentation

	marshaledServiceAccount, err := json.Marshal(serviceAccount)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledServiceAccount)
}

// ServiceAccountMutator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (a *ServiceAccountMutator) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}
