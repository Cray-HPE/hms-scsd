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
	"testing"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	compcreds "github.com/Cray-HPE/hms-compcredentials"
	sstorage "github.com/Cray-HPE/hms-securestorage"
)


func TestDoCredsGet(t *testing.T) {
	var bmcCredsData bmcCredsReturn

	loggerSetup()

	//Set up fake cred store

	xnameList := []string{"x0c0s0b0","x1c0s0b0","x2c0s0b0","x3c0s0b0"}
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

	//Do an HTTP call

	rr1 := httptest.NewRecorder()
	handler1 := http.HandlerFunc(doCredsGet)
	req1_payload := bytes.NewBufferString("")
	req1,err1 := http.NewRequest("GET","http://localhost:8080/v1/bmc/creds?targets=x0c0s0b0",req1_payload)
	if (err1 != nil) {
		t.Fatal(err1)
	}
	handler1.ServeHTTP(rr1,req1)
	if (rr1.Code != http.StatusOK) {
		t.Fatalf("HTTP handler 'doCredsGet' got bad status: %d, want %d",
			rr1.Code,http.StatusOK)
	}
	body,berr := ioutil.ReadAll(rr1.Body)
	if (berr != nil) {
		t.Fatalf("ERROR reading GET response body: %v",berr)
	}
	berr = json.Unmarshal(body,&bmcCredsData)
	if (berr != nil) {
		t.Fatalf("ERROR unmarshalling GET response: %v",berr)
	}

	if (len(bmcCredsData.Targets) != 1) {
		t.Fatalf("ERROR Target list length should be 1, is: %d",
			len(bmcCredsData.Targets))
	}

	expXname := "x0c0s0b0"
	expUser := "user_"+expXname
	expPass := "pw_"+expXname

	if (bmcCredsData.Targets[0].Xname != expXname) {
		t.Errorf("ERROR mismatched xname, exp: '%s', got: '%s'",
			expXname,bmcCredsData.Targets[0].Xname)
	}
	if (bmcCredsData.Targets[0].Username != expUser) {
		t.Errorf("ERROR mismatched username, exp: '%s', got: '%s'",
			expUser,bmcCredsData.Targets[0].Username)
	}
	if (bmcCredsData.Targets[0].Password != expPass) {
		t.Errorf("ERROR mismatched password, exp: '%s', got: '%s'",
			expPass,bmcCredsData.Targets[0].Password)
	}

	appParams.VaultEnable = nil
}

