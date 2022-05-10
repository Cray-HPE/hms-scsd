// MIT License
//
// (C) Copyright [2022] Hewlett Packard Enterprise Development LP
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
)

func createRfChassis(uris ...string) rfChassis {
	chassis := rfChassis{}
	for _, uri := range uris {
		chassis.Members = append(chassis.Members, rfChassisMembers{ID: uri})
	}
	return chassis
}

func createRfSystems(uris ...string) rfSystems {
	systems := rfSystems{}
	for _, uri := range uris {
		systems.Members = append(systems.Members, rfSystemsMember{ID: uri})
	}
	return systems
}

func TestToXnames(t *testing.T) {
	targets := make([]targInfo, 0)
	xnames := toXnames(targets)
	if len(xnames) != 0 {
		t.Errorf("Expected 0 xnames but instead found %d", len(xnames))
	}

	xname0 := "x0"
	xname1 := "x1"
	targets = append(targets, targInfo{target: xname0})
	targets = append(targets, targInfo{target: xname1})
	xnames = toXnames(targets)
	if len(xnames) != 2 {
		t.Fatalf("Expected 2 xnames but instead found %d", len(xnames))
	}
	if xnames[0] != xname0 {
		t.Errorf("Expected xnames[0] to be %s but was instead %s", xname0, xnames[0])
	}
	if xnames[1] != xname1 {
		t.Errorf("Expected xnames[1] to be %s but was instead %s", xname1, xnames[1])
	}
}

func TestGetManufacturerType(t *testing.T) {
	chassis := createRfChassis()
	manufactureType := getManufacturerType(&chassis)
	if manufactureType != unknown {
		t.Errorf("test1: Expected unknown type but instead got %d", manufactureType)
	}

	chassis = createRfChassis("/redfish/v1/Chassis/junk")
	manufactureType = getManufacturerType(&chassis)
	if manufactureType != unknown {
		t.Errorf("test2: Expected unknown type but instead got %d", manufactureType)
	}

	chassis = createRfChassis("/redfish/v1/Chassis/junk", "/redfish/v1/Chassis/Enclosure")
	manufactureType = getManufacturerType(&chassis)
	if manufactureType != cray {
		t.Errorf("Expected cray type but instead got %d", manufactureType)
	}

	chassis = createRfChassis("/redfish/v1/Chassis/Self")
	manufactureType = getManufacturerType(&chassis)
	if manufactureType != gigabyte {
		t.Errorf("Expected gigabyte type but instead got %d", manufactureType)
	}

	chassis = createRfChassis("/redfish/v1/Chassis/1")
	manufactureType = getManufacturerType(&chassis)
	if manufactureType != hpe {
		t.Errorf("Expected hpe type but instead got %d", manufactureType)
	}

	chassis = createRfChassis("/redfish/v1/Chassis/RackMount")
	manufactureType = getManufacturerType(&chassis)
	if manufactureType != intel {
		t.Errorf("Expected intel type but instead got %d", manufactureType)
	}
}

func TestGetSystemUri(t *testing.T) {
	systems := createRfSystems()
	xname := "x0" // getSystemuri only uses the xname in the error message

	uri, err, _ := getSystemUri(xname, 0, hpe, &systems)
	if err == nil {
		t.Errorf("Expected to get an error for an empty list of systems")
	}
	// ---- cray ----
	// /redfish/v1/Systems/Node0
	// /redfish/v1/Systems/Node1
	node0 := "/redfish/v1/Systems/Node0"
	node1 := "/redfish/v1/Systems/Node1"
	systems = createRfSystems(node1, node0)
	uri, err, _ = getSystemUri(xname, 0, cray, &systems)
	if err != nil {
		t.Errorf("Unexpected error for cray node 0. error: %v ", err)
	}
	if uri != node0 {
		t.Errorf("Expected uri, %s, but was %s", node0, uri)
	}

	uri, err, _ = getSystemUri(xname, 1, cray, &systems)
	if err != nil {
		t.Errorf("Unexpected error for cray node 1. error: %v ", err)
	}
	if uri != node1 {
		t.Errorf("Expected uri, %s, but was %s", node0, uri)
	}

	// ---- gigabyte ----
	// /redfish/v1/Systems/Self
	node0 = "/redfish/v1/Systems/Self"
	node1 = "junk" // A second node is probably not possible with gigabyte
	systems = createRfSystems(node1, node0)
	uri, err, _ = getSystemUri(xname, 0, gigabyte, &systems)
	if err != nil {
		t.Errorf("Unexpected error for gigabyte node 0. error: %v ", err)
	}
	if uri != node0 {
		t.Errorf("Expected uri, %s, but was %s", node0, uri)
	}

	uri, err, _ = getSystemUri(xname, 1, gigabyte, &systems)
	if err == nil {
		t.Errorf("Expected error for gigabyte node 1.")
	}

	// ---- hpe ----
	// /redfish/v1/Systems/1
	node0 = "/redfish/v1/Systems/1"
	node1 = "/redfish/v1/Systems/2" // A second node is probably not possible with gigabyte
	systems = createRfSystems(node1, node0)
	uri, err, _ = getSystemUri(xname, 0, hpe, &systems)
	if err != nil {
		t.Errorf("Unexpected error for hpe node 0. error: %v ", err)
	}
	if uri != node0 {
		t.Errorf("Expected uri, %s, but was %s", node0, uri)
	}

	uri, err, _ = getSystemUri(xname, 1, hpe, &systems)
	if err != nil {
		t.Errorf("Unexpected error for hpe node 1. error: %v ", err)
	}
	if uri != node1 {
		t.Errorf("Expected uri, %s, but was %s", node1, uri)
	}

	uri, err, _ = getSystemUri(xname, 2, hpe, &systems)
	if err == nil {
		t.Errorf("Expected error for hpe node 2.")
	}

	// ---- intel ----
	// /redfish/v1/Systems/BQWF73500342
	node0 = "/redfish/v1/Systems/B0"
	node1 = "/redfish/v1/Systems/B1"
	systems = createRfSystems(node0, node1)
	uri, err, _ = getSystemUri(xname, 0, intel, &systems)
	if err != nil {
		t.Errorf("Unexpected error for intel node 0. error: %v ", err)
	}
	if uri != node0 {
		t.Errorf("Expected uri, %s, but was %s", node0, uri)
	}

	uri, err, _ = getSystemUri(xname, 1, intel, &systems)
	if err != nil {
		t.Errorf("Unexpected error for intel node 1. error: %v ", err)
	}
	if uri != node1 {
		t.Errorf("Expected uri, %s, but was %s", node1, uri)
	}

	uri, err, _ = getSystemUri(xname, 2, intel, &systems)
	if err == nil {
		t.Errorf("Expected error for intel node 2.")
	}

	uri, err, _ = getSystemUri(xname, -1, intel, &systems)
	if err == nil {
		t.Errorf("Expected error for intel node -1.")
	}
}

func TestGetNodeNumber(t *testing.T) {
	xname := "x3000c0s27b0n0"
	expected := 0
	num, err := getNodeNumber(xname)
	if err != nil {
		t.Fatalf("Unexpected error. xname: %s, error: %v", xname, err)
	}
	if num != expected {
		t.Errorf("Unexpected node number, xname: %s, expected: %d, number: %d", xname, expected, num)
	}

	xname = "x3000c0s27b0"
	_, err = getNodeNumber(xname)
	if err == nil {
		t.Errorf("Expected error. xname: %s", xname)
	}

	xname = "x3000c0r21b0n0"
	_, err = getNodeNumber(xname)
	if err == nil {
		t.Errorf("Expected error. xname: %s", xname)
	}
}

func TestValidateXname(t *testing.T) {
	goodXnames := [...]string{"x3001c0s33b0n0"}
	for _, xname := range goodXnames {
		_, err, _ := validateXname(xname)
		if err != nil {
			t.Errorf("Unexpected error for xname: %s, error: %v", xname, err)
		}
	}

	badXnames := [...]string{"", "v3000"}
	for _, xname := range badXnames {
		_, err, _ := validateXname(xname)
		if err == nil {
			t.Errorf("No error for xname %s", xname)
		}
	}
}
