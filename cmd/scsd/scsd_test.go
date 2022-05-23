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
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"testing"
)

func TestENV(t *testing.T) {
	var ival int
	hport := "1234"
	hret := "7"
	uuid := "11223344"
	dlev := "TRACE"
	lmode := "true"
	kurl := "localhost:9678"
	surl := "http://a.b.c.d/hsm/v2"
	vkey := "vault_keypath"
	venbl := "true"
	dflth := "yes"
	os.Setenv("SCSD_HTTP_LISTEN_PORT", hport)
	os.Setenv("SCSD_HTTP_RETRIES", hret)
	os.Setenv("SCSD_UUID", uuid)
	os.Setenv("SCSD_LOG_LEVEL", dlev)
	os.Setenv("SCSD_LOCAL_MODE", lmode)
	os.Setenv("SCSD_KAFKA_URL", kurl)
	os.Setenv("SCSD_SMD_URL", surl)
	os.Setenv("VAULT_KEYPATH", vkey)
	os.Setenv("VAULT_ENABLE", venbl)
	os.Setenv("SCSD_DEFAULT_HTTP", dflth)

	parseEnvVars()

	if appParams.HTTPListenPort != hport {
		t.Errorf("Mismatch of env HTTP listen port, exp: %s, got: %s\n",
			hport, appParams.HTTPListenPort)
	}
	ival, _ = strconv.Atoi(hret)
	if appParams.HTTPRetries != ival {
		t.Errorf("Mismatch of env HTTP retries, exp: %d, got: %d\n",
			ival, appParams.HTTPRetries)
	}
	if appParams.UUID != uuid {
		t.Errorf("Mismatch of env UUID, exp: %s, got: %s\n",
			uuid, appParams.UUID)
	}
	if appParams.LogLevel != dlev {
		t.Errorf("Mismatch of env LogLevel, exp: %s got: %s\n",
			dlev, appParams.LogLevel)
	}
	if appParams.LocalMode == false {
		t.Errorf("Mismatch of env LocalMode, exp: true, got: false\n")
	}
	if appParams.KafkaURL != kurl {
		t.Errorf("Mismatch of env KafkaURL, exp: %s got: %s\n",
			kurl, appParams.KafkaURL)
	}
	if appParams.SmdURL != surl {
		t.Errorf("Mismatch of env SMD url, exp: %s, got: %s\n",
			surl, appParams.UUID)
	}
	if VaultKeypath != vkey {
		t.Errorf("Mismatch of env Vault keypath, exp: %s, got: %s\n",
			vkey, VaultKeypath)
	}
	if *appParams.VaultEnable == false {
		t.Errorf("Mismatch of env VaultEnable, exp: true, got: false\n")
	}
	if dfltHTTP == false {
		t.Errorf("Mismatch of env default http, exp: true, got: false\n")
	}
}

func printStuff() {
	logger.Tracef("LOGGER: TRACE")
	logger.Infof("LOGGER: INFO")
	logger.Debugf("LOGGER: DEBUG")
	logger.Warnf("LOGGER: WARN")
	logger.Errorf("LOGGER: ERROR")
}

func loggerSetup() {
	logger = logrus.New()
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
}

func TestSetLogLevel(t *testing.T) {
	loggerSetup()

	appParams.LogLevel = "TRACE"
	setLogLevel()
	printStuff()
	appParams.LogLevel = "INFO"
	setLogLevel()
	printStuff()
	appParams.LogLevel = "DEBUG"
	setLogLevel()
	printStuff()
	appParams.LogLevel = "WARN"
	setLogLevel()
	printStuff()
	appParams.LogLevel = "ERROR"
	setLogLevel()
	printStuff()
}
