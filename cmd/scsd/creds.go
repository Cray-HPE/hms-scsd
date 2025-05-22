// MIT License
//
// (C) Copyright [2020-2022,2024-2025] Hewlett Packard Enterprise Development LP
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
	"io/ioutil"
	"net/http"
	neturl "net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	base "github.com/Cray-HPE/hms-base/v2"
	trsapi "github.com/Cray-HPE/hms-trs-app-api/pkg/trs_http_api"
	"github.com/Cray-HPE/hms-xname/xnametypes"
	"github.com/gorilla/mux"
)

type credsData struct {
	Username string `json:"Username"`
	Password string `json:"Password"`
}

type credsTarg struct {
	Xname string    `jtag:"Xname"`
	Creds credsData `jtag:"Creds"`
}

type credsPost struct {
	Force   bool        `json:"Force"`
	Targets []credsTarg `json:"Targets"`
}

type globalCredsPost struct {
	Force    bool     `json:"Force"`
	Username string   `json:"Username"`
	Password string   `json:"Password"`
	Targets  []string `json:"Targets"`
}

type credsPostSingle struct {
	Force bool      `json:"Force"`
	Creds credsData `jtag:"Creds"`
}

type bmcCredsData struct {
	Xname      string `json:"Xname"`
	Username   string `json:"Username,omitempty"`
	Password   string `json:"Password,omitempty"`
	StatusCode int    `json:"StatusCode"`
	StatusMsg  string `json:"StatusMsg"`
}

type bmcCredsReturn struct {
	Targets []bmcCredsData `json:"Targets"`
}

// Used by HSM query endpoints

type hsmComponentQuery struct {
	ComponentIDs []string `json:"ComponentIDs`
	Type         []string `json:"type,omitempty"`
	StateOnly    bool     `json:"stateonly,omitempty"`
}

//Redfish account data

type rfAcctSvcData struct {
	ID string `json:"@odata.id"`
}

type rfAccountService struct {
	AccountService rfAcctSvcData `json:"AccountService"`
}

type rfAccountsAccounts struct {
	ID string `json:"@odata.id"`
}
type rfAccounts struct {
	Accounts rfAccountsAccounts `json:"Accounts"`
}

type rfAccountMember struct {
	ID string `json:"@odata.id"`
}

type rfAccountMembers struct {
	Members []rfAccountMember `json:"Members"`
}

type rfAccountData struct {
	UserName string `json:"UserName,omitempty"`
	Password string `json:"Password,omitempty"`
	RoleId   string `json:"RoleId,omitempty"`
	Etag     string `json:"@odata.etag,omitempty"`
}

// Used for tracking Account URLs

type acctID struct {
	index   int
	baseURL string
	IDs     []int
}

// Payload to send for HSM discovery
type discoverPayload struct {
	Xnames []string `json:"xnames"`
	Force  bool     `json:"force"`
}

const (
	EMPTY = "<empty>"
)

// Fix an Etag so we discard any decorations.

func fixEtag(etag string) string {
	toks := strings.Split(etag, "W/")
	//If W/ is present, we'll get 2 tokens, first one empty, 2nd with the etag.
	//Else, we'll get 1 token, with the tag.

	return strings.Trim(toks[len(toks)-1], `"`)
}

// Convenience func.  Given a task list, update each URL with the next
// account URL in its list.

func updateAccountURLs(taskList []trsapi.HttpTask, acctIDList []acctID) {
	for jj := 0; jj < len(taskList); jj++ {
		targ := targFromTask(&taskList[jj])
		aid := strconv.Itoa(acctIDList[jj].IDs[acctIDList[jj].index])
		url := dfltProtocol + "://" + targ + acctIDList[jj].baseURL + "/" + aid
		taskList[jj].Request.URL, _ = neturl.Parse(url)
		logger.Tracef("Target %s, trying next URL: '%s' -- '%s'",
			targ, url, taskList[jj].Request.URL.Path)
	}
}

// Fetch the user account matching the username we are looking for, on each
// target controller in a task list.
//
// The only way to get at the creds stuff on a RF endpoint is to read in each
// account, check if it's the root account, and then update that.   This
// means that we have to have a multiply-iterated list of endpoints until
// all of them resolve, then make the target list from the result of that.
// Some BMCs could have 20 user accts, some may have only 1.  There can be
// gaps in the account numbers as well.
//
// taskList(inout): Task list to execute to find user account URL.
// username(in):    List of account usernames, one per task, to match.
// retEtags(out):   Returned list of Etags, one per target.
// Return:          Error if a failure occurs, else nil.

func fetchTargAccount(taskList []trsapi.HttpTask, username []string, retEtags *[]string) error {
	var err error
	var luserName string
	var acctMembers rfAccountMembers
	var acctData rfAccountData
	var acctSvc rfAccountService
	var acctAccounts rfAccounts
	var maxAcctNum int

	if len(username) > 1 {
		if len(taskList) != len(username) {
			return fmt.Errorf("ERROR: Internal error, username array len != task array len.")
		}
		if len(*retEtags) != len(username) {
			return fmt.Errorf("ERROR: Internal error, etag array len != task array len.")
		}
	}

	unameLen := len(username)

	// First get the URL of the account service for each task.  This is derived
	// From: /redfish/v1/
	// Expecting, e.g.: /redfish/v1/AccountService

	err = doOp(taskList)
	if err != nil {
		logger.Errorf("Problem executing account service fetch task list: %v", err)
		return err
	}
	err = checkStatusCodes(taskList)
	if err != nil {
		return err
	}

	//Set the URL to the next stage

	for ii := 0; ii < len(taskList); ii++ {
		targ := targFromTask(&taskList[ii])
		if taskList[ii].Request.Response == nil {
			ferr := fmt.Errorf("ERROR: no response body from '%s'",
				taskList[ii].Request.URL.Path)
			return ferr
		}
		body, berr := ioutil.ReadAll(taskList[ii].Request.Response.Body)
		if berr != nil {
			logger.Errorf("Problem reading response body from '%s': %v",
				taskList[ii].Request.URL.Path, berr)
			return berr
		}

		err = json.Unmarshal(body, &acctSvc)
		if err != nil {
			logger.Errorf("Problem unmarshalling response from '%s': %v",
				taskList[ii].Request.URL.Path, err)
			return err
		}
		logger.Tracef("fetchTargAccount(1), '%s': acctSvc: '%v'",
			taskList[ii].Request.URL.Path, acctSvc)

		//Set the task's URL to the account area
		url := dfltProtocol + "://" + targ + acctSvc.AccountService.ID
		taskList[ii].Request.URL, _ = neturl.Parse(url)
		logger.Tracef("fetchTargAccount(1a): url: '%s', '%s'",
			url, taskList[ii].Request.URL.Path)
	}

	//The next call will tell us how many accounts there are.
	// From, e.g.: /redfish/v1/AccountService
	// Exp, e.g.:  /redfish/v1/AccountService/Accounts

	err = doOp(taskList)
	if err != nil {
		logger.Errorf("Problem fetching account service counts: %v", err)
		return err
	}
	err = checkStatusCodes(taskList)
	if err != nil {
		return err
	}

	for ii := 0; ii < len(taskList); ii++ {
		targ := targFromTask(&taskList[ii])
		if taskList[ii].Request.Response == nil {
			ferr := fmt.Errorf("ERROR: response from '%s' has no body.",
				taskList[ii].Request.URL.Path)
			return ferr
		}
		body, berr := ioutil.ReadAll(taskList[ii].Request.Response.Body)
		if berr != nil {
			logger.Errorf("Problem reading response body from '%s': %v",
				taskList[ii].Request.URL.Path, berr)
			return berr
		}
		err = json.Unmarshal(body, &acctAccounts)
		if err != nil {
			logger.Errorf("Problem unmarshalling response from '%s': %v",
				taskList[ii].Request.URL.Path, err)
			return err
		}

		//Set the task's URL to the account area
		url := dfltProtocol + "://" + targ + acctAccounts.Accounts.ID
		taskList[ii].Request.URL, _ = neturl.Parse(url)
		logger.Tracef("fetchTargAccount(2) Setting new URL: '%s'/'%s'",
			url, taskList[ii].Request.URL.Path)
	}

	//The next call will tell us how many accounts there are.
	// From, e.g.: /redfish/v1/AccountService/Accounts
	// Exp, e.g.:  /redfish/v1/AccountService/Accounts/1

	err = doOp(taskList)
	if err != nil {
		logger.Errorf("Problem fetching account service counts: %v", err)
		return err
	}
	err = checkStatusCodes(taskList)
	if err != nil {
		return err
	}

	//Parse each returned payload to get the max account ID.

	acctIDList := make([]acctID, len(taskList))
	maxAcctNum = -1
	for ii := 0; ii < len(taskList); ii++ {
		targ := targFromTask(&taskList[ii])
		if taskList[ii].Request.Response == nil {
			ferr := fmt.Errorf("ERROR: response from '%s' has no body.",
				taskList[ii].Request.URL.Path)
			return ferr
		}
		body, berr := ioutil.ReadAll(taskList[ii].Request.Response.Body)
		if berr != nil {
			logger.Errorf("Problem reading response body from '%s': %v",
				taskList[ii].Request.URL.Path, berr)
			return berr
		}
		err = json.Unmarshal(body, &acctMembers)
		if err != nil {
			logger.Errorf("Problem unmarshalling response from '%s': %v",
				taskList[ii].Request.URL.Path, err)
			return err
		}

		//Check for no members!

		if len(acctMembers.Members) == 0 {
			return fmt.Errorf("ERROR: No account members found for '%s'.",
				taskList[ii].Request.URL.Path)
		}

		max := -1
		acctMaxURL := ""
		for jj := 0; jj < len(acctMembers.Members); jj++ {
			//iLO bug: URIs can have trailing '/' sometimes, must trim.
			uri := strings.TrimRight(acctMembers.Members[jj].ID, "/")
			toks := strings.Split(uri, "/")
			ord, err := strconv.Atoi(toks[len(toks)-1])
			if err != nil {
				return fmt.Errorf("ERROR: Can't get account number from '%s'",
					uri)
			}
			if ord > max {
				max = ord
				acctMaxURL = uri
				logger.Tracef("fetchTargAccount(3) Replacing: %s, max: %d, ID: '%s'",
					targ, max, acctMaxURL)
			}
			acctIDList[ii].IDs = append(acctIDList[ii].IDs, ord)
		}

		if max > maxAcctNum {
			maxAcctNum = max
		}

		//Place holder account URL, lop off the account number, leaving
		//e.g. /redfish/v1/AccountService/Accounts

		toks := strings.Split(acctMaxURL, "/")
		acctIDList[ii].baseURL = strings.Join(toks[0:len(toks)-1], "/")
		sort.Sort(sort.Reverse(sort.IntSlice(acctIDList[ii].IDs)))
	}

	//Iterate over the list of accounts for each BMC until the target account
	//is found.  Don't go past the last one -- if there isn't a target acct,
	//it's an error.  We'll start at the end and never go below /1.

	allFound := false
	for ii := maxAcctNum; ii >= 0; ii-- {
		logger.Tracef("Acct Loop: %d=================", ii)
		updateAccountURLs(taskList, acctIDList)
		err := doOp(taskList)
		if err != nil {
			emsg := fmt.Sprintf("Problem fetching valid account URL: %v", err)
			logger.Errorf("%s", emsg)
			return fmt.Errorf("%s", emsg)
		}
		err = checkStatusCodes(taskList)
		if err != nil {
			return fmt.Errorf("Task list returned bad status code(s): %v", err)
		}

		//Read the returned payload to see if it contains our target username.

		numFound := 0
		for jj := 0; jj < len(taskList); jj++ {
			if taskList[jj].Request.Response == nil {
				ferr := fmt.Errorf("ERROR: response from '%s' has no body.",
					taskList[jj].Request.URL.Path)
				return ferr
			}
			body, berr := ioutil.ReadAll(taskList[jj].Request.Response.Body)
			if berr != nil {
				logger.Errorf("Problem reading response body from '%s': %v",
					taskList[jj].Request.URL.Path, berr)
				return berr
			}
			targ := targFromTask(&taskList[jj])
			err = json.Unmarshal(body, &acctData)
			if err != nil {
				logger.Errorf("Problem unmarshalling response from '%s': %v",
					taskList[jj].Request.URL.Path, err)
				return err
			}
			logger.Tracef("Account data read: '%v'", acctData)

			if unameLen == 1 {
				luserName = username[0]
			} else {
				luserName = username[jj]
			}
			if acctData.UserName == luserName {
				numFound++
				if acctData.Etag != "" {
					(*retEtags)[jj] = fixEtag(acctData.Etag)
				}
				logger.Tracef("Target %s, username Match! '%s'",
					targ, taskList[jj].Request.URL.Path)
			} else {
				//No match, increment to the next account name.  Don't go past
				//the end of the list.
				if acctIDList[jj].index < (len(acctIDList[jj].IDs) - 1) {
					acctIDList[jj].index++
				}
			}
		}

		if numFound == len(taskList) {
			allFound = true
			logger.Infof("Found all relevant Account Service URLs.")
			break
		}
	}

	if !allFound {
		return fmt.Errorf("Target(s) had no matching Redfish account.")
	}
	return nil
}

// Set the credit info on each target BMC/controller and also in Vault.
// Each URL in the task list is presumed to point to the target account URL.
//
// taskList:  Task list of targets to set creds for.
// password:  List of account passwords, one per task/target.  If the length
//            of this array is 1, then use the same password for all targets.
// etags:     List of etags, one per target.  These will always be unique to
//            each target.
// Return:    Error string on failure, else nil.

func setCreds(taskList []trsapi.HttpTask, password []string, etags []string) error {
	var accData rfAccountData

	//Set the data in e.g. /redfish/v1/AccountService/Accounts/2.  All the
	//URLs should now be pointing to the correct place.

	pwlen := len(password)

	for ii := 0; ii < len(taskList); ii++ {
		if pwlen == 1 {
			accData.Password = password[0]
		} else {
			accData.Password = password[ii]
		}
		ba, berr := json.Marshal(&accData)
		if berr != nil {
			return berr
		}
		url := dfltProtocol + "://" + targFromTask(&taskList[ii]) + taskList[ii].Request.URL.Path
		taskList[ii].Request, _ = http.NewRequest("PATCH", url, bytes.NewBuffer(ba))
		taskList[ii].Request.Header.Set(CT_TYPE, CT_APPJSON)
		if etags[ii] != "" {
			taskList[ii].Request.Header.Add(ET_IFNONE, etags[ii])
		}
		logger.Tracef("setCreds(): task[%d] URL: '%s' - '%s', pld: '%s'",
			ii, taskList[ii].Request.Host, taskList[ii].Request.URL.Path,
			string(ba))
	}

	err := doOp(taskList)
	if err != nil {
		return err
	}
	err = checkStatusCodes(taskList)
	if err != nil {
		return err
	}

	return nil
}

// Convenience func to make a target descriptor list from a list of credential
// targets received from the creds API.

func makeTarglistFromCreds(clist []credsTarg) []string {
	tlist := make([]string, len(clist))
	for ii := 0; ii < len(clist); ii++ {
		tlist[ii] = clist[ii].Xname
	}
	return tlist
}

// Update the Redfish credentials in vault for a given target.

func updateCreds(targ, uname, pw string) string {
	if (appParams.VaultEnable != nil) && *appParams.VaultEnable {
		creds, err := compCredStore.GetCompCred(targ)
		if err != nil {
			//Can't update!!
			return fmt.Sprintf("ERROR: Unable to read RF creds from vault for '%s': %v",
				targ, err)
		}

		creds.Username = uname
		creds.Password = pw
		err = compCredStore.StoreCompCred(creds)
		if err != nil {
			//Can't update!!
			return fmt.Sprintf("ERROR: Unable to write RF creds to vault for '%s': %v",
				targ, err)
		}
	}
	return ""
}

// /v1/bmc/discreetcreds POST

func doDiscreetCredsPost(w http.ResponseWriter, r *http.Request) {
	var jdata credsPost
	var retData loadCfgPostRsp

	defer base.DrainAndCloseRequestBody(r)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		emsg := fmt.Sprintf("ERROR: Problem reading request body: %v", err)
		sendErrorRsp(w, "Bad request body read", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &jdata)
	if err != nil {
		emsg := fmt.Sprintf("ERROR: Problem unmarshalling request: %v", err)
		sendErrorRsp(w, "Request unmarshal error", emsg, r.URL.Path,
			http.StatusBadRequest)
		return
	}

	//Make a map using the target name to get at the creds later, since
	//not all of the inbound targs may end up being used.

	credMap := make(map[string]*credsTarg)
	for ii := 0; ii < len(jdata.Targets); ii++ {
		credMap[jdata.Targets[ii].Xname] = &jdata.Targets[ii]
	}

	//Verify targets with HSM

	tl := makeTarglistFromCreds(jdata.Targets)
	targData := makeTargData(tl)
	expTargData, terr := hsmVerify(targData, jdata.Force, false)

	if terr != nil {
		emsg := fmt.Sprintf("ERROR: Problem verifying target states: %v.", terr)
		sendErrorRsp(w, "Target state error", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	var tlist []string
	tdMap := make(map[string]*targInfo)

	for ii := 0; ii < len(expTargData); ii++ {
		tdMap[expTargData[ii].target] = &expTargData[ii]
		if goodHSMState(expTargData[ii].state.String()) {
			tlist = append(tlist, expTargData[ii].target)
		}
	}

	if len(tlist) == 0 {
		emsg := fmt.Sprintf("ERROR: No valid targets.")
		sendErrorRsp(w, "No valid targets", emsg, r.URL.Path,
			http.StatusBadRequest)
		return
	}

	//Store the creds in the HW.  Keep track of failures, and only update
	//Vault with the ones that succeeded.

	var sourceTL trsapi.HttpTask
	unArray := make([]string, len(tlist))
	pwArray := make([]string, len(tlist))
	etagArray := make([]string, len(tlist))

	sourceTL.Timeout = time.Duration(appParams.HTTPTimeout) * time.Second
	sourceTL.Request, _ = http.NewRequest("GET", "", nil)
	taskList := tloc.CreateTaskList(&sourceTL, len(tlist))

	for ii := 0; ii < len(taskList); ii++ {
		url := dfltProtocol + "://" + tlist[ii] + RFROOT_API
		taskList[ii].Request, _ = http.NewRequest("GET", url, nil)
		logger.Tracef("dctCredPost(), url: '%s' / '%s'",
			url, taskList[ii].Request.URL)
		cp, ok := credMap[tlist[ii]]
		if !ok {
			emsg := fmt.Sprintf("ERROR: Problem retrieving auth info for '%s'",
				tlist[ii])
			sendErrorRsp(w, "Cred retrieval error", emsg, url,
				http.StatusInternalServerError)
			return
		}
		unArray[ii] = cp.Creds.Username
		pwArray[ii] = cp.Creds.Password
	}

	//Fetch the target account URLs

	err = fetchTargAccount(taskList, unArray, &etagArray)
	if err != nil {
		emsg := fmt.Sprintf("ERROR: Problem retrieving user accounts: %v", err)
		sendErrorRsp(w, "User account fetch error", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	//Now that we have all of the URLs in place, perform the operation.

	err = setCreds(taskList, pwArray, etagArray)
	if err != nil {
		//This error means that NOTHING worked.  Just return an error msg.
		emsg := fmt.Sprintf("ERROR: Problem attempting to set user creds: %v, none were changed.", err)
		sendErrorRsp(w, "Cred set error", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	//This is the tricky part.  For each creds-set task that succeeded,
	//update the cred store.  Failed ones, don't do dat.

	retData.Targets = make([]loadCfgPostRspElem, len(taskList))

	var discoveryTargets []string
	numBad := 0
	for ii := 0; ii < len(taskList); ii++ {
		ecode := getStatusCode(&taskList[ii])
		targ := targFromTask(&taskList[ii])
		tdMap[targ].statusCode = ecode
		if statusCodeOK(ecode) {
			logger.Infof("INFO: RF creds for '%s' successfully updated.", targ)
			discoveryTargets = append(discoveryTargets, targ)
			errStr := updateCreds(targ, unArray[ii], pwArray[ii])
			if errStr != "" {
				tdMap[targ].statusCode = http.StatusPreconditionFailed
				tdMap[targ].err = fmt.Errorf("%s", errStr)
				//TODO: NOTE: if we can't store creds, then the HW and the cred
				//store are out of sync.  Will need to be able to un-do this
				//at some point.
				continue
			}
		} else {
			emsg := fmt.Sprintf("ERROR: RF cred set operation failed for '%s'/'%s', creds unchanged.",
				targ, taskList[ii].Request.URL.Path)
			logger.Errorf("%s", emsg)
			tdMap[targ].err = fmt.Errorf("%s", emsg)
			numBad++
		}

		retData.Targets[ii].Xname = targ
		retData.Targets[ii].StatusCode = ecode
		retData.Targets[ii].StatusMsg = http.StatusText(ecode)
	}

	//Add in the bad targs (hsm bad state, etc.) to the return data.

	for ii := 0; ii < len(expTargData); ii++ {
		if expTargData[ii].groupMatched {
			continue
		}
		if !goodHSMState(expTargData[ii].state.String()) {
			elm := loadCfgPostRspElem{Xname: expTargData[ii].target,
				StatusCode: http.StatusUnprocessableEntity,
			}
			if expTargData[ii].err != nil {
				elm.StatusMsg = fmt.Sprintln(expTargData[ii].err)
			} else {
				elm.StatusMsg = fmt.Sprintf("Target '%s' in bad HSM state: %s",
					expTargData[ii].target, string(expTargData[ii].state))
			}
			retData.Targets = append(retData.Targets, elm)
		}
	}

	ba, berr := json.Marshal(&retData)
	if berr != nil {
		emsg := fmt.Sprintf("ERROR: Problem marshaling return data; %d targets had errors setting creds.", numBad)
		sendErrorRsp(w, "Return data marshal error", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	doHSMDiscover(discoveryTargets)

	w.Header().Set(CT_TYPE, CT_APPJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

// /v1/bmc/globalcreds POST

func doGlobalCredsPost(w http.ResponseWriter, r *http.Request) {
	var jdata globalCredsPost
	var retData loadCfgPostRsp

	defer base.DrainAndCloseRequestBody(r)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		emsg := fmt.Sprintf("ERROR: Problem reading request body: %v", err)
		sendErrorRsp(w, "Bad request body read", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &jdata)
	if err != nil {
		emsg := fmt.Sprintf("ERROR: Problem unmarshalling request: %v", err)
		sendErrorRsp(w, "Request unmarshal error", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	//Verify targets with HSM

	targData := makeTargData(jdata.Targets)
	expTargData, terr := hsmVerify(targData, jdata.Force, true)
	var tlist []string

	if terr != nil {
		emsg := fmt.Sprintf("ERROR: Problem verifying target states: %v.", terr)
		sendErrorRsp(w, "Target state validation error", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	tdMap := make(map[string]*targInfo)
	for ii := 0; ii < len(expTargData); ii++ {
		tdMap[expTargData[ii].target] = &expTargData[ii]
		if goodHSMState(expTargData[ii].state.String()) {
			tlist = append(tlist, expTargData[ii].target)
		}
	}

	if len(tlist) == 0 {
		emsg := fmt.Sprintf("ERROR: No valid targets.")
		sendErrorRsp(w, "No valid targets", emsg, r.URL.Path,
			http.StatusBadRequest)
		return
	}

	//Store the creds in the HW.  Keep track of failures, and only update
	//Vault with the ones that succeeded.

	var sourceTL trsapi.HttpTask
	sourceTL.Timeout = time.Duration(appParams.HTTPTimeout) * time.Second
	sourceTL.Request, _ = http.NewRequest("GET", "", nil)
	taskList := tloc.CreateTaskList(&sourceTL, len(tlist))
	populateTaskList(taskList, tlist, RFROOT_API, http.MethodGet, nil)

	//Fetch the target account URLs

	etagArray := make([]string, len(tlist))

	err = fetchTargAccount(taskList, []string{jdata.Username}, &etagArray)
	if err != nil {
		emsg := fmt.Sprintf("ERROR: Problem retrieving user accounts: %v", err)
		sendErrorRsp(w, "User account retrieval error", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	//Now that we have all of the URLs in place, perform the operation.

	err = setCreds(taskList, []string{jdata.Password}, etagArray)
	if err != nil {
		//This error means that NOTHING worked.  Just return an error msg.
		emsg := fmt.Sprintf("ERROR: Problem attempting to set user creds: %v, none were changed.", err)
		sendErrorRsp(w, "User cred set error", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	//This is the tricky part.  For each creds-set task that succeeded,
	//update the cred store.  Failed ones, don't do dat.

	retData.Targets = make([]loadCfgPostRspElem, len(taskList))

	var discoveryTargets []string
	numBad := 0
	for ii := 0; ii < len(taskList); ii++ {
		ecode := getStatusCode(&taskList[ii])
		targ := targFromTask(&taskList[ii])
		if statusCodeOK(ecode) {
			logger.Infof("INFO: RF creds for '%s' successfully updated.", targ)
			discoveryTargets = append(discoveryTargets, targ)
			errStr := updateCreds(targ, jdata.Username, jdata.Password)
			if errStr != "" {
				logger.Errorf("%s", errStr)
				//TODO: NOTE: if we can't store creds, then the HW and the cred
				//store are out of sync.  Will need to be able to un-do this
				//at some point.
				continue
			}
		} else {
			logger.Infof("INFO: RF cred set operation failed for '%s'/'%s', creds unchanged.",
				targ, taskList[ii].Request.URL.Path)
			numBad++
		}

		retData.Targets[ii].Xname = targ
		retData.Targets[ii].StatusCode = ecode
		retData.Targets[ii].StatusMsg = http.StatusText(ecode)
	}

	//Add in the bad targs (hsm bad state, etc.) to the return data.

	for ii := 0; ii < len(expTargData); ii++ {
		if expTargData[ii].groupMatched {
			continue
		}
		if !goodHSMState(expTargData[ii].state.String()) {
			elm := loadCfgPostRspElem{Xname: expTargData[ii].target,
				StatusCode: http.StatusUnprocessableEntity,
			}
			if expTargData[ii].err != nil {
				elm.StatusMsg = fmt.Sprintln(expTargData[ii].err)
			} else {
				elm.StatusMsg = fmt.Sprintf("Target '%s' in bad HSM state: %s",
					expTargData[ii].target, string(expTargData[ii].state))
			}
			retData.Targets = append(retData.Targets, elm)
		}
	}
	doHSMDiscover(discoveryTargets)

	ba, berr := json.Marshal(&retData)
	if berr != nil {
		emsg := fmt.Sprintf("ERROR: Problem marshaling return data; %d targets had errors setting creds.", numBad)
		sendErrorRsp(w, "Cred op return data marshal error", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	w.Header().Set(CT_TYPE, CT_APPJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

func doCredsPostOne(w http.ResponseWriter, r *http.Request) {
	var jdata credsPostSingle
	var retData cfgSingleRsp

	defer base.DrainAndCloseRequestBody(r)

	mvars := mux.Vars(r)
	XName := xnametypes.NormalizeHMSCompID(mvars["xname"])

	var discoveryTargets []string
	discoveryTargets = append(discoveryTargets, XName)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		emsg := fmt.Sprintf("ERROR: Problem reading request body: %v", err)
		sendErrorRsp(w, "Bad request body read", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &jdata)
	if err != nil {
		emsg := fmt.Sprintf("ERROR: Problem unmarshalling request: %v", err)
		sendErrorRsp(w, "Request data unmarshal error", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	targData := makeTargData([]string{XName})
	expTargData, terr := hsmVerify(targData, jdata.Force, false)
	if terr != nil {
		emsg := fmt.Sprintf("ERROR: Problem verifying target states: %v.", terr)
		sendErrorRsp(w, "HSM state validation error", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	if !goodHSMState(expTargData[0].state.String()) {
		emsg := fmt.Sprintf("ERROR: Target '%s' in incorrect state: %s",
			expTargData[0].target, string(expTargData[0].state))
		sendErrorRsp(w, "Target in bad state", emsg, r.URL.Path,
			http.StatusUnprocessableEntity)
		return
	}

	//Store the creds in the HW.  Keep track of failures, and only update
	//Vault with the ones that succeeded.

	var sourceTL trsapi.HttpTask

	sourceTL.Timeout = time.Duration(appParams.HTTPTimeout) * time.Second
	sourceTL.Request, _ = http.NewRequest("GET", "", nil)
	taskList := tloc.CreateTaskList(&sourceTL, 1)
	populateTaskList(taskList, []string{XName}, RFROOT_API, http.MethodGet, nil)

	//Fetch the target account URLs

	etagArray := make([]string, len(taskList))

	err = fetchTargAccount(taskList, []string{jdata.Creds.Username}, &etagArray)
	if err != nil {
		emsg := fmt.Sprintf("ERROR: Problem retrieving user accounts: %v", err)
		sendErrorRsp(w, "User account retrieval error", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	//Now that we have all of the URLs in place, perform the operation.

	err = setCreds(taskList, []string{jdata.Creds.Password}, etagArray)
	if err != nil {
		//This error means that NOTHING worked.  Just return an error msg.
		emsg := fmt.Sprintf("ERROR: Problem attempting to set user creds: %v, none were changed.", err)
		sendErrorRsp(w, "User cred set error", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	//If this succeeded, update the cred store.  Else, don't do dat.

	ecode := getStatusCode(&taskList[0])
	targ := targFromTask(&taskList[0])
	sawErr := false
	if statusCodeOK(ecode) {
		logger.Infof("INFO: RF creds for '%s' successfully updated.", targ)
		errStr := updateCreds(targ, jdata.Creds.Username, jdata.Creds.Password)
		if errStr != "" {
			logger.Errorf("%s", errStr)
			//TODO: NOTE: if we can't store creds, then the HW and the cred
			//store are out of sync.  Will need to be able to un-do this
			//at some point.
		}
	} else {
		logger.Infof("INFO: RF cred set operation failed for '%s'/'%s', creds unchanged.",
			targ, taskList[0].Request.URL.Path)
		sawErr = true
	}

	retData.StatusMsg = http.StatusText(ecode)
	ba, berr := json.Marshal(&retData)
	if berr != nil {
		emsg := "ERROR: Problem marshaling return data"
		if sawErr {
			emsg += "  Target had errors setting creds."
		}
		sendErrorRsp(w, "Return data marshal error", emsg, r.URL.Path,
			http.StatusInternalServerError)
		return
	}
	doHSMDiscover(discoveryTargets)

	w.Header().Set(CT_TYPE, CT_APPJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

func doCredsGet(w http.ResponseWriter, r *http.Request) {
	var xnames, retXnames []string
	var compType string
	var retData bmcCredsReturn

	defer base.DrainAndCloseRequestBody(r)

	if (appParams.VaultEnable == nil) || !(*appParams.VaultEnable) {
		logger.Tracef("doCredsGet(), Vault is disabled, no creds available.")
		sendErrorRsp(w, "Vault not available",
			"ERROR: Vault access is disabled.",
			r.URL.Path, http.StatusGone)
		return
	}

	//Get the params.  If none, we get all BMC creds.

	qvals := r.URL.Query()
	targlist, ok := qvals["targets"]
	typelist, tlok := qvals["type"]

	if ok && ((len(targlist) == 0) || (targlist[0] == "")) {
		sendErrorRsp(w, "Invalid query parameter",
			"ERROR: URL query parameter is empty.",
			r.URL.Path, http.StatusBadRequest)
		return
	}

	if ok {
		xlist := strings.Split(targlist[0], ",")
		elist := []string{}

		//Verify the name formats
		for ii := 0; ii < len(xlist); ii++ {
			xn := xnametypes.VerifyNormalizeCompID(xlist[ii])
			if xn == "" {
				logger.Errorf("Invalid XName: '%s'", xlist[ii])
				elist = append(elist, xlist[ii])
			} else {
				xnames = append(xnames, xn)
			}
		}

		if len(elist) != 0 {
			bxn := strings.Join(elist, ",")
			sendErrorRsp(w, "Bad XName(s) entered",
				fmt.Sprintf("ERROR: Invalid Xnames: %s.", bxn),
				r.URL.Path, http.StatusInternalServerError)
			return
		}
	}

	if tlok {
		//We'll only allow one type
		toks := strings.Split(typelist[0], ",")
		if len(toks) > 1 {
			sendErrorRsp(w, "Invalid query parameter 'type'",
				"ERROR: URL query parameter 'type' can only be a single value.",
				r.URL.Path, http.StatusBadRequest)
			return
		}
		compType = xnametypes.VerifyNormalizeType(toks[0])
		if compType == "" {
			sendErrorRsp(w, "Invalid query parameter 'type'",
				"ERROR: URL query parameter 'type' is invalid component type.",
				r.URL.Path, http.StatusBadRequest)
			return
		}
	}

	//Get list of XNames from HSM.

	urlTail := ""
	var rsp []byte
	var rerr error

	if len(xnames) == 0 {
		urlTail = "/State/Components"
		if compType == "" {
			urlTail = urlTail + "?type=NodeBMC&type=ChassisBMC&type=RouterBMC&type=CabinetBMC&stateonly=true"
		} else {
			urlTail = urlTail + "?type=" + compType + "&stateonly=true"
		}
		rsp, rerr = doHSMGet(appParams.SmdURL + urlTail)
	} else {
		urlTail = "/State/Components/Query"
		jdata := hsmComponentQuery{ComponentIDs: xnames, StateOnly: true}
		if compType != "" {
			jdata.Type = []string{compType}
		}
		ba, baerr := json.Marshal(&jdata)
		if baerr != nil {
			sendErrorRsp(w, "Error marshalling HSM query data",
				"ERROR: problem marshalling HSM query data.",
				r.URL.Path, http.StatusInternalServerError)
			return
		}
		rsp, rerr = doHSMPutPostPatchDel(appParams.SmdURL+urlTail, http.MethodPost, ba)
	}
	if rerr != nil {
		sendErrorRsp(w, "Can't get HSM component data",
			"ERROR: problem getting component info from HSM.",
			r.URL.Path, http.StatusInternalServerError)
		return
	}
	if rsp == nil {
		sendErrorRsp(w, "No HSM component data",
			"ERROR: Nil response data from HSM.",
			r.URL.Path, http.StatusInternalServerError)
		return
	}

	var compData hsmComponentList
	rerr = json.Unmarshal(rsp, &compData)
	if rerr != nil {
		sendErrorRsp(w, "Can't unmarshall HSM component data",
			"ERROR: Problem unmarshaling HSM data.",
			r.URL.Path, http.StatusInternalServerError)
		return
	}

	for ii := 0; ii < len(compData.Components); ii++ {
		if xnametypes.IsHMSTypeController(xnametypes.GetHMSType(compData.Components[ii].ID)) {
			if goodHSMState(compData.Components[ii].State) {
				retXnames = append(retXnames, compData.Components[ii].ID)
			}
		}
	}

	//For each XName, get the BMC creds from vault.  NOTE: this is SLOW
	//on larger systems.  Nothing we can really do about that.
	for ii := 0; ii < len(retXnames); ii++ {
		creds, err := compCredStore.GetCompCred(retXnames[ii])
		if err != nil {
			logger.Errorf("Error getting credentials for '%s': %v",
				retXnames[ii], err)
			retData.Targets = append(retData.Targets, bmcCredsData{Xname: retXnames[ii],
				StatusCode: http.StatusInternalServerError,
				StatusMsg:  "No credentials found.",
			})
		} else {
			statusCode := http.StatusOK
			statusMsg := "OK"
			un := EMPTY
			pw := EMPTY
			if creds.Username != "" {
				un = creds.Username
			}
			if creds.Password != "" {
				pw = creds.Password
			}

			if (un == EMPTY || pw == EMPTY) && creds.Xname == "" {
				statusCode = http.StatusNotFound
				statusMsg = "Not Found"
			}

			retData.Targets = append(retData.Targets, bmcCredsData{
				Xname:      retXnames[ii],
				Username:   un,
				Password:   pw,
				StatusCode: statusCode,
				StatusMsg:  statusMsg,
			})
		}
	}

	ba, berr := json.Marshal(&retData)
	if berr != nil {
		sendErrorRsp(w, "Return data marshal error", "ERROR: problem marshaling return data.",
			r.URL.Path, http.StatusInternalServerError)
		return
	}

	w.Header().Set(CT_TYPE, CT_APPJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

func doHSMDiscover(xnames []string) ([]byte, error) {
	if len(xnames) == 0 {
		return nil, nil
	}
	if hsmClient == nil {
		hsmTransport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		hsmClient = &http.Client{Transport: hsmTransport}
	}
	url := appParams.SmdURL + "/Inventory/Discover"

	var payloadS discoverPayload
	payloadS.Xnames = xnames
	payloadS.Force = true
	payload, marErr := json.Marshal(payloadS)
	if marErr != nil {
		logger.Errorf("attemtped to marshal JSON payload but failed: %s", marErr)
		return nil, marErr
	}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	base.SetHTTPUserAgent(req, serviceName)
	req.Header.Add("Content-Type", "application/json")
	rsp, err := hsmClient.Do(req)
	defer base.DrainAndCloseResponseBody(rsp)
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
