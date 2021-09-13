/*
Copyright 2021 anny UG (haftungsbeschränkt).

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

package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	admission "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	tlsDir      = `/run/secrets/tls`
	tlsCertFile = `tls.crt`
	tlsKeyFile  = `tls.key`
)

var (
	serviceaccountResource = metav1.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}
	podResource            = metav1.GroupVersionResource{Version: "v1", Resource: "pods"}
	logger                 *zap.SugaredLogger
)

// augmentImagePullSecrets constructs a new slice of LocalObjectReferences containing any existing and new ones
func augmentImagePullSecrets(existingPullSecrets []corev1.LocalObjectReference) []corev1.LocalObjectReference {

	imagePullSecrets := []corev1.LocalObjectReference{}

	if existingPullSecrets != nil || len(existingPullSecrets) > 0 {
		imagePullSecrets = existingPullSecrets
	}

	secretNames := viper.GetStringSlice("secretNames")

	for _, secretName := range secretNames {
		secret := &corev1.LocalObjectReference{
			Name: secretName,
		}
		if !contains(imagePullSecrets, secretName) {
			imagePullSecrets = append(imagePullSecrets, *secret)
		}
	}

	return imagePullSecrets
}

// contains checks if a slice of LocalObjectReferences already contains a secret name
// and returns true if so, otherwise false
func contains(s []corev1.LocalObjectReference, e string) bool {
	for _, secretObj := range s {
		if secretObj.Name == e {
			return true
		}
	}
	return false
}

// addImagePullSecretToPod runs the actual admission controller logic, i.e.
// for all configured secrets, checks if the pods resource requesting admission already contains the
// imagePullSecret or adds it otherwise
func augmentPod(req *admission.AdmissionRequest) ([]patchOperation, error) {

	if req.Resource != podResource {
		logger.Errorf("expect resource to be %s", podResource)
		return nil, nil
	}

	raw := req.Object.Raw
	pod := corev1.Pod{}

	if _, _, err := universalDeserializer.Decode(raw, nil, &pod); err != nil {
		logger.Errorf("could not deserialize pod object: %v", err)
		return nil, fmt.Errorf("could not deserialize pod object: %v", err)
	}

	newImagePullSecrets := augmentImagePullSecrets(pod.Spec.ImagePullSecrets)
	// convert the Spec to a JSON patch
	patches := []patchOperation{}

	patches = append(patches, patchOperation{
		Op:    "replace",
		Path:  "spec/imagePullSecrets",
		Value: newImagePullSecrets,
	})

	return patches, nil
}

// addImagePullSecretToServiceAccount runs the actual admission controller logic, i.e.
// for all configured secrets, checks if the serviceaccount resource requesting admission already contains the
// imagePullSecret or adds it otherwise
func augmentServiceAccount(req *admission.AdmissionRequest) ([]patchOperation, error) {

	if req.Resource != serviceaccountResource {
		logger.Errorf("expect resource to be %s", serviceaccountResource)
		return nil, nil
	}

	raw := req.Object.Raw
	serviceAccount := corev1.ServiceAccount{}

	if _, _, err := universalDeserializer.Decode(raw, nil, &serviceAccount); err != nil {
		logger.Errorf("could not deserialize serviceaccount object: %v", err)
		return nil, fmt.Errorf("could not deserialize serviceaccount object: %v", err)
	}

	newImagePullSecrets := augmentImagePullSecrets(serviceAccount.ImagePullSecrets)

	// convert the Spec to a JSON patch
	patches := []patchOperation{}
	patches = append(patches, patchOperation{
		Op:    "replace",
		Path:  "imagePullSecrets",
		Value: newImagePullSecrets,
	})

	return patches, nil
}

// main loads the configuration file and starts an HTTPS server with the provided certificates or fails fatally
func main() {
	zap, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
		return
	}
	defer zap.Sync()
	logger = zap.Sugar()

	certPath := filepath.Join(tlsDir, tlsCertFile)
	keyPath := filepath.Join(tlsDir, tlsKeyFile)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/cheiron/")
	viper.AddConfigPath(".")

	viper.OnConfigChange(func(e fsnotify.Event) {
		logger.Infof("config file changed", "file", e.Name)
	})
	viper.WatchConfig()

	err = viper.ReadInConfig()

	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Info("started without config file, won't do anything until secrets are added to the config file")
		} else {
			logger.Fatal("failed to load configuration file")
			return
		}
	}

	mux := http.NewServeMux()

	mux.Handle("/mutate-v1-service-account", admitFuncHandler(augmentPod))
	mux.Handle("/mutate-v1-pod", admitFuncHandler(augmentServiceAccount))

	mux.Handle("/healthz", ping)
	mux.Handle("/readyz", ping)

	server := &http.Server{
		Addr:    ":8443",
		Handler: mux,
	}

	logger.Fatalf("couldn't start server: %v", server.ListenAndServeTLS(certPath, keyPath))
}
