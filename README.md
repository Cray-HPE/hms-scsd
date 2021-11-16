# System Configuration Tool

Various BMC parameters are set during the system HW discovery process.   There will be times, however, when certain parameters need to be set outside of the discovery process.  Parameters commonly needing to be set:

* SSH keys
* NTP server
* Syslog server
* BMCs Redfish admin account passwords
* TLS certs for BMCs

**SCSD** is a service which allow these sorts of parameters to be set (or read) at any time.  SCSD can be accessed via its REST API or via the *cray* command-line tool.

## Operation

The SCSD service will present a REST API to facilitate parameter get/set operations.  It will typically contact the Hardware State Manager to verify targets as being correct and in a valid state.  Once it has a list of targets, SCSD will perform the needed Redfish operations in parallel.  Any credentials needed will be fetched from Vault secure storage.

*NOTE: in all POST operation payloads there is an optional "Force" field.  If present, and set to 'true', then the Hardware State Manager will not be utilized; the Redfish operations will be attempted without verifying they are present or in a good state.   If the "Force" field is not present or is present but set to 'false', target states will be verified, and any targets not in acceptable states will not be included in the operation.*

The specified targets can be BMC XNames, or Hardware State Manager Group IDs.  If BMCs are grouped in the Hardware State Manager, the usage of this tool becomes much easier since single targets can be used rather than long lists.

## Network Protocol Parameters

Network Protocol parameters set on BMCs include:

* SSH keys
* SSH Console keys
* NTP server 
* Syslog server

### v1/bmc/dumpcfg  (POST)

Get the current Network Protocol parameters and  boot order for the target BMCs in the payload.  Note that all fields are only applicable to Olympus BMCs.  Trying to set them for COTS BMCs will be ignored, and fetching them for COTS BMCs will return empty strings.  JSON payload data for POST operations to set parameters can omit parameters as desired.  So, for example, if only the NTP server info is to be set, only the "NTPServer" key/value has to be present.

**Payload**

```
POST (to fetch current info from targets):
 
{
    "Force": true,
    "Targets": [
        "x0c0s0b0",
        "x0c0s1b0"
    ],
    "Params": [
        "NTPServerInfo",
        "SyslogServerInfo",
        "SSHKey",
        "SSHConsoleKey"
    ]
}
 
 
Return data:
{
    "Targets": [
    {
        "StatusCode": 200,
        "StatusMsg": "OK",
        "Xname": "x0c0s0b0",
        "Params":
        {
              "NTPServerInfo":
              {
                  "NTPServers": "sms-ncn-w001",
                  "Port": 123,
                  "ProtocolEnabled": true
              },
              "SyslogServerInfo":
              {
                  "SyslogServers": "sms-ncn-w001",
                  "Port":514,
                  "ProtocolEnabled": true
              },
              "SSHKey": "xxxxyyyyzzzz",
              "SSHConsoleKey": "aaaabbbbcccc"
        }
    },
    {
        "StatusCode": 200,
        "StatusMsg": "OK",
        "Xname": "x0c0s0b0",
        "Params":
        {
              "NTPServerInfo":
              {
                  "NTPServers": "sms-ncn-w001",
                  "Port": 123,
                  "ProtocolEnabled": true
              },
              "SyslogServerInfo":
              {
                  "SyslogServers": "sms-ncn-w001",
                  "Port":514,
                  "ProtocolEnabled": true
              },
              "SSHKey": "xxxxyyyyzzzz",
              "SSHConsoleKey": "aaaabbbbcccc"
         }
    }
    ]
}
```

### /v1/bmc/loadcfg (POST)

Set the Syslog, NTP server info or SSH key on a set of target BMCs.

**Payloads:**

```
POST (to set params on targets):
 
{
    "Force": false,
    "Targets": [
        "x0c0s0b0",
        "x0c0s1b0"
    ],
    "Params":
    {
        "NTPServerInfo":
        {
            "NTPServers": "sms-ncn-w001",
            "Port": 123,
            "ProtocolEnabled": true
        },
        "SyslogServerInfo":
        {
            "SyslogServers": "sms-ncn-w001",
            "Port":514,
            "ProtocolEnabled": true
        },
        "SSHKey": "xxxxyyyyzzzz",
        "SSHConsoleKey": "aaaabbbbcccc"
    }
}
 
POST Response:
 
{
    "Targets": [
        {
            "Xname": "x0c0s0b0",
            "StatusCode": 200,
            "StatusMsg": "OK"
        },
        {
            "Xname": "x0c0s1b0",
            "StatusCode": 405,
            "StatusMsg": "Only GET operations permitted"
        }
    ]
}
```

### v1/bmc/cfg/{xname}  (GET/POST)

Same as *v1/bmc/nwprotocol*, but for a specific XName.   If no query parameters are specified in the URL, all NW protocol data is returned for the target, otherwise only the specified data is returned.

**Payloads:**

```
GET  /v1/bmc/cfg/x0c0s0b0n0?params=Force+NTPServerInfo+SyslogServerInfo+SSHKey+SSHConsoleKey
 
GET Response:
 
{
    "Force": true,
    "Params":
    {
        "NTPServerInfo":
        {
            "NTPServers": "sms-ncn-w001",
            "Port": 123,
            "ProtocolEnabled": true
        },
        "SyslogServerInfo":
        {
            "SyslogServers": "sms-ncn-w001",
            "Port":514,
            "ProtocolEnabled": true
        },
        "SSHKey": "xxxxyyyyzzzz",
        "SSHConsoleKey": "aaaabbbbcccc"
    }
}
 
POST /v1/bmc/nwp/cfg/x0c0s0b0n0:
 
{
    "Force": true,
    "Params":
    {
        "NTPServerInfo":
        {
            "NTPServers": "sms-ncn-w001",
            "Port": 123,
            "ProtocolEnabled": true
        },
        "SyslogServerInfo":
        {
            "SyslogServers": "sms-ncn-w001",
            "Port":514,
            "ProtocolEnabled": true
        },
        "SSHKey": "xxxxyyyyzzzz",
        "SSHConsoleKey": "aaaabbbbcccc"
    }
}
 
POST Response:
 
{
    "StatusMsg": "OK"
}
```


## BMC Credentials

BMC credentials are username/password pairs used for Redfish administrative accounts.  BMC Redfish access used by Shasta requires the use of the administrative account and thus Shasta SW is responsible for setting and using the correct account credentials for BMC Redfish access.

BMC admin account passwords are created by the admin and set on the BMCs using SCSD using the following APIs.


### /v1/bmc/discreetcreds (POST)

This endpoint is used for setting redfish credentials for BMCs.  Note that this is different than SSH keys (only used on Olympus BMCs) – these credentials are for Redfish access, not SSH access into a BMC.   

This API allows the setting of different creds for each target within one call.  See */v1/bmc/globalcreds* for a mechanism to set multiple BMCs' creds to the same value.

Note that it is not possible to "fetch" creds via this API.  Only setting them is allowed, for security reasons.

To set a group of BMCs' creds, set up a group in Hardware State Manager and use the group ID rather than BMC xname(s).

**Payloads:**

```
POST (to set creds on targets):
 
{
    "Force": false,
    "Targets": [
        {
            "Xname": "x0c0s0b0",
            "Creds": {
                "Username": "root",
                "Password": "admin-pw"
            }
        },
        {
            "Xname": "x0c0s1b0",
            "Creds": {
                "Username": "root",
                "Password": "admin-pw"
            }
        }
    ]
}
 
 
POST response:
{
    "Targets": [
        {
            "Xname": "x0c0s0b0",
            "StatusCode": 200,
            "StatusMsg": "OK"
        },
        {
            "Xname": "x0c0s1b0",
            "StatusCode": 405,
            "StatusMsg": "Only POST operations permitted"
        }
    ]
}
```

### /v1/bmc/creds/{xname} (POST)

Same as */v1/bmc/discreetcreds*, but for a single target.

**Payloads:**

```
POST:
 
{
    "Force": true,
    "Creds": {
        "Username": "root",
        "Password": "admin-pw"
    }
}
 
POST Response:
 
{
    "StatusMsg": "OK"
}
```

### /v1/bmc/creds/{xname} (POST)

Fetches the BMC creds of selected targets.  Targets are specified as
URL query parameters.

** Query Parameters **

```
GET /v1/bmc/creds?targets=list&type=type

list     Comma-separated list of BMC XNames.  If ommitted, 
         all BMCs are targeted.

type     Component type.  Can be one of:

         NodeBMC
         RouterBMC
         ChassisBMC

         If ommitted, all of these types are targeted.
```

### /v1/bmc/globalcreds (POST only)

This interface allows the caller to set the same access creds on all target BMCs' admin accounts.  

**Payloads:**

```
POST (to set the same creds on all targets):
 
{
    "Force": false,
    "Username": "user",
    "Password": "user-pw",
    "Targets": [
        "x0c0s0b0",
        "x0c0s1b0"
    ]
}
 
 
POST response:
 
{
    "Targets": [
        {
            "Xname": "x0c0s0b0",
            "StatusCode": 200,
            "StatusMsg": "OK"
        },
        {
            "Xname": "x0c0s1b0",
            "StatusCode": 200,
            "StatusMsg": "OK"
        }
    ]
}
```

## TLS Cert Management

To facilitate validated HTTPS communications to Redfish BMCs, TLS certs need to be created and placed onto the BMCs.  The following APIs provide the means to create, fetch, delete, and place TLS certs onto target BMCs.


### /v1/bmc/createcerts

This API creates TLS cert/key pairs for a set of BMCs.  It does not apply these certs to BMCs -- it only creates them and places them into Vault secure storage for later use.

TLS certs are created to cover "domains".  The most common domain is the cabinet/rack domain.  The admin creates a single TLS cert per cabinet/rack which can later be applied to any BMC in that cabinet.  Other domains are possible -- chassis, blade, and BMC; however the "lower" the domain, the more TLS certs must be created to cover the system.


**Payloads:**

```

POST request:

{
  "Domain": "Cabinet",
  "DomainIDs": [
    "x0",
    "x1"
  ]
}

POST response:

{
  "DomainIDs": [
    {
      "ID": "x0",
      "StatusCode": 200,
      "StatusMsg": "OK"
    },
    {
      "ID": "x1",
      "StatusCode": 200,
      "StatusMsg": "OK"
    }
  ]
}
```

### /v1/bmc/deletecerts

This API deletes previously created TLS cert/key pairs.  It does not remove these certs from BMCs -- it only deletes them from Vault secure storage.

**Payloads:**

```

POST request:

{
  "Domain": "Cabinet",
  "DomainIDs": [
    "x1"
  ]
}

POST response:

{
  "DomainIDs": [
    {
      "ID": "x1",
      "StatusCode": 200,
      "StatusMsg": "OK"
    }
  ]
}
```

### /v1/bmc/fetchcerts

This API fetches previously-created TLS cert/key pairs from Vault secure storage -- not the BMCs themselves -- and displays them.  The payload specifies the domain and the targets, which can be any BMCs within that domain.  For example, if a cabinet-level cert was generated, the simplest thing to do is fetch the cert for that cabinet using that cabinet's XName, since all BMCs in the cabinet have the same cert.

**Payload:**

```

POST request:

{
  "Domain": "Cabinet",
  "DomainIDs": [
    "x0"
  ]
}

POST response:

{
  "DomainIDs": [
    {
      "ID": "x0",
      "StatusCode": 200,
      "StatusMsg": "OK",
      "Cert": {
        "CertType": "PEM",
        "CertData": "-----BEGIN CERTIFICATE-----...-----END CERTIFICATE-----"
      }
    }
  ]
}
```

### /v1/bmc/setserts

This API applies previously-generated TLS certs onto BMCs.  The following example will set a cabinet-domain cert, previously generated, onto a single BMC.

**Payloads:**

```

POST request:

{
  "Force": false,
  "CertDomain": "Cabinet",
  "Targets": [
    "x0c0s0b0"
  ]
}

POST response:

{
  "Targets": [
    {
      "ID": "x0c0s0b0",
      "StatusCode": 200,
      "StatusMsg": "OK"
    }
  ]
}
```

### /v1/bmc/setcert/{xname}

Same as */v1/bmc/setserts*, but for a single BMC.  URL query parameters can include:

* Force -- Don't check with State Manager for target BMC state -- just do it.
* Domain -- Set cert domain.  Default is cabinet.

```

POST /v1/bmc/setcert/x0c0s0b0?Force=false,Domain=Cabinet

POST response is the HTTP status code.  200 == cert was successfully set.

```


## Service Health

Includes *health*, *liveness* and *readiness* APIs.


### /v1/health (GET only)

Get the current health status of SCSD.

**Payload:**

```
{
    "TaskRunnerStatus": "Status String",
    "TaskRunnerMode": "Local"
}
```

### /v1/readiness (GET only)

Get the current readiness status of SCSD.  Note that there is no payload; the return is one of:

* 201 (No Content) - ready
* 503 (Service Unavailable)


### /v1/liveness (GET only)

Get the current liveness state of SCSD.   Note that there is no payload; the return is one of:

* 405 (Not allowed)
* 404 (Not available, returned due to service not exiting)
* 201 (No Content) - alive


### /v1/version (GET only)

Get the current build version.  Will be v1.maj.min format.

**Payload:**

```
GET:
 
 
{
    "Version": "v1.2.3"
}
```

## CLI

The SCSD CLI will be the standard Cray CLI based on the Swagger specification for the service.  Thus, all of the usual syntaxes apply for things like authentication, etc.   This style of CLI will make the use of SCSD consistent with all other Cray CLIs.

Since the Cray CLI for SCSD is the same as all other CLIs, and is self-documenting, its usage will not be outlined in this document.

Quick Example:   cray scsd syslog update --syslog sms-ncn-w001:514 x0c0s0b0

This updates the syslog server info on BMC x0c0s0b0, having it now point to sms-ncn-w001 port 514.


## Use Cases

Following are various use cases for SCSD.   Many of them require a list of valid BMCs, which can be obtained via the following script:

```bash
#!/bin/bash

valid=""
invalid=""

for fff in `cray hsm inventory redfishEndpoints list --format json | jq '.RedfishEndpoints[] | select(.FQDN | contains("-rts") | not) | select(.DiscoveryInfo.LastDiscoveryStatus == "DiscoverOK") | select(.Enabled==true) | .ID' | sed 's/"//g'`; do
    echo "Pinging ${fff}..." ;
    curl -k https://${fff}/redfish/v1/ > /dev/null 2>&1
    if [[ $? == 0 ]]; then
        echo "${fff} PRESENT"
        valid="${valid},${fff}"
    else
        echo "${fff} NOT PRESENT"
        invalid="${invalid},${fff}"
    fi
done

valTargs=`echo ${valid} | sed 's/^,//' | sed 's/,$//'`
invalTargs=`echo ${invalid} | sed 's/^,//' | sed 's/,$//'`

echo " "
echo "VALID:"
echo $valTargs
echo " "
echo "INVALID:"
echo $invalTargs
echo " "

```

Output looks something like:

```
Pinging x0c0s0b0...
x0c0s0b0 PRESENT
Pinging x0c0s1b0...
x0c0s1b0 PRESENT
Pinging x0c0s2b0...
x0c0s2b0 NOT PRESENT

VALID:
x0c0s0b0,x0c0s1b0

INVALID:
x0c0s2b0
```

The VALID list of BMCs can be used in various payloads' "Target" elements.


### Set BMC Redfish Credentials: All BMCs Use the Same Creds

**1. Run the above script to determine the live BMCs in the system.**

**2. Use the output from the script to create a JSON file containing the BMC creds, placing the items in the VALID list into the "Targets" array, e.g.:**

```
bmc_creds_glb.json
 
{
  "Force": false,
  "Username": "root",
  "Password": "new.root.password"
  "Targets": [
    "x0c0s0b0", "x0c0s1b0"
  ]
}
```

**3. Run the cray cli to have SCSD apply the global creds to all target BMCs:**

```bash
# cray scsd bmc globalcreds create ./bmc_creds_glb.json
```

If the above cray cli command has any components that do not have the status of "OK", these must be retried until they work, or retries are exhausted and noted as failures.  Failed modules need to be taken out of the system until they are fixed.

### Set BMC Redfish Credentials: All BMCs Have Different Creds

**1. Get valid BMCs using the script above.**

**2. Use this list to create a JSON file containing the BMC creds, e.g.:**

```
bmc_creds_dsc.json
{
  "Force": true,
  "Targets": [
    {
      "Xname": "x0c0s0b0",
      "Creds": {
        "Username": "root",
        "Password": "pw-x0c0s0b0"
      }
    },
    {
      "Xname": "x0c0s0b1",
      "Creds": {
        "Username": "root",
        "Password": "pw-x0c0s0b1"
      }
    }
  ]
}
```

**3. Run the cray cli to have SCSD apply the discrete creds:**

```bash
# cray scsd bmc discreetcreds create ./bmc_creds_dsc.json
```

If the above cray cli command has any components that do not have the status of "OK", these must be retried until they work, or retries are exhausted and noted as failures.  Failed modules need to be taken out of the system until they are fixed.


### Network Parameters: Fetch Current Settings On BMCs

**1. Run the above script to determine the live BMCs in the system.**

**2. Create a JSON file containing the BMCs and parameters to fetch.**

```
bmc_nwp_fetch.json

{
  "Targets": [
    "x0c0s0b0","x0c0s1b0"
  ],
  "Params": [
    "NTPServerInfo"
  ]
}
```

**3. Execute the CLI to dump the current requested parameters.**

```bash
# cray scsd bmc dumpcfg create --json bmc_nwp_fetch.json
```

Output example:

```
{
  "Targets": [
    {
      "StatusCode": 0,
      "StatusMsg": "string",
      "Xname": "x0c0s0b0",
      "Params": {
        "NTPServerInfo": {
          "NTPServers": "sms-ncn-w001",
          "Port": 0,
          "ProtocolEnabled": true
        }
      }
    }
  ]
}
```

If the above cray cli command has any components that do not have the status of "OK", these must be retried until they work, or retries are exhausted and noted as failures.  Failed modules need to be taken out of the system until they are fixed.


### Network Parameters: Set Settings On BMCs

**1. Run the above script to determine the live BMCs in the system.**

**2. Create a JSON file containing the BMCs and parameters to set.**

```
bmc_nwp_set.json

{
  "Force": true,
  "Targets": [
    "x0c0s0b0"
  ],
  "Params": {
    "NTPServer": {
      "NTPServers": "sms-ncn-w001",
      "Port": 0,
      "ProtocolEnabled": true
    },
    "SyslogServer": {
      "SyslogServers": "sms-ncn-w001",
      "Port": 0,
      "ProtocolEnabled": true
    },
    "SSHKey": "xyzabc123...",
    "SSHConsoleKey": "xyzabc123..."
  }
}
```

**3. Execute the CLI to set the requested parameters onto target BMCs.**

```bash
# cray scsd bmc loadcfg create bmc_nwp_set.json
```

Output example:

```
{
  "Targets": [
    {
      "Xname": "x0c0s0b0",
      "StatusCode": 200,
      "StatusMsg": "OK"
    },
    {
      "Xname": "x0c0s1b0",
      "StatusCode": 200,
      "StatusMsg": "OK"
    }
  ]
}
```


### TLS Certs: Generate And Place TLS Certs On BMCs

**1. Use SCSD to Generate TLS Certs**

First create a JSON file containing all cabinet level cert creation info:

```
{
  "Domain": "Cabinet",
  "DomainIDs": [ "x0", "x1", "x2", "x3"]
}
 
Save this to 'cert_create.json'
```

```bash
# cray scsd bmc createcerts create --format json cert_create.json
{
  "DomainIDs": [
    {
      "ID": "x0",
      "StatusCode": 200,
      "StatusMsg": "OK"
    },
    {
      "ID": "x1",
      "StatusCode": 200,
      "StatusMsg": "OK"
    },
    {
      "ID": "x2",
      "StatusCode": 200,
      "StatusMsg": "OK"
    },
    {
      "ID": "x3",
      "StatusCode": 200,
      "StatusMsg": "OK"
    }
  ]
}
```

**2. Use SCSD To Apply TLS Certs To Target BMCs**

Eventually this step will include all BMCs.  For the near future (1.4), only Mountain BMCs are supported.

First create a JSON file specifying the endpoints:

```
{
  "Force": false,
  "CertDomain": "Cabinet",
  "Targets": [
    "x0c0s0b0","x0c0s1b0","x0c0s2b0", "x0c0s3b0"
  ]
}
```

Execute SCSD to set the certs on the target BMCs:

```
# cray scsd bmc setcerts create --format json cert_set.json
{
  "Targets": [
    {
      "ID": "x0c0s0b0",
      "StatusCode": 200,
      "StatusMsg": "OK"
    },
    {
      "ID": "x0c0s1b0",
      "StatusCode": 200,
      "StatusMsg": "OK"
    },
    {
      "ID": "x0c0s2b0",
      "StatusCode": 200,
      "StatusMsg": "OK"
    },
    {
      "ID": "x0c0s3b0",
      "StatusCode": 200,
      "StatusMsg": "OK"
    }
  ]
}
```

### TLS Certs: TLS Cert and CA Trust Bundle Rolling

At any point the TLS certs can be re-generated and replaced on Redfish BMCs.   The CA trust bundle can also be modified at any time.

When this is to be done, the following steps are needed:

1. Modify the CA trust bundle.  This is outside the scope of this document.  Please reference the Vault PKI docs in the reference list below.
2. Once the CA trust bundle is modified, each service will automatically pick up the new CA bundle data.  There is no manual step.
3. Re-generate the TLS cabinet-level certs as in step 1 above.
Place the TLS certs onto the Redfish BMCs as in step 2 above.

### SCSD CT Testing

This repository builds and publishes hms-scsd-ct-test RPMs along with the service itself containing tests that verify SCSD on the
NCNs of live Shasta systems. The tests require the hms-ct-test-base RPM to also be installed on the NCNs in order to execute.
The version of the test RPM installed on the NCNs should always match the version of SCSD deployed on the system.
