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
	"fmt"
	"net/http"
	"encoding/json"
	"io/ioutil"
	"strings"
)

type healthData struct {
	TaskRunnerStatus string `json:"TaskRunnerStatus"`
	TaskRunnerMode   string `json:"TaskRunnerMode"`
	VaultStatus      string `json:"VaultStatus"`
}

type versionData struct {
	Version string `json:"Version"`
}


// /v1/params GET

func doParamsGet(w http.ResponseWriter, r *http.Request) {
	ba,baerr := json.Marshal(appParams)
	if (baerr != nil) {
		emsg := fmt.Sprintf("ERROR: Problem marshal params data: %v",baerr)
		sendErrorRsp(w,"Parameter data marshal error",emsg,r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	w.Header().Set(CT_TYPE,CT_APPJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

// /v1/params PATCH

func doParamsPatch(w http.ResponseWriter, r *http.Request) {
	var jdata opParams

	body,err := ioutil.ReadAll(r.Body)
	if (err != nil) {
		emsg := fmt.Sprintf("ERROR: Problem reading request body: %v",err)
		sendErrorRsp(w,"Bad request body read",emsg,r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	//verify the umarshaller touched these

	jdata.LogLevel = "xxx"
	jdata.HTTPRetries = -1
	jdata.HTTPTimeout = -1

	err = json.Unmarshal(body,&jdata)
	if (err != nil) {
		emsg := fmt.Sprintf("ERROR: Problem unmarshalling request body: %v",err)
		sendErrorRsp(w,"Request data unmarshal error",emsg,r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	//All is OK, take action.  The only thing settable is debug level

	if (jdata.LogLevel != "xxx") {
		appParams.LogLevel = jdata.LogLevel
		setLogLevel()
	}
	if (jdata.HTTPRetries != -1) {
		appParams.HTTPRetries = jdata.HTTPRetries
	}
	if (jdata.HTTPTimeout != -1) {
		appParams.HTTPTimeout = jdata.HTTPTimeout
	}
	oldve := appParams.VaultEnable
	if (jdata.VaultEnable != nil) {
		ve := *jdata.VaultEnable
		appParams.VaultEnable = &ve
	}

	if ((appParams.VaultEnable != nil) && *appParams.VaultEnable) {
		if ((oldve == nil) || (*oldve == false)) {
			//We're enabling the vault cred store, initialize it if
			//it's not already initialized.
			go setupVault()
		}
	}

	ba,baerr := json.Marshal(appParams)
	if (baerr != nil) {
		emsg := fmt.Sprintf("ERROR: Problem marshal params data: %v",baerr)
		sendErrorRsp(w,"Parameter data marshal error",emsg,r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	w.Header().Set(CT_TYPE,CT_APPJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

// /v1/health GET

func doHealthGet(w http.ResponseWriter, r *http.Request) {
	var health healthData

	alive,err := tloc.Alive()
	if ((err == nil) && alive) {
		health.TaskRunnerStatus = "Running"
	} else {
		health.TaskRunnerStatus = "Not Running"
	}
	logger.Infof("Health: task runner status: %s",health.TaskRunnerStatus)

	if (appParams.LocalMode) {
		health.TaskRunnerMode = "Local"
	} else {
		health.TaskRunnerMode = "Worker"
	}
	logger.Infof("Health: task runner mode: %s",health.TaskRunnerMode)

	if ((appParams.VaultEnable != nil) && *appParams.VaultEnable) {
		if (compCredStore == nil) {
			health.VaultStatus = "Not Connected"
		} else {
			health.VaultStatus = "Connected"
		}
	} else {
		health.VaultStatus = "Disabled"
	}
	logger.Infof("Health: Vault status: %s",health.VaultStatus)

	ba,baerr := json.Marshal(health)
	if (baerr != nil) {
		emsg := fmt.Sprintf("ERROR: Problem marshal health data: %v",baerr)
		sendErrorRsp(w,"Health data marshal error",emsg,r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	w.Header().Set(CT_TYPE,CT_APPJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

// /v1/liveness GET

func doLivenessGet(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// /v1/readiness GET

func doReadinessGet(w http.ResponseWriter, r *http.Request) {
	ready := true

	alive,err := tloc.Alive()
	if (!alive || (err != nil)) {
		logger.Infof("Readiness check: TRS not alive, err: %v",err)
		ready = false
	}

	if ((appParams.VaultEnable != nil) && *appParams.VaultEnable && (compCredStore == nil)) {
		logger.Infof("Readiness check: Vault not connected")
		ready = false
	}

	if (ready) {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}

// /v1/health GET

func doVersionGet(w http.ResponseWriter, r *http.Request) {
	var ver versionData

	content,err := ioutil.ReadFile("/var/run/scsd_version.txt")
	if (err != nil) {
		emsg := fmt.Sprintf("ERROR: Problem reading version file: %v",err)
		sendErrorRsp(w,"Version data error",emsg,r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	ver.Version = strings.TrimSpace(string(content))
	ba,baerr := json.Marshal(&ver)
	if (baerr != nil) {
		emsg := fmt.Sprintf("ERROR: Problem marshalling version data: %v",baerr)
		sendErrorRsp(w,"Health data marshal error",emsg,r.URL.Path,
			http.StatusInternalServerError)
		return
	}

	w.Header().Set(CT_TYPE,CT_APPJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

