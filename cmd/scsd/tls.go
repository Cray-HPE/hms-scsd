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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	neturl "net/url"
	"strconv"
	"strings"
	"time"

	base "github.com/Cray-HPE/hms-base"
	"github.com/Cray-HPE/hms-certs/pkg/hms_certs"
	trsapi "github.com/Cray-HPE/hms-trs-app-api/pkg/trs_http_api"
	"github.com/gorilla/mux"
)

///////// Cray cert service RF endpoints

// Cray Certificate Service: Actions

type crayCertificateServiceActions struct {
	ReplaceCert crayCertificateServiceActionsReplaceCert `json:"#CertificateService.ReplaceCertificate"`
}

type crayCertificateServiceActionsReplaceCert struct {
	Target string `json:"target"`
}

// Cray Certificate Service

type crayCertificateService struct {
	Etag                 string `json:"@odata.etag,omitempty"`
	Actions              crayCertificateServiceActions
	CertificateLocations crayCertificateServiceLocations
}

type crayCertificateServiceLocations struct {
	ID string `json:"@odata.id"`
}

// Cray Certificate Service Locations

type crayCertificateLocations struct {
	Links crayCertificateLocationLinks
}

type crayCertificateLocationLinks struct {
	Certificates []crayCertificateLocationIDs
}

type crayCertificateLocationIDs struct {
	ID string `json:"@odata.id"`
}

// Certificate Service: certificate payload, works for Cray and HPE

type CertificatePayload struct {
	Certificate       *string         `json:"Certificate"`       //HPE only
	CertificateString *string         `json:"CertificateString"` //all other vendors
	CertificateType   *string         `json:"CertificateType,omitempty"`
	CertificateUri    *CertificateURI `json:"CertificateUri,omitempty"`
}

type CertificateURI struct {
	Uri string `json:"@odata.id"`
}

////// HPE Cert service endpoints

// HPE Managers

type hpeManagers struct {
	Members []hpeManagersMembers `json:"Members"`
}

type hpeManagersMembers struct {
	ID string `json:"@odata.id"`
}

type hpeManagerData struct {
	Oem hpeManagerDataOem `json:"Oem"`
}

type hpeManagerDataOem struct {
	HPE hpeManagerDataOemHPE `json:"HPE"`
}

type hpeManagerDataOemHPE struct {
	Links hpeManagerDataOemHPELinks `json:"Links"`
}

type hpeManagerDataOemHPELinks struct {
	SecurityService hpeManagerDataOemLinksSecurityService `json:"SecurityService"`
}

type hpeManagerDataOemLinksSecurityService struct {
	ID string `json:"@odata.id"`
}

// HPE Security Service

type hpeSecurityService struct {
	Links hpeSecurityServiceLinks `json:"Links"`
}

type hpeSecurityServiceLinks struct {
	HttpsCert hpeSecurityServiceLinksHttpsCert `json:"HttpsCert"`
}

type hpeSecurityServiceLinksHttpsCert struct {
	ID string `json:"@odata.id"`
}

type hpeSecurityServiceHttpsCert struct {
	Actions hpeSecurityServiceHttpsCertActions `json:"Actions"`
}

type hpeSecurityServiceHttpsCertActions struct {
	ImportCertificate hpeSecurityServiceHttpsCertActionsImport `json:"#HpeHttpsCert.ImportCertificate"`
}

type hpeSecurityServiceHttpsCertActionsImport struct {
	Target string `json:"target"`
}

// Redfish Service root

type rfServiceRoot struct {
	CertificateService crayRootCertService
}

type crayRootCertService struct {
	ID string `json:"@odata.id"`
}

// Redfish Chassis

type rfChassis struct {
	Members []rfChassisMembers
}

type rfChassisMembers struct {
	ID string `json:"@odata.id"`
}

// SCSD Cert API payload

type bmcCertData struct {
	Cert string `json:"Cert"`
	Key  string `json:"Key,omitempty"`
}

//////// SCSD CERT ENDPOINTS /////////

// Used for /bmc/managecerts POST and DELETE

type bmcManageCertPost struct {
	Domain    string   `json:"Domain"`
	DomainIDs []string `json:"DomainIDs"`
	FQDN      string   `json:"FQDN,omitempty"`
}

type bmcManageCertPostRsp struct {
	DomainIDs []certRsp `json:"DomainIDs"`
}

//Used for: /bmc/fetchcerts rsp
//          /bmc/managecerts rsp
//          /bmc/rfcerts rsp

type certRsp struct {
	ID         string    `json:"ID"`
	StatusCode int       `json:"StatusCode"`
	StatusMsg  string    `json:"StatusMsg"`
	Cert       *certData `json:"Cert,omitempty"`
}

type certData struct {
	CertType string `json:"CertType"`
	CertData string `json:"CertData"`
	FQDN     string `json:"FQDN"`
}

// Used for /bmc/rfcerts

type rfCertPost struct {
	Force      bool     `json:"Force"`
	CertDomain string   `json:"CertDomain"` //"Cabinet", "Chassis", "BMC", etc.
	Targets    []string `json:"Targets"`
}

type rfCertPostRsp struct {
	Targets []certRsp `json:"Targets"`
}

const (
	VendorCray     = "Cray"
	VendorHPE      = "HPE"
	VendorIntel    = "Intel"
	VendorGigabyte = "GB"
)

// Fetch the certificate URIs needed for cert mgmt, on each target controller
// in a task list.
//
// The algo is:
//   o Verify target using the service root
//   o Check for /redfish/v1/CertificateService.  If present, may be Intel or
//     Cray; if not, it's HPE or GigaByte.
//   o Unambiguate.  Cray:    /redfish/v1/Chassis/Enclosure
//                   Intel:   /redfish/v1/Chassis/RackMount (NOT SUPPORTED)
//                   HPE:     /redfish/v1/Chassis/1
//                   GB:      /redfish/v1/Chassis/Self (NOT SUPPORTED)
//                   RTS/PDU: /redfish/v1/Chassis returns {}.  Also check the
//                            URL, it will have -rts:port.  Same RF cert
//                            schema as Cray mountain!
//   o Call Cray or HPE cert func, and mark the rest as unsupported.
//
// taskList(inout): Task list to execute on which to perform cert replacement.
// certs(in):       TLS cert/key data (leaf cert).
// retData(out):    Returned data for REST return.
// Return:          Error if a failure occurs, else nil.

func setCerts(taskList []trsapi.HttpTask, certs []bmcCertData,
	retData *rfCertPostRsp) error {
	var err error
	var sourceTL trsapi.HttpTask
	var certsCray, certsHPE []bmcCertData

	funcName := "setCerts()"

	if len(taskList) != len(certs) {
		return fmt.Errorf("%s: ERROR: Internal error, key array len != task array len.",
			funcName)
	}

	//Set the URL to /redfish/v1/Chassis

	for ii := 0; ii < len(taskList); ii++ {
		targ := targFromTask(&taskList[ii])
		url := dfltProtocol + "://" + targ + RFCHASSIS_API
		taskList[ii].Request.URL, _ = neturl.Parse(url)
	}

	logger.Tracef("%s: Fetching Chassis data.", funcName)
	err = doOp(taskList)
	if err != nil {
		logger.Errorf("%s: Problem executing chassis data fetch: %v",
			funcName, err)
		return err
	}

	//Set the return data for any chassis-get ops that failed, and also set
	//the ignore flag for subsequent ops.

	for ii := 0; ii < len(taskList); ii++ {
		if taskList[ii].Ignore {
			continue
		}
		ecode := getStatusCode(&taskList[ii])
		if !statusCodeOK(ecode) {
			crsp := certRsp{ID: targFromTask(&taskList[ii]),
				StatusMsg:  getStatusMsg(&taskList[ii]),
				StatusCode: ecode}
			retData.Targets = append(retData.Targets, crsp)
			taskList[ii].Ignore = true
		}
	}

	//Get the results, unambiguate, mark unsupported ones, and call the func
	//corresponding to the target vendor (Cray vs. HPE).

	var tlistCray, tlistHPE, tlistUns []string
	var jdata rfChassis

	for ii := 0; ii < len(taskList); ii++ {
		if taskList[ii].Ignore {
			continue
		}

		targ := targFromTask(&taskList[ii])
		err = grabTaskRspData(funcName, &taskList[ii], &jdata)
		if err != nil {
			logger.Errorf("%s: Problem getting/parsing response from '%s': %v",
				funcName, taskList[ii].Request.URL.Path, err)
			return err
		}

		//Check for RTS PDUs.  These have -rts:port in the hostname portion of
		//the URL.  They should also have no members in the Chassis endpoint.
		//RTS PDUs use the same RF schema as Cray Mountain, so add it to the
		//Cray list, and skip the other checks.

		if strings.Contains(taskList[ii].Request.Host, "-rts") ||
			(len(jdata.Members) == 0) {
			tlistCray = append(tlistCray, targ)
			certsCray = append(certsCray, bmcCertData{Cert: certs[ii].Cert, Key: certs[ii].Key})
			continue
		}

		found := false
		for member := 0; member < len(jdata.Members); member++ {
			logger.Tracef("Member[%d]: '%s", member, jdata.Members[member])
			if strings.Contains(jdata.Members[member].ID, "Enclosure") {
				logger.Tracef("Cray")
				tlistCray = append(tlistCray, targ)
				certsCray = append(certsCray, bmcCertData{Cert: certs[ii].Cert, Key: certs[ii].Key})
				logger.Tracef("%s: Adding '%s' to Cray list", funcName, targ)
				found = true
				break
			} else if strings.Contains(jdata.Members[member].ID, "RackMount") ||
				strings.Contains(jdata.Members[member].ID, "Self") {
				logger.Tracef("Unsupported: intel or GB")
				tlistUns = append(tlistUns, targ)
				logger.Tracef("%s: Adding '%s' to unsupported-vendor list",
					funcName, targ)
				found = true
				break
			} else {
				logger.Tracef("Might be HPE")
				toks := strings.Split(strings.Trim(jdata.Members[member].ID, "/"), "/")
				_, err := strconv.Atoi(toks[len(toks)-1])
				if err == nil {
					logger.Tracef("HPE")
					tlistHPE = append(tlistHPE, targ)
					certsHPE = append(certsHPE, bmcCertData{Cert: certs[ii].Cert, Key: certs[ii].Key})
					logger.Tracef("%s: Adding '%s' to iLO list", funcName, targ)
					found = true
					break
				}
			}
		}
		if !found {
			tlistUns = append(tlistUns, targ)
			logger.Tracef("%s: Adding '%s' to unsupported-vendor list",
				funcName, targ)
		}
	}

	sourceTL.Timeout = time.Duration(appParams.HTTPTimeout) * time.Second
	sourceTL.Request, _ = http.NewRequest("GET", "", nil)

	if len(tlistCray) > 0 {
		logger.Tracef("%s: Setting Cray certs.", funcName)
		taskListCray := tloc.CreateTaskList(&sourceTL, len(tlistCray))
		err = doCrayCerts(taskListCray, tlistCray, certsCray)
		if err != nil {
			logger.Errorf("%s: Problem setting TLS certs on Cray target(s): %v",
				funcName, err)
			return err
		}

		setRetData(taskListCray, retData)
	}

	if len(tlistHPE) > 0 {
		logger.Tracef("%s: Setting HPE certs.", funcName)
		taskListHPE := tloc.CreateTaskList(&sourceTL, len(tlistHPE))
		err = doHPECerts(taskListHPE, tlistHPE, certsHPE)
		if err != nil {
			logger.Errorf("%s: Problem setting TLS certs on HPE target(s): %v",
				funcName, err)
			return err
		}

		setRetData(taskListHPE, retData)
	}

	//Populate unsupported ones

	for ix := 0; ix < len(tlistUns); ix++ {
		elm := certRsp{ID: tlistUns[ix],
			StatusCode: http.StatusNotImplemented,
			StatusMsg:  "Unsupported vendor"}
		retData.Targets = append(retData.Targets, elm)
	}

	return nil
}

// Convenience func to reduce code duplication.  Takes task list
// results and appends them onto a return data struct to return
// to the REST caller.

func setRetData(taskList []trsapi.HttpTask, retData *rfCertPostRsp) {
	for ix := 0; ix < len(taskList); ix++ {
		targ := targFromTask(&taskList[ix])
		ecode := getStatusCode(&taskList[ix])
		emsg := getStatusMsg(&taskList[ix])
		elm := certRsp{ID: targ, StatusCode: ecode, StatusMsg: emsg}
		retData.Targets = append(retData.Targets, elm)
	}
}

//Massage a cert and key into a usable JSON payload.

func makeRFCertPayload(vendor string, cert bmcCertData, certURI string, certType string) []byte {
	var pld CertificatePayload
	var certStr string

	//Replace all CRLF with literal "\n" tuples in the cert and key.

	//TODO: only take the first one if multiple certs?

	certStr = hms_certs.NewlineToTuple(cert.Cert)
	if cert.Key != "" {
		certStr = certStr + "\\n" + hms_certs.NewlineToTuple(cert.Key)
	}

	if vendor == VendorHPE {
		pld.Certificate = &certStr
	} else {
		pld.CertificateString = &certStr
	}
	if certURI != "" {
		pld.CertificateUri = &CertificateURI{Uri: certURI}
	}
	if certType != "" {
		pld.CertificateType = &certType
	}
	ba, _ := json.Marshal(&pld)

	//The #$%^@!! marshaller escapes '\n' sequences into '\\n'.  This is
	//hacky, but it works.
	st := strings.Replace(string(ba), "\\\\", "\\", -1)
	return []byte(st)
}

// Cray cert replacement.  This also works with RTS PDUs.
//
// Algo:
//
// o Get the contents of /redfish/v1/CertificateService to verify the target
//   will take a PEM cert, and that it will do a cert replace operation.
//   Also get the "action" URI.
// o Get contents of /redfish/v1/CertificateService/CertificateLocations to
//   get the URI of the current cert.  There may be multiples -- use the
//   highest numbered one??
// o POST the new cert using that URI (e.g. /redfish/v1/Managers/BMC/NetworkProtocol/HTTPS/Certificates/1)
//   at the action URI /redfish/v1/CertificateService/Actions/CertificateService.ReplaceCertificate
// o May need to deal with Etags too!

func doCrayCerts(taskList []trsapi.HttpTask, targList []string, certs []bmcCertData) error {
	var err error
	funcName := "doCrayCerts()"

	//First check /redfish/v1/CertificateService to verify we're using PEM
	//certs and that we have the cert replacement URI

	logger.Tracef("%s: Fetching CertificateService data.", funcName)
	populateTaskList(taskList, targList, CRAY_CERTSVC_API, "GET", nil)
	err = doOp(taskList)
	if err != nil {
		logger.Errorf("%s: Problem executing cert set: %v",
			funcName, err)
		return err
	}
	ignoreBadTasks(funcName, taskList)

	//Parse the results, set up the next stage.

	//etags := make([]string,len(taskList))
	actionURIs := make([]string, len(taskList))

	logger.Tracef("%s: Parsing Chassis data, setting up for CertificateLocation data.",
		funcName)

	for ii := 0; ii < len(taskList); ii++ {
		if taskList[ii].Ignore {
			continue
		}

		var jdata crayCertificateService
		targ := targFromTask(&taskList[ii])
		err = grabTaskRspData(funcName, &taskList[ii], &jdata)
		if err != nil {
			logger.Errorf("%s: Problem getting response from '%s': %v",
				funcName, taskList[ii].Request.URL.Path, err)
			return err
		}

		//Get the action URI and the CertificateLocations URI

		actionURIs[ii] = jdata.Actions.ReplaceCert.Target
		url := dfltProtocol + "://" + targ + jdata.CertificateLocations.ID
		taskList[ii].Request, _ = http.NewRequest("GET", url, nil)
	}

	err = doOp(taskList)
	if err != nil {
		logger.Errorf("%s: Problem executing cert service data: %v",
			funcName, err)
		return err
	}
	ignoreBadTasks(funcName, taskList)

	//Parse the cert location URI, set stage for the cert write.

	logger.Tracef("%s: Parsing cert location data, set up for cert replace.",
		funcName)

	for ii := 0; ii < len(taskList); ii++ {
		if taskList[ii].Ignore {
			continue
		}

		var jdata crayCertificateLocations
		targ := targFromTask(&taskList[ii])
		err = grabTaskRspData(funcName, &taskList[ii], &jdata)
		if err != nil {
			logger.Errorf("%s: Problem getting response from '%s': %v",
				funcName, taskList[ii].Request.URL.Path, err)
			return err
		}

		//Create URL and payload for writing cert.  Note that Cray only
		//supports 1 cert URI even though it's represented as an array.
		//TODO: this may change eventually.

		certURI := jdata.Links.Certificates[0].ID
		url := dfltProtocol + "://" + targ + actionURIs[ii]
		pld := makeRFCertPayload(VendorCray, certs[ii], certURI, "PEM")

		//This is needed for testing with older mountain BMC FW which
		//uses a different URL than the CertificateLocations says.
		//
		//aaa := strings.Replace(certURI,"/Certificates","",1)
		//logger.Tracef("Cray Cert URI: '%s'",aaa)
		//pld := makeRFCertPayload(VendorCray,certs[ii],aaa,"PEM")

		taskList[ii].Request, _ = http.NewRequest("POST", url, bytes.NewBuffer(pld))
		taskList[ii].Request.Header.Set(CT_TYPE, CT_APPJSON)
		/*etag := jdata.Etag	//might be empty, that's OK
		if (etag != nil) {
			taskList[ii].Request.Header.Add(ET_IFNONE,fixEtag(etag))
		}*/
		logger.Tracef("%s: url: '%s', '%s'",
			funcName, url, taskList[ii].Request.URL.Path)
	}

	//Do the POST to write the new certs

	err = doOp(taskList)
	if err != nil {
		logger.Errorf("%s: Problem setting certificate: %v", funcName, err)
		return err
	}
	ignoreBadTasks(funcName, taskList)
	logger.Tracef("%s: Finished cert replacement.", funcName)

	return nil
}

// HPE blade cert replacement.
//
// Algo:
//
// o GET /redfish/v1/Managers  to get the manager ID (should be only one entry)
// o GET /redfish/v1/Managers/X to get Oem SecurityService URL.
// o GET /redfish/v1/Managers/X/SecurityService to get ETAG and URL of
//   cert mod endpoint from the "Links" data.
// o GET /redfish/v1/Managers/X/SecurityService/HttpsCert
//   This should get the action in the Actions section.
//
// xxxxxxxxx below is current behavior, will change in 1.5 xxxxxxxxxxxxx
// o POST /redfish/v1/Managers/1/SecurityService/HttpsCert/Actions/HpeHttpsCert.GenerateCSR
//       Data is a standard blade-level CSR.
// o GET /redfish/v1/Managers/1/SecurityService/HttpsCert
//       Look for a present and populated CertificateSigningRequest field.
//       Loop over the GET operations until this is populated.
// o Post CSR to the Vault PKI via hms_certs.
// o Get the generated cert from Vault PKI.
// o Store cert in Vault K/V.
// xxxxxxxxx
//
// o POST /redfish/v1/Managers/1/SecurityService/HttpsCert/Actions/HpeHttpsCert.ImportCertificate/
//   Data is cert from CSR. No private key.

func doHPECerts(taskList []trsapi.HttpTask, targList []string, certs []bmcCertData) error {
	var err error
	funcName := "doHPECerts()"

	// GET /redfish/v1/Managers  to get the manager ID (should be only
	// one entry)

	logger.Tracef("%s: Fetching Managers data.", funcName)
	populateTaskList(taskList, targList, HPE_MGR_API, "GET", nil)
	err = doOp(taskList)
	if err != nil {
		logger.Errorf("%s: Problem executing cert set: %v",
			funcName, err)
		return err
	}
	ignoreBadTasks(funcName, taskList)

	//Get the Manager ID and set up the next stage

	logger.Tracef("%s: Parsing Managers data, setup for target manager fetch.",
		funcName)

	for ii := 0; ii < len(taskList); ii++ {
		if taskList[ii].Ignore {
			continue
		}

		var jdata hpeManagers
		targ := targFromTask(&taskList[ii])
		err = grabTaskRspData(funcName, &taskList[ii], &jdata)
		if err != nil {
			logger.Errorf("%s: Problem getting response from '%s': %v",
				funcName, taskList[ii].Request.URL.Path, err)
			return err
		}

		logger.Tracef("%s, '%s': Managers: '%v'",
			funcName, taskList[ii].Request.URL.Path, jdata)

		//TODO: will it always just be one manager?
		url := dfltProtocol + "://" + targ + jdata.Members[0].ID
		taskList[ii].Request.URL, _ = neturl.Parse(url)
	}

	err = doOp(taskList)
	if err != nil {
		logger.Errorf("%s: Problem executing cert set: %v",
			funcName, err)
		return err
	}
	ignoreBadTasks(funcName, taskList)

	//Get the Oem data from the manager account

	logger.Tracef("%s: Parsing target Manager data, setup for security service data fetch.",
		funcName)
	for ii := 0; ii < len(taskList); ii++ {
		if taskList[ii].Ignore {
			continue
		}

		var jdata hpeManagerData
		targ := targFromTask(&taskList[ii])
		err = grabTaskRspData(funcName, &taskList[ii], &jdata)
		if err != nil {
			logger.Errorf("%s: Problem getting response from '%s': %v",
				funcName, taskList[ii].Request.URL.Path, err)
			return err
		}

		url := dfltProtocol + "://" + targ + jdata.Oem.HPE.Links.SecurityService.ID
		taskList[ii].Request.URL, _ = neturl.Parse(url)
	}

	err = doOp(taskList)
	if err != nil {
		logger.Errorf("%s: Problem executing cert set: %v",
			funcName, err)
		return err
	}
	ignoreBadTasks(funcName, taskList)

	// Get the Cert URL

	logger.Tracef("%s: Parsing security service data, setup for cert info fetch.",
		funcName)
	for ii := 0; ii < len(taskList); ii++ {
		if taskList[ii].Ignore {
			continue
		}

		var jdata hpeSecurityService
		targ := targFromTask(&taskList[ii])
		err = grabTaskRspData(funcName, &taskList[ii], &jdata)
		if err != nil {
			logger.Errorf("%s: Problem getting response from '%s': %v",
				funcName, taskList[ii].Request.URL.Path, err)
			return err
		}

		url := dfltProtocol + "://" + targ + jdata.Links.HttpsCert.ID
		taskList[ii].Request.URL, _ = neturl.Parse(url)
	}

	err = doOp(taskList)
	if err != nil {
		logger.Errorf("%s: Problem executing cert set: %v",
			funcName, err)
		return err
	}
	ignoreBadTasks(funcName, taskList)

	// Get the action info, set the stage for the cert POST

	logger.Tracef("%s: Parsing cert info, setup for cert replacement.",
		funcName)

	for ii := 0; ii < len(taskList); ii++ {
		if taskList[ii].Ignore {
			continue
		}

		var jdata hpeSecurityServiceHttpsCert
		targ := targFromTask(&taskList[ii])
		err = grabTaskRspData(funcName, &taskList[ii], &jdata)
		if err != nil {
			logger.Errorf("%s: Problem getting response from '%s': %v",
				funcName, taskList[ii].Request.URL.Path, err)
			return err
		}

		//Set up for the cert post

		pld := makeRFCertPayload(VendorHPE, certs[ii], "", "")
		url := dfltProtocol + "://" + targ + jdata.Actions.ImportCertificate.Target
		taskList[ii].Request, _ = http.NewRequest("POST", url, bytes.NewBuffer(pld))
		taskList[ii].Request.Header.Set(CT_TYPE, CT_APPJSON)
	}

	err = doOp(taskList)
	if err != nil {
		logger.Errorf("%s: Problem executing cert set: %v",
			funcName, err)
		return err
	}

	ignoreBadTasks(funcName, taskList)
	logger.Tracef("%s: Finished cert replacement.", funcName)

	return nil
}

// Convert a user-supplied domain name to one hms_certs can understand.

var dom2domMap = map[string]string{"cabinet": hms_certs.CertDomainCabinet,
	"chassis": hms_certs.CertDomainChassis,
	"blade":   hms_certs.CertDomainBlade,
	"bmc":     hms_certs.CertDomainBMC}

func userDomainToCertDomain(domIN string) (string, error) {
	val, ok := dom2domMap[strings.ToLower(domIN)]
	if !ok {
		return "", fmt.Errorf("Invalid domain name: '%s'", domIN)
	}
	return val, nil
}

// Create leaf cert/key pair(s), and store in Vault.

func doBMCCreateCertsPost(w http.ResponseWriter, r *http.Request) {
	var jdata bmcManageCertPost
	var retData bmcManageCertPostRsp
	funcName := "doBMCCreateCertsPost"

	err := getReqData(funcName, r, &jdata)
	if err != nil {
		emsg := fmt.Sprintf("ERROR: Problem getting request data: %v", err)
		sendErrorRsp(w, "Bad request data", emsg, r.URL.Path,
			http.StatusInternalServerError)
		logger.Errorf("%s: %s", funcName, emsg)
		return
	}

	//For each domain, create the certs

	domainName, derr := userDomainToCertDomain(jdata.Domain)
	if derr != nil {
		emsg := fmt.Sprintf("%v", err)
		sendErrorRsp(w, "Bad request data", emsg, r.URL.Path,
			http.StatusBadRequest)
		logger.Errorf("%s: %s", funcName, emsg)
		return
	}

	retData.DomainIDs = make([]certRsp, len(jdata.DomainIDs))
	domMap := make(map[string][]int)
	certMap := make(map[string]*hms_certs.VaultCertData)

	//Reduce all DomainIDs to a list of actual domain IDs

	for ix := 0; ix < len(jdata.DomainIDs); ix++ {
		retData.DomainIDs[ix].ID = jdata.DomainIDs[ix]
		retData.DomainIDs[ix].StatusCode = http.StatusOK //start with success
		retData.DomainIDs[ix].StatusMsg = "OK"

		domID, err := hms_certs.CheckDomain([]string{jdata.DomainIDs[ix]}, domainName)
		if err != nil {
			logger.Tracef("%s: Domain check for '%s' failed: %v",
				funcName, jdata.DomainIDs[ix], err)
			retData.DomainIDs[ix].StatusCode = http.StatusBadRequest
			retData.DomainIDs[ix].StatusMsg = fmt.Sprintf("ID '%s' not in stated domain: %v",
				jdata.DomainIDs[ix], err)
			continue
		}

		domMap[domID] = append(domMap[domID], ix)
	}

	//Create certs for all mapped domainIDs

	for k, _ := range domMap {
		logger.Tracef("%s: Creating cert for cert domain '%s'", funcName, k)

		vcert := new(hms_certs.VaultCertData)
		err = hms_certs.CreateCert([]string{k}, domainName, jdata.FQDN, vcert)
		if err != nil {
			logger.Tracef("%s: ERROR creating cert for '%s': %v",
				funcName, k, err)

			for xx := 0; xx < len(domMap[k]); xx++ {
				retData.DomainIDs[domMap[k][xx]].StatusCode = http.StatusInternalServerError
				retData.DomainIDs[domMap[k][xx]].StatusMsg = fmt.Sprintf("Error creating cert for '%s', domain '%s: %v",
					jdata.DomainIDs[domMap[k][xx]], jdata.Domain, err)
			}
			delete(domMap, k)
			continue
		}

		certMap[k] = vcert
	}

	//Store the certs

	for k, _ := range domMap {
		logger.Tracef("%s: Storing cert data for '%s'.", funcName, k)
		err := hms_certs.StoreCertData(k, *certMap[k])
		if err != nil {
			logger.Tracef("%s: Cert store for '%s' failed: %v",
				funcName, k, err)
			for xx := 0; xx < len(domMap[k]); xx++ {
				retData.DomainIDs[domMap[k][xx]].StatusCode = http.StatusInternalServerError
				retData.DomainIDs[domMap[k][xx]].StatusMsg = fmt.Sprintf("ERROR storing cert for '%s', domain '%s': %v",
					jdata.DomainIDs[domMap[k][xx]], jdata.Domain, err)
			}
		}
	}

	//Marshal return data and send.

	ba, baerr := json.Marshal(&retData)
	if baerr != nil {
		emsg := fmt.Sprintf("ERROR marshalling response data: %v", baerr)
		sendErrorRsp(w, "JSON marshal error", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	w.Header().Set(CT_TYPE, CT_APPJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

// Delete a leaf cert/key from Vault.

func doBMCDeleteCertsPost(w http.ResponseWriter, r *http.Request) {
	var jdata bmcManageCertPost
	var retData bmcManageCertPostRsp
	funcName := "doBMCDeleteCertsPost"

	err := getReqData(funcName, r, &jdata)
	if err != nil {
		emsg := fmt.Sprintf("ERROR: Problem getting request data: %v", err)
		sendErrorRsp(w, "Bad request data", emsg, r.URL.Path,
			http.StatusInternalServerError)
		logger.Errorf("%s: %s", funcName, emsg)
		return
	}

	domainName, derr := userDomainToCertDomain(jdata.Domain)
	if derr != nil {
		emsg := fmt.Sprintf("%v", err)
		sendErrorRsp(w, "Bad request data", emsg, r.URL.Path,
			http.StatusBadRequest)
		logger.Errorf("%s: %s", funcName, emsg)
		return
	}

	retData.DomainIDs = make([]certRsp, len(jdata.DomainIDs))
	domMap := make(map[string][]int)

	//Reduce all DomainIDs to a list of actual domain IDs

	for ix := 0; ix < len(jdata.DomainIDs); ix++ {
		retData.DomainIDs[ix].ID = jdata.DomainIDs[ix]
		retData.DomainIDs[ix].StatusCode = http.StatusOK //start with success
		retData.DomainIDs[ix].StatusMsg = "OK"

		domID, err := hms_certs.CheckDomain([]string{jdata.DomainIDs[ix]}, domainName)
		if err != nil {
			logger.Tracef("%s: Domain check for '%s' failed: %v",
				funcName, jdata.DomainIDs[ix], err)
			retData.DomainIDs[ix].StatusCode = http.StatusBadRequest
			retData.DomainIDs[ix].StatusMsg = fmt.Sprintf("ID '%s' not in stated domain: %v",
				jdata.DomainIDs[ix], err)
			continue
		}

		domMap[domID] = append(domMap[domID], ix)
	}

	//Delete certs for all mapped domainIDs

	for k, _ := range domMap {
		logger.Tracef("%s: Deleting cert for cert domain '%s'", funcName, k)

		err = hms_certs.DeleteCertData(k, false)
		if err != nil {
			logger.Tracef("%s: ERROR deleting cert for '%s': %v",
				funcName, k, err)

			for xx := 0; xx < len(domMap[k]); xx++ {
				retData.DomainIDs[domMap[k][xx]].StatusCode = http.StatusInternalServerError
				retData.DomainIDs[domMap[k][xx]].StatusMsg = fmt.Sprintf("Error deleting cert for '%s', domain '%s: %v",
					jdata.DomainIDs[domMap[k][xx]], jdata.Domain, err)
			}
		}
	}

	//Marshal return data and send.

	ba, baerr := json.Marshal(&retData)
	if baerr != nil {
		emsg := fmt.Sprintf("ERROR marshalling response data: %v", baerr)
		sendErrorRsp(w, "JSON marshal error", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	w.Header().Set(CT_TYPE, CT_APPJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

// Fetch and display a cert/key from Vault.

func doBMCFetchCerts(w http.ResponseWriter, r *http.Request) {
	var jdata bmcManageCertPost
	var retData bmcManageCertPostRsp
	funcName := "doBMCFetchCerts"

	err := getReqData(funcName, r, &jdata)
	if err != nil {
		emsg := fmt.Sprintf("ERROR: Problem getting response data: %v", err)
		sendErrorRsp(w, "Bad response data", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	domainName, derr := userDomainToCertDomain(jdata.Domain)
	if derr != nil {
		emsg := fmt.Sprintf("%v", err)
		sendErrorRsp(w, "Bad request data", emsg, r.URL.Path,
			http.StatusBadRequest)
		logger.Errorf("%s: %s", funcName, emsg)
		return
	}

	retData.DomainIDs = make([]certRsp, len(jdata.DomainIDs))
	domMap := make(map[string][]int)

	//Reduce all DomainIDs to a list of actual domain IDs

	for ix := 0; ix < len(jdata.DomainIDs); ix++ {
		retData.DomainIDs[ix].ID = jdata.DomainIDs[ix]
		retData.DomainIDs[ix].StatusCode = http.StatusOK //start with success
		retData.DomainIDs[ix].StatusMsg = "OK"

		domID, err := hms_certs.CheckDomain([]string{jdata.DomainIDs[ix]}, domainName)
		if err != nil {
			logger.Tracef("%s: Domain check for '%s' failed: %v",
				funcName, jdata.DomainIDs[ix], err)
			retData.DomainIDs[ix].StatusCode = http.StatusBadRequest
			retData.DomainIDs[ix].StatusMsg = fmt.Sprintf("ID '%s' not in stated domain: %v",
				jdata.DomainIDs[ix], err)
			continue
		}

		domMap[domID] = append(domMap[domID], ix)
	}

	//Fetch certs for all mapped domainIDs

	for k, _ := range domMap {
		logger.Tracef("%s: Fetching cert for cert domain '%s'", funcName, k)

		vcert, err := hms_certs.FetchCertData(k, domainName)
		if err != nil {
			logger.Tracef("%s: ERROR fetching cert for '%s': %v",
				funcName, k, err)

			for xx := 0; xx < len(domMap[k]); xx++ {
				retData.DomainIDs[domMap[k][xx]].StatusCode = http.StatusInternalServerError
				retData.DomainIDs[domMap[k][xx]].StatusMsg = fmt.Sprintf("Error creating cert for '%s', domain '%s: %v",
					jdata.DomainIDs[domMap[k][xx]], jdata.Domain, err)
			}
			continue
		}

		//Set return data for each mapped target

		for xx := 0; xx < len(domMap[k]); xx++ {
			retData.DomainIDs[domMap[k][xx]].Cert = &certData{CertType: "PEM",
				CertData: hms_certs.NewlineToTuple(vcert.Data.Certificate),
				FQDN:     vcert.Data.FQDN}
		}
	}

	//Marshal return data and send.

	ba, baerr := json.Marshal(&retData)
	if baerr != nil {
		emsg := fmt.Sprintf("ERROR marshalling response data: %v", baerr)
		sendErrorRsp(w, "JSON marshal error", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	w.Header().Set(CT_TYPE, CT_APPJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

// Fetch leaf cert(s) from Vault and apply them to RF targets.

func doBMCSetCertsPost(w http.ResponseWriter, r *http.Request) {
	var jdata rfCertPost
	var retData rfCertPostRsp
	var certDomain string

	funcName := "doBMCSetCertsPost"

	err := getReqData(funcName, r, &jdata)
	if err != nil {
		emsg := fmt.Sprintf("ERROR: Problem getting request data: %v", err)
		sendErrorRsp(w, "Bad request data", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	//Verify targets

	td := makeTargData(jdata.Targets)
	_, tderr := hsmVerify(td, jdata.Force, false)
	if tderr != nil {
		emsg := fmt.Sprintf("ERROR: Problem verifying target/states: %v.", tderr)
		sendErrorRsp(w, "Indeterminate target/state", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	//Verify the cert domain

	certDomain, derr := userDomainToCertDomain(jdata.CertDomain)
	if derr != nil {
		emsg := fmt.Sprintf("%v", err)
		sendErrorRsp(w, "Bad request data", emsg, r.URL.Path,
			http.StatusBadRequest)
		logger.Errorf("%s: %s", funcName, emsg)
		return
	}

	bads := make([]bool, len(jdata.Targets))
	certMap := make(map[string]*hms_certs.VaultCertData)
	domMap := make(map[string]*hms_certs.VaultCertData)
	var vcert hms_certs.VaultCertData
	var vcerr error

	//Resolve all targets to domainIDs; make a map 'twixt targets and
	//domainIDs; fetch all relevant domain certs and place into the map.

	for ix := 0; ix < len(jdata.Targets); ix++ {
		domID, err := hms_certs.CheckDomain([]string{jdata.Targets[ix]}, certDomain)
		if err != nil {
			crsp := certRsp{ID: jdata.Targets[ix],
				StatusCode: http.StatusBadRequest,
				StatusMsg: fmt.Sprintf("Cert target %s not found in domain %s",
					jdata.Targets[ix], jdata.CertDomain)}
			retData.Targets = append(retData.Targets, crsp)
			bads[ix] = true
			continue
		}

		//Grab the cert from Vault

		dmap, ok := domMap[domID]
		if ok {
			logger.Tracef("%s: '%s' using already-fetched map for '%s'",
				funcName, jdata.Targets[ix], domID)
			certMap[jdata.Targets[ix]] = dmap
		} else {
			logger.Tracef("%s: '%s' fetching map for '%s'",
				funcName, jdata.Targets[ix], domID)

			vcert, vcerr = hms_certs.FetchCertData(domID, certDomain)
			if vcerr != nil {
				crsp := certRsp{ID: jdata.Targets[ix],
					StatusCode: http.StatusInternalServerError,
					StatusMsg: fmt.Sprintf("ERROR fetching cert for '%s', domain '%s': %v",
						jdata.Targets[ix], jdata.CertDomain, vcerr)}
				retData.Targets = append(retData.Targets, crsp)
				bads[ix] = true
				continue
			}
			domMap[domID] = &vcert
			certMap[jdata.Targets[ix]] = &vcert
		}
	}

	//Call setCerts() to do the dirty work using the map to get the right cert
	//data.

	var tlist []string
	var sourceTL trsapi.HttpTask

	for ix := 0; ix < len(jdata.Targets); ix++ {
		if bads[ix] {
			continue
		}
		tlist = append(tlist, jdata.Targets[ix])
	}

	sourceTL.Timeout = time.Duration(appParams.HTTPTimeout) * time.Second
	sourceTL.Request, _ = http.NewRequest(http.MethodGet, "", nil)
	taskList := tloc.CreateTaskList(&sourceTL, len(tlist))
	populateTaskList(taskList, tlist, RFROOT_API, http.MethodGet, nil)

	//Make cert/key list

	certs := make([]bmcCertData, len(taskList))

	for ii := 0; ii < len(taskList); ii++ {
		if bads[ii] {
			taskList[ii].Ignore = true
			continue
		}
		targ := targFromTask(&taskList[ii])
		certs[ii].Cert = certMap[targ].Data.Certificate
		certs[ii].Key = certMap[targ].Data.PrivateKey
	}

	certErr := setCerts(taskList, certs, &retData)

	if certErr != nil {
		emsg := fmt.Sprintf("ERROR: Certificate set operation failed: %v",
			certErr)
		sendErrorRsp(w, "Certificate set operation error", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	ba, berr := json.Marshal(&retData)
	if berr != nil {
		emsg := fmt.Sprintf("ERROR: Problem marshaling return data.")
		sendErrorRsp(w, "JSON marshal error", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	w.Header().Set(CT_TYPE, CT_APPJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

func getRFPostParams(r *http.Request) (bool, string) {
	var qvals []string
	var ok bool

	force := false
	cdom := "cabinet"

	queryValues := r.URL.Query()

	qvals, ok = queryValues["Force"]
	if !ok {
		qvals, ok = queryValues["force"]
	}
	if ok {
		force = true
	}

	qvals, ok = queryValues["Domain"]
	if !ok {
		qvals, ok = queryValues["domain"]
	}
	if ok {
		cdom = qvals[0]
	}

	return force, cdom
}

// Fetch leaf cert from Vault and apply it to a single RF target.

func doBMCSetCertsPostSingle(w http.ResponseWriter, r *http.Request) {
	var certDomain string
	var retData rfCertPostRsp

	funcName := "doBMCSetCertsPostSingle"
	vars := mux.Vars(r)
	targ := base.NormalizeHMSCompID(vars["xname"])

	force, cdom := getRFPostParams(r)

	//Verify the target

	td := makeTargData([]string{targ})
	_, tderr := hsmVerify(td, force, false)
	if tderr != nil {
		emsg := fmt.Sprintf("ERROR: Invalid xname: '%s'.",
			vars["xname"])
		sendErrorRsp(w, "Bad target name", emsg, r.URL.Path,
			http.StatusBadRequest)
		return
	}

	//Verify the cert domain

	switch strings.ToLower(cdom) {
	case "cabinet":
		certDomain = hms_certs.CertDomainCabinet
	case "chassis":
		certDomain = hms_certs.CertDomainChassis
	case "blade":
		certDomain = hms_certs.CertDomainBlade
	case "bmc":
		certDomain = hms_certs.CertDomainBMC
	default:
		emsg := fmt.Sprintf("ERROR: Invalid cert domain: '%s'.",
			cdom)
		sendErrorRsp(w, "Bad cert domain", emsg, r.URL.Path,
			http.StatusBadRequest)
		return
	}

	//Resolve target to domainID.

	domID, err := hms_certs.CheckDomain([]string{targ}, certDomain)
	if err != nil {
		emsg := fmt.Sprintf("Cert target %s not found in domain %s",
			targ, cdom)
		sendErrorRsp(w, "Bad cert domain", emsg, r.URL.Path,
			http.StatusBadRequest)
		logger.Errorf("%s: %s", funcName, emsg)
		return
	}

	//Grab the cert from Vault

	vcert, vcerr := hms_certs.FetchCertData(domID, certDomain)
	if vcerr != nil {
		emsg := fmt.Sprintf("ERROR fetching cert for '%s', domain '%s': %v",
			targ, cdom, vcerr)
		sendErrorRsp(w, "Bad cert domain", emsg, r.URL.Path,
			http.StatusInternalServerError)
		logger.Errorf("%s: %s", funcName, emsg)
		return
	}

	//Call setCerts() to do the dirty work using the map to get the right cert
	//data.

	var tlist = []string{targ}
	var sourceTL trsapi.HttpTask

	sourceTL.Timeout = time.Duration(appParams.HTTPTimeout) * time.Second
	sourceTL.Request, _ = http.NewRequest(http.MethodGet, "", nil)
	taskList := tloc.CreateTaskList(&sourceTL, len(tlist))
	populateTaskList(taskList, tlist, RFROOT_API, http.MethodGet, nil)

	//Make cert/key list

	certs := make([]bmcCertData, 1)

	certs[0].Cert = vcert.Data.Certificate
	certs[0].Key = vcert.Data.PrivateKey

	certErr := setCerts(taskList, certs, &retData)

	if certErr != nil {
		emsg := fmt.Sprintf("ERROR: Certificate set operation failed: %v",
			certErr)
		sendErrorRsp(w, "Certificate set operation error", emsg, r.URL.Path,
			http.StatusBadRequest)
		logger.Errorf("%s: %s", funcName, emsg)
		return
	}

	ba, berr := json.Marshal(&retData.Targets[0])
	if berr != nil {
		emsg := fmt.Sprintf("ERROR: Problem marshaling return data.")
		sendErrorRsp(w, "JSON data marshal error", emsg, r.URL.Path,
			http.StatusInternalServerError)
		logger.Errorf("%s: %s", funcName, emsg)
		return
	}

	//Since we only return data for a single target, if it failed we will
	//return an error status code.

	w.Header().Set(CT_TYPE, CT_APPJSON)
	w.WriteHeader(retData.Targets[0].StatusCode)
	w.Write(ba)
}
