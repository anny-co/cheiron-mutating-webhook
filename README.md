# anny-co/cheiron-mutating-webhook

> NOTE: Cheiron is currently in very early stages of development and and far
> from anything usable. Feel free to contribute if you want to, an(n)y
> contributions are welcome!

The smaller brother of the Cheiron operator, that acts only as a (Golang) webserver that mutates
AdmissionReview objects from Kubernetes s.t. all service accounts have a set of imagePullSecrets specified at deployment.

> NOTE that different to the full operator, this webhook **CANNOT** guarantee or enforce that the imagePullSecrets exist in 
> the namespace of the object patched by the webhook! This exceeds the capabilities of this tiny controller and should be 
> done by the user in-charge!

## Combining Cheiron and the Webhook

Cheiron is capable of creating imagePullSecrets in any namespace. Hence it can be used and configured to *only create the specified imagePullSecret objects in all namespaces and leave the rest to the webhook*. This way, we can ensure that a secret exists before the webhook mutates the object with a LocalObjectReference that does not exist yet.

The Operator-SDK scaffolding provides means to inject Kubernetes clients such that we can theoretically fetch secrets from the API and only add those secrets available in the namespace.

> `TODO(feat): add client call to get dockerconfigjson secrets from API and add those to the JSON Patch`