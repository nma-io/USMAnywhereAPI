# Powershell Example
# Works under powershell core, should work on all OS's.
# NMA.IO 2018


$url="alienvault.cloud/api/2.0"
$domain=""  # put your subdomain here.
$user="" # Your application ID
$passwd="" # Your Secret Key

function Authenticate {
    $encodedCreds=[System.Convert]::ToBase64String([System.Text.Encoding]::ASCII.GetBytes($user+":"+$passwd))
    $req="https://" + $domain + "." + $url + "/oauth/token?grant_type=client_credentials"
    Invoke-WebRequest $req -Headers @{"Authorization"="Basic $encodedCreds"} -Method POST | ConvertFrom-Json | select-object -ExpandProperty access_token
}


function Alarms {
    $alarmreq="https://" + $domain + "." + $url + "/alarms/?page=1&size=20&suppressed=false&status=open"
    Invoke-WebRequest $alarmreq -Headers @{"Authorization"="Bearer " + $args[0]} -Method GET | ConvertFrom-Json | ForEach-Object _embedded | ForEach-Object alarms | Select-Object rule_method, alarm_sources
}

$token=Authenticate
$alarms=Alarms $token

echo $Alarms
