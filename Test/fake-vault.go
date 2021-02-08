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
    "encoding/json"
    "net/http"
	"log"
    "io/ioutil"
	"os"
	"strings"
)


type vaultTokStuff struct {
	Auth vtsAuth `json:"auth"`
}

type vtsAuth struct {
	ClientToken string `json:"client_token"`
}

type vaultCertReq struct {
	CommonName string `json:"common_name"`
	TTL        string `json:"ttl"`
	AltNames   string `json:"alt_names"`
}

type VaultCertData struct {
	RequestID     string   `json:"request_id"`
	LeaseID       string   `json:"lease_id"`
	Renewable     bool     `json:"renewable"`
	LeaseDuration int      `json:"lease_duration"`
	Data          CertInfo `json:"data"`
}

type CertInfo struct {
	CAChain        []string `json:"ca_chain"`
	Certificate    string   `json:"certificate"`
	Expiration     int      `json:"expiration"`
	IssuingCA      string   `json:"issuing_ca"`
	PrivateKey     string   `json:"private_key"`
	PrivateKeyType string   `json:"private_key_type"`
	SerialNumber   string   `json:"serial_number"`
}


func k8sLogin(w http.ResponseWriter, r *http.Request) {
	if (r.Method != "POST") {
        log.Printf("ERROR: request is not a POST.\n")
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

	_,err := ioutil.ReadAll(r.Body)
	if (err != nil) {
		log.Printf("ERROR reading req body: %v",err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//Don't care about the request data.  Just gin up the response.

	jdata := vaultTokStuff{Auth: vtsAuth{ClientToken: "CLIENT_TOKEN",},}
	ba,baerr := json.Marshal(&jdata)
	if (baerr != nil) {
		log.Printf("ERROR marshalling rsp data: %v",baerr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

func pkiCerts(w http.ResponseWriter, r *http.Request) {
	var jdata vaultCertReq
	var pkiDataStr = `
{
  "request_id": "dead1562-4a3c-6828-9951-d85d1997e0ce",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "ca_chain": [
      "-----BEGIN CERTIFICATE-----\naaa\n-----END CERTIFICATE-----",
      "-----BEGIN CERTIFICATE-----\nbbb\n-----END CERTIFICATE-----"
    ],
    "certificate": "-----BEGIN CERTIFICATE-----\nccc\n-----END CERTIFICATE-----",
    "expiration": 1627423464,
    "issuing_ca": "-----BEGIN CERTIFICATE-----\nddd\n-----END CERTIFICATE-----",
    "private_key": "-----BEGIN RSA PRIVATE KEY-----\neee\n-----END RSA PRIVATE KEY-----",
    "private_key_type": "rsa",
    "serial_number": "4f:fe:98:c2:0d:d4:1e:bb:50:75:8b:94:fe:b9:48:89:b6:d4:7f:86"
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}`


	if (r.Method != "POST") {
        log.Printf("ERROR: request is not a POST.\n")
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

	body,err := ioutil.ReadAll(r.Body)
	if (err != nil) {
		log.Printf("ERROR reading req body: %v",err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body,&jdata)
	if (err != nil) {
		log.Printf("ERROR un-marshalling req data: %v",err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if (jdata.CommonName == "") {
		log.Printf("ERROR: Cert request has no common name.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if (jdata.TTL == "") {
		log.Printf("ERROR: Cert request has no TTL.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if (jdata.AltNames == "") {
		log.Printf("ERROR: Cert request has no SANs.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("Cert Req: CommonName: '%s'",jdata.CommonName)
	log.Printf("Cert Req: TTL:        '%s'",jdata.TTL)
	log.Printf("Cert Req: AltNames:   '%s'",jdata.AltNames)

	//Send back the fake cert data

	rstr := strings.Replace(pkiDataStr,"aaa",(jdata.CommonName+"aaa"),-1)
	rstr  = strings.Replace(rstr,"bbb",(jdata.CommonName+"bbb"),-1)
	rstr  = strings.Replace(rstr,"ccc",(jdata.CommonName+"ccc"),-1)
	rstr  = strings.Replace(rstr,"ddd",(jdata.CommonName+"ddd"),-1)
	rstr  = strings.Replace(rstr,"eee",(jdata.CommonName+"eee"),-1)

	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(rstr))
}

func pkiCAChain(w http.ResponseWriter, r *http.Request) {
	if (r.Method != "GET") {
        log.Printf("ERROR: request is not a GET.\n")
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

	caChainStr := `-----BEGIN CERTIFICATE-----\n11223344\n55667788\n-----END CERTIFICATE-----\n`

	//w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(caChainStr))
}

func main() {
	port := ":8200"
	envstr := os.Getenv("PORT")
	if (envstr != "") {
		port = envstr
		if (!strings.Contains(port,":")) {
			port = ":" + port
		}
	}

	urlFront := "http://10.0.2.15"+port

	klogURL := "/v1/auth/kubernetes/login"
	pkiCertURL := "/v1/pki_common/issue/pki-common"
	caChainURL := "/v1/pki_common/ca_chain"
	http.HandleFunc(klogURL,k8sLogin)
	http.HandleFunc(pkiCertURL,pkiCerts)
	http.HandleFunc(caChainURL,pkiCAChain)

	log.Printf("Listening on: %s",urlFront)
	log.Printf("URLs: %s",klogURL)
	log.Printf("      %s",pkiCertURL)
	log.Printf("      %s",caChainURL)

	srv := &http.Server{Addr: port,}
	srv.ListenAndServe()
}


