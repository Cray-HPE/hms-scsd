// MIT License
//
// (C) Copyright [2021] Hewlett Packard Enterprise Development LP
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
	"testing"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"

	compcreds "github.com/Cray-HPE/hms-compcredentials"
	sstorage "github.com/Cray-HPE/hms-securestorage"
)


func popAllCompData(rdata *hsmComponentList) {
	comp := hsmComponent{ID:"x0c0s0b0",Type:"NodeBMC",State:"On",Flag:"OK"}
	rdata.Components = append(rdata.Components,comp)
	comp = hsmComponent{ID:"x3c0s0b0",Type:"NodeBMC",State:"On",Flag:"OK"}
	rdata.Components = append(rdata.Components,comp)
	comp = hsmComponent{ID:"x1c0b0",Type:"ChassisBMC",State:"On",Flag:"OK"}
	rdata.Components = append(rdata.Components,comp)
	comp = hsmComponent{ID:"x2c0r0b0",Type:"RouterBMC",State:"On",Flag:"OK"}
	rdata.Components = append(rdata.Components,comp)
}

func smCompStuff(w http.ResponseWriter, req *http.Request) {
	var retData hsmComponentList

	if (strings.Contains(req.URL.Path,"Query")) {
		var jdata hsmComponentQuery

		//This is a /State/Components/Query POST.  Get the JSON data
		//and return the appropriate stuff.  If ComponentIDs has one entry
		//and it's "all", return a fake "everything" list of the requested
		//type.  Else, return what they ask for.

		if (req.Method != http.MethodPost) {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body,err := ioutil.ReadAll(req.Body)
		if (err != nil) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = json.Unmarshal(body,&jdata)
		if (err != nil) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if (len(jdata.ComponentIDs) < 1) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if (strings.ToLower(jdata.ComponentIDs[0]) == "all") {
			popAllCompData(&retData)
		} else {
			//Use the components in the request
			for ix,_ := range(jdata.ComponentIDs) {
				comp := hsmComponent{ID:jdata.ComponentIDs[ix],Type:"NodeBMC",State:"On",Flag:"OK"}
				retData.Components = append(retData.Components,comp)
			}
		}

		ba,baerr := json.Marshal(&retData)
		if (baerr != nil) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(ba)
	} else {
		//This is a /State/Components GET.  Parse out the query params.
		xnames := req.URL.Query()["id"]

		if (len(xnames) == 0) {
			popAllCompData(&retData)
		} else {
			for ix,_ := range(xnames) {
				comp := hsmComponent{ID:xnames[ix],Type:"NodeBMC",State:"On",Flag:"OK"}
				retData.Components = append(retData.Components,comp)
			}
		}

		ba,baerr := json.Marshal(&retData)
		if (baerr != nil) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(ba)
	}
}

func TestDoCredsGet(t *testing.T) {
	var bmcCredsRet bmcCredsReturn

	loggerSetup()
appParams.LogLevel = "TRACE"
	routes := generateRoutes()
	router := newRouter(routes)
	if len(routes) == 0 {
		t.Errorf("No routes generated.")
	}
	if router == nil {
		t.Errorf("No router generated.")
	}


	//Set up fake cred store.  NOTE: these have to be in the same order
	//that they are expected to be read out!

	xnameList := []string{"x0c0s0b0","x3c0s0b0","x1c0b0","x2c0r0b0"}
	ss,adapter := sstorage.NewMockAdapter()
	compCredStore = compcreds.NewCompCredStore("secret/hms-cred",ss)
	mockLData := []sstorage.MockLookup{}

	for _,xn := range(xnameList) {
		ssItem := sstorage.MockLookup{
			Output: sstorage.OutputLookup{
				Output: &compcreds.CompCredentials{Xname: xn,
				                                   Username: "user_"+xn,
				                                   Password: "pw_"+xn,
				},
				Err: nil,
			},
		}

		mockLData = append(mockLData,ssItem)
	}

	adapter.LookupNum = 0
	adapter.LookupData = mockLData

	ve := true
	appParams.VaultEnable = &ve

	//set up a fake HSM receiver func and pass the URL into appParams.SmdURL.
	smServer := httptest.NewServer(http.HandlerFunc(smCompStuff))
	appParams.SmdURL = smServer.URL
	defer smServer.Close()

	//Do an HTTP call

	rr1 := httptest.NewRecorder()
	//handler1 := http.HandlerFunc(doCredsGet)
	req1_payload := bytes.NewBufferString("")
	req1,err1 := http.NewRequest("GET","http://localhost:8080/v1/bmc/creds?targets=x0c0s0b0,x3c0s0b0&type=NodeBMC",req1_payload)
	if (err1 != nil) {
		t.Fatal(err1)
	}
	//handler1.ServeHTTP(rr1,req1)
	router.ServeHTTP(rr1,req1)
	if (rr1.Code != http.StatusOK) {
		t.Fatalf("HTTP handler 'doCredsGet' got bad status: %d, want %d",
			rr1.Code,http.StatusOK)
	}
	body,berr := ioutil.ReadAll(rr1.Body)
	if (berr != nil) {
		t.Fatalf("ERROR reading GET response body: %v",berr)
	}
	berr = json.Unmarshal(body,&bmcCredsRet)
	if (berr != nil) {
		t.Fatalf("ERROR unmarshalling GET response: %v",berr)
	}

	if (len(bmcCredsRet.Targets) != 2) {
		t.Fatalf("ERROR Target list length should be 2, is: %d",
			len(bmcCredsRet.Targets))
	}

	exps := []string{"x0c0s0b0","x3c0s0b0"}
	actMap := make(map[string]*bmcCredsData)
	for ix,_ := range(bmcCredsRet.Targets) {
		actMap[bmcCredsRet.Targets[ix].Xname] = &bmcCredsRet.Targets[ix]
	}

	for _,expXname := range(exps) {
		expUser := "user_"+expXname
		expPass := "pw_"+expXname

		ptr,ok := actMap[expXname]
		if (!ok) {
			t.Fatalf("ERROR, component missing in return data map: '%s'",expXname)
		}
		if (ptr.Xname != expXname) {
			t.Errorf("ERROR mismatched xname, exp: '%s', got: '%s'",
				expXname,ptr.Xname)
		}
		if (ptr.Username != expUser) {
			t.Errorf("ERROR mismatched username, exp: '%s', got: '%s'",
				expUser,ptr.Username)
		}
		if (ptr.Password != expPass) {
			t.Errorf("ERROR mismatched password, exp: '%s', got: '%s'",
				expPass,ptr.Password)
		}
	}

	//Do an "all" call

	adapter.LookupNum = 0
	rr1 = httptest.NewRecorder()
	req1,err1 = http.NewRequest("GET","http://localhost:8080/v1/bmc/creds",req1_payload)
	if (err1 != nil) {
		t.Fatal(err1)
	}
	//handler1.ServeHTTP(rr1,req1)
	router.ServeHTTP(rr1,req1)
	if (rr1.Code != http.StatusOK) {
		t.Fatalf("HTTP handler 'doCredsGet' got bad status: %d, want %d",
			rr1.Code,http.StatusOK)
	}
	body,berr = ioutil.ReadAll(rr1.Body)
	if (berr != nil) {
		t.Fatalf("ERROR reading GET response body: %v",berr)
	}
	berr = json.Unmarshal(body,&bmcCredsRet)
	if (berr != nil) {
		t.Fatalf("ERROR unmarshalling GET response: %v",berr)
	}

	if (len(bmcCredsRet.Targets) == 0) {
		t.Fatalf("ERROR Target list length should not be 0")
	}

	exps = []string{"x0c0s0b0","x3c0s0b0","x1c0b0","x2c0r0b0"}
	actMap2 := make(map[string]*bmcCredsData)
	for ix,_ := range(bmcCredsRet.Targets) {
		actMap2[bmcCredsRet.Targets[ix].Xname] = &bmcCredsRet.Targets[ix]
	}
	for _,expXname := range(exps) {
		expUser := "user_"+expXname
		expPass := "pw_"+expXname

		ptr,ok := actMap2[expXname]
		if (!ok) {
			t.Fatalf("ERROR, component missing in return data map: '%s'",expXname)
		}
		if (ptr.Xname != expXname) {
			t.Errorf("ERROR mismatched xname, exp: '%s', got: '%s'",
				expXname,ptr.Xname)
		}
		if (ptr.Username != expUser) {
			t.Errorf("ERROR mismatched username, exp: '%s', got: '%s'",
				expUser,ptr.Username)
		}
		if (ptr.Password != expPass) {
			t.Errorf("ERROR mismatched password, exp: '%s', got: '%s'",
				expPass,ptr.Password)
		}
	}

	appParams.VaultEnable = nil
}

