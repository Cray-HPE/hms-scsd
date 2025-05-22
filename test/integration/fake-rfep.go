// MIT License
// 
// (C) Copyright [2020-2021] Hewlett Packard Enterprise Development LP
// 
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package main

import (
	"encoding/json"
	"net/http"
	"fmt"
	"log"
	"io/ioutil"
	"os"
	"strings"
	"strconv"
	"time"
	"context"
)

// Disclaimer: this code is relatively awful, and does a huge amount of stuff.
// Certainly not production-quality; but, it's only used for testing.
//
// This is an app that exposes a fake collection Redfish endpoints.  This
// allows the functional testing of SCSD's API.  Lots of environment variables
// are used to steer the app's behavior -- how to display various URIs, 
// data payloads, etc., as follows:
//
// BMCURL  Determines how /redfish/v1/Managers and its members look.  Typically
//         will be set to "BMC" or "1".
//
// NWPTYPE Determines which NetworkProtocol data items will be in play.  Values
//         are:  NONE   Don't use any NWP data
//               NTP    Only use NTP data
//               ???    Any other value == use ALL NWP data.
//
// VENDOR  'cray' or 'hpe'.  Determines the behavior of Certificate Service.
//
// BMCPORT Determines the port used in the host:port of the app instance.
//         Defaults to 20000 e.g. http://${X_S0_HOST}:2000/redfish/v1/...
//
// XNAME   Determines which BMC XName the app is pretending to be.  Defaults is
//         ${X_S0_HOST}.
//
// BADACCT Used for debugging, to force AccountService data to be incorrect.
//
// HTTPS   Not set: use http.  
//         1: use https, don't replace certs.  
//         2: use https, replace certs and restart server



//NWP data struct

type SyslogData struct {
	ProtocolEnabled bool   `json:"ProtocolEnabled,omitempty"`
	SyslogServers []string `json:"SyslogServers,omitempty"`
	Transport string       `json:"Transport,omitempty"`
	Port int               `json:"Port,omitempty"`
}
type SSHAdminData struct {
	AuthorizedKeys string `json:"AuthorizedKeys"`
}

type OemData struct {
	Syslog SyslogData `json:"Syslog"`
	SSHAdmin SSHAdminData   `json:"SSHAdmin"`
	SSHConsole SSHAdminData `json:"SSHConsole"`
}
type NTPData struct {
	NTPServers []string  `json:"NTPServers,omitempty"`
	ProtocolEnabled bool `json:"ProtocolEnabled,omitempty"`
	Port int             `json:"Port,omitempty"`
}

type RedfishNWProtocol struct {
	valid bool      //not marshallable
	Oem *OemData   `json:"Oem,omitempty"`
	NTP *NTPData   `json:"NTP,omitempty"`
}

type rfLockouts struct {
	AccountLockoutThreshold int `json:"AccountLockoutThreshold"`
}

type rfManagerMember struct {
	ID string `json:"@odata.id"`
}

type rfManagers struct {
	Members []rfManagerMember
	Etag    string `json:"@odata.etag,omitempty"`
}

type rfManagersBMCNWP struct {
	ID string `json:"@odata.id,omitempty"`
}

type rfManagersBMC struct {
	NetworkProtocol rfManagersBMCNWP
	//HPE stuff
	Oem *rfManagersOem `json:"Oem,omitempty"`
}

type rfManagersOem struct {
	HPE rfManagersOemHPE `json:"HPE"`
}

type rfManagersOemHPE struct {
	Links rfManagersOemHPELinks `json:"Links"`
}

type rfManagersOemHPELinks struct {
	SecurityService rfManagersOemHPELinksSecurityService `json:"SecurityService"`
}

type rfManagersOemHPELinksSecurityService struct {
	ID string `json:"@odata.id"`
}

// Certificate Service Root: Actions
type CertificateServiceActionsReplaceCert struct {
	Target string `json:"target"`
}

type CertificateServiceActions struct {
	ReplaceCert CertificateServiceActionsReplaceCert
}

// Certificate Service Root: CertificateLocations

type CertificateServiceLocations struct {
	ID string `json:"@odata.id"`
}

// Certificate Service: Root

type CertificateService struct {
	Etag string `json:"@odata.etag,omitempty"`
	Actions CertificateServiceActions
	CertificateLocations CertificateServiceLocations
}

// Certificate Service: CertificateLocations

type CertificateLocations struct {
	Links CertificateLocationLinks
}

type CertificateLocationLinks struct {
	Certificates []CertificateLocationIDs
}

type CertificateLocationIDs struct {
	ID string `json:"@odata.id"`
}

// Certificate Service: CertificateService.ReplaceCertificate payload

type CertificatePayload struct {
	CertificateString string `json:"CertificateString"`
	CertificateType   string `json:"CertificateType,omitempty"`
	CertificateURI    *CertificateURI
}

type CertificateURI struct {
	Uri string `json:"@odata.id"`
}


type httpStuff struct {
    stuff string
}

//Canned NetworkProtocol payloads

var nwpData = RedfishNWProtocol{valid: true,
                                Oem: &OemData{Syslog: SyslogData{ProtocolEnabled: true,
                                                                 SyslogServers: []string{"10.11.12.13","110.111.112.113",},
                                                                 Transport: "udp",
                                                                 Port: 123,
                                                                },
                                              SSHAdmin: SSHAdminData{"aabbccdd"},
                                              SSHConsole: SSHAdminData{"eeffgghh"},
                                                 },
                                   NTP: &NTPData{NTPServers: []string{"120.121.122.123","130.131.132.133"},
                                                 ProtocolEnabled: true,
                                                 Port: 789,
                                                },
                                   }

var nwpDataNo = RedfishNWProtocol{valid: true,}

var nwpDataNTPOnly = RedfishNWProtocol{valid:true,
                                       NTP: &NTPData{NTPServers: []string{"120.121.122.123","130.131.132.133"},
                                                     ProtocolEnabled: true,
                                                     Port: 789,
                                                    },
                                      }

type acctData struct {
	Enabled      bool   `json:"Enabled"`
	Id           string `json:"Id"`
	Name         string `json:"Name"`
	UserName     string `json:"UserName"`
	RoleId       string `json:"RoleId"`
	Description  string `json:"Description"`
	Password    *string `json:"Password"`
	Etag        *string `json:"@odata.etag"`
}

// Canned Account data

var invalidAcct = acctData{Enabled: false,
                           Id: "Z",
                           Name: "User Account",
                           UserName: "anonymous",
                           RoleId: "NoAccess",
                           Description: "User Account Description",
}
var validAcctNoEtag = acctData{Enabled: true,
                               Id: "A",
                               Name: "User Account",
                               UserName: "root",
                               RoleId: "NoAccess",
                               Description: "root User Account",
}

var wetag string = `W/"12345678"`
var etag string = `"aabbccdd"`

var validAcctWithEtag = acctData{Enabled: true,
                                 Id: "B",
                                 Name: "User Account",
                                 UserName: "root",
                                 RoleId: "NoAccess",
                                 Description: "root User Account",
                                 Etag: &wetag,
                                 //Password: null,
}

var noAcctRdat = `{
   "Enabled" : false,
   "Links" : {
      "Role" : {
         "@odata.id" : "/redfish/v1/AccountService/Roles/NoAccess"
      }
   },
   "Id" : "1",
   "Name" : "User Account",
   "@odata.context" : "/redfish/v1/$metadata#ManagerAccount.ManagerAccount",
   "@odata.id" : "/redfish/v1/AccountService/Accounts/1",
   "UserName" : "anonymous",
   "@odata.type" : "#ManagerAccount.v1_1_1.ManagerAccount",
   "RoleId" : "NoAccess",
   "Description" : "User Account",
   "Password" : null
   }`

var validAcctRdat = `{
   "Enabled" : false,
   "Links" : {
      "Role" : {
         "@odata.id" : "/redfish/v1/AccountService/Roles/NoAccess"
      }
   },
   "Id" : "1",
   "Name" : "User Account",
   "@odata.context" : "/redfish/v1/$metadata#ManagerAccount.ManagerAccount",
   "@odata.id" : "/redfish/v1/AccountService/Accounts/1",
   "UserName" : "root",
   "@odata.type" : "#ManagerAccount.v1_1_1.ManagerAccount",
   "RoleId" : "NoAccess",
   "Description" : "User Account",
   "Password" : null
   }`

//CertificateService canned payloads

var certSvcCrayPld = `{
  "@odata.context": "/redfish/v1/$metadata",
  "@odata.etag": "W/\"1594406384\"",
  "@odata.id": "/redfish/v1/CertificateService",
  "Actions": {
    "#CertificateService.ReplaceCertificate": {
      "@Redfish.ActionInfo": "/redfish/v1/CertificateService/ReplaceCertificateActionInfo",
      "target": "/redfish/v1/CertificateService/Actions/CertificateService.ReplaceCertificate"
    }
  },
  "CertificateLocations": {
    "@odata.id": "/redfish/v1/CertificateService/CertificateLocations"
  },
  "Id": "CertificateService",
  "Name": "Certificate Service"
}`


// Variables that steer data generation

var burl = "BMC"
var nwpType = ""
var isHPE = true
var ishttps = false
var replaceCert = false
var tlsCertFile = "/tmp/server.crt"
var tlsKeyFile = "/tmp/server.key"
var port = ":20000"

var httpsrv *http.Server


func printReqHdrs(fname string, r *http.Request) {
	for k,v := range(r.Header) {
		log.Printf("%s():  URL: '%s', HDR: '%s'/'%s'",fname,r.URL.Path,k,v)
	}
}


// RF CertificateService

func (p *httpStuff) certificateService(w http.ResponseWriter, r *http.Request) {
	pld := certSvcCrayPld

	if (r.Method != "GET") {
        fmt.Printf("ERROR: request is not a GET.\n")
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

	printReqHdrs("certificateService",r)
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(pld))
}


func (p *httpStuff) certificateLocations(w http.ResponseWriter, r *http.Request) {
	pld := `{
  "@odata.context": "/redfish/v1/$metadata#CertificateLocations.CertificateLocations",
  "@odata.id": "/redfish/v1/CertificateService/CertificateLocations",
  "@odata.type": "#CertificateLocations.v1_0_1.CertificateLocations",
  "Id": "CertificateLocations",
  "Name": "Certificate Locations",
  "Links": {
    "Certificates": [
      {
        "@odata.id": "/redfish/v1/Managers/BMC/NetworkProtocol/HTTPS/Certificates/1"
      }
    ]
  }
}`

	if (r.Method != "GET") {
        fmt.Printf("ERROR: request is not a GET.\n")
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

	printReqHdrs("certificateLocations",r)
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(pld))
}

func (p *httpStuff) certificateReplace(w http.ResponseWriter, r *http.Request) {
	if (r.Method != "POST") {
        fmt.Printf("ERROR: request is not a POST.\n")
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

	printReqHdrs("certificateReplace",r)
	var jdata CertificatePayload
	body,err := ioutil.ReadAll(r.Body)
	if (err != nil) {
		fmt.Println("ERROR reading req body:",err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("Cert replace payload: '%s'",string(body))

	err = json.Unmarshal(body,&jdata)
	if (err != nil) {
		fmt.Println("ERROR unmarshalling data:",err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//actually try to replace the cert and restart the HTTP server if
	//the flag says to do so.

	if (replaceCert && ishttps) {
		//Disassemble cert
		tlsCertFile = "/tmp/newserver.crt"
		tlsKeyFile = "/tmp/newserver.key"
		derr := dumpCertInfo(jdata.CertificateString)
		if (derr != nil) {
			log.Printf("ERROR dumping cert info: %v",derr)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		time.AfterFunc((2 * time.Second),restartServer)
	}

	//TODO: any JSON payload?
	w.WriteHeader(http.StatusOK)
}

func (p *httpStuff) Chassis(w http.ResponseWriter, r *http.Request) {
	crayPld := `{
  "@odata.context": "/redfish/v1/$metadata#ChassisCollection.ChassisCollection",
  "@odata.etag": "W/\"1592448856\"",
  "@odata.id": "/redfish/v1/Chassis",
  "@odata.type": "#ChassisCollection.ChassisCollection",
  "Description": "The Collection for Chassis",
  "Members": [
    {
      "@odata.id": "/redfish/v1/Chassis/Enclosure"
    },
    {
      "@odata.id": "/redfish/v1/Chassis/Node0"
    },
    {
      "@odata.id": "/redfish/v1/Chassis/Node1"
    }
  ],
  "Members@odata.count": 3,
  "Name": "Chassis Collection"
}`

	hpePld := `{
  "@odata.context": "/redfish/v1/$metadata#ChassisCollection.ChassisCollection",
  "@odata.etag": "W/\"AA6D42B0\"",
  "@odata.id": "/redfish/v1/Chassis/",
  "@odata.type": "#ChassisCollection.ChassisCollection",
  "Description": "Computer System Chassis View",
  "Members": [
    {
      "@odata.id": "/redfish/v1/Chassis/1"
    }
  ],
  "Members@odata.count": 1,
  "Name": "Computer System Chassis"
}`

	var pld string

	if (isHPE) {
		pld = hpePld
	} else {
		pld = crayPld
	}

	if (r.Method != "GET") {
        fmt.Printf("ERROR: request is not a GET.\n")
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

	printReqHdrs("Chassis",r)
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(pld))
}

func restartServer() {
	log.Printf("Shutting down HTTP server.")
	httpsrv.Shutdown(context.TODO())
	time.Sleep(2 * time.Second)
	log.Printf("Restarting HTTP server.")
	httpsrv = startHTTPServer()
	log.Printf("HTTP server running.")
}

//Chassis, pretends to be a mountain endpoint

func (p *httpStuff) chassisEnclosure(w http.ResponseWriter, r *http.Request) {
	printReqHdrs("chassisEnclosure",r)
	pld := `{"Architecture":"This_Is_Mountain"}`
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(pld))
}

//Chassis, pretends to be a river intel endpoint

func (p *httpStuff) chassisRackmount(w http.ResponseWriter, r *http.Request) {
	printReqHdrs("chassisRackmount",r)
	pld := `{"Architecture":"This_Is_River_Intel"}`
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(pld))
}

//Chassis, pretends to be a river GB endpoint

func (p *httpStuff) chassisSelf(w http.ResponseWriter, r *http.Request) {
	printReqHdrs("chassisSelf",r)
	pld := `{"Architecture":"This_Is_River_Gigabyte"}`
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(pld))
}

// Redfish root

func (p *httpStuff) rfroot(w http.ResponseWriter, r *http.Request) {
	//OK this is hokey.  Since we're using the simple HTTP hookups,
	//if we get a bad URL, we somehow respond to the last handler
	//func, which is for /redfish/v1/.  THEREFORE, if the URL we
	//see in this func isn't the service root URL, we should return
	//404 since we're not really supposed to be responding.

	if ((r.URL.Path != "/redfish/v1/") && (r.URL.Path != "/redfish/v1")) {
		fmt.Printf("Responding to a URL we don't have, forcing 404.\n")
		w.WriteHeader(http.StatusNotFound)
		return
	}

    if (r.Method != "GET") {
        fmt.Printf("ERROR: request is not a GET.\n")
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

	printReqHdrs("rfroot",r)

	rdat := `{
   "@odata.id" : "/redfish/v1/",
   "RedfishVersion" : "1.2.0",
   "EventService" : {
      "@odata.id" : "/redfish/v1/EventService"
   },
   "AccountService" : {
      "@odata.id" : "/redfish/v1/AccountService"
   } }`
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(rdat))
}

// Convenience func to copy NWP data.

func dc(dest,src *RedfishNWProtocol) {
	dest.valid = src.valid
	if ((dest.Oem == nil) && (src.Oem != nil)) {
		dest.Oem = &OemData{}
	}
	if (src.Oem != nil) {
		if (src.Oem.Syslog.ProtocolEnabled) {
			dest.Oem.Syslog = src.Oem.Syslog
		}
		if (src.Oem.SSHAdmin.AuthorizedKeys != "") {
			dest.Oem.SSHAdmin.AuthorizedKeys = src.Oem.SSHAdmin.AuthorizedKeys
		}
		if (src.Oem.SSHConsole.AuthorizedKeys != "") {
			dest.Oem.SSHConsole.AuthorizedKeys = src.Oem.SSHConsole.AuthorizedKeys
		}
	}
	if ((dest.NTP == nil) && (src.NTP != nil)) {
		dest.NTP = &NTPData{}
	}
	if (src.NTP != nil) {
		if (len(src.NTP.NTPServers) > 0) {
			dest.NTP.NTPServers = src.NTP.NTPServers
		}
		dest.NTP.ProtocolEnabled = src.NTP.ProtocolEnabled
		dest.NTP.Port = src.NTP.Port
	}
}

func (p *httpStuff) acctService(w http.ResponseWriter, r *http.Request) {
    if (r.Method == "GET") {
        rdat := `{
           "@odata.context" : "/redfish/v1/$metadata#AccountService.AccountService",
           "@odata.id" : "/redfish/v1/AccountService",
           "Roles" : {
              "@odata.id" : "/redfish/v1/AccountService/Roles"
           },
           "MinPasswordLength" : 1,
           "ServiceEnabled" : true,
           "@odata.type" : "#AccountService.v1_2_2.AccountService",
           "Id" : "AccountService",
           "Status" : {
              "Health" : "OK",
              "State" : "Enabled",
              "HealthRollup" : "OK"
           },
           "Description" : "BMC User Accounts",
           "MaxPasswordLength" : 20,
           "Accounts" : {
              "@odata.id" : "/redfish/v1/AccountService/Accounts"
           },
           "Name" : "Account Service"
        }`

	    w.Header().Set("Content-Type","application/json")
	    w.WriteHeader(http.StatusOK)
	    w.Write([]byte(rdat))
	} else if (r.Method == "PATCH") {
		var jdata rfLockouts
		body,err := ioutil.ReadAll(r.Body)
		if (err != nil) {
			fmt.Println("ERROR reading req body:",err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(body,&jdata)
		if (err != nil) {
			fmt.Println("ERROR unmarshalling data:",err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Setting accounts lockout threshold for %s %s to: %d",
			r.URL.Host,r.URL.Path,jdata.AccountLockoutThreshold)
        w.WriteHeader(http.StatusNoContent)
	} else {
		log.Printf("Invalid method %s, must be GET or PATCH.",r.Method)
        w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (p *httpStuff) acctAccounts(w http.ResponseWriter, r *http.Request) {
	var rdat string
	rdat1 := `{
   "Description" : "BMC User Accounts",
   "Name" : "Accounts Collection",
   "@odata.type" : "#ManagerAccountCollection.ManagerAccountCollection",
   "@odata.context" : "/redfish/v1/$metadata#ManagerAccountCollection.ManagerAccountCollection",
   "Members@odata.count" : 2,`
	rdat2 := `"Members" : [
      {
         "@odata.id" : "/redfish/v1/AccountService/Accounts/1"
      }
   ],`
	rdat3 := `"Members" : [
      {
         "@odata.id" : "/redfish/v1/AccountService/Accounts/1"
      },
      {
         "@odata.id" : "/redfish/v1/AccountService/Accounts/2"
      }
   ],`
	rdat4 := `"Members" : [
      {
         "@odata.id" : "/redfish/v1/AccountService/Accounts/1"
      },
      {
         "@odata.id" : "/redfish/v1/AccountService/Accounts/2"
      },
      {
         "@odata.id" : "/redfish/v1/AccountService/Accounts/3"
      }
   ],`
	rdat5 := `"Members" : [
      {
         "@odata.id" : "/redfish/v1/AccountService/Accounts/1"
      },
      {
         "@odata.id" : "/redfish/v1/AccountService/Accounts/2"
      },
      {
         "@odata.id" : "/redfish/v1/AccountService/Accounts/4"
      }
   ],`

	rdat10 := `"@odata.id" : "/redfish/v1/AccountService/Accounts" }`

	envstr := os.Getenv("NACCTS")
	if (envstr == "1") {
		rdat = rdat1+rdat2+rdat10
	} else if (envstr == "2") {
		rdat = rdat1+rdat3+rdat10
	} else if (envstr == "3") {
		rdat = rdat1+rdat4+rdat10
	} else if (envstr == "4") {
		rdat = rdat1+rdat5+rdat10
	} else {
		rdat = rdat1+rdat2+rdat10
	}

	printReqHdrs("acctAccounts",r)
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(rdat))
}

func doAccount(num int, w http.ResponseWriter, r *http.Request) {
	ord := 1
	etagOrd := -1
	var ap *acctData
	var vacp acctData
	var envstr string

	envstr = os.Getenv("BADACCT")
	if (envstr != "") {
		rval,_ := strconv.Atoi(envstr)
		log.Printf("Intentional bad return value: %d",rval)
		w.WriteHeader(rval)
		return
	}

	for k,v := range(r.Header) {
		log.Printf("doAccount():  acct: %d, URL: '%s', HDR: '%s'/'%s'",num,r.URL.Path,k,v)
	}

	//Special case: If this is account 3, and NACCTS is 4, we have to return
	//a 404, since we want to emulate 3 accounts labeled 1, 2, and 4.

	envstr = os.Getenv("NACCTS")
	if ((envstr == "4") && (num == 3)) {
		log.Printf("Intentional missing account '3'.")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	envstr = os.Getenv("GOODACCT")
	if (envstr != "") {
		ord,_ = strconv.Atoi(envstr)
		if (ord > 3) {
			ord = 3
		}
	}
	envstr = os.Getenv("ETAGACCT")
	if (envstr != "") {
		etagOrd,_ = strconv.Atoi(envstr)
		if (etagOrd > 3) {
			etagOrd = 3
		}
	}

	if (ord == num) {
		if (etagOrd == num) {
			vacp = validAcctWithEtag
			if ((num & 1) != 0) {	//if acct number is odd, use string etag
				vacp.Etag = &etag
			}
			ap = &vacp
		} else {
			ap = &validAcctNoEtag
		}
	} else {
		ap = &invalidAcct
	}

	if (r.Method == "GET") {
		ba,err := json.Marshal(ap)
		if (err != nil) {
			fmt.Println("ERROR marshalling data:",err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type","application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(ba)
	} else if (r.Method == "PATCH") {
		rsp_p1 := `{"@odata.type": "#ManagerAccount.v1_1_3.ManagerAccount", "Id": "`
		rsp_p2 := `", "Name": "User Account", "Description": "User Account", "Enabled": true, "Password": null, "UserName": "Administrator", "RoleId": "Administrator", "Locked": false, "Links": { "Role": { "@odata.id": "/redfish/v1/AccountService/Roles/Administrator" } }, "@odata.context": "/redfish/v1/$metadata#ManagerAccount.ManagerAccount", "@odata.id": "/redfish/v1/AccountService/Accounts/`
		rsp_p3 := `"}`

		var jdata acctData
		body,err := ioutil.ReadAll(r.Body)
		if (err != nil) {
			fmt.Println("ERROR reading req body:",err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(body,&jdata)
		if (err != nil) {
			fmt.Println("ERROR unmarshalling data:",err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ((jdata.Password != nil) && (*jdata.Password != "")) {
			fmt.Printf("Changed password for '%s' to '%s'\n",
				r.URL.Path,*jdata.Password)
			//tstr := *jdata.Password
			//ap.Password = &tstr
		}
		w.WriteHeader(http.StatusOK)
		rspRaw := fmt.Sprintf("%s%d%s%d%s",rsp_p1,num,rsp_p2,num,rsp_p3)
		w.Write([]byte(rspRaw))
	} else {
		fmt.Printf("Bad req method: %s\n",r.Method)
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (p *httpStuff) targAccount1(w http.ResponseWriter, r *http.Request) {
	doAccount(1,w,r)
}

func (p *httpStuff) targAccount2(w http.ResponseWriter, r *http.Request) {
	doAccount(2,w,r)
}

func (p *httpStuff) targAccount3(w http.ResponseWriter, r *http.Request) {
	doAccount(3,w,r)
}

func (p *httpStuff) targAccount4(w http.ResponseWriter, r *http.Request) {
	doAccount(4,w,r)
}


func (p *httpStuff) Managers(w http.ResponseWriter, r *http.Request) {
	if (r.Method != "GET") {
        fmt.Printf("ERROR: request is not a GET.\n")
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

	printReqHdrs("Managers",r)
	var jdata rfManagers

	jdata.Members = make([]rfManagerMember,1)
	jdata.Members[0].ID = "/redfish/v1/Managers/"+burl
	envstr := os.Getenv("ETAGACCT")
	if (envstr != "") {
		jdata.Etag = wetag
	}
	ba,err := json.Marshal(&jdata)
	if (err != nil) {
		fmt.Println("ERROR marshalling data:",err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

func (p *httpStuff) ManagersBMC(w http.ResponseWriter, r *http.Request) {
	if (r.Method != "GET") {
        fmt.Printf("ERROR: request is not a GET.\n")
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
	printReqHdrs("ManagersBMC",r)
	var jdata rfManagersBMC

	jdata.NetworkProtocol.ID = "/redfish/v1/Managers/"+burl+"/NetworkProtocol"

	if (isHPE) {
		jdata.Oem = &rfManagersOem{HPE: rfManagersOemHPE{Links: rfManagersOemHPELinks{SecurityService: rfManagersOemHPELinksSecurityService{ID: "/redfish/v1/Managers/1/SecurityService",},},},}
	}

	ba,err := json.Marshal(&jdata)
	if (err != nil) {
		fmt.Println("ERROR marshalling data:",err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

func (p *httpStuff) hpeSecurityService(w http.ResponseWriter, r *http.Request) {
	if (r.Method != "GET") {
        fmt.Printf("ERROR: request is not a GET.\n")
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

	printReqHdrs("hpeSecurityService",r)
	pld := `{"Links":{"HttpsCert":{"@odata.id":"/redfish/v1/Managers/1/SecurityService/HttpsCert"}}}`
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(pld))
}

func (p *httpStuff) hpeSecurityServiceHttpsCert(w http.ResponseWriter, r *http.Request) {
	if (r.Method != "GET") {
        fmt.Printf("ERROR: request is not a GET.\n")
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
	printReqHdrs("hpeSecurityServiceHttpsCert",r)

	pld := `{
  "Actions": {
    "#HpeHttpsCert.GenerateCSR": {
      "target": "/redfish/v1/Managers/1/SecurityService/HttpsCert/Actions/HpeHttpsCert.GenerateCSR"
    },
    "#HpeHttpsCert.ImportCertificate": {
      "target": "/redfish/v1/Managers/1/SecurityService/HttpsCert/Actions/HpeHttpsCert.ImportCertificate"
    }
  }}`
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(pld))
}

func (p *httpStuff) hpeImportCert(w http.ResponseWriter, r *http.Request) {
	if (r.Method != "POST") {
        fmt.Printf("ERROR: request is not a POST.\n")
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

	printReqHdrs("hpeImportCert",r)
	var jdata CertificatePayload
	body,err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body,&jdata)
	if (err != nil) {
		log.Printf("ERROR unmarshalling cert data: %v",err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("POST HPE Cert: '%s'",jdata.CertificateString)
	w.WriteHeader(http.StatusOK)
}

func (p *httpStuff) nwp_rcv(w http.ResponseWriter, r *http.Request) {
	printReqHdrs("nwp_rcv",r)
    if (r.Method == "PATCH") {
		var jdata RedfishNWProtocol
		body,err := ioutil.ReadAll(r.Body)
		err = json.Unmarshal(body,&jdata)
		if (err != nil) {
			fmt.Println("ERROR unmarshalling data:",err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		fmt.Printf("#########################################\n")
		fmt.Printf("Received NWP data:\n")
		if (jdata.Oem != nil) {
			fmt.Println("Oem:",*jdata.Oem)
		}
		if (jdata.NTP != nil) {
			fmt.Println("NTP:",*jdata.NTP)
		}
		fmt.Printf("#########################################\n")

		dc(&nwpData,&jdata)
		if (jdata.Oem != nil) {
			log.Println("dcOem:",*nwpData.Oem)
		}
		if (jdata.NTP != nil) {
			log.Println("dcNTP:",*nwpData.NTP)
		}

		w.WriteHeader(http.StatusOK)
	} else if (r.Method == "GET") {
		var ba []byte
		var err error
		if (nwpType == "NONE") {
			ba,err = json.Marshal(&nwpDataNo)
		} else if (nwpType == "NTP") {
			ba,err = json.Marshal(&nwpDataNTPOnly)
		} else {
			ba,err = json.Marshal(&nwpData)
		}

		if (err != nil) {
			fmt.Printf("ERROR marshalling NWP data: %v\n",err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type","application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(ba)
	} else {
        fmt.Printf("ERROR: request is not a GET or PATCH.\n")
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
}

func dumpCertInfo(cert string) error {
	//Split into cert and key sections.  Look for ----BEGIN pieces.

	splitter := "-----BEGIN RSA PRIVATE KEY-----"
	toks := strings.Split(cert,splitter)
	certData := strings.Replace(toks[0],"\\n","\n",-1)
	keyData := strings.Replace(splitter+toks[1],"\\n","\n",-1)
	err := ioutil.WriteFile(tlsCertFile,[]byte(certData),0666)
	if (err != nil) {
		return err
	}
	err = ioutil.WriteFile(tlsKeyFile,[]byte(keyData),0666)
	if (err != nil) {
		return err
	}
	return nil
}

func startHTTPServer() *http.Server {
	var err error
	srv := &http.Server{Addr: port,}

	go func() {
		if (ishttps) {
			log.Printf("Starting ListenAndServeTLS().")
			err = srv.ListenAndServeTLS(tlsCertFile,tlsKeyFile)
			log.Printf("ListenAndServeTLS() exit: %v.",err)
		} else {
			log.Printf("Starting ListenAndServe().")
			err = srv.ListenAndServe()
			log.Printf("ListenAndServe() exit: %v.",err)
		}
		if (err != http.ErrServerClosed) {
			log.Fatalf("ListenAndServe(): %v",err)
		}
		log.Printf("Exiting startHTTPServer gofunc.")
	}()

	return srv
}


func main() {
	xname := "${X_S0_HOST}"
	port = ":20000"
	envstr := os.Getenv("BMCURL")
	if (envstr != "") {
		burl = envstr
	}
	envstr = os.Getenv("NWPTYPE")
	if (envstr != "") {
		nwpType = envstr
	}
	envstr = os.Getenv("VENDOR")
	if (strings.ToLower(envstr) != "hpe") {
		isHPE = false
	}
	envstr = os.Getenv("BMCPORT")
	if (envstr != "") {
		port = envstr
		if (!strings.Contains(port,":")) {
			port = ":" + port
		}
	}
	envstr = os.Getenv("XNAME")
	if (envstr != "") {
		xname = envstr
	}
	envstr = os.Getenv("HTTPS")
	if (envstr == "1") {
		ishttps = true
		replaceCert = false
	} else if (envstr == "2") {
		ishttps = true
		replaceCert = true
	}
	urlFront := "http://"+xname+port
	if (ishttps) {
		urlFront = "https://"+xname+port
	}
    hstuff := new(httpStuff)
    fmt.Printf("Listening on: %s\n",urlFront)

    http.HandleFunc("/redfish/v1/Managers",hstuff.Managers)
    http.HandleFunc("/redfish/v1/Managers/"+burl,hstuff.ManagersBMC)
    http.HandleFunc("/redfish/v1/Managers/"+burl+"/NetworkProtocol",hstuff.nwp_rcv)
    http.HandleFunc("/redfish/v1/Managers/"+burl+"/SecurityService",hstuff.hpeSecurityService)
    http.HandleFunc("/redfish/v1/Managers/"+burl+"/SecurityService/HttpsCert",hstuff.hpeSecurityServiceHttpsCert)
    http.HandleFunc("/redfish/v1/Managers/"+burl+"/SecurityService/HttpsCert/Actions/HpeHttpsCert.ImportCertificate",hstuff.certificateReplace)
	http.HandleFunc("/redfish/v1/AccountService",hstuff.acctService)
	http.HandleFunc("/redfish/v1/AccountService/Accounts",hstuff.acctAccounts)
	http.HandleFunc("/redfish/v1/AccountService/Accounts/1",hstuff.targAccount1)
	http.HandleFunc("/redfish/v1/AccountService/Accounts/2",hstuff.targAccount2)
	http.HandleFunc("/redfish/v1/AccountService/Accounts/3",hstuff.targAccount3)
	http.HandleFunc("/redfish/v1/AccountService/Accounts/4",hstuff.targAccount4)
	http.HandleFunc("/redfish/v1/Chassis",hstuff.Chassis)
	http.HandleFunc("/redfish/v1/Chassis/Enclosure",hstuff.chassisEnclosure)
	http.HandleFunc("/redfish/v1/Chassis/Rackmount",hstuff.chassisRackmount)
	http.HandleFunc("/redfish/v1/Chassis/Self",hstuff.chassisSelf)
	http.HandleFunc("/redfish/v1/CertificateService",hstuff.certificateService)
	http.HandleFunc("/redfish/v1/CertificateService/Actions/CertificateService.ReplaceCertificate",hstuff.certificateReplace)
	http.HandleFunc("/redfish/v1/CertificateService/CertificateLocations",hstuff.certificateLocations)
	http.HandleFunc("/redfish/v1/",hstuff.rfroot)

	httpsrv = startHTTPServer()
	for {
		time.Sleep(5 * time.Second)
	}
}


