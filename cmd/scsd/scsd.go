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
	"os"
	"os/signal"
	"syscall"
	"strconv"
	"strings"
	"sync"
	"net/http"
	"context"
	"time"

	"github.com/sirupsen/logrus"
	trsapi "stash.us.cray.com/HMS/hms-trs-app-api/pkg/trs_http_api"
	compcreds "stash.us.cray.com/HMS/hms-compcredentials"
	sstorage "stash.us.cray.com/HMS/hms-securestorage"
	"stash.us.cray.com/HMS/hms-certs/pkg/hms_certs"
	"stash.us.cray.com/HMS/hms-base"
)


// Operational parameters

type opParams struct {
	LogLevel       string `json:"LogLevel"`
	LocalMode      bool   `json:"LocalMode"`		//read-only
	KafkaURL       string `json:"KafkaURL"`			//read-only
	SmdURL         string `json:"SmdURL"`			//read-only
	HTTPListenPort string `json:"HTTPListenPort"`	//read-only
	HTTPRetries    int    `json:"HTTPRetries"`
	HTTPTimeout    int    `json:"HTTPTimeout"`
	UUID           string `json:"UUID"`				//read-only
	VaultEnable    *bool  `json:"VaultEnable"`
}

const (
	LOGLVL_TRACE = "TRACE"
	LOGLVL_INFO  = "INFO"
	LOGLVL_DEBUG = "DEBUG"
	LOGLVL_WARN  = "WARN"
	LOGLVL_ERROR = "ERROR"
	LOGLVL_FATAL = "FATAL"
	LOGLVL_PANIC = "PANIC"
)


var appParams = opParams{LogLevel: LOGLVL_ERROR,
                         LocalMode: true,
                         KafkaURL: "",
                         SmdURL: "http://cray-smd/hsm/v1",
                         HTTPListenPort: ":25309",
                         HTTPRetries: 5,
                         HTTPTimeout: 15,
                         UUID: "0",
                         VaultEnable: new(bool),
}
var tloc trsapi.TrsAPI
var tlocLocal trsapi.TRSHTTPLocal
var tlocRemote trsapi.TRSHTTPRemote
var VaultKeypath string
var Running = true
var dfltHTTP = false // for testing
var caURI string
var vaultCAURL string
var vaultPKIURL string
var dfltProtocol = "https"
var serviceName = "scsd"
var logger *logrus.Logger

//Test stuff
var test_k8sAuthUrl string
var test_vaultJWTFile string
var test_vaultPKIUrl string
var test_vaultCAUrl string

var compCredStore *compcreds.CompCredStore

var rfClientLock sync.Mutex
var caUpdateCount int


/////////////////////////////////////////////////////////////////////////////
// Convenience function to parse an integer-based environment variable.
//
// envvar(in): Env variable string
// pval(out):  Ptr to an integer to hold the result.
// Return:     true if env var was seen, else false.
/////////////////////////////////////////////////////////////////////////////

func __env_parse_int(envvar string, pval *int) bool {
	var val string
	if val = os.Getenv(envvar); val != "" {
		ival, err := strconv.ParseUint(val, 0, 64)
		if err != nil {
			logger.Errorf("Invalid %s value '%s'.", envvar, val)
		} else {
			*pval = int(ival)
		}
		return true
	}
	return false
}

/////////////////////////////////////////////////////////////////////////////
// Convenience function to parse a boolean-based environment variable.
//
// envvar(in): Env variable string
// pval(out):  Ptr to an integer to hold the result.
// Return:     true if env var was seen, else false.
/////////////////////////////////////////////////////////////////////////////

func __env_parse_bool(envvar string, pval *bool) bool {
	var val string
	if val = os.Getenv(envvar); val != "" {
		lcut := strings.ToLower(val)
		if (lcut == "0") || (lcut == "no") || (lcut == "off") || (lcut == "false") {
			*pval = false
		} else if (lcut == "1") || (lcut == "yes") || (lcut == "on") || (lcut == "true") {
			*pval = true
		} else {
			logger.Errorf("Invalid %s value '%s'.", envvar, val)
		}
		return true
	}
	return false
}

/////////////////////////////////////////////////////////////////////////////
// Convenience function to parse a string-based environment variable.
//
// envvar(in): Env variable string
// pval(out):  Ptr to an integer to hold the result.
// Return:     true if env var was seen, else false.
/////////////////////////////////////////////////////////////////////////////

func __env_parse_string(envvar string, pval *string) bool {
	var val string
	if val = os.Getenv(envvar); val != "" {
		*pval = val
		return true
	}
	return false
}

/////////////////////////////////////////////////////////////////////////////
// Parse environment variables.  Process any that have the "HMNFD_" prefix
// and set operating parameters with their values.
//
// Args, Return: None.
/////////////////////////////////////////////////////////////////////////////

func parseEnvVars() {
	__env_parse_string("SCSD_HTTP_LISTEN_PORT", &appParams.HTTPListenPort)
	__env_parse_int("SCSD_HTTP_RETRIES", &appParams.HTTPRetries)
	__env_parse_int("SCSD_HTTP_TIMEOUT", &appParams.HTTPTimeout)
	__env_parse_string("SCSD_UUID", &appParams.UUID)
	__env_parse_string("SCSD_LOG_LEVEL", &appParams.LogLevel)
	__env_parse_bool("SCSD_LOCAL_MODE", &appParams.LocalMode)
	__env_parse_string("SCSD_KAFKA_URL", &appParams.KafkaURL)
	__env_parse_string("SCSD_SMD_URL", &appParams.SmdURL)
	__env_parse_bool("SCSD_DEFAULT_HTTP", &dfltHTTP)
	__env_parse_string("SCSD_CA_URI", &caURI)
	__env_parse_string("SCSD_VAULT_CA_URL", &vaultCAURL)
	__env_parse_string("SCSD_VAULT_PKI_URL", &vaultPKIURL)

	//These env vars are for vault and need to be named without SCSD_
	//since libraries use them too.

	var ve bool
	__env_parse_string("VAULT_KEYPATH", &VaultKeypath) 
	veseen := __env_parse_bool("VAULT_ENABLE", &ve)
	if (veseen) {
		appParams.VaultEnable = &ve
	}

	//If testing in off-line mode, the following env vars will need to be set:
	//  CRAY_VAULT_JWT_FILE    # e.g. /tmp/k8stoken
	//  CRAY_VAULT_ROLE_FILE   # e.g. also /tmp/k8stoken

	//The following are used only for testing

	__env_parse_string("SCSD_TEST_K8S_AUTH_URL",&test_k8sAuthUrl)
	__env_parse_string("SCSD_TEST_VAULT_JWT_FILE",&test_vaultJWTFile)
	__env_parse_string("SCSD_TEST_VAULT_PKI_URL",&test_vaultPKIUrl)
	__env_parse_string("SCSD_TEST_VAULT_CA_URL",&test_vaultCAUrl)
}

func setupVault() {
	logger.Infof("Connecting to secure store (Vault)...")

	for Running {
		ss,err := sstorage.NewVaultAdapter("")
		if (err != nil) {
			logger.Errorf("Connecting to Vault: %v -- retry in 1 second...",
				err)
			time.Sleep(time.Second)
		} else {
			logger.Infof("Connected to vault.")
			compCredStore = compcreds.NewCompCredStore(VaultKeypath,ss)
			break
		}
	}
}

func setupTRSCA() error {
	if (caURI != "") {
		logger.Infof("setupTRSCA(): Using CA bundle from '%s'",caURI)
	} else {
		logger.Infof("setupTRSCA(): No CA bundle specified.")
	}

	caChain,err := hms_certs.FetchCAChain(caURI)
	if (err != nil) {
		return fmt.Errorf("ERROR fetching CA chain from '%s', falling back to unvalidated https for Redfish communications: %v",
			caURI,err)
	}

	rfClientLock.Lock()
	defer rfClientLock.Unlock()
	rawCert := hms_certs.TupleToNewline(caChain)
	err = tloc.SetSecurity(trsapi.TRSHTTPLocalSecurity{CACertBundleData:rawCert,})
	if (err != nil) {
		return fmt.Errorf("ERROR setting CA chain in HTTP interface, falling back to unvalidated https for Redfish communications.")
	}
	caUpdateCount ++
	return nil
}

func caCB(cbData string) {
	logger.Infof("Updating CA bundle for Redfish HTTP transports.")
	logger.Infof("All API threads paused.")
	err := setupTRSCA()
	if (err != nil) {
		logger.Errorf("%v",err)
	}
	logger.Infof("CA bundle and HTTP transports updated.")
}

func setLogLevel() {
	switch (appParams.LogLevel) {
		case LOGLVL_TRACE:
			logrus.SetLevel(logrus.TraceLevel)
			logger.SetLevel(logrus.TraceLevel)
		case LOGLVL_DEBUG:
			logrus.SetLevel(logrus.DebugLevel)
			logger.SetLevel(logrus.DebugLevel)
		case LOGLVL_INFO:
			logrus.SetLevel(logrus.InfoLevel)
			logger.SetLevel(logrus.InfoLevel)
		case LOGLVL_WARN:
			logrus.SetLevel(logrus.WarnLevel)
			logger.SetLevel(logrus.WarnLevel)
		case LOGLVL_ERROR:
			logrus.SetLevel(logrus.ErrorLevel)
			logger.SetLevel(logrus.ErrorLevel)
		case LOGLVL_FATAL:
			logrus.SetLevel(logrus.FatalLevel)
			logger.SetLevel(logrus.FatalLevel)
		case LOGLVL_PANIC:
			logrus.SetLevel(logrus.PanicLevel)
			logger.SetLevel(logrus.PanicLevel)
		default:
			logger.Printf("Unknown log level '%s', setting to ERROR.",
				appParams.LogLevel)
			appParams.LogLevel = LOGLVL_ERROR
			logrus.SetLevel(logrus.ErrorLevel)
			logger.SetLevel(logrus.ErrorLevel)
	}
}

func main() {
	var err error

	logger = logrus.New()
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true,})
	logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true,})

	venb := true
	appParams.VaultEnable = &venb
	parseEnvVars()

	setLogLevel()

	tloc = &tlocLocal
	if (!appParams.LocalMode) {
		tloc = &tlocRemote
	}

	//Set the default protocol.  Defaults to "https" unless specified otherwise.
	if (dfltHTTP) {
		dfltProtocol = "http"
	}

	//Fetch the service instance name

	serviceName,err = base.GetServiceInstanceName()
	if (err != nil) {
		logger.Errorf("Can't get service instance (hostname)!  Setting to 'SCSD'")
		serviceName = "SCSD"
	}
	logger.Printf("Service Instance Name: '%s'",serviceName)

	err = tloc.Init(serviceName,logger)
	if (err != nil) {
		logger.Errorf("TRS API Init() failed: %v",err)
		//Don't panic -- liveness/readiness will handle this
	}

	// For testing.  ENV VARS relevant:
	//   SCSD_TEST_K8S_AUTH_URL
	//   SCSD_TEST_VAULT_PKI_URL
	//   SCSD_TEST_VAULT_CA_URL
	//   SCSD_TEST_VAULT_JWT_FILE
	//   See also: CRAY_VAULT_JWT_FILE and CRAY_VAULT_ROLE_FILE

	if (test_k8sAuthUrl != "") {
		logger.Infof("Overriding k8s auth url with: '%s'",test_k8sAuthUrl)
		hms_certs.ConfigParams.K8SAuthUrl = test_k8sAuthUrl
	}
	if (test_vaultPKIUrl != "") {
		logger.Infof("Overriding PKI url with: '%s'",test_vaultPKIUrl)
		hms_certs.ConfigParams.VaultPKIUrl = test_vaultPKIUrl
	}
	if (test_vaultCAUrl != "") {
		logger.Infof("Overriding CA url with: '%s'",test_vaultCAUrl)
		hms_certs.ConfigParams.VaultCAUrl = test_vaultCAUrl
	}
	if (test_vaultJWTFile != "") {
		logger.Infof("Overriding Vault JWT file with: '%s'",test_vaultJWTFile)
		hms_certs.ConfigParams.VaultJWTFile = test_vaultJWTFile
	}
	estr := os.Getenv("CRAY_VAULT_JWT_FILE")
	if (estr != "") {
		logger.Infof("Overriding JWT file with: '%s'",estr)
	}
	estr = os.Getenv("CRAY_VAULT_ROLE_FILE")
	if (estr != "") {
		logger.Infof("Overriding ROLE file with: '%s'",estr)
	}
	hms_certs.InitInstance(logger,serviceName)

	if (appParams.LocalMode && (caURI != "")) {
		if (vaultCAURL != "") {
			logger.Infof("Setting Vault CA URL to: '%s'",vaultCAURL)
			hms_certs.ConfigParams.VaultCAUrl = vaultCAURL
		}
		if (vaultPKIURL != "") {
			logger.Infof("Setting Vault PKI URL to: '%s'",vaultPKIURL)
			hms_certs.ConfigParams.VaultPKIUrl = vaultPKIURL
		}

		//Set up TRS cert security stuff and register CA chain update callback

		go func() {
			var ix int
			for ix = 1; ix <= 10; ix ++ {
				err := setupTRSCA()
				if (err == nil) {
					logger.Infof("CA trust bundle loaded successfully.")
					break
				}
				logger.Errorf("Can't get CA trust bundle (attempt %d): %v",
					ix,err)
				time.Sleep(3 * time.Second)
			}

			if (ix >= 10) {
				logger.Errorf("CA trust bundle load attempts exhausted, no trusted TLS operations possible.")
				return
			}
			err := hms_certs.CAUpdateRegister(caURI,caCB)
			if (err != nil) {
				logger.Errorf("Can't register CA bundle watcher: %v",err)
				logger.Errorf("    This means CA bundle updates will not update Redfish HTTP transports!")
			} else {
				logger.Infof("CA bundle watcher registered.")
			}
		}()

	}

	routes := generateRoutes()
	router := newRouter(routes)

	port := appParams.HTTPListenPort
	if (!strings.Contains(port,":")) {
		port = ":"+appParams.HTTPListenPort
	}
	srv := &http.Server{Addr: port, Handler: router,}

	//Set up signal handling for graceful kill

	c := make(chan os.Signal,1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	idleConnsClosed := make(chan struct{})

	go func() {
		<-c
		Running = false

		//Gracefully shutdown the HTTP server
		lerr := srv.Shutdown(context.Background())
		if (lerr != nil) {
			logger.Errorf("HTTP server shutdown error: %v",err)
		}
		close(idleConnsClosed)
	}()

	logger.Infof("%-11s UUID: %s",serviceName,appParams.UUID)
	logger.Infof("HTTP Listen port: %s",appParams.HTTPListenPort)
	logger.Infof("HTTP retries:     %d",appParams.HTTPRetries)
	logger.Infof("Log level:        %s",appParams.LogLevel)
	logger.Infof("TRS mode local:   %t",appParams.LocalMode)
	logger.Infof("TRS kafka URL:    '%s'",appParams.KafkaURL)
	logger.Infof("Vault enabled:    %t",*appParams.VaultEnable)
	logger.Infof("Vault keypath:    '%s'",VaultKeypath)

	if (*appParams.VaultEnable) {
		setupVault()
	}

	logger.Infof("Starting up HTTP server.")
	err = srv.ListenAndServe()
	if (err != http.ErrServerClosed) {
		logger.Fatalf("FATAL: HTTP server ListenandServe failed: %v",err)
	}

	logger.Infof("HTTP Server shutdown, waiting for idle connections to close...")
	<-idleConnsClosed
	logger.Infof("Done.  Exiting.")
	os.Exit(0)
}

