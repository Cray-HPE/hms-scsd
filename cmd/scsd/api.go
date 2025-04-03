// MIT License
//
// (C) Copyright [2020-2022,2025] Hewlett Packard Enterprise Development LP
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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	base "github.com/Cray-HPE/hms-base/v2"
	trsapi "github.com/Cray-HPE/hms-trs-app-api/pkg/trs_http_api"
	"github.com/Cray-HPE/hms-xname/xnametypes"
	"github.com/gorilla/mux"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

// Describes many aspects of a Redfish target

type targInfo struct {
	target       string        //Can be an XName or a group name
	group        string        //Group target is found in
	groupMatched bool          //Target is a groupname, matched HSM, had BMCs
	state        base.HMSState //State of target
	username     string        //Target's Redfish admin account username
	password     string        //Target's Redfish admin account password
	isMountain   bool          //Indicates target is a mountain controller
	statusCode   int           //Status of most recent RF operation
	err          error         //Error message of most recent RF operation
}

// HSM Component bare-bones info.

type hsmComponent struct {
	ID    string `json:"ID"`
	Type  string `json:"Type"`
	State string `json:"State"`
	Flag  string `json:"Flag"`
}

type hsmComponentList struct {
	Components []hsmComponent `json:"Components"`
}

type grpMembers struct {
	IDS []string `jtag:"ids"`
}

type hsmGroup struct {
	Label          string     `jtag:"label"`
	Description    string     `jtag:"description"`
	Tags           []string   `jtag:"tags"`
	ExclusiveGroup string     `jtag:"exclusiveGroup"`
	Members        grpMembers `jtag:"members"`
}

type hsmGroupList []hsmGroup

// This service's API endpoints

const (
	API_ROOT        = "/v1"
	API_DUMPCFG     = API_ROOT + "/bmc/dumpcfg"
	API_LOADCFG     = API_ROOT + "/bmc/loadcfg"
	API_CFG         = API_ROOT + "/bmc/cfg"
	API_DCREDS      = API_ROOT + "/bmc/discreetcreds"
	API_CREDS       = API_ROOT + "/bmc/creds"
	API_GLB_CREDS   = API_ROOT + "/bmc/globalcreds"
	API_CRT_CERTS   = API_ROOT + "/bmc/createcerts"
	API_DEL_CERTS   = API_ROOT + "/bmc/deletecerts"
	API_FETCH_CERTS = API_ROOT + "/bmc/fetchcerts"
	API_SET_CERTS   = API_ROOT + "/bmc/setcerts"
	API_SET_CERT    = API_ROOT + "/bmc/setcert"
	API_BIOS        = API_ROOT + "/bmc/bios"
	API_HEALTH      = API_ROOT + "/health"
	API_LIVENESS    = API_ROOT + "/liveness"
	API_READINESS   = API_ROOT + "/readiness"
	API_VERSION     = API_ROOT + "/version"
	API_PARAMS      = API_ROOT + "/params"
)

// Commonly used Redfish endpoints

const (
	RFROOT_API       = "/redfish/v1/"
	RFCHASSIS_API    = "/redfish/v1/Chassis"
	IS_MT_API        = "/redfish/v1/Chassis/Enclosure"
	RFMANAGERS_API   = "/redfish/v1/Managers"
	RFSYSTEMS_API    = "/redfish/v1/Systems"
	RFREGISTRIES_API = "/redfish/v1/Registries"
	MT_NWP_API       = "/redfish/v1/Managers/BMC/NetworkProtocol"
	RV_NWP_API       = "/redfish/v1/Managers/Self/NetworkProtocol"
	ACCTSVC_API      = "/redfish/v1/Account"
	CRAY_CERTSVC_API = "/redfish/v1/CertificateService"
	HPE_MGR_API      = "/redfish/v1/Managers"
)

// JSON PUT/POST/PATCH header

const (
	CT_TYPE    = "Content-Type"
	CT_APPJSON = "application/json"
	ET_IFNONE  = "If-None-Match"
)

// Generate the API routes
func newRouter(routes []Route) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	// If the 'pprof' build tag is set, then this will register pprof handlers,
	// otherwise this function is stubbed and will do nothing.
	RegisterPProfHandlers(router)

	return router
}

// Create the API route descriptors.

func generateRoutes() Routes {
	return Routes{
		Route{"doParamsGet",
			strings.ToUpper("Get"),
			API_PARAMS,
			doParamsGet,
		},
		Route{"doParamsPatch",
			strings.ToUpper("Patch"),
			API_PARAMS,
			doParamsPatch,
		},
		Route{"doDumpCfgPost",
			strings.ToUpper("Post"),
			API_DUMPCFG,
			doDumpCfgPost,
		},
		Route{"doLoadCfgPost",
			strings.ToUpper("Post"),
			API_LOADCFG,
			doLoadCfgPost,
		},
		Route{"doCfgGet",
			strings.ToUpper("Get"),
			API_CFG + "/{xname}",
			doCfgGet,
		},
		Route{"doCfgPost",
			strings.ToUpper("Post"),
			API_CFG + "/{xname}",
			doCfgPost,
		},
		Route{"doDiscreetCredsPost",
			strings.ToUpper("Post"),
			API_DCREDS,
			doDiscreetCredsPost,
		},
		Route{"doCredsPostOne",
			strings.ToUpper("Post"),
			API_CREDS + "/{xname}",
			doCredsPostOne,
		},
		Route{"doGlobalCredsPost",
			strings.ToUpper("Post"),
			API_GLB_CREDS,
			doGlobalCredsPost,
		},
		Route{"doCredsGet",
			strings.ToUpper("Get"),
			API_CREDS,
			doCredsGet,
		},
		Route{"doBMCCreateCertsPost",
			strings.ToUpper("Post"),
			API_CRT_CERTS,
			doBMCCreateCertsPost,
		},
		Route{"doBMCDeleteCertsPost",
			strings.ToUpper("Post"),
			API_DEL_CERTS,
			doBMCDeleteCertsPost,
		},
		Route{"doBMCFetchCerts",
			strings.ToUpper("Post"),
			API_FETCH_CERTS,
			doBMCFetchCerts,
		},
		Route{"doBMCSetCertsPost",
			strings.ToUpper("Post"),
			API_SET_CERTS,
			doBMCSetCertsPost,
		},
		Route{"doBMCSetCertsPostSingle",
			strings.ToUpper("Post"),
			API_SET_CERT + "/{xname}",
			doBMCSetCertsPostSingle,
		},
		Route{"doBiosTpmStateGet",
			strings.ToUpper("Get"),
			API_BIOS + "/{xname}/tpmstate",
			doBiosTpmStateGet,
		},
		Route{"doBiosTpmStatePatch",
			strings.ToUpper("Patch"),
			API_BIOS + "/{xname}/tpmstate",
			doBiosTpmStatePatch,
		},
		Route{"doHealthGet",
			strings.ToUpper("Get"),
			API_HEALTH,
			doHealthGet,
		},
		Route{"doLivenessGet",
			strings.ToUpper("Get"),
			API_LIVENESS,
			doLivenessGet,
		},
		Route{"doReadinessGet",
			strings.ToUpper("Get"),
			API_READINESS,
			doReadinessGet,
		},
		Route{"doVersionGet",
			strings.ToUpper("Get"),
			API_VERSION,
			doVersionGet,
		},
	}
}

// Given an HTTP status code, returns if the operation was considered
// successful.

func statusCodeOK(code int) bool {
	switch code {
	case http.StatusOK:
		fallthrough
	case http.StatusCreated:
		fallthrough
	case http.StatusAccepted:
		fallthrough
	case http.StatusNoContent:
		return true
	default:
		return false
	}
	return false
}

// Given an HTTP status code, returns the corresponding status message.

func statusMsg(code int) string {
	if statusCodeOK(code) {
		return "OK"
	} else if code >= 600 {
		return http.StatusText(http.StatusInternalServerError)
	}
	return http.StatusText(int(code))
}

// Return the HTTP status code from a completed TRS task.  This is a little
// funky... if a valid transaction yields no response, there will be no
// Response structure.  This would basically be a 204.  But if there is an
// error, you also get no Response struct, and thus no return code.  The
// "Err" field should have something in it, but the messages don't give good
// details.
//
// So for now, if there is no Response data and Err is nil, we will consider
// that a 204.  If there is no Response and Err is populated, it will be a 500.

func getStatusCode(tp *trsapi.HttpTask) int {
	ecode := 600
	if tp.Request.Response != nil {
		ecode = int(tp.Request.Response.StatusCode)
	} else {
		if tp.Err != nil {
			logger.Tracef("getStatusCode, no response, err: '%v'", *tp.Err)
			ecode = int(http.StatusInternalServerError)
		} else {
			ecode = int(http.StatusNoContent)
		}
	}
	return ecode
}

func getStatusMsg(tp *trsapi.HttpTask) string {
	ecode := getStatusCode(tp)
	if statusCodeOK(ecode) {
		return "OK"
	}

	smsg := fmt.Sprintf("%s, URL: %s", http.StatusText(ecode),
		tp.Request.URL.Path)
	return smsg
}

func ignoreBadTasks(funcName string, taskList []trsapi.HttpTask) {
	for ii := 0; ii < len(taskList); ii++ {
		if taskList[ii].Ignore {
			continue
		}
		ecode := getStatusCode(&taskList[ii])
		if !statusCodeOK(ecode) {
			targ := targFromTask(&taskList[ii])
			logger.Tracef("%s: Target '%s' got bad error code: %d from '%s'",
				funcName, targ, ecode, taskList[ii].Request.URL.Path)
			taskList[ii].Ignore = true
		}
	}
}

// Create a full URL spec from a standard GO request object.

func urlFromReq(req *http.Request) string {
	if (req == nil) || (req.URL == nil) {
		return "NOURL"
	}
	rv := dfltProtocol + "://" + req.URL.Host + req.URL.Path
	return rv
}

// Return the target (hostname/XName) of a TRS task.

func targFromTask(tp *trsapi.HttpTask) string {
	if tp.Request == nil {
		return "BADURL"
	}
	return tp.Request.URL.Host
}

func stripPort(targ string) string {
	if strings.Contains(targ, ":") {
		toks := strings.Split(targ, ":")
		return toks[0]
	}

	return targ
}

// Given a TRS task list, check all status codes. If any of them were
// considered failures, return an error message.

func checkStatusCodes(taskList []trsapi.HttpTask) error {
	for ii := 0; ii < len(taskList); ii++ {
		ecode := getStatusCode(&taskList[ii])
		if !statusCodeOK(ecode) {
			return fmt.Errorf("Bad return code from '%s': %d - %s",
				urlFromReq(taskList[ii].Request), ecode, http.StatusText(ecode))
		}
	}
	return nil
}

func __getRBodyData(funcName string, url string, RBody io.ReadCloser, jdata interface{}) error {
	body, berr := ioutil.ReadAll(RBody)
	if berr != nil {
		logger.Errorf("%s: Problem reading response body from '%s': %v",
			funcName, url, berr)
		return berr
	}

	berr = json.Unmarshal(body, &jdata)
	if berr != nil {
		logger.Errorf("%s: Problem unmarshalling response from '%s': %v",
			funcName, url, berr)
		return berr
	}
	return nil
}

func getReqData(funcName string, r *http.Request, jdata interface{}) error {
	return __getRBodyData(funcName, r.URL.Path, r.Body, jdata)
}

func grabTaskRspData(funcName string, tp *trsapi.HttpTask, jdata interface{}) error {
	if tp.Request.Response == nil {
		ferr := fmt.Errorf("ERROR: response from '%s' has no body.",
			tp.Request.URL.Path)
		return ferr
	}

	return __getRBodyData(funcName, tp.Request.URL.Path, tp.Request.Response.Body, jdata)
}

// Populate Redfish credentials for a TRS task.  These are fetched from
// Vault, unless Vault is disabled, in which case defaults are used.

func popRFCreds(task *trsapi.HttpTask) error {
	if task.Request == nil {
		return fmt.Errorf("Task request is NIL!")
	}

	if (appParams.VaultEnable == nil) || !(*appParams.VaultEnable) {
		logger.Tracef("popRFCreds(), Vault is disabled, no creds.")
		return nil
	}

	logger.Tracef("popRFCreds(), setting vault creds for '%s'-'%s'",
		task.Request.Host, task.Request.URL.Path)
	targ := targFromTask(task)
	creds, err := compCredStore.GetCompCred(targ)
	if err != nil {
		return fmt.Errorf("Can't get RF creds for '%s': %v", targ, err)
	}
	if strings.ToLower(creds.Xname) != strings.ToLower(targ) {
		logger.Infof("Creds for '%s' don't match XName: '%s'",
			targ, creds.Xname)
	}

	task.Request.SetBasicAuth(creds.Username, creds.Password)
	return nil
}

// Convenience func to populate a task list with URL, method, etc.

func populateTaskList(taskList []trsapi.HttpTask, targs []string, urlTail string, method string, pld []byte) {
	for ii := 0; ii < len(taskList); ii++ {
		if taskList[ii].Ignore {
			continue
		}
		url := dfltProtocol + "://" + targs[ii] + urlTail
		if method == http.MethodGet {
			taskList[ii].Request, _ = http.NewRequest(method, url, nil)
		} else {
			taskList[ii].Request, _ = http.NewRequest(method, url, bytes.NewBuffer(pld))
			taskList[ii].Request.Header.Set(CT_TYPE, CT_APPJSON)
		}
	}
}

// Convienience func to send an HTTP error response.

func sendErrorRsp(w http.ResponseWriter, title string, emsg string, url string, ecode int) {
	logger.Errorf("%s", emsg)
	pdet := base.NewProblemDetails("about:blank", title, emsg, url, ecode)
	base.SendProblemDetails(w, pdet, 0)
}

// Query HSM to verify targets and/or expand groups

var hsmClient *http.Client
var hsmTransport *http.Transport

func doHSMGet(url string) ([]byte, error) {
	if hsmClient == nil {
		hsmTransport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		hsmClient = &http.Client{Transport: hsmTransport}
	}
	req, _ := http.NewRequest("GET", url, nil)
	base.SetHTTPUserAgent(req, serviceName)
	rsp, err := hsmClient.Do(req)
	if err != nil {
		logger.Errorf("Problem contacting state manager: %v", err)
		return nil, err
	}
	rspPayload, plerr := ioutil.ReadAll(rsp.Body)
	if plerr != nil {
		logger.Errorf("Problem reading request body: %v", plerr)
		return nil, plerr
	}

	if !statusCodeOK(rsp.StatusCode) {
		emsg := fmt.Errorf("Bad return status from state manager: %d",
			rsp.StatusCode)
		logger.Println(emsg)
		return nil, emsg
	}

	return rspPayload, nil
}

func doHSMPutPostPatchDel(url string, method string, pld []byte) ([]byte, error) {
	if hsmClient == nil {
		hsmTransport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		hsmClient = &http.Client{Transport: hsmTransport}
	}
	req, _ := http.NewRequest(strings.ToUpper(method), url, bytes.NewBuffer(pld))
	base.SetHTTPUserAgent(req, serviceName)
	rsp, err := hsmClient.Do(req)
	if err != nil {
		logger.Errorf("Problem contacting state manager: %v", err)
		return nil, err
	}
	rspPayload, plerr := ioutil.ReadAll(rsp.Body)
	if plerr != nil {
		logger.Errorf("Problem reading request body: %v", plerr)
		return nil, plerr
	}

	if !statusCodeOK(rsp.StatusCode) {
		emsg := fmt.Errorf("Bad return status from state manager: %d",
			rsp.StatusCode)
		logger.Println(emsg)
		return nil, emsg
	}

	return rspPayload, nil
}

// Given an HSM state returns if it is a viable state.

func goodHSMState(state string) bool {
	switch state {
	case base.StateOn.String():
		fallthrough
	case base.StateReady.String():
		return true
	default:
		return false
	}
	return false
}

//Given a list of targets (XNames) return a list of target info descriptors.

func makeTargData(targList []string) []targInfo {
	tdata := make([]targInfo, len(targList))
	for ii := 0; ii < len(targList); ii++ {
		tdata[ii].target = targList[ii]
	}
	return tdata
}

// Use the HSM to verify state of a list of targets.
//
// inList:    List of target descriptors
// force:     If true, use HSM to verify; if not, just verify the XName syntaxes
// expGroups: If true, expand groups.  If force is false this is n/a.
// Return:    List of target descriptors in good state; error if encountered.

func hsmVerify(inList []targInfo, force bool, expGroups bool) ([]targInfo, error) {
	checkList := make([]targInfo, len(inList))
	inMap := make(map[string]*targInfo)
	checkMap := make(map[string]*targInfo)

	if force {
		var nfCheckList []targInfo
		//Just copy the input to the result.  Verify XNames with the lib func.
		for ii := 0; ii < len(inList); ii++ {
			logger.Tracef("hsmVerify, inList[%d]: '%v'", ii, inList[ii])
			hstate := base.StateUnknown
			xn := xnametypes.VerifyNormalizeCompID(stripPort(inList[ii].target))
			if xn != "" {
				hstate = base.StateReady
			}
			ht := targInfo{target: inList[ii].target, group: "", state: hstate}
			nfCheckList = append(nfCheckList, ht)
		}
		return nfCheckList, nil
	}

	//Go through the list and look for non-XNames.  Flag these to be
	//checked as group names.

	checkGroups := false

	for ii := 0; ii < len(inList); ii++ {
		xn := xnametypes.VerifyNormalizeCompID(stripPort(inList[ii].target))
		logger.Tracef("hsmVerify() xn: '%s', targ: '%s'", xn, inList[ii].target)
		if xn == "" {
			checkGroups = true
		}
		checkList[ii] = targInfo{target: inList[ii].target, group: "", state: base.StateUnknown}
		inMap[inList[ii].target] = &checkList[ii]
		logger.Tracef("Added %p '%s' to checkList", &inList[ii], inList[ii].target)
	}

	if checkGroups && expGroups {
		// 1. Expand every non-XName as a group.
		// 2. If it succeeds, check all the members to see if any of them are
		//    BMCs.  Any that are, add to the list.  If none of them are, "fail"
		//    that target in the inList.

		rsp, err := doHSMGet(appParams.SmdURL + "/groups")
		if err != nil {
			logger.Errorf("Getting group info from HSM: %v", err)
			return nil, err
		}
		if rsp == nil {
			emsg := fmt.Errorf("Nil response data from HSM.")
			logger.Error(emsg)
			return nil, emsg
		}

		var groupData hsmGroupList
		err = json.Unmarshal(rsp, &groupData)
		if err != nil {
			logger.Errorf("Problem unmarshalling HSM group data: %v", err)
			return nil, err
		}

		//Look at each group name returned by HSM and if it matches AND contains
		//BMCs, tag the targData element as group-matched.

		var expGroupTargs []targInfo

		for ii := 0; ii < len(groupData); ii++ {
			logger.Tracef("group[%d]: label: '%s', '%v'",
				ii, groupData[ii].Label, groupData[ii])
			_, ok := inMap[groupData[ii].Label]
			if !ok {
				continue
			}

			//Matched the group label.  If there are no BMC members, "fail" this
			//target in inList.  Any that  are BMCs add to the list to program.

			clLen := len(expGroupTargs)
			for jj := 0; jj < len(groupData[ii].Members.IDS); jj++ {
				//In some cases there can be :port especially for testing.
				//Gotta lop that off before checking for valid type.
				targ := stripPort(groupData[ii].Members.IDS[jj])
				ctype := xnametypes.GetHMSType(targ)
				if (ctype != xnametypes.HMSTypeInvalid) && xnametypes.IsHMSTypeController(ctype) {
					expGroupTargs = append(expGroupTargs, targInfo{
						target: groupData[ii].Members.IDS[jj],
						group:  groupData[ii].Label,
						state:  base.StateUnknown})
					logger.Tracef("Added group '%s' member: '%s'",
						groupData[ii].Label, groupData[ii].Members.IDS[jj])
				}
			}

			if len(expGroupTargs) > clLen {
				(*inMap[groupData[ii].Label]).groupMatched = true
			}
		}

		//Append matched group targets onto the checklist

		if len(expGroupTargs) > 0 {
			checkList = append(checkList, expGroupTargs...)
		}
	}

	//Map checklist array by xname.  Don't include group-matched group names.

	for ii := 0; ii < len(checkList); ii++ {
		if checkList[ii].groupMatched {
			continue
		}
		logger.Tracef("Mapping checklist member %d: '%v'", ii, checkList[ii])
		checkMap[checkList[ii].target] = &checkList[ii]
	}

	//TODO: Check for locked components.  This needs to wait for the new
	//lock manager.

	//Do an HSM call to get all BMCs and their states.  Check this against
	//'checkList'.  Record the states of each component.

	rsp, err := doHSMGet(appParams.SmdURL + "/State/Components?type=NodeBMC&type=ChassisBMC&type=RouterBMC&type=CabinetBMC&stateonly=true")
	if err != nil {
		logger.Errorf("Problem getting component info from HSM: %v", err)
		return nil, err
	}
	if rsp == nil {
		emsg := fmt.Errorf("ERROR: Nil response data from HSM.")
		logger.Error(emsg)
		return nil, emsg
	}

	var compData hsmComponentList
	err = json.Unmarshal(rsp, &compData)
	if err != nil {
		logger.Errorf("Problem unmarshalling HSM data: %v", err)
		return nil, err
	}

	//Iterate over the results, grab the good ones, tag the bad ones.

	for ii := 0; ii < len(compData.Components); ii++ {
		vp, ok := checkMap[compData.Components[ii].ID]
		if !ok {
			continue
		}
		if goodHSMState(compData.Components[ii].State) {
			vp.state = base.HMSState(compData.Components[ii].State)
		}
	}

	if appParams.LogLevel == LOGLVL_TRACE {
		logger.Tracef("CHECKLIST:")
		for ii := 0; ii < len(checkList); ii++ {
			logger.Tracef("  [%2d]: %p tg: '%s', gp: '%s', gpm: %t, st: %s, isM: %t sc: %d err: %v",
				ii, &checkList[ii], checkList[ii].target,
				checkList[ii].group, checkList[ii].groupMatched,
				checkList[ii].state.String(), checkList[ii].isMountain,
				checkList[ii].statusCode, checkList[ii].err)
		}
	}

	return checkList, nil
}

//TODO: Lock components before updating them

func lockComponents(taskList []trsapi.HttpTask) error {
	return nil
}

//TODO: Unlock components after updating them

func unlockComponents(taskList []trsapi.HttpTask) error {
	return nil
}

// Convenience func.  Launches a task list and waits for all tasks to complete.
// Returns an error if the launch fails (NOT if any of the tasks fail).

func doOp(taskList []trsapi.HttpTask) error {
	rfClientLock.Lock()
	defer rfClientLock.Unlock()
	nTasks := 0
	for ii := 0; ii < len(taskList); ii++ {
		if taskList[ii].Ignore {
			continue
		}
		nTasks++
		err := popRFCreds(&taskList[ii])
		if err != nil {
			return fmt.Errorf("Error getting RF creds for '%s': %v",
				targFromTask(&taskList[ii]), err)
		}
	}

	if nTasks == 0 {
		logger.Debugf("doOp(): No tasks to perform, all ignored.")
		return nil
	}

	rchan, lerr := tloc.Launch(&taskList)
	if lerr != nil {
		logger.Errorf("Launch() failed: %v", lerr)
		return lerr
	}

	nDone := 0

	for {
		task := <-rchan
		logger.Debugf("Task complete, URL: '%s', status code: %d",
			task.Request.URL.Path, getStatusCode(task))

		nDone++
		if nDone >= nTasks {
			break
		}
	}

	tloc.Close(&taskList)

	return nil
}
