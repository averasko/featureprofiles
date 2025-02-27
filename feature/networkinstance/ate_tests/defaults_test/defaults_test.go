/*
 Copyright 2022 Google LLC

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package ni_address_families_test

import (
	"testing"
	"time"

	"github.com/openconfig/featureprofiles/internal/attrs"
	"github.com/openconfig/featureprofiles/internal/deviations"
	"github.com/openconfig/featureprofiles/internal/fptest"
	"github.com/openconfig/ondatra"
	"github.com/openconfig/ondatra/gnmi"
	"github.com/openconfig/ondatra/gnmi/oc"
	"github.com/openconfig/ygot/ygot"
)

func TestMain(m *testing.M) {
	fptest.RunTests(m)
}

func assignPort(t *testing.T, d *oc.Root, intf, niName string, a *attrs.Attributes) {
	t.Helper()
	ni := d.GetOrCreateNetworkInstance(niName)
	if niName != *deviations.DefaultNetworkInstance {
		ni.Type = oc.NetworkInstanceTypes_NETWORK_INSTANCE_TYPE_L3VRF
	}
	if niName != *deviations.DefaultNetworkInstance || *deviations.ExplicitInterfaceInDefaultVRF {
		niIntf := ni.GetOrCreateInterface(intf)
		niIntf.Interface = ygot.String(intf)
		niIntf.Subinterface = ygot.Uint32(0)
	}

	ocInt := a.ConfigOCInterface(&oc.Interface{})
	ocInt.Name = ygot.String(intf)

	if err := d.AppendInterface(ocInt); err != nil {
		t.Fatalf("AddInterface(%v): cannot configure interface %s, %v", ocInt, intf, err)
	}
}

var (
	dutPort1 = &attrs.Attributes{
		IPv4:    "192.0.2.0",
		IPv4Len: 31,
		IPv6:    "2001:db8::1",
		IPv6Len: 64,
	}
	dutPort2 = &attrs.Attributes{
		IPv4:    "192.0.2.2",
		IPv4Len: 31,
		IPv6:    "2001:db8:1::1",
		IPv6Len: 64,
	}
	atePort1 = &attrs.Attributes{
		Name:    "port1",
		IPv4:    "192.0.2.1",
		IPv4Len: 31,
		IPv6:    "2001:db8::2",
		IPv6Len: 64,
		MAC:     "02:00:01:01:01:01",
	}
	atePort2 = &attrs.Attributes{
		Name:    "port2",
		IPv4:    "192.0.2.3",
		IPv4Len: 31,
		IPv6:    "2001:db8:1::2",
		IPv6Len: 64,
		MAC:     "02:00:02:01:01:01",
	}
)

// TestDefaultAddressFamilies verifies that both IPv4 and IPv6 are enabled by default without a need for additional
// configuration within a network instance. It does so by validating that simple IPv4 and IPv6 flows do not experience
// loss.
func TestDefaultAddressFamilies(t *testing.T) {
	ate := ondatra.ATE(t, "ate")
	top := ate.Topology().New()

	p1 := ate.Port(t, "port1")
	p2 := ate.Port(t, "port2")

	i1 := atePort1.AddToATE(top, p1, dutPort1)
	i2 := atePort2.AddToATE(top, p2, dutPort2)
	// Create an IPv4 flow between ATE port 1 and ATE port 2.
	ipHeader := ondatra.NewIPv4Header().
		WithSrcAddress(atePort1.IPv4).
		WithDstAddress(atePort2.IPv4)
	v4Flow := ate.Traffic().NewFlow("ipv4").
		WithSrcEndpoints(i1).
		WithDstEndpoints(i2).
		WithHeaders(ondatra.NewEthernetHeader().WithSrcAddress(atePort1.MAC), ipHeader)

	// Create an IPv6 flow between ATE port 1 and ATE port 2.
	ip6Header := ondatra.NewIPv6Header().
		WithSrcAddress(atePort1.IPv6).
		WithDstAddress(atePort2.IPv6)
	v6Flow := ate.Traffic().NewFlow("ipv6").
		WithSrcEndpoints(i1).
		WithDstEndpoints(i2).
		WithHeaders(ondatra.NewEthernetHeader().WithSrcAddress(atePort1.MAC), ip6Header)

	// Push ATE config.
	top.Push(t)

	cases := []struct {
		desc   string
		niName string
	}{
		{
			desc:   "Default network instance",
			niName: *deviations.DefaultNetworkInstance,
		},
		{
			desc:   "Non default network instance",
			niName: "xyz",
		},
	}
	dut := ondatra.DUT(t, "dut")
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			d := &oc.Root{}
			// Assign two ports into the network instance.
			assignPort(t, d, dut.Port(t, "port1").Name(), tc.niName, dutPort1)
			assignPort(t, d, dut.Port(t, "port2").Name(), tc.niName, dutPort2)

			fptest.LogQuery(t, "test configuration", gnmi.OC().Config(), d)
			gnmi.Update(t, dut, gnmi.OC().Config(), d)

			top.StartProtocols(t)
			time.Sleep(10 * time.Second)

			ate.Traffic().Start(t, v4Flow, v6Flow)
			time.Sleep(15 * time.Second)
			ate.Traffic().Stop(t)

			// Check that we did not lose any packets for the IPv4 and IPv6 flows.
			for _, flow := range []string{"ipv4", "ipv6"} {
				m := gnmi.Get(t, ate, gnmi.OC().Flow(flow).State())
				tx := m.GetCounters().GetOutPkts()
				rx := m.GetCounters().GetInPkts()
				loss := tx - rx
				lossPct := loss * 100 / tx
				if got := lossPct; got > 0 {
					t.Errorf("LossPct for flow %s: got %v, want 0", flow, got)
				}
			}
			top.StopProtocols(t)
		})
	}
}
