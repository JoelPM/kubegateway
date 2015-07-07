/*
Copyright 2014 The Kubernetes Authors All rights reserved.

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

package validation

import (
	"github.com/joelpm/kubegateway/Godeps/_workspace/src/github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/joelpm/kubegateway/Godeps/_workspace/src/github.com/GoogleCloudPlatform/kubernetes/pkg/util"
	errs "github.com/joelpm/kubegateway/Godeps/_workspace/src/github.com/GoogleCloudPlatform/kubernetes/pkg/util/fielderrors"
)

// ValidateEvent makes sure that the event makes sense.
func ValidateEvent(event *api.Event) errs.ValidationErrorList {
	allErrs := errs.ValidationErrorList{}
	// TODO: There is no namespace required for minion
	if event.InvolvedObject.Kind != "Node" &&
		event.Namespace != event.InvolvedObject.Namespace {
		allErrs = append(allErrs, errs.NewFieldInvalid("involvedObject.namespace", event.InvolvedObject.Namespace, "namespace does not match involvedObject"))
	}
	if !util.IsDNS1123Subdomain(event.Namespace) {
		allErrs = append(allErrs, errs.NewFieldInvalid("namespace", event.Namespace, ""))
	}
	return allErrs
}
