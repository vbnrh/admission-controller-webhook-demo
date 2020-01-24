/*
Copyright (c) 2019 StackRox Inc.

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
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"


	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	tlsDir      = `/etc/webhook`
	tlsCertFile = `tls.crt`
	tlsKeyFile  = `tls.key`
)

var (
	podResource = metav1.GroupVersionResource{Version: "v1", Resource: "pods"}
)

func validatePodByName(req *v1beta1.AdmissionRequest) ([]patchOperation, error) {
	// This handler should only get called on Pod objects as per the ValidatingWebhookConfiguration in the YAML file.
	// However, if (for whatever reason) this gets invoked on an object of a different kind, issue a log message but
	// let the object request pass through otherwise.
	if req.Resource != podResource {
		log.Printf("expect resource to be %s", podResource)
		return nil, nil
	}

	// Parse the Pod object.
	raw := req.Object.Raw
	pod := corev1.Pod{}

	if _, _, err := universalDeserializer.Decode(raw, nil, &pod); err != nil {
		return nil, fmt.Errorf("could not deserialize pod object: %v", err)
	}

	var patches []patchOperation
	if pod.ObjectMeta.Name == "reject-pod" {
		fmt.Println(pod.ObjectMeta.Name + "\n" + pod.ObjectMeta.Namespace)
		return nil, errors.New("pod-to-be-rejected found ! creation failed ")
	}

	return patches, nil
}

func main() {
	//certPath := filepath.Join(tlsDir, tlsCertFile)
	//keyPath := filepath.Join(tlsDir, tlsKeyFile)

	cert := flag.String("tls-cert-file", "cert.pem", "File containing the default x509 Certificate for HTTPS.")
	key := flag.String("tls-private-key-file", "key.pem", "File containing the default x509 private key matching --tls-cert-file.")

	flag.Parse()

	//cert := "/etc/webhook/cert.pem"
	//key := "/etc/webhook/key.pem"

	keyPair, err := NewTlsKeypairReloader(*cert, *key)

	if err != nil {
		log.Printf("error load certificate: %v", err.Error())
	}

	

	mux := http.NewServeMux()
	mux.Handle("/validate", admitFuncHandler(validatePodByName))
	/* server := &http.Server{
		// We listen on port 8443 such that we do not need root privileges or extra capabilities for this server.
		// The Service object will take care of mapping this port to the HTTPS port 443.
		Addr:    ":8443",
		Handler: mux,
	} */

	var httpServer *http.Server
	httpServer = &http.Server{
		Addr: ":8443",
		TLSConfig: &tls.Config{
			GetCertificate: keyPair.GetCertificateFunc(),
		},
		Handler: mux,
	}
	//log.Fatal(server.ListenAndServeTLS(certPath, keyPath))
	log.Fatal(httpServer.ListenAndServeTLS("", ""))
}
