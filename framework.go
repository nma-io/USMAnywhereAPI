package main

/*
Quick Alienvault USM Anywhere client for Go.
This was VERY hackish; I need to figure out a better solution
for JSON problems, this re-defining interfaces is ugly and
confusing. Recommending the python code for simplicity,
but I'm leaving this here and soliciting pull requests to show
me how to handle the json better.

NMA 2018.
*/
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var URL string = "alienvault.cloud/api/2.0"
var HOST string = "yoursubdomain"

func Auth(apiUser, apiKey string) (token string) {
	var json_data map[string]interface{}
	client := &http.Client{}
	v := url.Values{}
	apilink := fmt.Sprintf("https://" + HOST + "." + URL + "/oauth/token?grant_type=client_credentials")
	method := "POST"
	req, err := http.NewRequest(method, apilink, strings.NewReader(v.Encode()))

	req.SetBasicAuth(apiUser, apiKey)
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		log.Fatal("Alienvault OAuth Error.")
		return
	}
	defer resp.Body.Close()
	z, _ := ioutil.ReadAll(resp.Body)

	if strings.Contains(string(z), "Unauthorized") {
		log.Fatal(string(z)) // Bad API Key
		return
	}
	json.Unmarshal([]byte(z), &json_data)
	if len(json_data["access_token"].(string)) > 0 {
		token = json_data["access_token"].(string)
	}
	return
}

func Alarms(accessToken string) []byte {
	var json_data map[string]interface{}
	client := &http.Client{}
	v := url.Values{}
	apilink := fmt.Sprintf("https://" + HOST + "." + URL + "/alarms/?page=1&size=20&suppressed=false&status=open")
	method := "GET"
	req, err := http.NewRequest(method, apilink, strings.NewReader(v.Encode()))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	z, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(z, &json_data)
	alarms := json_data["_embedded"].(map[string]interface{})
	jsonOutput, _ := json.Marshal(alarms)
	return jsonOutput
}

func Events() {
	// Not implemented
}

func main() {
	key := Auth("apiclientname", "secret")
	var jdata map[string]interface{}
	x := Alarms(key)
	err := json.Unmarshal(x, &jdata)
	if err != nil {
		log.Fatal(err)
	}
	for _, k := range jdata["alarms"].([]interface{}) {
		rule := k.(map[string]interface{})["rule_method"]
		src := k.(map[string]interface{})["alarm_sources"]

		fmt.Println(rule, src)

	}

}
