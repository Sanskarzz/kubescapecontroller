package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

type ControlReport struct {
	Control struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Remediation string `json:"remediation"`
		Severity    string `json:"severity"`
	} `json:"control"`
	Summary struct {
		TotalResourcesCount    int `json:"totalResourcesCount"`
		FailedResourcesCount   int `json:"failedResourcesCount"`
		ExcludedResourcesCount int `json:"excludedResourcesCount"`
		SeverityScore          int `json:"severityScore"`
	} `json:"summary"`
}

type KubescapeResult struct {
	ControlReports []ControlReport `json:"controlReports"`
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not read request body: %v", err), http.StatusBadRequest)
		return
	}

	var admissionReviewReq admissionv1.AdmissionReview
	_, _, err = scheme.Codecs.UniversalDeserializer().Decode(body, nil, &admissionReviewReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not deserialize request: %v", err), http.StatusBadRequest)
		return
	}

	// Run Kubescape scan on the request body
	cmd := exec.Command("kubescape", "scan", "body")
	cmd.Stdin = bytes.NewReader(body)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error running Kubescape: %s\n", err)
		fmt.Printf("Stderr: %s\n", stderr.String())
		return
	}

	// Get the scan results
	results := stdout.String()

	// Decode the Kubescape results and extract the severityScore
	var result KubescapeResult
	err = json.Unmarshal([]byte(results), &result)
	if err != nil {
		panic(err)
	}

	// Show the results to the user and ask for validation
	fmt.Println("Kubescape scan results:")
	for _, report := range result.ControlReports {
		fmt.Printf("Control: %s, Severity: %s, Severity Score: %d\n",
			report.Control.Name,
			report.Control.Severity,
			report.Summary.SeverityScore)
	}

	var allowed bool
	fmt.Print("Allow request? (y/n): ")
	var input string
	fmt.Scanf("%s", &input)

	if input == "y" || input == "Y" {
		allowed = true
	} else {
		allowed = false
	}

	admissionReviewResponse := &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Response: &admissionv1.AdmissionResponse{
			UID:     admissionReviewReq.Request.UID,
			Allowed: allowed,
		},
	}

	responseInBytes, err := json.Marshal(admissionReviewResponse)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not marshal response: %v", err),
			http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseInBytes)
}
