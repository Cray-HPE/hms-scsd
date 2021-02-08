/*
 * MIT License
 *
 * (C) Copyright [2020-2021] Hewlett Packard Enterprise Development LP
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 * THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 * OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 * ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 * OTHER DEALINGS IN THE SOFTWARE.
 */

package main

import (
    "testing"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	trsapi "stash.us.cray.com/HMS/hms-trs-app-api/pkg/trs_http_api"
	base "stash.us.cray.com/HMS/hms-base"
	"stash.us.cray.com/HMS/hms-certs/pkg/hms_certs"
)

func TestRouting(t *testing.T) {
	loggerSetup()
	routes := generateRoutes()
	router := newRouter(routes)
	if (len(routes) == 0) {
		t.Errorf("No routes generated.")
	}
	if (router == nil) {
		t.Errorf("No router generated.")
	}
}

func TestStatusCodeOK(t *testing.T) {
	loggerSetup()
	if (!statusCodeOK(http.StatusOK)) {
		t.Errorf("StatusOK failed.")
	}
	if (!statusCodeOK(http.StatusCreated)) {
		t.Errorf("StatusCreated failed.")
	}
	if (!statusCodeOK(http.StatusAccepted)) {
		t.Errorf("StatusAccepted failed.")
	}
	if (!statusCodeOK(http.StatusNoContent)) {
		t.Errorf("StatusNoContent failed.")
	}
	if (statusCodeOK(http.StatusBadRequest)) {
		t.Errorf("StatusBadRequest did NOT fail.")
	}
}

func TestStatusMsg(t *testing.T) {
	loggerSetup()
	msg := statusMsg(http.StatusOK)
	if (msg != "OK") {
		t.Errorf("StatusOK generated incorrect message: %s",msg)
	}
	msg = statusMsg(600)
	if (msg != http.StatusText(http.StatusInternalServerError)) {
		t.Errorf("Error code 600 generated bad msg: '%s'",msg)
	}
}

func TestGetStatusCode(t *testing.T) {
	loggerSetup()
	var task trsapi.HttpTask
	var code int
	var err error
	task.Request,err = http.NewRequest("GET","http://blah/blah",nil)
	if (err != nil) {
		t.Errorf("Got error from NewRequest: %v",err)
	}
	code = getStatusCode(&task)
	if (code != 204) {
		t.Errorf("Code should be zero, got: %d",code)
	}
	task.Request.Response = &http.Response{StatusCode: 200}
	code = getStatusCode(&task)
	if (code != 200) {
		t.Errorf("Code should be 200, got: %d",code)
	}
	lerr := fmt.Errorf("Error place holder")
	task.Err = &lerr
	task.Request.Response = nil
	code = getStatusCode(&task)
	if (code != 500) {
		t.Errorf("Code should be 500, got: %d",code)
	}
}

func TestTargFromTask(t *testing.T) {
	loggerSetup()
	var task trsapi.HttpTask
	var err error
	targ := targFromTask(&task)
	if (targ != "BADURL") {
		t.Errorf("Target should be BADURL, got: '%s'",targ)
	}
	task.Request,err = http.NewRequest("GET","http://x0c1s2b3/blah",nil)
	if (err != nil) {
		t.Errorf("Got error from NewRequest: %v",err)
	}
	targ = targFromTask(&task)
	if (targ != "x0c1s2b3") {
		t.Errorf("Target should be x0c1s2b3, got: '%s'",targ)
	}
}

func TestCheckStatusCodes(t *testing.T) {
	loggerSetup()
	var err error
	var tasks = make([]trsapi.HttpTask,3)

	for ii := 0; ii < 3; ii++ {
		url := fmt.Sprintf("http:x0c0s%db0/blah",ii)
		tasks[ii].Request,err = http.NewRequest("GET",url,nil)
		if (err != nil) {
			t.Errorf("Got error from NewRequest: %v",err)
		}
		tasks[ii].Request.Response = &http.Response{StatusCode: 200+(ii*50)}
	}

	err = checkStatusCodes(tasks)
	if (err == nil) {
		t.Errorf("Bad status codes should have failed.")
	}
}

func TestPopulateTaskList(t *testing.T) {
	loggerSetup()
	var targs = []string{"x0c0s0b0","x0c0s1b0","x0c0s2b0",}
	urlTail := "/redfish/v1/"
	var tasks = make([]trsapi.HttpTask,3)

	ve := false
	appParams.VaultEnable = &ve
	populateTaskList(tasks,targs,urlTail,"GET",nil)

	for ii := 0; ii < 3; ii++ {
		if (tasks[ii].Request.Method != "GET") {
			t.Errorf("Task %d bad method: %s.",ii,tasks[ii].Request.Method)
		}
		if (tasks[ii].Request.URL.Path != urlTail) {
			t.Errorf("Task %d bad url, exp: '%s', got '%s'",
				ii,urlTail,tasks[ii].Request.URL.Path)
		}
	}
}

func TestGoodHSMState(t *testing.T) {
	loggerSetup()
	if (!goodHSMState(base.StateOn.String())) {
		t.Errorf("HSM state ON is not good.")
	}
	if (!goodHSMState(base.StateReady.String())) {
		t.Errorf("HSM state READY is not good.")
	}
	if (goodHSMState(base.StateOff.String())) {
		t.Errorf("HSM state OFF should not be good.")
	}
}

func TestMakeTargData(t *testing.T) {
	loggerSetup()
	var targs = []string{"x0c0s0b0","x0c0s1b0",}

	tinfo := makeTargData(targs)
	if (len(tinfo) != 2) {
		t.Errorf("Targ list len should be 2, is %d",len(tinfo))
	}
	for ii := 0; ii < 2; ii++ {
		if (tinfo[ii].target != targs[ii]) {
			t.Errorf("Targ item %d wrong value: exp '%s', got:'%s'",
				ii,targs[0],tinfo[0].target)
		}
	}
}

func TestRemoveBadTargs(t *testing.T) {
	loggerSetup()
	var err error
	var tasks = make([]trsapi.HttpTask,3)

	for ii := 0; ii < 3; ii++ {
		url := fmt.Sprintf("http://x0c0s%db0/blah",ii)
		tasks[ii].Request,err = http.NewRequest("GET",url,nil)
		if (err != nil) {
			t.Errorf("Got error from NewRequest: %v",err)
		}
		tasks[ii].Request.Response = &http.Response{StatusCode: 200+(ii*50)}
	}

	newTargs := removeBadTargs(tasks)

	if (len(newTargs) != 1) {
		t.Errorf("Too many non-bad targs, exp: 1, got: %d",len(newTargs))
	}
	if (newTargs[0] != "x0c0s0b0") {
		t.Errorf("Wrong non-bad targ, exp: x0c0s0b0, got: %s",newTargs[0])
	}
}

func TestMakeTarglistFromCreds(t *testing.T) {
	var creds = make([]credsTarg,2)
	creds[0] = credsTarg{Xname: "x0c0s0b0",
	                     Creds: credsData{Username: "u_s0b0",
	                                      Password: "p_s0b0",},
	}
	creds[1] = credsTarg{Xname: "x0c0s1b0",
	                     Creds: credsData{Username: "u_s1b0",
	                                      Password: "p_s1b0",},
	}

	tlist := makeTarglistFromCreds(creds)

	if (len(tlist) != 2) {
		t.Errorf("Wrong length targlist, exp: 2, got: %d",len(tlist))
	}
	if (tlist[0] != "x0c0s0b0") {
		t.Errorf("Wrong targlist[0], exp x0c0s0b0, got %s",tlist[0])
	}
	if (tlist[1] != "x0c0s1b0") {
		t.Errorf("Wrong targlist[1], exp x0c0s0b0, got %s",tlist[0])
	}
}

func TestCARoll(t *testing.T) {
	caURI = "/tmp/fakeCA.crt"
	capld1 := `-----BEGIN FAKE CERT-----\nxyzzy_11111_blah\n-----END FAKE CERT-----`
	capld2 := `-----BEGIN FAKE CERT-----\nxyzzy_22222_blah\n-----END FAKE CERT-----`

	tloc = &tlocLocal
	err := tloc.Init("SCSD_TEST",nil)
	if (err != nil) {
		t.Errorf("Error initializing TRS API: %v",err)
	}

	//Create fake CA file
	err = ioutil.WriteFile(caURI,[]byte(capld1),0777)
	if (err != nil) {
		t.Errorf("Error writing CA file '%s': %v",caURI,err)
	}
	err = os.Chmod(caURI,0777)
	if (err != nil) {
		t.Errorf("Can't change '%s' to 0777: %v",caURI,err)
	}

	caUpdateCount = 0
	setupTRSCA()

	if (caUpdateCount != 1) {
		t.Errorf("ERROR, initial CA set didn't take.")
	}

	err = hms_certs.CAUpdateRegister(caURI,caCB)
	if (err != nil) {
		t.Errorf("Error setting up CA update CB: %v",err)
	}

	time.Sleep(1 * time.Second)

	err = ioutil.WriteFile(caURI,[]byte(capld2),0777)
	if (err != nil) {
		t.Errorf("Error writing new CA file '%s': %v",caURI,err)
	}
	err = os.Chmod(caURI,0777)
	if (err != nil) {
		t.Errorf("Can't change '%s' to 0777: %v",caURI,err)
	}

	time.Sleep(40 * time.Second)
	if (caUpdateCount <= 1) {
		t.Errorf("ERROR, CA update didn't work.")
	}
}

