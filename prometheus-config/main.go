package main

// bash
// unset https_proxy
// export KUBE_TOKEN=$(</var/run/secrets/kubernetes.io/serviceaccount/token)
// curl -sSk -H "Authorization: Bearer $KUBE_TOKEN" https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_PORT_443_TCP_PORT/api/v1/namespaces/api-factory-asia-qa/pods
// curl -sSk -H "Authorization: Bearer $KUBE_TOKEN" https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_PORT_443_TCP_PORT/api/v1/namespaces/api-factory-asia-monitoring-dev/pods

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func main() {

	namespace := getenv("prometheus_namespace", "api-factory-asia-qa")
	podList := getPodList(namespace)

	var f interface{}
	err := json.Unmarshal(podList, &f)
	if err != nil {
		errorf("JSON parse error: %v\n", err)
	}

	fmt.Print(f)

	// When decoding into an interface{} variable, the JSON module represents
	// dictionaries as map[string]interface{} maps
	data := f.(map[string]interface{})

	// When decoding into an interface{} variable, the JSON module represents
	// arrays as []interface{} slices
	items := data["items"].([]interface{})

	itemsTotal := len(items)
	// fmt.Println(itemsTotal)

	var targets []string

	for i := 0; i < itemsTotal; i++ {
		// fmt.Print(items[i])
		pod := items[i].(map[string]interface{})
		name := pod["metadata"].(map[string]interface{})["name"].(string)

		r1, _ := regexp.Compile("myaxa-apifactory")
		r2, _ := regexp.Compile("(build|deploy)")

		if r1.MatchString(name) && !r2.MatchString(name) {
			ip := podIP(items[i])
			// fmt.Printf("name: %s = %s\n", name, ip)
			targets = append(targets, ip)
		}

	}
	// fmt.Printf("Targets: %s", targets)
	// targets = append(targets, "111.111.111.111")
	// targets = append(targets, "222.222.222.222")

	doTmpl(targets)

}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func getAuthToken() string {
	tokenPath := "/var/run/secrets/kubernetes.io/serviceaccount/token"

	if _, err := os.Stat(tokenPath); os.IsNotExist(err) {
		return "authTokenDefault"
	}

	data, err2 := ioutil.ReadFile(tokenPath)
	if err2 != nil {
		errorf("Error reading Token file: %v\n", err2)
	}

	// trim new line just to be sure
	return strings.Trim(string(data), "\n")
}

func getPodList(namespace string) []byte {
	kubernetesServiceHost := getenv("KUBERNETES_SERVICE_HOST", "127.0.0.1")
	kubernetesPort := getenv("KUBERNETES_PORT_443_TCP_PORT", "3010")
	protocol := "http"
	if kubernetesPort == "443" {
		protocol = "https"
	}
	url := fmt.Sprintf("%s://%s:%s/api/v1/namespaces/%s/pods", protocol, kubernetesServiceHost, kubernetesPort, namespace)

	// Disable TLS verify for internal self signed certificates
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	// client := &http.Client{}

	req, e := http.NewRequest("GET", url, nil)
	if e != nil {
		errorf("Error setting new request %v\n", e)
	}
	req.Header.Add("Bearer", getAuthToken())

	res, err := client.Do(req)
	if err != nil {
		errorf("Get Pod List error: %v\n", err)
	}

	defer res.Body.Close()

	body, err2 := ioutil.ReadAll(res.Body)
	if err2 != nil {
		errorf("Get Pod List Error: %v\n", err2)
	}

	// Print body if not success
	if res.StatusCode != 200 {
		errorf("Response Error: %v\nBody: \n%v\n", res.Status, string(body))
	}

	return body
}

func podIP(item interface{}) string {
	podIP := item.(map[string]interface{})["status"].(map[string]interface{})["podIP"].(string)
	return podIP
}

func doTmpl(targets []string) {
	tmpl := template.Must(template.New("main").ParseFiles("./prometheus.tpl.yml"))

	e := tmpl.ExecuteTemplate(os.Stdout, "prometheus.tpl.yml", targets)
	if e != nil {
		errorf("JSON parse error: %v\n", e)
	}
}

func errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}
