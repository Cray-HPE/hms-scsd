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
	"fmt"
	"log"
	"net/http"

	"github.com/Cray-HPE/hms-certs/pkg/hms_certs"
)

var sessionAuthPath = "/redfish/v1/SessionService/Sessions"
var accountsPath = "/redfish/v1/AccountService/Accounts"
var rfClient *hms_certs.HTTPClientPair

const (
	dfltMaxHTTPRetries = 5
	dfltMaxHTTPTimeout = 40
	dfltMaxHTTPBackoff = 8
)

func HPEPDUPasswordChange(pdu, user, oldpass, newpass string) {
	rfClient, _ = hms_certs.CreateRetryableHTTPClientPair("", dfltMaxHTTPTimeout, dfltMaxHTTPRetries, dfltMaxHTTPBackoff)
	sessionAuthToken := GetAuthToken(pdu, user, oldpass)
	tempUser := AddUser(pdu, user+"_temp", oldpass, sessionAuthToken)
	sessionAuthToken = GetAuthToken(pdu, tempUser, oldpass)
	DeleteUser(pdu, user, sessionAuthToken)
	AddUser(pdu, user, newpass, sessionAuthToken)
	sessionAuthToken = GetAuthToken(pdu, user, newpass)
	DeleteUser(pdu, tempUser, sessionAuthToken)
}

func AddUser(host, newuser, password, sessionAuthToken string) (newusername string) {
	addUserBody := fmt.Sprintf(`{"username":"%s","password":"%s","email":"none@none.com","chkenable":true,"frpasschk":true,"rolename":"admin","temperature":1}`, newuser, password)
	addUserPath := fmt.Sprintf("https://%s%s", host, accountsPath)
	log.Printf("AddUser with: POST %s, Data: %s", addUserPath, addUserBody)
	// create the request
	req, err := http.NewRequest("POST", addUserPath, bytes.NewBuffer([]byte(addUserBody)))
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Token", sessionAuthToken)

	// execute the request
	//rfClientLock.RLock()
	rsp, err := rfClient.Do(req)
	//rfClientLock.RUnlock()
	if err != nil {
		log.Printf("POST %s\n Body           --> %s\nNetwork Error: %s",
			addUserPath, addUserBody, err)
		return
	}
	defer rsp.Body.Close()
	newusername = newuser

	return newusername
}

func DeleteUser(host, user, sessionAuthToken string) {
	deleteUserPath := fmt.Sprintf("https://%s%s/%s", host, accountsPath, user)
	log.Printf("DeleteUser with: DELETE %s", deleteUserPath)
	// create the request
	req, err := http.NewRequest("DELETE", deleteUserPath, bytes.NewBuffer([]byte("")))
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Token", sessionAuthToken)

	// execute the request
	//rfClientLock.RLock()
	rsp, err := rfClient.Do(req)
	//rfClientLock.RUnlock()
	if err != nil {
		log.Printf("DELETE %s\nNetwork Error: %s",
			deleteUserPath, err)
		return
	}
	defer rsp.Body.Close()

	return
}

func GetAuthToken(host, user, password string) (sessionAuthToken string) {
	sessionAuthToken = ""
	sessionAuthBody := fmt.Sprintf(`{"username":"%s","password":"%s"}`, user, password)
	getAuthPath := fmt.Sprintf("https://%s%s", host, sessionAuthPath)
	req, err := http.NewRequest("POST", getAuthPath, bytes.NewBuffer([]byte(sessionAuthBody)))
	log.Printf("GetAuthToken with: POST %s, Data: %s", getAuthPath, sessionAuthBody)
	req.SetBasicAuth(user, password)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-Type", "application/json")
	// execute the request
	//	rfClientLock.RLock()
	rsp, err := rfClient.Do(req)
	//	rfClientLock.RUnlock()
	if err != nil {
		log.Printf("POST %s\nNetwork Error: %s",
			getAuthPath, err)
		return
	}
	defer rsp.Body.Close()
	sessionAuthToken = rsp.Header.Get("X-Auth-Token")
	log.Printf("SessionAuthToken %s", sessionAuthToken)
	return
}
