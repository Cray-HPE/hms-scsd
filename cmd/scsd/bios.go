// MIT License
//
// (C) Copyright [2022,2025] Hewlett Packard Enterprise Development LP
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
	"net/http"
	"strconv"
	"strings"
	"time"

	base "github.com/Cray-HPE/hms-base/v2"
	trsapi "github.com/Cray-HPE/hms-trs-app-api/pkg/trs_http_api"
	"github.com/Cray-HPE/hms-xname/xnametypes"
	"github.com/gorilla/mux"
)

type Bios struct {
	common   *BiosCommon
	cray     *BiosCray
	gigabyte *BiosGigabyte
	hpe      *BiosHpe
	intel    *BiosIntel
}

type BiosCommon struct {
	xname            string
	bmcXname         string
	targets          []targInfo
	nodeNumber       int
	manufacturerType manufacturerType
	systemUri        string
	biosUri          string
	chassis          *rfChassis
	systems          *rfSystems
	system           *rfSystem
}

type BiosHpe struct {
	current   *rfBiosHpe
	future    *rfBiosHpe
	futureUri string
}

type BiosGigabyte struct {
	current        *rfBiosGigabyte
	future         *rfBiosSDGigabyte
	futureUri      string
	biosAttributes *rfBiosAttributesRegistry
}

type BiosCray struct {
	current   *rfBiosCray
	future    *rfBiosSDCray
	futureUri string
}

type BiosHpeRegistries struct {
	biosRegistryUri          string
	biosRegistryEnUri        string // english uri
	registries               *rfRegistries
	biosAttributesRegistries *rfBiosAttributesRegistries
	biosAttributes           *rfBiosAttributesRegistry
}

type BiosIntel struct {
	current   *rfBiosIntel
	future    *rfBiosIntel
	futureUri string
}

type BiosTpmState struct {
	Current TpmState `json:"Current"`
	Future  TpmState `json:"Future"`
}

type BiosTpmStatePatch struct {
	Future TpmState `json:"Future"`
}

// SCSD rest interface values
type TpmState string

const (
	TpmStateEnabled    TpmState = "Enabled"
	TpmStateDisabled   TpmState = "Disabled"
	TpmStateNotPresent TpmState = "NotPresent"
)

// BIOS Attribute Names from the redfish interface
type BiosAttributeName string

const (
	TpmStateAttributeCray     BiosAttributeName = "TPM Control"
	TpmStateAttributeGigabyte BiosAttributeName = "TPM State"
	TpmStateAttributeHpe      BiosAttributeName = "TpmState"
	TpmStateAttributeIntel    BiosAttributeName = "TpmOperation"
)

// BIOS Attribute Values from the redfish interface
const (
	EnabledCray      string = "Enabled"
	DisabledCray     string = "Disabled"
	EnabledGigabyte  string = "Enabled"
	DisabledGigabyte string = "Disabled"
	EnabledHpe       string = "PresentEnabled" // these values are for TPM State new const may be needed for other Attributes
	DisabledHpe      string = "PresentDisabled"
	NotPresentHpe    string = "NotPresent"
	EnabledIntel     int    = 1
	DisabledIntel    int    = 0
)

type rfManagersMembers struct {
	ID string `json:"@odata.id"`
}

type rfSystems struct {
	Members []rfSystemsMember
}

type rfSystemsMember struct {
	ID string `json:"@odata.id"`
}

type rfSystem struct {
	Bios rfBios `json:"Bios"`
}

type rfBios struct {
	ID string `json:"@odata.id"`
}

type rfBiosIntel struct {
	Attributes map[string]interface{}
}

type rfBiosHpe struct {
	Attributes rfBiosHpeAttributes
}

type rfBiosHpeAttributes struct {
	TpmState string
}

type PatchAttributeName struct {
	cray     BiosAttributeName
	gigabyte BiosAttributeName
	hpe      BiosAttributeName
	intel    BiosAttributeName
}

type PatchAttributeValue struct {
	cray     interface{}
	gigabyte interface{}
	hpe      interface{}
	intel    interface{}
}

/*
Example for rfBiosGigabyte from /redfish/v1/Systems/Self/Bios
{
  "@Redfish.Settings": {
    "@odata.type": "#Settings.v1_2_1.Settings",
    "SettingsObject": {
      "@odata.id": "/redfish/v1/Systems/Self/Bios/SD"
    }
  },
  "@odata.etag": "W/\"1652393956\"",
  "Attributes": {
    "FBO001": "UEFI",
	...
	"GBT0140": "5",
	...
	"NWSK004": 4,
	...
  }
}
*/
type rfBiosGigabyte struct {
	Settings   rfRedfishSettings      `json:"@Redfish.Settings"`
	ETag       string                 `json:"@odata.etag"`
	Attributes map[string]interface{} `json:"Attributes"` // the values can be either string or int
}

/*
Example for rfBiosSDGigabyte from /redfish/v1/Systems/Self/Bios
The Attributes map will only contain the values that are going to change on reboot.
The Attributes field is not present when there are no changes.
{
  "@odata.etag": "W/\"1652393956\"",
  "Attributes": {
	"TCG001": "Disabled"
  }
}
*/
type rfBiosSDGigabyte struct {
	ETag       string                 `json:"@odata.etag"`
	Attributes map[string]interface{} `json:"Attributes"` // this can be nil. the values can be either string or int
}

/*
Example for rfBiosCray
{
  "@odata.etag": "W/\"1665087620\"",
  "@odata.id": "/redfish/v1/Systems/Node0/Bios",
  "Attributes": {
    "TPM Control": {
      "AllowableValues": [
        "Disabled",
        "Enabled"
      ],
      "DataType": "string",
      "current_value": "Enabled",
      "default_value": "Enabled",
      "menu_type": "Debug",
      "reset_type": "Cold"
    }
  },
  "Id": "Bios",
  "Name": "Current BIOS Settings"
}
*/
type rfBiosCray struct {
	ETag       string                         `json:"@odata.etag"`
	Attributes map[string]rfBiosAttributeCray `json:"Attributes"`
}

type rfBiosAttributeCray struct {
	AllowableValues []interface{} `json:"AllowableValues"`
	DataType        string        `json:"DataType"`
	CurrentValue    interface{}   `json:"current_value"`
}

/*
Example for rfBiosSDCray
{
  "@odata.etag": "W/\"1668554829\"",
  "@odata.id": "/redfish/v1/Systems/Node0/Bios/SD",
  "@odata.type": "#Bios.v1_0_2.Bios",
  "Attributes": {
    "TPM Control": "Enabled"
  },
  "Description": "Future BIOS Settings",
  "Id": "SD",
  "Name": "Future BIOS Settings"
}
*/
type rfBiosSDCray struct {
	ETag       string                 `json:"@odata.etag"`
	Attributes map[string]interface{} `json:"Attributes"`
}

type rfRedfishSettings struct {
	Type           string           `json:"@odata.type"`
	SettingsObject rfSettingsObject `json:"SettingsObject"`
}

type rfSettingsObject struct {
	ID string `json:"@odata.id"`
}

type rfRegistries struct {
	Members []rfManagersMembers
}

type rfBiosAttributesRegistries struct {
	Location []rfRegistriesLocations
}
type rfRegistriesLocations struct {
	Language string `json:"Language"`
	Uri      string `json:"Uri"`
}

/*
Gigabyte example for rfBiosAttributesRegistry from /redfish/v1/Registries/BiosAttributeRegistry.json
{
  "RegistryEntries": {
    "Attributes": [
      {
        "AttributeName": "TCG001",
        "DefaultValue": "Enabled",
        "DisplayName": "  TPM State",
        "HelpText": "Enable/Disable Security Device. NOTE: Your Computer will reboot during restart in order to change State of the Device.",
        "ReadOnly": false,
        "Type": "Enumeration",
        "Value": [
          {
            "ValueDisplayName": "Disabled",
            "ValueName": "Disabled"
          },
          {
            "ValueDisplayName": "Enabled",
            "ValueName": "Enabled"
          }
        ]
      },
    ]
}
*/
type rfBiosAttributesRegistry struct {
	RegistryEntries rfRegistryEntry `json:"RegistryEntries"`
}
type rfRegistryEntry struct {
	Attributes []rfRegistryAttribute `json:"Attributes"`
}
type rfRegistryAttribute struct {
	AttributeName string            `json:"AttributeName"`
	DisplayName   string            `json:"DisplayName"`
	Type          string            `json:"Type"`
	Value         []rfRegistryValue `json:"Value"`
}
type rfRegistryValue struct {
	ValueName        string `json:"ValueName"`
	ValueDisplayName string `json:"ValueDisplayName"`
}

type manufacturerType int

const (
	unknown manufacturerType = iota
	cray
	gigabyte
	hpe
	intel
)

func toXnames(targets []targInfo) []string {
	xnames := make([]string, len(targets), len(targets))
	for i, target := range targets {
		xnames[i] = target.target
	}
	return xnames
}

func getManufacturerType(chassis *rfChassis) manufacturerType {
	for _, member := range chassis.Members {
		switch {
		case strings.EqualFold(member.ID, "/redfish/v1/Chassis/Enclosure"):
			return cray
		case strings.EqualFold(member.ID, "/redfish/v1/Chassis/Self"):
			return gigabyte
		case strings.EqualFold(member.ID, "/redfish/v1/Chassis/1"):
			return hpe
		case strings.EqualFold(member.ID, "/redfish/v1/Chassis/RackMount"):
			return intel
		}
	}
	return unknown
}

func getRedfish(targets []targInfo, uri string) (tasks []trsapi.HttpTask, err error, httpCode int) {
	tasks, _, _, _ = getRedfishNoCheck(targets, uri)

	err = checkStatusCodes(tasks)
	if err != nil {
		err = fmt.Errorf("ERROR: Call failed %s. %v", uri, err)
		return tasks, err, http.StatusInternalServerError
	}

	return tasks, nil, http.StatusOK
}

func getRedfishNoCheck(targets []targInfo, uri string) (tasks []trsapi.HttpTask, codes map[string]int, err error, httpCode int) {
	codes = make(map[string]int)
	xnames := toXnames(targets)

	var sourceTL trsapi.HttpTask
	sourceTL.Timeout = time.Duration(appParams.HTTPTimeout) * time.Second
	sourceTL.Request, _ = http.NewRequest("GET", "", nil)
	tasks = tloc.CreateTaskList(&sourceTL, len(targets))
	populateTaskList(tasks, xnames, uri, http.MethodGet, nil)

	err = doOp(tasks)
	if err != nil {
		err = fmt.Errorf("ERROR: Call failed %s. %v", uri, err)
		return tasks, codes, err, http.StatusInternalServerError
	}

	for _, task := range tasks {
		code := getStatusCode(&task)
		codes[task.Request.Host] = code
	}

	return tasks, codes, nil, http.StatusOK
}

func getRedfishAndParseResponse(
	description string, host string, targets []targInfo, uri string, responseData interface{}) (
	err error, httpCode int) {

	tasks, err, httpCode := getRedfish(targets, uri)
	if err != nil {
		return
	}

	err, httpCode = parseResponseForHost(description, host, tasks, &responseData)
	if err != nil {
		return
	}
	return
}

func patchRedfish(targets []targInfo, uri string, requestBody []byte) (tasks []trsapi.HttpTask, err error, httpCode int) {
	return patchRedfishEtag(targets, uri, requestBody, []string{})
}

func patchRedfishEtag(targets []targInfo, uri string, requestBody []byte, etags []string) (tasks []trsapi.HttpTask, err error, httpCode int) {
	httpCode = http.StatusOK

	xnames := toXnames(targets)

	var sourceTL trsapi.HttpTask
	sourceTL.Timeout = time.Duration(appParams.HTTPTimeout) * time.Second
	sourceTL.Request, _ = http.NewRequest("PATCH", "", nil)
	tasks = tloc.CreateTaskList(&sourceTL, len(targets))
	populateTaskList(tasks, xnames, uri, http.MethodPatch, requestBody)

	if len(tasks) == len(etags) {
		for i, task := range tasks {
			task.Request.Header.Set("If-Match", etags[i])
		}
	} else if len(etags) > 0 {
		err = fmt.Errorf(
			"ERROR: Mismatched number of etags (%d) with the number of tasks (%d). Patch %s for xnames: %v",
			len(etags), len(tasks), uri, xnames)
		return tasks, err, http.StatusInternalServerError
	}

	err = doOp(tasks)
	if err != nil {
		err = fmt.Errorf("ERROR: Patch call %s failed for xnames: %v with error: %v", uri, xnames, err)
		return tasks, err, http.StatusInternalServerError
	}

	err = checkStatusCodes(tasks)
	if err != nil {
		err = fmt.Errorf("ERROR: Patch call %s returned failure status for xnames: %v with error: %v", uri, xnames, err)
		return tasks, err, http.StatusInternalServerError
	}

	return tasks, nil, httpCode
}

func getTask(host string, tasks []trsapi.HttpTask) (task *trsapi.HttpTask, err error, code int) {
	foundTasks := make([]string, len(tasks))
	for _, t := range tasks {
		taskHost := t.Request.Host
		if host == taskHost {
			task = &t
			return
		}
		foundTasks = append(foundTasks, taskHost)
	}
	err = fmt.Errorf("ERROR: failed to find task for %s. tasks: %v", host, foundTasks)
	code = http.StatusInternalServerError
	return
}

func parseResponse(description string, task *trsapi.HttpTask, data interface{}) (err error, code int) {
	err = grabTaskRspData(description, task, data)
	if err != nil {
		err = fmt.Errorf("ERROR: parsing response from %v gave the error: %v", task.Request.URL, err)
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

func parseResponseForHost(description string, host string, tasks []trsapi.HttpTask, data interface{}) (err error, httpCode int) {
	task, err, httpCode := getTask(host, tasks)
	if err != nil {
		return
	}
	err, httpCode = parseResponse(description, task, data)
	return
}

func getSystemUri(xname string, nodeNumber int, manufacturer manufacturerType, systems *rfSystems) (uri string, err error, httpCode int) {
	httpCode = http.StatusOK
	switch manufacturer {
	case hpe:
		hpeId := nodeNumber + 1
		suffix := "/" + strconv.Itoa(hpeId)
		for _, member := range systems.Members {
			if strings.HasSuffix(member.ID, suffix) {
				return member.ID, nil, httpCode
			}
		}
	case cray:
		suffix := "/node" + strconv.Itoa(nodeNumber)
		for _, member := range systems.Members {
			if strings.HasSuffix(strings.ToLower(member.ID), suffix) {
				return member.ID, nil, httpCode
			}
		}
	case gigabyte:
		if nodeNumber == 0 {
			suffix := "/self"
			for _, member := range systems.Members {
				if strings.HasSuffix(strings.ToLower(member.ID), suffix) {
					return member.ID, nil, httpCode
				}
			}
		}
	case intel:
		if len(systems.Members) > nodeNumber && nodeNumber >= 0 {
			return systems.Members[nodeNumber].ID, nil, httpCode
		}
	}

	uris := make([]string, len(systems.Members))
	for i, member := range systems.Members {
		uris[i] = member.ID
	}
	err = fmt.Errorf("Error: Could not find redfish system URI that matches the node %s. Known system URIs %v", xname, uris)
	return "", err, http.StatusNotFound
}

func getNodeNumber(xname string) (int, error) {
	xnameType := xnametypes.GetHMSType(xname)
	if xnameType != xnametypes.Node {
		err := fmt.Errorf("Error: Xname %s is not of the type Node, but is instead %s", xname, xnameType.String())
		return 0, err
	}

	regex, err := xnametypes.GetHMSTypeRegex(xnameType)
	if err != nil {
		return 0, err
	}

	matchedGroups := regex.FindSubmatch([]byte(xname))
	if matchedGroups == nil {
		err := fmt.Errorf("Error: Xname %s did not match the pattern %s", xname, regex.String())
		return 0, err
	}

	if len(matchedGroups) < 6 {
		err := fmt.Errorf(
			"Error: Less than 6 matched groups for xname %s using the pattern %s.",
			xname, regex.String())
		return 0, err
	}

	nodeNumberStr := string(matchedGroups[5])
	nodeNumber, err := strconv.Atoi(nodeNumberStr)
	if err != nil {
		wrappedErr := fmt.Errorf(
			"Error: Unable to parse, %s, as an integer for xname %s using group 5 of the pattern %s. Conversion error: %s",
			nodeNumberStr, xname, regex.String(), err.Error())
		return 0, wrappedErr
	}

	return nodeNumber, nil
}

func validateXname(xname string) (normalizedXname string, err error, httpCode int) {
	if !xnametypes.IsHMSCompIDValid(xname) {
		err = fmt.Errorf("ERROR: %s is an invalid xname", xname)
		httpCode = http.StatusBadRequest
		return
	}

	normalizedXname = xnametypes.NormalizeHMSCompID(xname)

	xnameType := xnametypes.GetHMSType(xname)
	if xnameType != xnametypes.Node {
		exampleString := "xXcCsSbBnN"
		t := xnametypes.GetHMSCompRecognitionTable()
		if value, found := t[strings.ToLower(xnametypes.Node.String())]; found {
			exampleString = value.ExampleString
		}

		err = fmt.Errorf("ERROR: %s is not the xname of a node. The xname should match: %s", xname, exampleString)
		httpCode = http.StatusBadRequest
		return
	}

	return
}

func getAttribute(name string, attributes *rfBiosAttributesRegistry) (attribute rfRegistryAttribute, found bool) {
	n := strings.ToLower(name)
	for _, attribute = range attributes.RegistryEntries.Attributes {
		displayName := strings.ToLower(strings.TrimSpace(attribute.DisplayName))
		if n == displayName {
			return attribute, true
		}
	}
	return attribute, false
}

func getBiosCommon(xname string) (bios *BiosCommon, err error, httpCode int) {
	httpCode = http.StatusOK
	bios = &BiosCommon{
		xname:    xname,
		bmcXname: xnametypes.GetHMSCompParent(xname),
	}

	bios.nodeNumber, err = getNodeNumber(xname)
	if err != nil {
		httpCode = http.StatusBadRequest
		return
	}

	xnames := []string{bios.bmcXname}

	bios.targets, err = hsmVerify(makeTargData(xnames), true, true) // verify with hsm and fill in extra data
	if err != nil {
		err = fmt.Errorf("ERROR: Problem verifying target states: %v.", err)
		httpCode = http.StatusInternalServerError
		return
	}

	// ---- /redfish/v1/Chassis ----

	var chassis rfChassis
	err, httpCode = getRedfishAndParseResponse("Chassis", bios.bmcXname, bios.targets, RFCHASSIS_API, &chassis)
	if err != nil {
		return
	}
	bios.chassis = &chassis

	// ---- /redfish/v1/Systems ----

	var systems rfSystems
	err, httpCode = getRedfishAndParseResponse("Systems", bios.bmcXname, bios.targets, RFSYSTEMS_API, &systems)
	if err != nil {
		return
	}
	bios.systems = &systems

	bios.manufacturerType = getManufacturerType(bios.chassis)
	if bios.manufacturerType == unknown {
		members := make([]string, len(bios.chassis.Members))
		for i, member := range bios.chassis.Members {
			members[i] = member.ID
		}
		logger.Errorf("Unknown manufacturer for %s where its chassis has these members: %v", bios.xname, members)
		err = fmt.Errorf("ERROR: BIOS calls not supported for this type of hardware. xname: %s ", bios.xname)
		httpCode = http.StatusBadRequest
		return
	}

	bios.systemUri, err, httpCode = getSystemUri(xname, bios.nodeNumber, bios.manufacturerType, bios.systems)
	if err != nil {
		return
	}

	// ---- /redfish/v1/Systems/1 ----
	// ---- /redfish/v1/Systems/Node0 ----
	// ---- /redfish/v1/Systems/Node1 ----
	// ---- /redfish/v1/Systems/Self ----
	// ---- /redfish/v1/Systems/BQWF73500342 ----

	var system rfSystem
	err, httpCode = getRedfishAndParseResponse("System", bios.bmcXname, bios.targets, bios.systemUri, &system)
	if err != nil {
		return
	}
	bios.system = &system

	bios.biosUri = bios.system.Bios.ID

	return
}

func getBiosHpe(biosCommon *BiosCommon) (biosHpe *BiosHpe, err error, httpCode int) {
	httpCode = http.StatusOK
	biosHpe = &BiosHpe{}

	// ---- /redfish/v1/Systems/1/Bios ----

	var current rfBiosHpe
	err, httpCode = getRedfishAndParseResponse(
		"System/1/Bios", biosCommon.bmcXname, biosCommon.targets, biosCommon.biosUri, &current)
	if err != nil {
		return
	}
	biosHpe.current = &current

	// ---- /redfish/v1/Systems/1/Bios/Settings ----

	biosHpe.futureUri = biosCommon.biosUri + "/Settings"

	var future rfBiosHpe
	err, httpCode = getRedfishAndParseResponse(
		"System/1/Bios/Settings", biosCommon.bmcXname, biosCommon.targets, biosHpe.futureUri, &future)
	if err != nil {
		return
	}
	biosHpe.future = &future

	return
}

func getBiosRegistriesHpe(biosCommon *BiosCommon) (biosRegistries *BiosHpeRegistries, err error, httpCode int) {
	httpCode = http.StatusOK
	biosRegistries = &BiosHpeRegistries{}

	// ---- /redfish/v1/Registries ----

	var registries rfRegistries
	err, httpCode = getRedfishAndParseResponse(
		"Registries", biosCommon.bmcXname, biosCommon.targets, RFREGISTRIES_API, &registries)
	if err != nil {
		return
	}
	biosRegistries.registries = &registries

	for _, member := range biosRegistries.registries.Members {
		// looking for a uri like: /redfish/v1/Registries/BiosAttributeRegistryA43.v1_2_40
		if strings.HasPrefix(strings.ToLower(member.ID), "/redfish/v1/registries/biosattributeregistry") {
			biosRegistries.biosRegistryUri = member.ID
			break
		}
	}

	if biosRegistries.biosRegistryUri == "" {
		err = fmt.Errorf(
			"ERROR: Could not find bios registries for %s from %s %s. Known registries %v",
			biosCommon.xname, biosCommon.bmcXname, RFREGISTRIES_API, biosRegistries.registries.Members)
		httpCode = http.StatusInternalServerError
		return
	}

	// ---- /redfish/v1/Registries/BiosAttributeRegistryA43.v1_2_40 ----

	var biosAttributesRegistries rfBiosAttributesRegistries
	err, httpCode = getRedfishAndParseResponse(
		"AttributesRegistries", biosCommon.bmcXname, biosCommon.targets, biosRegistries.biosRegistryUri, &biosAttributesRegistries)
	if err != nil {
		return
	}
	biosRegistries.biosAttributesRegistries = &biosAttributesRegistries

	for _, location := range biosAttributesRegistries.Location {
		if strings.ToLower(location.Language) == "en" {
			biosRegistries.biosRegistryEnUri = location.Uri
			break
		}
	}

	if biosRegistries.biosRegistryEnUri == "" {
		err = fmt.Errorf(
			"ERROR: Could not find bios registries english uri for %s from %s %s. Known registries %v",
			biosCommon.xname, biosCommon.bmcXname, biosRegistries.biosRegistryUri, biosAttributesRegistries.Location)
		httpCode = http.StatusInternalServerError
		return
	}

	// ---- /redfish/v1/registrystore/registries/en/biosattributeregistrya43.v1_2_40 ----

	var biosAttributesRegistry rfBiosAttributesRegistry
	err, httpCode = getRedfishAndParseResponse(
		"AttributesRegistry", biosCommon.bmcXname, biosCommon.targets, biosRegistries.biosRegistryEnUri, &biosAttributesRegistry)
	if err != nil {
		return
	}
	biosRegistries.biosAttributes = &biosAttributesRegistry

	return
}

func patchBiosHpe(biosCommon *BiosCommon, name PatchAttributeName, value PatchAttributeValue) (err error, httpCode int) {
	biosHpe, err, httpCode := getBiosHpe(biosCommon)
	if err != nil {
		return
	}

	futureValue := fmt.Sprintf("%v", value.hpe)
	attributeName := string(name.hpe)

	biosHpeRegistries, err, httpCode := getBiosRegistriesHpe(biosCommon)
	if err != nil {
		return
	}

	hardwareSupportsFutureValue := false
	for _, attribute := range biosHpeRegistries.biosAttributes.RegistryEntries.Attributes {
		if strings.ToLower(attribute.AttributeName) == strings.ToLower(attributeName) {
			for _, value := range attribute.Value {
				if value.ValueName == futureValue {
					hardwareSupportsFutureValue = true
					break
				}
			}
			break
		}
	}

	if !hardwareSupportsFutureValue {
		// todo change message to contain the value that was passed to the scsd interface
		// instead of the value being passed to redfish
		err = fmt.Errorf("ERROR: Hardware does not support Future state %s", futureValue)
		httpCode = http.StatusMethodNotAllowed
		return
	}

	rfRequestBody := "{\"Attributes\":{\"" + attributeName + "\":\"" + futureValue + "\"}}"

	tasks, err, httpCode := patchRedfish(biosCommon.targets, biosHpe.futureUri, []byte(rfRequestBody))
	if err != nil {
		return
	}

	for _, task := range tasks {
		statusCode := getStatusCode(&task)
		if !statusCodeOK(statusCode) {
			err = fmt.Errorf("ERROR: Redfish patch failed %s %d", biosHpe.futureUri, statusCode)
			httpCode = http.StatusInternalServerError
			return
		}
	}
	return
}

func getBiosGigabyte(biosCommon *BiosCommon) (biosGigabyte *BiosGigabyte, err error, httpCode int) {
	httpCode = http.StatusOK
	biosGigabyte = &BiosGigabyte{}

	// ---- /redfish/v1/Systems/Self/Bios ----

	var current rfBiosGigabyte
	err, httpCode = getRedfishAndParseResponse(
		"Systems/Self/Bios", biosCommon.bmcXname, biosCommon.targets, biosCommon.biosUri, &current)
	if err != nil {
		return
	}
	biosGigabyte.current = &current

	// ---- /redfish/v1/Systems/Self/Bios/SD ----

	biosGigabyte.futureUri = biosGigabyte.current.Settings.SettingsObject.ID
	if biosGigabyte.futureUri == "" {
		biosGigabyte.futureUri = biosCommon.biosUri + "/SD"
	}

	biosFutureTasks, codes, err, httpCode := getRedfishNoCheck(biosCommon.targets, biosGigabyte.futureUri)
	if err != nil {
		return
	}

	code, ok := codes[biosCommon.bmcXname]
	if !ok {
		err = fmt.Errorf(
			"ERROR: missing http return code for %s from %s %s",
			biosCommon.xname, biosCommon.bmcXname, biosGigabyte.futureUri)
		httpCode = http.StatusInternalServerError
		return
	}

	if !statusCodeOK(code) && code != http.StatusNotFound {
		err = fmt.Errorf(
			"ERROR: Unexpected http return code, %d, for %s from %s %s",
			code, biosCommon.xname, biosCommon.bmcXname, biosGigabyte.futureUri)
		httpCode = http.StatusInternalServerError
		return
	}

	if code == http.StatusNotFound {
		// If there are no pending changes to the bios the /SD redfish call can return 404
		biosGigabyte.future = &rfBiosSDGigabyte{
			Attributes: make(map[string]interface{}),
		}
	} else {
		var future rfBiosSDGigabyte
		err, httpCode = parseResponseForHost("Systems/Self/Bios/SD", biosCommon.bmcXname, biosFutureTasks, &future)
		if err != nil {
			return
		}
		biosGigabyte.future = &future
	}

	// ---- /redfish/v1/Registries/BiosAttributeRegistry.json ----

	registryUri := "/redfish/v1/Registries/BiosAttributeRegistry.json"

	var biosAttributes rfBiosAttributesRegistry
	err, httpCode = getRedfishAndParseResponse(
		"BiosAttributeRegistry.json", biosCommon.bmcXname, biosCommon.targets, registryUri, &biosAttributes)
	if err != nil {
		return
	}
	biosGigabyte.biosAttributes = &biosAttributes

	return
}

func patchBiosGigabyte(biosCommon *BiosCommon, attributeName PatchAttributeName, attrbiuteValue PatchAttributeValue) (err error, httpCode int) {
	biosGigabyte, err, httpCode := getBiosGigabyte(biosCommon)

	name := string(attributeName.gigabyte)
	futureValue := fmt.Sprintf("%v", attrbiuteValue.gigabyte)

	attribute, found := getAttribute(name, biosGigabyte.biosAttributes)
	if !found {
		err = fmt.Errorf("%s not supported in the BIOS", name)
		httpCode = http.StatusMethodNotAllowed
		return
	}

	foundValueDefined := false
	for _, value := range attribute.Value {
		if value.ValueName == futureValue {
			foundValueDefined = true
			break
		}
	}
	if !foundValueDefined {
		logger.Errorf(
			"Tried to set %s with redfish value %s, but redfish only supports %v", name, futureValue, attribute.Value)
		err = fmt.Errorf(
			"BIOS for the field %s does not support %s", name, futureValue)
		httpCode = http.StatusMethodNotAllowed
		return
	}

	rfRequestBody := "{\"Attributes\":{\"" + attribute.AttributeName + "\":\"" + futureValue + "\"}}"

	etag := biosGigabyte.future.ETag
	if etag == "" {
		// when SD (i.e. the future settings) is not avialble the If-Match header should match anything
		// gigabyte will reject any patch request that does not have a If-Match header
		etag = "*"
	}
	tasks, err, httpCode := patchRedfishEtag(biosCommon.targets, biosGigabyte.futureUri, []byte(rfRequestBody), []string{etag})
	if err != nil {
		return
	}

	for _, task := range tasks {
		statusCode := getStatusCode(&task)
		if !statusCodeOK(statusCode) {
			err = fmt.Errorf("ERROR: Redfish patch failed %s %d", biosGigabyte.futureUri, statusCode)
			httpCode = http.StatusInternalServerError
			return
		}
	}

	return
}

func getBiosCray(biosCommon *BiosCommon) (biosCray *BiosCray, err error, httpCode int) {
	httpCode = http.StatusOK
	biosCray = &BiosCray{}

	// ---- /redfish/v1/Systems/Node0/Bios ----
	// ---- /redfish/v1/Systems/Node1/Bios ----

	var current rfBiosCray
	err, httpCode = getRedfishAndParseResponse(
		"Systems/Node[0-1]/Bios", biosCommon.bmcXname, biosCommon.targets, biosCommon.biosUri, &current)
	if err != nil {
		return
	}
	biosCray.current = &current

	// ---- /redfish/v1/Systems/Node0/Bios/SD ----
	// ---- /redfish/v1/Systems/Node1/Bios/SD ----

	biosCray.futureUri = biosCommon.biosUri + "/SD"

	biosFutureTasks, codes, err, httpCode := getRedfishNoCheck(biosCommon.targets, biosCray.futureUri)
	if err != nil {
		return
	}

	code, ok := codes[biosCommon.bmcXname]
	if !ok {
		err = fmt.Errorf(
			"ERROR: missing http return code for %s from %s %s",
			biosCommon.xname, biosCommon.bmcXname, biosCray.futureUri)
		httpCode = http.StatusInternalServerError
		return
	}

	if !statusCodeOK(code) && code != http.StatusNotFound {
		err = fmt.Errorf(
			"ERROR: Unexpected http return code, %d, for %s from %s %s",
			code, biosCommon.xname, biosCommon.bmcXname, biosCray.futureUri)
		httpCode = http.StatusInternalServerError
		return
	}

	if code == http.StatusNotFound {
		// If there are no pending changes to the bios the /SD redfish call can return 404
		biosCray.future = &rfBiosSDCray{
			Attributes: make(map[string]interface{}),
		}
	} else {
		var future rfBiosSDCray
		err, httpCode = parseResponseForHost("Systems/Node[0-1]/Bios/SD", biosCommon.bmcXname, biosFutureTasks, &future)
		if err != nil {
			return
		}
		biosCray.future = &future
	}

	return
}

func patchBiosCray(biosCommon *BiosCommon, attributeName PatchAttributeName, attrbiuteValue PatchAttributeValue) (err error, httpCode int) {
	biosCray, err, httpCode := getBiosCray(biosCommon)
	if err != nil {
		return
	}

	name := string(attributeName.cray)
	futureValue := fmt.Sprintf("%v", attrbiuteValue.cray)

	attribute, found := biosCray.current.Attributes[name]
	if !found {
		err = fmt.Errorf("%s not supported in the BIOS", name)
		httpCode = http.StatusMethodNotAllowed
		return
	}

	foundValueDefined := false
	for _, value := range attribute.AllowableValues {
		if value == futureValue {
			foundValueDefined = true
			break
		}
	}
	if !foundValueDefined {
		logger.Errorf(
			"Tried to set %s with redfish value %s, but redfish only supports %v",
			name, futureValue, attribute.AllowableValues)
		err = fmt.Errorf("BIOS %s value %s is not supported", name, futureValue)
		httpCode = http.StatusMethodNotAllowed
		return
	}

	rfRequestBody := "{\"Attributes\":{\"" + name + "\":\"" + futureValue + "\"}}"

	etag := biosCray.future.ETag
	if etag == "" {
		// when SD (i.e. the future settings) is not avialble the If-Match header should match anything
		// This is not strictly required because cray hardware does not currently require the etag
		etag = "*"
	}
	tasks, err, httpCode := patchRedfishEtag(biosCommon.targets, biosCray.futureUri, []byte(rfRequestBody), []string{etag})
	if err != nil {
		return
	}

	for _, task := range tasks {
		statusCode := getStatusCode(&task)
		if !statusCodeOK(statusCode) {
			err = fmt.Errorf("ERROR: Redfish patch failed %s %d", biosCray.futureUri, statusCode)
			httpCode = http.StatusInternalServerError
			return
		}
	}

	return
}

func getBiosIntel(biosCommon *BiosCommon) (biosIntel *BiosIntel, err error, httpCode int) {
	httpCode = http.StatusOK

	biosIntel = &BiosIntel{}
	biosIntel.current = &rfBiosIntel{}
	biosIntel.future = &rfBiosIntel{}
	biosIntel.futureUri = biosCommon.biosUri + "/Settings"

	// ---- /redfish/v1/Systems/BQWF73500342/Bios ----

	err, httpCode = getRedfishAndParseResponse(
		"Systems/*/Bios", biosCommon.bmcXname, biosCommon.targets, biosCommon.biosUri, biosIntel.current)
	if err != nil {
		return
	}

	// ---- /redfish/v1/Systems/BQWF73500342/Bios/Settings ----

	err, httpCode = getRedfishAndParseResponse(
		"Systems/*/Bios/Settings", biosCommon.bmcXname, biosCommon.targets, biosIntel.futureUri, biosIntel.future)
	if err != nil {
		return
	}

	return
}

func getBios(r *http.Request) (bios *Bios, err error, httpCode int) {
	bios = &Bios{}

	mvars := mux.Vars(r)
	xnameOriginal := mvars["xname"]

	xname, err, httpCode := validateXname(xnameOriginal)
	if err != nil {
		return
	}

	bios.common, err, httpCode = getBiosCommon(xname)
	if err != nil {
		return
	}

	switch bios.common.manufacturerType {
	case cray:
		bios.cray, err, httpCode = getBiosCray(bios.common)
	case gigabyte:
		bios.gigabyte, err, httpCode = getBiosGigabyte(bios.common)
	case hpe:
		bios.hpe, err, httpCode = getBiosHpe(bios.common)
	case intel:
		bios.intel, err, httpCode = getBiosIntel(bios.common)
	default:
		logger.Errorf(
			"Getting BIOS has not been implmented for the hardware. type: %d, xname: %s",
			bios.common.manufacturerType, xname)
		err = fmt.Errorf("Modifications not supported by BMC at %s", xname)
		httpCode = http.StatusBadRequest
	}
	return
}

func patchBios(r *http.Request, attributeName PatchAttributeName, attributeValue PatchAttributeValue) (err error, httpCode int) {
	mvars := mux.Vars(r)
	xnameOriginal := mvars["xname"]

	xname, err, httpCode := validateXname(xnameOriginal)
	if err != nil {
		return
	}

	biosCommon, err, httpCode := getBiosCommon(xname)
	if err != nil {
		return
	}

	switch biosCommon.manufacturerType {
	case cray:
		err, httpCode = patchBiosCray(biosCommon, attributeName, attributeValue)
	case gigabyte:
		err, httpCode = patchBiosGigabyte(biosCommon, attributeName, attributeValue)
	case hpe:
		err, httpCode = patchBiosHpe(biosCommon, attributeName, attributeValue)
	case intel:
		// todo implement this
		logger.Errorf(
			"Modifications for %s has not been implmented for intel hardware. xname: %s",
			attributeName.intel, xname)
		err = fmt.Errorf("Modifications not supported by BMC at %s", xname)
		httpCode = http.StatusBadRequest
	default:
		logger.Errorf(
			"Modifications for %s has not been implmented for hardware. type: %d, xname: %s",
			attributeName.intel, biosCommon.manufacturerType, xname)
		err = fmt.Errorf("Modifications not supported by BMC at %s", xname)
		httpCode = http.StatusBadRequest
	}
	return
}

// TPM State functions

func toTpmStateHpe(bios *BiosHpe) BiosTpmState {
	state := BiosTpmState{
		Current: TpmStateNotPresent,
		Future:  TpmStateNotPresent,
	}

	switch bios.current.Attributes.TpmState {
	case EnabledHpe:
		state.Current = TpmStateEnabled
	case DisabledHpe:
		state.Current = TpmStateDisabled
	case NotPresentHpe:
		state.Current = TpmStateNotPresent
	}

	switch bios.future.Attributes.TpmState {
	case EnabledHpe:
		state.Future = TpmStateEnabled
	case DisabledHpe:
		state.Future = TpmStateDisabled
	case NotPresentHpe:
		state.Future = TpmStateNotPresent
	}

	return state
}

func toTpmStateValueGigabyte(value interface{}) TpmState {
	str := fmt.Sprintf("%v", value)
	switch str {
	case EnabledGigabyte:
		return TpmStateEnabled
	case DisabledGigabyte:
		return TpmStateDisabled
	default:
		return TpmStateNotPresent
	}
}

func toTpmStateGigabyte(bios *BiosGigabyte) BiosTpmState {
	state := BiosTpmState{
		Current: TpmStateNotPresent,
		Future:  TpmStateNotPresent,
	}
	attributeName := string(TpmStateAttributeGigabyte)
	attribute, found := getAttribute(attributeName, bios.biosAttributes)
	if found {
		value, ok := bios.current.Attributes[attribute.AttributeName]
		if ok {
			state.Current = toTpmStateValueGigabyte(value)
			valueFuture, okFuture := bios.future.Attributes[attribute.AttributeName]
			if okFuture {
				state.Future = toTpmStateValueGigabyte(valueFuture)
			} else {
				state.Future = state.Current
			}
		} else {
			logger.Errorf(
				"ERROR: found attrbiute for '%s' in the registry attributes, "+
					"but did not find it in the current bios settings. { Attribute: %v } { Bios: %v }",
				attributeName, attribute, bios.current.Attributes)
		}
	}
	return state
}

func toTpmStateValueCray(value interface{}) TpmState {
	str := fmt.Sprintf("%v", value)
	switch str {
	case EnabledCray:
		return TpmStateEnabled
	case DisabledCray:
		return TpmStateDisabled
	default:
		return TpmStateNotPresent
	}
}

func toTpmStateCray(bios *BiosCray) BiosTpmState {
	state := BiosTpmState{
		Current: TpmStateNotPresent,
		Future:  TpmStateNotPresent,
	}
	tpmStateKey := string(TpmStateAttributeCray)
	attribute, ok := bios.current.Attributes[tpmStateKey]
	if ok {
		state.Current = toTpmStateValueCray(attribute.CurrentValue)

		attributeFuture, okFuture := bios.future.Attributes[tpmStateKey]
		if okFuture {
			state.Future = toTpmStateValueCray(attributeFuture)
		} else {
			state.Future = state.Current
		}
	}
	return state
}

func toInt(value interface{}, defaultValue int) int {
	if value == nil {
		return defaultValue
	}
	str := fmt.Sprintf("%v", value)
	result, err := strconv.Atoi(str)
	if err != nil {
		return defaultValue
	}
	return result
}

func toTpmStateIntel(bios *BiosIntel) BiosTpmState {
	state := BiosTpmState{
		Current: TpmStateNotPresent,
		Future:  TpmStateNotPresent,
	}
	// The interface{} values in the Attributes map will likely be float64
	// but should only be either 0 or 1
	foundCurrentValue := false
	tpmOperation := -1
	value, ok := bios.current.Attributes["TpmOperation"]
	if ok {
		foundCurrentValue = true
		tpmOperation = toInt(value, tpmOperation)
	}

	tpm2Operation := -1
	value, ok = bios.current.Attributes["Tpm2Operation"]
	if ok {
		foundCurrentValue = true
		tpm2Operation = toInt(value, tpm2Operation)
	}

	tpmOperationFuture := tpmOperation
	value, ok = bios.future.Attributes["TpmOperation"]
	if ok {
		tpmOperationFuture = toInt(value, tpmOperationFuture)
	}

	tpm2OperationFuture := tpm2Operation
	value, ok = bios.future.Attributes["Tpm2Operation"]
	if ok {
		tpm2OperationFuture = toInt(value, tpm2OperationFuture)
	}

	if foundCurrentValue {
		if EnabledIntel == tpmOperation || EnabledIntel == tpm2Operation {
			state.Current = TpmStateEnabled
		} else {
			state.Current = TpmStateDisabled
		}
		if EnabledIntel == tpmOperationFuture || EnabledIntel == tpm2OperationFuture {
			state.Future = TpmStateEnabled
		} else {
			state.Future = TpmStateDisabled
		}
	} else {
		state.Current = TpmStateNotPresent
		state.Future = TpmStateNotPresent
	}
	return state
}

func doBiosTpmStateGet(w http.ResponseWriter, r *http.Request) {
	title := "Get BIOS TPM State"

	defer base.DrainAndCloseRequestBody(r)

	bios, err, httpCode := getBios(r)
	if err != nil {
		sendErrorRsp(w, title, err.Error(), r.URL.Path, httpCode)
		return
	}
	var tpmState BiosTpmState
	tpmState.Current = TpmStateNotPresent
	tpmState.Future = TpmStateNotPresent

	switch bios.common.manufacturerType {
	case cray:
		tpmState = toTpmStateCray(bios.cray)
	case gigabyte:
		tpmState = toTpmStateGigabyte(bios.gigabyte)
	case hpe:
		tpmState = toTpmStateHpe(bios.hpe)
	case intel:
		tpmState = toTpmStateIntel(bios.intel)
	}

	ba, baerr := json.Marshal(tpmState)
	if baerr != nil {
		emsg := fmt.Sprintf("ERROR: Problem marshaling TPM data: %v", baerr)
		sendErrorRsp(w, "Get TPM State Error", emsg, r.URL.Path, http.StatusInternalServerError)
		return
	}

	w.Header().Set(CT_TYPE, CT_APPJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

func doBiosTpmStatePatch(w http.ResponseWriter, r *http.Request) {
	title := "Patch BIOS TPM State"

	defer base.DrainAndCloseRequestBody(r)

	var requestBody BiosTpmStatePatch

	err := getReqData(title, r, &requestBody)
	if err != nil {
		emsg := fmt.Sprintf("ERROR: Problem getting request data: %v", err)
		sendErrorRsp(w, "Bad request data", emsg, r.URL.Path, http.StatusBadRequest)
		return
	}

	attributeName := PatchAttributeName{
		cray:     TpmStateAttributeCray,
		gigabyte: TpmStateAttributeGigabyte,
		hpe:      TpmStateAttributeHpe,
		intel:    TpmStateAttributeIntel,
	}
	var attributeValue PatchAttributeValue
	switch requestBody.Future {
	case TpmStateEnabled:
		attributeValue.cray = EnabledCray
		attributeValue.gigabyte = EnabledGigabyte
		attributeValue.hpe = EnabledHpe
		attributeValue.intel = EnabledIntel
	case TpmStateDisabled:
		attributeValue.cray = DisabledCray
		attributeValue.gigabyte = DisabledGigabyte
		attributeValue.hpe = DisabledHpe
		attributeValue.intel = DisabledIntel
	default:
		emsg := fmt.Sprintf("ERROR: Invalid future value: %s", requestBody.Future)
		sendErrorRsp(w, "Bad request data", emsg, r.URL.Path, http.StatusBadRequest)
		return
	}

	err, httpCode := patchBios(r, attributeName, attributeValue)
	if err != nil {
		sendErrorRsp(w, title, err.Error(), r.URL.Path, httpCode)
		return
	}

	w.Header().Set(CT_TYPE, CT_APPJSON)
	w.WriteHeader(http.StatusNoContent)
}
