package main

/*
Alienvault USM Anywhere Alarm Relayer
(c) 2018, Nicholas Albright (w1rtp)

My team is leveraging a third party device to triage alarms.

We also leverage Microsoft Teams for alerting and sharing 
those alerts with our customers.

This application will pull Alarms from AT&T's Alienvault USM Anywhere
Product and send to Teams as well as a third party syslog collector
like Splunk or Graylog.

I run this from my third party collection server, and have that server
monitor tcp 2514 for interesting events, those are configurable below.
 
This runs well via cron every 5 min. Now you don't need USM Central.
*/

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

var URL string = "alienvault.cloud/api/2.0"
var HOST string

func talk(hook string, message string) (resp bool) {
	resp = true
	data := make(map[string]string)
	data["text"] = message
	mkjson, _ := json.Marshal(data)
	_, err := http.Post(hook, "application/json", bytes.NewBuffer(mkjson))
	if err != nil {
		resp = false
	}
	sendSyslog("tcp", "127.0.0.1:2514", message) // <-- Modify the remote syslog server
	return
}

func sendSyslog(proto string, server string, message string) (resp bool) {
	syslog, _ := net.Dial(proto, server)
	_, err := fmt.Fprintf(syslog, strings.Replace(message, "%", "%%", -1))
	if err != nil {
		log.Fatal(err)
	}
	defer syslog.Close()
	return

}

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

func Alarms(accessToken string, priority string) string {
	client := &http.Client{}
	v := url.Values{}

	apilink := fmt.Sprintf("https://"+HOST+"."+URL+"/alarms/?priority_label=%s&size=50&suppressed=false&status=open", priority)
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

func main() {
	a := flag.String("a", "high,medium", "Alarm types to alert on.")
	domain := flag.String("d", "", "Subdomain for Alienvault USM Anywhere.")
	ignore := flag.String("i", "USMA_ignore.txt", "File with plugins to ignore (one per line)")
	teamsHook := flag.String("t", "", "Teams Hook to send alert to.")
	apic := flag.String("c", "", "API Client")
	apik := flag.String("k", "", "API Key")
	flag.Parse()

	HOST = *domain
	if len(HOST) < 3 || len(*apic) < 3 || len(*apik) < 30 || len(*teamsHook) < 30 {
		log.Fatal("Alarm Relay Failure: use -h")
	}
	alarmArray := strings.Split(*a, ",")
	timecomp := (time.Now().Unix() - 300) * 1000
	key := Auth(*apic, *apik)
	ignorePatterns, _ := ioutil.ReadFile(*ignore)

	for _, alarmType := range alarmArray {
		x := Alarms(key, alarmType)

		rule := gjson.Get(x, "#.events.#.message.event_description")
		plugin := gjson.Get(x, "#.events.#.message.plugin")
		log := gjson.Get(x, "#.events.#.message.log")
		timestamp := gjson.Get(x, "#.events.#.message.timestamp_received")
		for i := range timestamp.Array() {
			if len(ignorePatterns) > 1 && strings.Contains(string(ignorePatterns), plugin.Array()[i].Array()[0].String()) {
				continue
			}
			TS := timestamp.Array()[i].Array()
			if timestamp.Array()[i].Array()[len(TS)-1].Int() >= timecomp {
				rawlog := strings.Replace(log.Array()[i].Array()[0].String(), "\\", "", -1)
				SndMSG := fmt.Sprintf("%s :: %s - %s %s", strings.ToUpper(alarmType), plugin.Array()[i].Array()[0].String(), rule.Array()[i].Array()[0].String(), "\n\n```\n"+rawlog+"\n```")
				talk(*teamsHook, SndMSG)
			}
		}
	}
}
