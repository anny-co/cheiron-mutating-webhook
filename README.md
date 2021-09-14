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

Then, installing the mutating webhook and passing the secrets explicitly as configuration will mutate each CREATE request 
of a service account (or pod) and add the secrets to the respective imagePullSecrets spec.

> NOTE that *you*, the cluster operator, need to ensure that the secrets exist in the namespace. For now, this webhook cannot 
> do that for you, as API client functionality is not fully integrated. Missing LocalObjectReferences will simply be omitted
> by the pod controller on startup of new pods.

## Reading Kubernetes secrets from this application

> `TODO(feat): add client call to get dockerconfigjson secrets from API and add those to the JSON Patch`

Due to Kubernetes scoping on RBAC-enabled clusters, secrets are not visible to the service account attached to a running pod
*by default*. We need to use a different service account with a different, more elevated clusterrole to obtain secrets