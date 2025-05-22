// MIT License
//
// (C) Copyright [2020-2021,2025] Hewlett Packard Enterprise Development LP
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
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	base "github.com/Cray-HPE/hms-base/v2"
	trsapi "github.com/Cray-HPE/hms-trs-app-api/pkg/trs_http_api"
	"github.com/Cray-HPE/hms-xname/xnametypes"
)

// Used by loadcfg, dumpcfg, and cfg/{xname}

type cfgParams struct {
	NTPServerInfo    *NTPData    `json:"NTPServerInfo,omitempty"`
	SyslogServerInfo *SyslogData `json:"SyslogServerInfo,omitempty"`
	SSHKey           string      `json:"SSHKey,omitempty"`
	SSHConsoleKey    string      `json:"SSHConsoleKey,omitempty"`
	BootOrder        string      `json:"BootOrder,omitempty"`
}

// Ued by cfg/{xname}

type cfgSingle struct {
	Force  bool      `json:"Force"`
	Params cfgParams `json:"Params"`
}

type cfgSingleRsp struct {
	StatusMsg string `json:"StatusMsg"`
}

// Used by /v1/bmc/dumpcfg POST to fetch config params

type dumpCfgPost struct {
	Force   bool     `json:"Force"`
	Targets []string `json:"Targets"`
	Params  []string `json:"Params"`
}

// Return of /v1/bmc/dumpcfg POST

type dumpCfgPostRspElem struct {
	Xname      string    `json:"Xname"`
	StatusCode int       `json:"StatusCode"`
	StatusMsg  string    `json:"StatusMsg"`
	Params     cfgParams `json:"Params:`
}

type dumpCfgPostRsp struct {
	Targets []dumpCfgPostRspElem `json:"Targets"`
}

// Used by /v1/bmc/loadcfg POST to set config params

type loadCfgPost struct {
	Force   bool      `json:"Force,omitempty"`
	Targets []string  `json:"Targets"`
	Params  cfgParams `json:"Params"`
}

// General purpose POST response

type loadCfgPostRspElem struct {
	Xname      string `json:"Xname"`
	StatusCode int    `json:"StatusCode"`
	StatusMsg  string `json:"StatusMsg"`
}

type loadCfgPostRsp struct {
	Targets []loadCfgPostRspElem `json:"Targets"`
}

// Redfish NWProtocol data returned from Mountain controllers

type SyslogData struct {
	ProtocolEnabled bool     `json:"ProtocolEnabled,omitempty"`
	SyslogServers   []string `json:"SyslogServers,omitempty"`
	Transport       string   `json:"Transport,omitempty"`
	Port            int      `json:"Port,omitempty"`
}

type SSHAdminData struct {
	AuthorizedKeys string `json:"AuthorizedKeys,omitempty"`
}

type OemData struct {
	Syslog     *SyslogData   `json:"Syslog,omitempty"`
	SSHAdmin   *SSHAdminData `json:"SSHAdmin,omitempty"`
	SSHConsole *SSHAdminData `json:"SSHConsole,omitempty"`
}

type NTPData struct {
	NTPServers      []string `json:"NTPServers,omitempty"`
	ProtocolEnabled bool     `json:"ProtocolEnabled,omitempty"`
	Port            int      `json:"Port,omitempty"`
}

type RedfishNWProtocol struct {
	Oem *OemData `json:"Oem,omitempty"`
	NTP *NTPData `json:"NTP,omitempty"`
}

// Remove targets which failed an operation from a task list.
//
// inList: List of tasks to check
// Return: List of target names that are still viable.

func removeBadTargs(inList []trsapi.HttpTask) []string {
	var rtargs []string

	for ii := 0; ii < len(inList); ii++ {
		ecode := getStatusCode(&inList[ii])
		if statusCodeOK(ecode) {
			rtargs = append(rtargs, targFromTask(&inList[ii]))
		}
	}
	return rtargs
}

// Check a list of targets to see if they are River or Mountain.
// Returns an error if things fail along the sequence.

func getRvMt(targData []targInfo) error {
	var sourceTL trsapi.HttpTask
	var tlist, tlist2 []string
	var err error

	sourceTL.Timeout = time.Duration(appParams.HTTPTimeout) * time.Second
	sourceTL.Request, _ = http.NewRequest("GET", "", nil)

	//Check if the RF service root is there to weed out bad targ URLs/ctlrs.

	tdMap := make(map[string]*targInfo)
	for ii := 0; ii < len(targData); ii++ {
		tdMap[targData[ii].target] = &targData[ii]
		logger.Tracef("getRvMt(0), targ: '%s', state: %s",
			targData[ii].target, targData[ii].state.String())
		if goodHSMState(targData[ii].state.String()) {
			tlist = append(tlist, targData[ii].target)
		}
	}

	if len(tlist) == 0 {
		logger.Errorf("getRvMt(): no valid targets.")
		emsg := fmt.Errorf("ERROR: No valid targets.")
		return emsg
	}

	taskList1 := tloc.CreateTaskList(&sourceTL, len(tlist))
	populateTaskList(taskList1, tlist, RFROOT_API, http.MethodGet, nil)

	err = doOp(taskList1)
	if err != nil {
		//Launch() failed or some such.  Bail.
		logger.Errorf("getRvMt(1) task launch failed.")
		return err
	}

	//Store results in the target data.  This is so that the 2nd operation
	//can use a reduced set (bad ones from the 1st operation weeded out) and
	//we can still map the result to the correct array elements.

	for ii := 0; ii < len(taskList1); ii++ {
		targ := targFromTask(&taskList1[ii])
		tdMap[targ].statusCode = getStatusCode(&taskList1[ii])
	}

	//Remove failed targets from 1st operation.

	tlist2 = removeBadTargs(taskList1)

	//Now hit the NetworkProtocol endpoint to see mt. versus rv.

	taskList2 := tloc.CreateTaskList(&sourceTL, len(tlist2))
	populateTaskList(taskList2, tlist2, IS_MT_API, http.MethodGet, nil)

	err = doOp(taskList2)
	if err != nil {
		//Launch() failed or some such.  Bail.
		logger.Errorf("getRvMt(2) task launch failed.")
		return err
	}

	//Combine the status info into the result array

	for ii := 0; ii < len(taskList2); ii++ {
		targ := targFromTask(&taskList2[ii])
		scode := getStatusCode(&taskList2[ii])
		(*(tdMap[targ])).statusCode = scode
		if statusCodeOK(scode) {
			(*(tdMap[targ])).isMountain = true
		}
	}

	return nil
}

// Get syslog, NTP servers, SSH keys, SSH console keys for a list of targets.

func getNWP(pmList []string, targData []targInfo) (dumpCfgPostRsp, error) {
	var sourceTL trsapi.HttpTask
	var rspData dumpCfgPostRsp
	var tlist []string

	tdMap := make(map[string]*targInfo)
	for ii := 0; ii < len(targData); ii++ {
		logger.Tracef("getNWP(): targ[%d]: '%s'  '%v'",
			ii, targData[ii].target, targData[ii])
		tdMap[targData[ii].target] = &targData[ii]
		if goodHSMState(targData[ii].state.String()) && targData[ii].isMountain {
			tlist = append(tlist, targData[ii].target)
		}
	}

	if len(tlist) == 0 {
		logger.Errorf("getNWP(): no valid targets.")
		emsg := fmt.Errorf("ERROR: No valid targets.")
		return rspData, emsg
	}

	sourceTL.Timeout = time.Duration(appParams.HTTPTimeout) * time.Second
	sourceTL.Request, _ = http.NewRequest("GET", "", nil)
	taskList := tloc.CreateTaskList(&sourceTL, len(tlist))
	populateTaskList(taskList, tlist, MT_NWP_API, http.MethodGet, nil)

	err := doOp(taskList)
	if err != nil {
		//Launch() failed or some such.  Bail.
		logger.Errorf("getNWP(1) task launch failed.")
		return rspData, err
	}

	//Get the list of params to be returned.

	var iNTP, iSyslog, iSSHKey, iSSHCKey, iBootOrder bool

	for _, prm := range pmList {
		if strings.ToLower(prm) == "ntpserverinfo" {
			iNTP = true
		} else if strings.ToLower(prm) == "syslogserverinfo" {
			iSyslog = true
		} else if strings.ToLower(prm) == "sshkey" {
			iSSHKey = true
		} else if strings.ToLower(prm) == "sshconsolekey" {
			iSSHCKey = true
		} else if strings.ToLower(prm) == "bootorder" {
			iBootOrder = true
		}
	}

	//Return the data back to the caller.

	var jdata RedfishNWProtocol
	for ii := 0; ii < len(taskList); ii++ {
		ecode := getStatusCode(&taskList[ii])
		targ := targFromTask(&taskList[ii])
		tdMap[targ].statusCode = ecode

		if !statusCodeOK(ecode) {
			emsg := fmt.Errorf("ERROR: Bad return status from '%s': %d",
				taskList[ii].Request.URL.Path, ecode)
			logger.Error(emsg)
			rsp := dumpCfgPostRspElem{StatusCode: ecode,
				Xname: targ}
			if tdMap[targ].err != nil {
				rsp.StatusMsg = fmt.Sprintf("%v", tdMap[targ].err)
			} else {
				rsp.StatusMsg = fmt.Sprintf("Target '%s' in bad HSM state: %s",
					targ, string(tdMap[targ].state))
			}
			rspData.Targets = append(rspData.Targets, rsp)
			tdMap[targ].err = emsg
			continue
		}
		if (taskList[ii].Request.Response == nil) || (taskList[ii].Request.Response.ContentLength == 0) {
			emsg := "ERROR: No payload from NWProtocol GET operation."
			logger.Errorf("%s", emsg)
			rsp := dumpCfgPostRspElem{StatusCode: http.StatusPreconditionFailed,
				Xname: targ}
			rsp.StatusMsg = "Target contains no NWProtocol data."
			rspData.Targets = append(rspData.Targets, rsp)
			tdMap[targ].statusCode = http.StatusPreconditionFailed
			tdMap[targ].err = fmt.Errorf("%s", emsg)
			continue
		}

		body, berr := ioutil.ReadAll(taskList[ii].Request.Response.Body)
		if berr != nil {
			emsg := fmt.Sprintf("ERROR: Problem reading GET response: '%v'", berr)
			logger.Errorf("%s", emsg)
			rsp := dumpCfgPostRspElem{StatusCode: http.StatusInternalServerError,
				StatusMsg: "Error reading response from server.",
				Xname:     targ}
			rspData.Targets = append(rspData.Targets, rsp)
			tdMap[targ].statusCode = http.StatusInternalServerError
			tdMap[targ].err = fmt.Errorf("%s", emsg)
			continue
		}
		err := json.Unmarshal(body, &jdata)
		if err != nil {
			emsg := fmt.Sprintf("ERROR: Problem unmarshaling GET response: '%v'", err)
			logger.Errorf("%s", emsg)
			rsp := dumpCfgPostRspElem{StatusCode: http.StatusInternalServerError,
				StatusMsg: "Error umnarshalling server data.",
				Xname:     targ}
			rspData.Targets = append(rspData.Targets, rsp)
			tdMap[targ].statusCode = http.StatusInternalServerError
			tdMap[targ].err = fmt.Errorf("%s", emsg)
			continue
		}

		rsp := &dumpCfgPostRspElem{StatusCode: http.StatusOK,
			StatusMsg: "OK", Xname: targ}

		if iNTP && (jdata.NTP != nil) {
			rsp.Params.NTPServerInfo = &NTPData{}
			rsp.Params.NTPServerInfo.NTPServers = make([]string, len(jdata.NTP.NTPServers))
			copy(rsp.Params.NTPServerInfo.NTPServers, jdata.NTP.NTPServers)
			rsp.Params.NTPServerInfo.Port = jdata.NTP.Port
			rsp.Params.NTPServerInfo.ProtocolEnabled = jdata.NTP.ProtocolEnabled
		}
		if iSyslog && (jdata.Oem != nil) {
			rsp.Params.SyslogServerInfo = &SyslogData{}
			rsp.Params.SyslogServerInfo.ProtocolEnabled = jdata.Oem.Syslog.ProtocolEnabled
			rsp.Params.SyslogServerInfo.Port = jdata.Oem.Syslog.Port
			rsp.Params.SyslogServerInfo.Transport = jdata.Oem.Syslog.Transport
			rsp.Params.SyslogServerInfo.SyslogServers = make([]string, len(jdata.Oem.Syslog.SyslogServers))
			copy(rsp.Params.SyslogServerInfo.SyslogServers, jdata.Oem.Syslog.SyslogServers)
		}
		if iSSHKey && (jdata.Oem != nil) {
			rsp.Params.SSHKey = jdata.Oem.SSHAdmin.AuthorizedKeys
		}
		if iSSHCKey && (jdata.Oem != nil) {
			rsp.Params.SSHConsoleKey = jdata.Oem.SSHConsole.AuthorizedKeys
		}
		if iBootOrder {
			//TODO
		}
		rspData.Targets = append(rspData.Targets, *rsp)
	}

	//Now add in all bad targets (non-mountain or bad-state).

	for ii := 0; ii < len(targData); ii++ {
		if targData[ii].groupMatched {
			continue
		}
		if !goodHSMState(targData[ii].state.String()) {
			elm := dumpCfgPostRspElem{Xname: targData[ii].target,
				StatusCode: http.StatusUnprocessableEntity,
			}
			if targData[ii].err != nil {
				elm.StatusMsg = fmt.Sprintf("%v", targData[ii].err)
			} else {
				elm.StatusMsg = fmt.Sprintf("Target '%s' in bad HSM state: %s",
					targData[ii].target, string(targData[ii].state))
			}
			rspData.Targets = append(rspData.Targets, elm)
		} else if !targData[ii].isMountain {
			elm := dumpCfgPostRspElem{Xname: targData[ii].target,
				StatusCode: http.StatusUnsupportedMediaType,
			}
			elm.StatusMsg = fmt.Sprintf("Target '%s' is COTS, not the correct type",
				targData[ii].target)
			rspData.Targets = append(rspData.Targets, elm)
		}
	}

	return rspData, nil
}

// Set NTPServer, SyslogServer, SSH key, SSH console key on a list of targets.
// Returns the data structs to return to the caller, and error info.

func setNWP(nwp cfgParams, targData []targInfo) (loadCfgPostRsp, error) {
	var sourceTL trsapi.HttpTask
	var rspData loadCfgPostRsp
	var jdata RedfishNWProtocol
	var tlist []string

	//Set up the Redfish version of the NWProtocol stuff

	if (nwp.SyslogServerInfo != nil) || (nwp.SSHKey != "") || (nwp.SSHConsoleKey != "") {
		jdata.Oem = &OemData{}
	}
	if nwp.SyslogServerInfo != nil {
		jdata.Oem.Syslog = &SyslogData{ProtocolEnabled: nwp.SyslogServerInfo.ProtocolEnabled,
			SyslogServers: nwp.SyslogServerInfo.SyslogServers,
			Transport:     nwp.SyslogServerInfo.Transport,
			Port:          nwp.SyslogServerInfo.Port}
	}
	if nwp.SSHKey != "" {
		//jdata.Oem.SSHAdmin.AuthorizedKeys = nwp.SSHKey
		jdata.Oem.SSHAdmin = &SSHAdminData{AuthorizedKeys: nwp.SSHKey}
	}
	if nwp.SSHConsoleKey != "" {
		//jdata.Oem.SSHConsole.AuthorizedKeys = nwp.SSHConsoleKey
		jdata.Oem.SSHConsole = &SSHAdminData{AuthorizedKeys: nwp.SSHConsoleKey}
	}

	if nwp.NTPServerInfo != nil {
		jdata.NTP = nwp.NTPServerInfo
	}

	ba, baerr := json.Marshal(&jdata)
	if baerr != nil {
		emsg := fmt.Sprintf("ERROR: Problem marshaling NWProtocol data: '%v'",
			baerr)
		logger.Errorf("%s", emsg)
		return rspData, fmt.Errorf("%s", emsg)
	}

	logger.Tracef("setNWP() NWP info: '%s'", string(ba))

	tdMap := make(map[string]*targInfo)
	for ii := 0; ii < len(targData); ii++ {
		tdMap[targData[ii].target] = &targData[ii]
		if goodHSMState(targData[ii].state.String()) && targData[ii].isMountain {
			tlist = append(tlist, targData[ii].target)
		}
	}

	if len(tlist) == 0 {
		emsg := fmt.Sprintf("ERROR: No valid targets.")
		logger.Errorf("setNWP(): %s", emsg)
		return rspData, fmt.Errorf("%s", emsg)
	}

	//Check if the RF service root is there to weed out bad targ URLs/ctlrs.

	sourceTL.Timeout = time.Duration(appParams.HTTPTimeout) * time.Second
	sourceTL.Request, _ = http.NewRequest("GET", "", nil)
	taskList := tloc.CreateTaskList(&sourceTL, len(tlist))
	populateTaskList(taskList, tlist, MT_NWP_API, http.MethodPatch, ba)

	err := doOp(taskList)
	if err != nil {
		//Launch() failed or some such.  Bail.
		logger.Errorf("setNWP() Config load task launch failed: %v", err)
		return rspData, err
	}

	for ii := 0; ii < len(taskList); ii++ {
		ecode := getStatusCode(&taskList[ii])
		rsp := loadCfgPostRspElem{Xname: targFromTask(&taskList[ii]),
			StatusCode: ecode,
			StatusMsg:  statusMsg(ecode),
		}
		rspData.Targets = append(rspData.Targets, rsp)
	}

	//Add in the bad targs (hsm bad state, etc.) to the return data.

	for ii := 0; ii < len(targData); ii++ {
		if targData[ii].groupMatched {
			continue
		}
		if !goodHSMState(targData[ii].state.String()) {
			elm := loadCfgPostRspElem{Xname: targData[ii].target,
				StatusCode: http.StatusUnprocessableEntity,
			}
			if targData[ii].err != nil {
				elm.StatusMsg = fmt.Sprintln(targData[ii].err)
			} else {
				elm.StatusMsg = fmt.Sprintf("Target '%s' in bad HSM state: %s",
					targData[ii].target, string(targData[ii].state))
			}
			rspData.Targets = append(rspData.Targets, elm)
		}
	}

	return rspData, nil
}

// /v1/bmc/dumpcfg POST

func doDumpCfgPost(w http.ResponseWriter, r *http.Request) {
	var jdata dumpCfgPost

	defer base.DrainAndCloseRequestBody(r)

	// Decode the JSON to see what we are to return

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		emsg := fmt.Sprintln("ERROR: Problem reading request body:", err)
		sendErrorRsp(w, "Bad request body read", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(body, &jdata)
	if err != nil {
		emsg := fmt.Sprintln("ERROR: Problem unmarshalling request body:", err)
		sendErrorRsp(w, "Unmarshal error", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	if len(jdata.Targets) == 0 {
		emsg := fmt.Sprintln("ERROR: No targets in request.")
		sendErrorRsp(w, "Bad request", emsg, r.URL.Path, http.StatusBadRequest)
		return
	}

	if len(jdata.Params) == 0 {
		emsg := fmt.Sprintln("ERROR: No parameters in request.")
		sendErrorRsp(w, "Missing parameters in request", emsg, r.URL.Path,
			http.StatusBadRequest)
		return
	}

	// Take a list of targs, make a list of targ info (combine rvmt and hsmTarg)
	// Use this list along the way to weed out bad/incorrect targs

	targData := makeTargData(jdata.Targets)
	expTargData, terr := hsmVerify(targData, jdata.Force, true)

	if terr != nil {
		emsg := fmt.Sprintf("ERROR: Problem verifying target states: %v.", terr)
		sendErrorRsp(w, "HSM verification failed", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	//Get river vs. mountain for the targ list

	rverr := getRvMt(expTargData)
	if rverr != nil {
		emsg := fmt.Sprintf("ERROR: Problem determining target architectures: %v.", rverr)
		sendErrorRsp(w, "Target architectures", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	//Next step is to get all of the requested data.  Look at the params
	//desired to be fetched.  These are all mountain-only.

	rdata, rerr := getNWP(jdata.Params, expTargData)
	if rerr != nil {
		emsg := fmt.Sprintln("ERROR: Problem getting NWP data:", rerr)
		sendErrorRsp(w, "NWP data fetch", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	ba, baerr := json.Marshal(rdata)
	if baerr != nil {
		emsg := fmt.Sprintln("ERROR: Problem marshalling NWP data:", baerr)
		sendErrorRsp(w, "Cannot marshal NWP data", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	w.Header().Set(CT_TYPE, CT_APPJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

// /v1/bmc/loadcfg POST

func doLoadCfgPost(w http.ResponseWriter, r *http.Request) {
	var jdata loadCfgPost

	defer base.DrainAndCloseRequestBody(r)

	// Decode the JSON to see what we are to return

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		emsg := fmt.Sprintln("ERROR: problem reading request body:", err)
		sendErrorRsp(w, "Bad request body read", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(body, &jdata)
	if err != nil {
		emsg := fmt.Sprintln("ERROR: problem unmarshalling request body:", err)
		sendErrorRsp(w, "Request data unmarshall error", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	targData := makeTargData(jdata.Targets)
	expTargData, terr := hsmVerify(targData, jdata.Force, true)
	if terr != nil {
		emsg := fmt.Sprintf("ERROR: Problem verifying target states: %v.", terr)
		sendErrorRsp(w, "Indeterminate target state", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	//Get river vs. mountain for the targ list

	rverr := getRvMt(expTargData)
	if rverr != nil {
		emsg := fmt.Sprintf("ERROR: Problem determining target architectures: %v.", rverr)
		sendErrorRsp(w, "Target architectures", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	rsp, rsperr := setNWP(jdata.Params, expTargData)
	if rsperr != nil {
		emsg := fmt.Sprintln("ERROR: problem loading NWP data:", rsperr)
		sendErrorRsp(w, "NWP data", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	ba, berr := json.Marshal(&rsp)
	if berr != nil {
		emsg := fmt.Sprintln("ERROR: problem marshalling NWP data:", berr)
		sendErrorRsp(w, "NWP data marshal error", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	w.Header().Set(CT_TYPE, CT_APPJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

// /v1/bmc/cfg/{xname} GET

func doCfgGet(w http.ResponseWriter, r *http.Request) {
	var qvstr string
	var qvals []string

	defer base.DrainAndCloseRequestBody(r)

	queryValues := r.URL.Query()
	qvstr = queryValues.Get("params")
	if len(qvstr) == 0 {
		qvstr = queryValues.Get("Params")
	}
	if len(qvstr) == 0 {
		//Get all params, no force
		qvstr = "NTPServerInfo SyslogServerInfo SSHKey SSHConsoleKey"
	}

	//We're so nice... allow params?force and get all with force

	if strings.ToLower(qvstr) == "force" {
		qvstr = "Force NTPServerInfo SyslogServerInfo SSHKey SSHConsoleKey"
	}

	qvals = strings.Split(qvstr, " ")
	force := strings.Contains(strings.ToLower(qvstr), "force")

	for _, tok := range qvals {
		if strings.ToLower(tok) == "force" {
			continue
		}
		if (strings.ToLower(tok) != "ntpserverinfo") &&
			(strings.ToLower(tok) != "syslogserverinfo") &&
			(strings.ToLower(tok) != "sshkey") &&
			(strings.ToLower(tok) != "sshconsolekey") {
			emsg := fmt.Sprintf("ERROR: unknown parameter: '%s'", tok)
			sendErrorRsp(w, "Unknown query parameter", emsg, r.URL.Path,
				http.StatusBadRequest)
			return
		}
	}

	vars := mux.Vars(r)
	targData := makeTargData([]string{vars["xname"]})
	expTargData, terr := hsmVerify(targData, force, false)
	if terr != nil {
		emsg := fmt.Sprintln("ERROR: Problem verifying target with HSM:",
			terr)
		sendErrorRsp(w, "HSM verification error", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	rverr := getRvMt(expTargData)
	if rverr != nil {
		emsg := fmt.Sprintf("ERROR: Problem determining target architectures: %v.", rverr)
		sendErrorRsp(w, "Target architectures", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	dfr, err := getNWP(qvals, expTargData)
	if err != nil {
		emsg := fmt.Sprintln("ERROR: Problem fetching NWP data:", err)
		sendErrorRsp(w, "NWP data", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	if !statusCodeOK(int(dfr.Targets[0].StatusCode)) {
		emsg := fmt.Sprintf("ERROR: Bad status fetching NWP data for '%s': %d",
			dfr.Targets[0].Xname, dfr.Targets[0].StatusCode)
		sendErrorRsp(w, "NWP data fetch status error", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	//Convert returned dumpCfgPostRsp to cfgSingle data

	var rdata cfgSingle
	rdata.Force = force
	rdata.Params = dfr.Targets[0].Params

	ba, berr := json.Marshal(&rdata)
	if berr != nil {
		emsg := fmt.Sprintln("ERROR: Problem marshalling NWP data:", berr)
		sendErrorRsp(w, "NWP data marshal error", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	w.Header().Set(CT_TYPE, CT_APPJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

// /v1/bmc/cfg/{xname} POST

func doCfgPost(w http.ResponseWriter, r *http.Request) {
	var jdata cfgSingle
	var rdata cfgSingleRsp

	defer base.DrainAndCloseRequestBody(r)

	vars := mux.Vars(r)
	targ := xnametypes.NormalizeHMSCompID(vars["xname"])

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		emsg := fmt.Sprintln("ERROR: Problem reading request data:", err)
		sendErrorRsp(w, "Request data read error", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &jdata)
	if err != nil {
		emsg := fmt.Sprintln("ERROR: Problem unmarshalling NWP data:", err)
		sendErrorRsp(w, "NWP unmarshal error", emsg, r.URL.Path,
			http.StatusBadRequest)
		return
	}

	//Check for mountain-ness

	targData := makeTargData([]string{targ})
	expTargData, terr := hsmVerify(targData, jdata.Force, true)
	if terr != nil {
		emsg := fmt.Sprintf("ERROR: Problem verifying target states: %v.", terr)
		sendErrorRsp(w, "Indeterminate target state", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	//Get river vs. mountain for the targ list

	rverr := getRvMt(expTargData)
	if rverr != nil {
		emsg := fmt.Sprintf("ERROR: Problem determining target architectures: %v.", rverr)
		sendErrorRsp(w, "Target architectures", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	cpr, cerr := setNWP(jdata.Params, expTargData)
	if cerr != nil {
		emsg := fmt.Sprintln("ERROR: Problem setting NWP data:", cerr)
		sendErrorRsp(w, "NWP data", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	//check the response code

	if len(cpr.Targets) == 0 {
		emsg := fmt.Sprintf("ERROR: No responses from NWP set operation.")
		sendErrorRsp(w, "Error setting NWP data", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	rdata.StatusMsg = statusMsg(int(cpr.Targets[0].StatusCode))
	ba, berr := json.Marshal(&rdata)
	if berr != nil {
		emsg := fmt.Sprintln("ERROR: Problem marshalling NWP data:", berr)
		sendErrorRsp(w, "NWP data marshal error", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	w.Header().Set(CT_TYPE, CT_APPJSON)
	w.WriteHeader(cpr.Targets[0].StatusCode)
	w.Write(ba)
}
