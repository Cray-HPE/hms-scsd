1.  [CASMHMS](index.html)
2.  [CASMHMS Home](CASMHMS-Home_119901124.html)
3.  [Design Documents](Design-Documents_127906417.html)

# <span id="title-text"> CASMHMS : SCSD Design </span>

Created by <span class="author"> Matt Kelly</span>, last modified by
<span class="editor"> Andrew Nieuwsma</span> on Jan 13, 2022

**Relevant JIRAs:**

<span class="jira-issue resolved" jira-key="CASMHMS-2805">
<a href="https://connect.us.cray.com/jira/browse/CASMHMS-2805?src=confmacro" class="jira-issue-key"><img src="https://connect.us.cray.com/jira/secure/viewavatar?size=xsmall&amp;avatarId=13315&amp;avatarType=issuetype" class="icon" />CASMHMS-2805</a>
- <span class="summary">set up NTP/rsyslog forwarding, ssh keys on River
Rosetta TORs</span> <span
class="aui-lozenge aui-lozenge-subtle aui-lozenge-success jira-macro-single-issue-export-pdf">Done</span>
</span>

<span class="jira-issue resolved" jira-key="CASM-1462">
<a href="https://connect.us.cray.com/jira/browse/CASM-1462?src=confmacro" class="jira-issue-key"><img src="https://connect.us.cray.com/jira/secure/viewavatar?size=xsmall&amp;avatarId=13307&amp;avatarType=issuetype" class="icon" />CASM-1462</a>
- <span class="summary">Scaling and Hardware System Configuration
Service</span> <span
class="aui-lozenge aui-lozenge-subtle aui-lozenge-success jira-macro-single-issue-export-pdf">Done</span>
</span>

  

Various BMC and controller parameters are set during discovery.   There
will be times, however, when certain parameters need to be set after
discovery (or perhaps before).  Parameters commonly needing to be set:

-   SSH keys
-   NTP server
-   Syslog server
-   BMC/Controller passwords

The need is for a tool to allow these sorts of parameters to be set (or
read) at any time, not just at discovery time.  A new tool is needed
that provides a CLI to do these operations.

## Options

There are two axes of options for implementation of this functionality:

1.  CLI-only vs. micro-service with a CLI
2.  Abstracted operations as listed above, or generic Redfish
    capability.

### CLI-only vs MicroService+CLI

Implementing this tool as a simple CLI would be fairly fast and easy. 
Less complication overall since there is no microservice required to run
it.  The downside is that customers would not have the ability to write
their own tool that do REST calls to set parameters.

Implementing a microservice with associated CLI offers customers
flexibility to write their own tools that can either use the Cray CLI or
directly use REST calls.  It also helps with things like AuthN/AuthZ
since the micro service would operate in the service mesh.  The
micro-service+Cray CLI is a solved problem already.  A stand-alone CLI
would have to deal with tokens separately, plus it would have to query
HSM for information from outside the service mesh.

### Abstracted Operations vs Generic Redfish Capability

Abstracting the operations makes the customer-facing interface simpler. 
For example, setting the NTP server on all controllers and BMCs is a
single command.  But, it lacks the flexibility to be able to do any
desired Redfish operation across multiple targets.

Allowing generic Redfish operations to be done will require the user to
know the Redfish endpoint URLs, and which pieces of hardware require
each URL.   It puts lots of burden on the user, but does provide
flexibility.

Another potential downside of allowing generic Redfish operations in
this tool is that it makes it VERY easy to screw up a system by going
"under the covers" and causing confusion between hardware states and
micro-service knowledge of state.

# SOLUTION

Given the pros and cons of each approach, it would seem the best option
would be:

-   Abstracted operations only (e.g. set-ntp, set-syslog, etc.)
-   Micro-service + Cray CLI implementation

## Micro Service (scsd)

The scsd service will present a REST API to facilitate parameter set
operations.  It will contact the HSM to verify targets as being correct
and in a valid HW state unless the "Force" flag is specified during a
transaction.  Once it has a list of targets, scsd will perform the
needed Redfish operations in parallel using TRS.  Any credentials needed
will be fetched from Vault.

The northbound REST API presents the endpoints as outlined below.

NOTE: in all POST operation payloads there is an optional "Force"
field.  If present, and set to 'true', then HSM will not be contacted;
the Redfish operations will be attempted without verifying they are
present or in a good state.   If the "Force" field is not present or is
present but set to 'false', HSM will be used.

The specified targets can be BMC or controller XNames, or HSM Group
IDs.  If BMCs and controllers are grouped in the HSM, the usage of this
tool becomes much easier since single targets can be used rather than
long lists.

### v1/bmc/dumpcfg  (POST)

Get  the Network Protocol parameters (NTP/syslog server, SSH keys) and 
boot order for the targets in the payload.  Note that all fields are
only applicable to Mountain controllers.  Trying to set them for river
BMCs will be ignored, getting them for river BMCs will return empty
strings.  JSON payload data for POST operations to set parameters can
omit parameters as desired.  So, for example, if only the NTP server
info is to be set, only the "NTPServer" key/value has to be present.

**Payloads:**

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
            "SSHConsoleKey",
            "BootOrder"
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
                  "SSHConsoleKey": "aaaabbbbcccc",
                  "BootOrder": ["Boot0",Boot1",Boot2",Boot3"]
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
                  "SSHConsoleKey": "aaaabbbbcccc",
                  "BootOrder": ["Boot0",Boot1",Boot2",Boot3"]
             }
        }
        ]
    }

### v1/bmc/loadcfg (POST)

Set Syslog, NTP server info or SSH key for a set of targets.

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
            "SSHConsoleKey": "aaaabbbbcccc",
            "BootOrder": ["Boot0","Boot1","Boot2","Boot3"]
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

  

### v1/bmc/cfg/{xname}  (GET/POST)

Same as v1/bmc/nwprotocol, but for a specific XName.   If no form data
is specified, all NW protocol data is returned for the target, otherwise
only the requested data is returned.

**Payloads:**

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
            "SSHConsoleKey": "aaaabbbbcccc",
            "BootOrder": ["Boot0","Boot1","Boot2","Boot3"]
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
            "SSHConsoleKey": "aaaabbbbcccc",     
            "BootOrder": ["Boot0","Boot1","Boot2","Boot3"]
        }
    }

    POST Response:

    {
        "StatusMsg": "OK"
    }

### /v1/bmc/discreetcreds (POST)

This endpoint is used for setting redfish credentials for BMCs and
controllers.  Note that this is different than SSH keys (only used on
controllers) – these credentials are for Redfish access, not SSH access
into a controller.   

This API allows the setting of different creds for each target within
one call.

Note that it is not possible to "fetch" creds via this API.  Only
setting them is allowed, for security reasons.

The payload for this API is amenable to setting different creds for
different targets all in one call.  To set a group of controllers'
creds, set up a group in HSM and use the group ID.

  

**Payload:**

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

  

### /v1/bmc/creds/{xname} (POST)

Same as /v1/bmc/discreetcreds, but for a single target.

**Payloads:**

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

  

### /v1/bmc/globalcreds (POST only)

This interface allows the caller to set all targets to the same
username/password.  

**Payloads:**

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
                "StatusCode": 405,
                "StatusMsg": "Only POST operations permitted"
            }
        ]
    }

  

### /v1/health (GET only)

Get the current health status of scsd.

**Payload:**

    {
        "TaskRunnerStatus": "Status String",
        "TaskRunnerMode": "Local"
    }

  

### /v1/readiness (GET only)

Get the current readiness status of scsd.  Note that there is no
payload; the return is one of:

-   201 (No Content) - ready
-   503 (Service Unavailable)

### /v1/liveness (GET only)

Get the current liveness state of scsd.   Note that there is no payload;
the return is one of:

-   405 (Not allowed)
-   404 (Not available, returned due to service not exiting)
-   201 (No Content) - alive

### /v1/version (GET only)

Get the current build version.  Will be v1.maj.min format.

    GET:


    {
        "Version": "v1.2.3"
    }

  

## CLI

The scsd CLI will be the standard Cray CLI based on the Swagger
specification for the service.  Thus, all of the usual syntaxes apply
for things like authentication, etc.   This style of CLI will make the
use of scsd consistent with all other Cray CLIs.

Example:   cray scsd syslog update --syslog sms-ncn-w001:514 x0c0s0n0b0

  

# OPEN QUESTIONS

-   Is this solution "too small"?  Should we be looking at a more
    generic service that can configure other things too?
-   If so, and given the rapidly approaching deadline, should we do the
    simpler "script" approach and make a more robust tool/service later?
-   Is the focus, even within the realm of HMS only, too narrow (only
    SSH keys, NTP, etc.)?
-   /v1/bmc/creds – how to make it possible to set different login/pw
    combinations for a large number of targets?

  

  

  

  

## Comments:

<table data-border="0" width="100%">
<colgroup>
<col style="width: 100%" />
</colgroup>
<tbody>
<tr class="odd">
<td><span id="comment-166370012"></span>
<p>The more I think about this, the more I'd like to see some updates to the REST interactions.</p>
<p>There are three generic entities you seem to be describing.  The first is a default template for all bmcs, the second is a group of BMCs that share a template, and the third is a set of individual bmcs. </p>
<p>This is a pretty common REST pattern and supports complex interactions.  It also allows for intuitive precedence.  Unless a change has been made to the individual bmc, fields from the template apply. </p>
<p>Additionally, this structure leaves the door open for future versioning of the templates and/or of the BMC objects themselves.</p>
<p><br />
</p>
<p>In my opinion, this also simplifies the CLI interactions.</p>
<p>cray scsd bmc update --syslog sms-ncn-w001:514 x0c0s0n0b0</p>
<p>cray scsd bmcgroup create -f my_xname_array.json partition_1_bmcgroup</p>
<p>cray scsd bmctemplate update --syslog sms-ncn-w001:514 default_partition_1</p>
<p>cray scsd bmc update --login root --password ******* x0c0s0n0b0</p>
<p><br />
</p>
<div class="smallfont" data-align="left" style="color: #666666; width: 98%; margin-bottom: 10px;">
<img src="images/icons/contenttypes/comment_16.png" width="16" height="16" /> Posted by alovelltr at Feb 04, 2020 09:11
</div></td>
</tr>
<tr class="even">
<td style="border-top: 1px dashed #666666"><span id="comment-166370162"></span>
<p>If we use templates, how do they get defined?</p>
<p>As to the grouping, we can use HSM groups as the target.</p>
<div class="smallfont" data-align="left" style="color: #666666; width: 98%; margin-bottom: 10px;">
<img src="images/icons/contenttypes/comment_16.png" width="16" height="16" /> Posted by mpkelly at Feb 04, 2020 11:38
</div></td>
</tr>
<tr class="odd">
<td style="border-top: 1px dashed #666666"><span id="comment-166371269"></span>
<p>I could see it go either way.  On the one hand, the template could simply be a portion of the group object and created/updated along with the list of nodes.  On the other, the group could contain a reference to the template object which is created with a POST.</p>
<p>cray scsd bmctemplate create -f bmc_template.json default_partition_1</p>
<p>I don't have a strong opinion on this one.  What do you think users would prefer?</p>
<div class="smallfont" data-align="left" style="color: #666666; width: 98%; margin-bottom: 10px;">
<img src="images/icons/contenttypes/comment_16.png" width="16" height="16" /> Posted by alovelltr at Feb 06, 2020 05:59
</div></td>
</tr>
</tbody>
</table>

Document generated by Confluence on Jan 14, 2022 07:17

[Atlassian](http://www.atlassian.com/)
