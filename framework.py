#!/usr/bin/env python
"""
Nicholas' Example API code for interacting with Alienvault API.

This is just Example code written by NMA.IO.

There isn't really much you can do with the API just yet, so
this will be a work in progress.
Grab your API key here:
    https://www.alienvault.com/documentation/usm-anywhere/api/alienvault-api.htm?cshid=1182

That said, you could use this to write an alerter bot.

* Note, this isn't an SDK, just some example code to get  people started...

"""
# import json  # uncomment if you want pretty printing.
import base64
import requests

URL = "alienvault.cloud/api/2.0"
HOST = ""  # put your subdomain here.


def Auth(apiuser, apikey):
    """Our Authentication Code.

    :params apiuser
    :params apikey

    :returns oauth_token
    """
    headers = {"Authorization": "Basic {}".format(base64.b64encode("%s:%s" % (apiuser, apikey)))}

    r = requests.post("https://{}.{}/oauth/token?grant_type=client_credentials".format(HOST, URL), headers=headers)
    if r.status_code is 200:
        return r.json()["access_token"]
    else:
        print("Authentication failed. Check username/Password")
        exit(1)


def Alarms(token):
    """Pull Alarms from the API Console."""
    headers = {"Authorization": "Bearer {}".format(token)}
    r = requests.get("https://{}.{}/alarms/?page=1&size=20&suppressed=false&status=open".format(HOST, URL), headers=headers)
    if r.status_code is not 200:
        print("Something went wrong. \n{}".format(r.content))
        exit(1)
    else: return (r.json())


def Events():
    """Nothing yet."""
    pass

if __name__ == "__main__":
    print("Simple API Integration with Alienvault USM Anywhere - 2018 NMA.IO")
    token = Auth("username", "password")
    jdata = Alarms(token)
    for item in jdata["_embedded"]["alarms"]:
        # print(json.dumps(item, indent=2)) # uncomment if you want a pretty version of the whole block.
        print(item["rule_method"] + ": " + " ".join(item["alarm_sources"]))
