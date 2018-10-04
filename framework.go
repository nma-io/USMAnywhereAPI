package main

/*
Quick Alienvault USM Anywhere client for Go.

Requires gjson for json parsing:
	go get github.com/tidwall/gjson

NMA 2018.
*/
import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/tidwall/gjson"
)

var URL string = "alienvault.cloud/api/2.0"
var HOST string = "yoursubdomain"

func Auth(apiUser, apiKey string) (token string) {
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
	json, _ := ioutil.ReadAll(resp.Body)

	if strings.Contains(string(json), "Unauthorized") {
		log.Fatal(string(json)) // Bad API Key
		return
	}
	token = gjson.Get(string(json), "access_token").String()
	if len(token) < 128 {
		log.Fatal("No Token Received.") // We shouldn't get here.
	}
	return
}

func Alarms(accessToken string) string {
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
	json, _ := ioutil.ReadAll(resp.Body)
	jsonOutput := gjson.Get(string(json), "_embedded.alarms").String()

	return jsonOutput
}

func Events() {
	// Not implemented ... But here is where you can get more data.
}

func main() {
	key := Auth("apiclientname", "secret")
	x := Alarms(key)
	rule := gjson.Get(x, "#.rule_method")
	src := gjson.Get(x, "#.alarm_sources")
	for i := range rule.Array() {
		fmt.Println(rule.Array()[i], src.Array()[i])
	}

}
