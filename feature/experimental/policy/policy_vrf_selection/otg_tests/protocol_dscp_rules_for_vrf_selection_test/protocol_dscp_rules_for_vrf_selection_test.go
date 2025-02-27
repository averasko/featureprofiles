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

package policy_based_vrf_selection_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/openconfig/featureprofiles/internal/attrs"
	"github.com/openconfig/featureprofiles/internal/deviations"
	"github.com/openconfig/featureprofiles/internal/fptest"
	"github.com/openconfig/featureprofiles/internal/otgutils"
	"github.com/openconfig/ondatra"
	"github.com/openconfig/ondatra/gnmi"
	"github.com/openconfig/ondatra/gnmi/oc"
	"github.com/openconfig/ygot/ygot"
)

const (
	trafficDuration = 1 * time.Minute
	ipv4PrefixLen   = 30
	ipv6PrefixLen   = 126
)

// testArgs holds the objects needed by a test case.
type testArgs struct {
	dut        *ondatra.DUTDevice
	ate        *ondatra.ATEDevice
	top        gosnappi.Config
	policyName string
	iptype     string
	protocol   oc.E_PacketMatchTypes_IP_PROTOCOL
}

var (
	dutPort1 = attrs.Attributes{
		Desc:    "dutPort1",
		MAC:     "02:00:01:01:01:01",
		IPv4:    "192.0.2.1",
		IPv4Len: ipv4PrefixLen,
		IPv6:    "2001:0db8::192:0:2:1",
		IPv6Len: ipv6PrefixLen,
	}

	atePort1 = attrs.Attributes{
		Name:    "atePort1",
		MAC:     "02:00:02:01:01:01",
		IPv4:    "192.0.2.2",
		IPv4Len: ipv4PrefixLen,
		IPv6:    "2001:0db8::192:0:2:2",
		IPv6Len: ipv6PrefixLen,
	}

	dutPort2 = attrs.Attributes{
		Desc:    "dutPort2",
		MAC:     "01:00:01:01:01:01",
		IPv4:    "192.0.2.5",
		IPv4Len: ipv4PrefixLen,
		IPv6:    "2001:0db8::192:0:2:5",
		IPv6Len: ipv6PrefixLen,
	}

	atePort2 = attrs.Attributes{
		Name:    "atePort2",
		MAC:     "01:00:02:01:01:01",
		IPv4:    "192.0.2.6",
		IPv4Len: ipv4PrefixLen,
		IPv6:    "2001:0db8::192:0:2:6",
		IPv6Len: ipv6PrefixLen,
	}

	dutPort2Vlan10 = attrs.Attributes{
		Desc:    "dutPort2Vlan10",
		MAC:     "01:00:01:01:01:01",
		IPv4:    "192.0.2.9",
		IPv4Len: ipv4PrefixLen,
		IPv6:    "2001:0db8::192:0:2:9",
		IPv6Len: ipv6PrefixLen,
	}

	atePort2Vlan10 = attrs.Attributes{
		Name:    "atePort2Vlan10",
		MAC:     "01:00:02:01:01:01",
		IPv4:    "192.0.2.10",
		IPv4Len: ipv4PrefixLen,
		IPv6:    "2001:0db8::192:0:2:a",
		IPv6Len: ipv6PrefixLen,
	}

	dutPort2Vlan20 = attrs.Attributes{
		Desc:    "dutPort2Vlan20",
		MAC:     "01:00:01:01:01:01",
		IPv4:    "192.0.2.13",
		IPv4Len: ipv4PrefixLen,
		IPv6:    "2001:0db8::192:0:2:d",
		IPv6Len: ipv6PrefixLen,
	}

	atePort2Vlan20 = attrs.Attributes{
		Name:    "atePort2Vlan20",
		MAC:     "01:00:02:01:01:01",
		IPv4:    "192.0.2.14",
		IPv4Len: ipv4PrefixLen,
		IPv6:    "2001:0db8::192:0:2:e",
		IPv6Len: ipv6PrefixLen,
	}

	dutPort2Vlan30 = attrs.Attributes{
		Desc:    "dutPort2Vlan30",
		MAC:     "01:00:01:01:01:01",
		IPv4:    "192.0.2.17",
		IPv4Len: ipv4PrefixLen,
		IPv6:    "2001:0db8::192:0:2:11",
		IPv6Len: ipv6PrefixLen,
	}

	atePort2Vlan30 = attrs.Attributes{
		Name:    "atePort2Vlan30",
		MAC:     "01:00:02:01:01:01",
		IPv4:    "192.0.2.18",
		IPv4Len: ipv4PrefixLen,
		IPv6:    "2001:0db8::192:0:2:12",
		IPv6Len: ipv6PrefixLen,
	}
)

func TestMain(m *testing.M) {
	fptest.RunTests(m)
}

// configureATE configures port1, port2 and vlans on port2 on the ATE.
func configureATE(t *testing.T, ate *ondatra.ATEDevice) gosnappi.Config {
	top := ate.OTG().NewConfig(t)

	p1 := ate.Port(t, "port1")
	top.Ports().Add().SetName(p1.ID())
	srcDev := top.Devices().Add().SetName(atePort1.Name)
	ethSrc := srcDev.Ethernets().Add().SetName(atePort1.Name + ".eth").SetMac(atePort1.MAC)
	ethSrc.Connection().SetChoice(gosnappi.EthernetConnectionChoice.PORT_NAME).SetPortName(p1.ID())
	ethSrc.Ipv4Addresses().Add().SetName(srcDev.Name() + ".ipv4").SetAddress(atePort1.IPv4).SetGateway(dutPort1.IPv4).SetPrefix(int32(atePort1.IPv4Len))
	ethSrc.Ipv6Addresses().Add().SetName(srcDev.Name() + ".ipv6").SetAddress(atePort1.IPv6).SetGateway(dutPort1.IPv6).SetPrefix(int32(atePort1.IPv6Len))

	p2 := ate.Port(t, "port2")
	top.Ports().Add().SetName(p2.ID())
	dstDev := top.Devices().Add().SetName(atePort2.Name)
	ethDst := dstDev.Ethernets().Add().SetName(atePort2.Name + ".eth").SetMac(atePort2.MAC)
	ethDst.Connection().SetChoice(gosnappi.EthernetConnectionChoice.PORT_NAME).SetPortName(p2.ID())
	ethDst.Ipv4Addresses().Add().SetName(dstDev.Name() + ".ipv4").SetAddress(atePort2.IPv4).SetGateway(dutPort2.IPv4).SetPrefix(int32(atePort2.IPv4Len))
	ethDst.Ipv6Addresses().Add().SetName(dstDev.Name() + ".ipv6").SetAddress(atePort2.IPv6).SetGateway(dutPort2.IPv6).SetPrefix(int32(atePort2.IPv6Len))

	// configure vlans on ATE port2
	dstDevVlan10 := top.Devices().Add().SetName(atePort2Vlan10.Name)
	ethDstVlan10 := dstDevVlan10.Ethernets().Add().SetName(atePort2Vlan10.Name + ".eth").SetMac(atePort2Vlan10.MAC)
	ethDstVlan10.Connection().SetChoice(gosnappi.EthernetConnectionChoice.PORT_NAME).SetPortName(p2.ID())
	ethDstVlan10.Vlans().Add().SetName(atePort2Vlan10.Name + "vlan").SetId(10)
	ethDstVlan10.Ipv4Addresses().Add().SetName(atePort2Vlan10.Name + ".ipv4").SetAddress(atePort2Vlan10.IPv4).SetGateway(dutPort2Vlan10.IPv4).SetPrefix(int32(atePort2Vlan10.IPv4Len))
	ethDstVlan10.Ipv6Addresses().Add().SetName(atePort2Vlan10.Name + ".ipv6").SetAddress(atePort2Vlan10.IPv6).SetGateway(dutPort2Vlan10.IPv6).SetPrefix(int32(atePort2Vlan10.IPv6Len))

	dstDevVlan20 := top.Devices().Add().SetName(atePort2Vlan20.Name)
	ethDstVlan20 := dstDevVlan20.Ethernets().Add().SetName(atePort2Vlan20.Name + ".eth").SetMac(atePort2Vlan20.MAC)
	ethDstVlan20.Connection().SetChoice(gosnappi.EthernetConnectionChoice.PORT_NAME).SetPortName(p2.ID())
	ethDstVlan20.Vlans().Add().SetName(atePort2Vlan20.Name + "vlan").SetId(20)
	ethDstVlan20.Ipv4Addresses().Add().SetName(atePort2Vlan20.Name + ".ipv4").SetAddress(atePort2Vlan20.IPv4).SetGateway(dutPort2Vlan20.IPv4).SetPrefix(int32(atePort2Vlan20.IPv4Len))
	ethDstVlan20.Ipv6Addresses().Add().SetName(atePort2Vlan20.Name + ".ipv6").SetAddress(atePort2Vlan20.IPv6).SetGateway(dutPort2Vlan20.IPv6).SetPrefix(int32(atePort2Vlan20.IPv6Len))

	dstDevVlan30 := top.Devices().Add().SetName(atePort2Vlan30.Name)
	ethDstVlan30 := dstDevVlan30.Ethernets().Add().SetName(atePort2Vlan30.Name + ".eth").SetMac(atePort2Vlan30.MAC)
	ethDstVlan30.Connection().SetChoice(gosnappi.EthernetConnectionChoice.PORT_NAME).SetPortName(p2.ID())
	ethDstVlan30.Vlans().Add().SetName(atePort2Vlan30.Name + "vlan").SetId(30)
	ethDstVlan30.Ipv4Addresses().Add().SetName(atePort2Vlan30.Name + ".ipv4").SetAddress(atePort2Vlan30.IPv4).SetGateway(dutPort2Vlan30.IPv4).SetPrefix(int32(atePort2Vlan30.IPv4Len))
	ethDstVlan30.Ipv6Addresses().Add().SetName(atePort2Vlan30.Name + ".ipv6").SetAddress(atePort2Vlan30.IPv6).SetGateway(dutPort2Vlan30.IPv6).SetPrefix(int32(atePort2Vlan30.IPv6Len))

	return top
}

// configNetworkInstance creates VRFs and subinterfaces and then applies VRFs on the subinterfaces.
func configNetworkInstance(t *testing.T, dut *ondatra.DUTDevice, vrfname string, intfname string, subint uint32) {
	// create empty subinterface
	si := &oc.Interface_Subinterface{}
	si.Index = ygot.Uint32(subint)
	gnmi.Replace(t, dut, gnmi.OC().Interface(intfname).Subinterface(subint).Config(), si)

	// create vrf and apply on subinterface
	v := &oc.NetworkInstance{
		Name: ygot.String(vrfname),
	}
	vi := v.GetOrCreateInterface(intfname + "." + strconv.Itoa(int(subint)))
	vi.Subinterface = ygot.Uint32(subint)
	gnmi.Replace(t, dut, gnmi.OC().NetworkInstance(vrfname).Config(), v)
}

// getSubInterface returns a subinterface configuration populated with IP addresses and VLAN ID.
func getSubInterface(dutPort *attrs.Attributes, index uint32, vlanID uint16) *oc.Interface_Subinterface {
	s := &oc.Interface_Subinterface{}
	// unshut sub/interface
	if *deviations.InterfaceEnabled {
		s.Enabled = ygot.Bool(true)
	}
	s.Index = ygot.Uint32(index)
	s4 := s.GetOrCreateIpv4()
	a := s4.GetOrCreateAddress(dutPort.IPv4)
	a.PrefixLength = ygot.Uint8(dutPort.IPv4Len)
	s6 := s.GetOrCreateIpv6()
	a6 := s6.GetOrCreateAddress(dutPort.IPv6)
	a6.PrefixLength = ygot.Uint8(dutPort.IPv6Len)
	if index != 0 {
		if *deviations.DeprecatedVlanID {
			s.GetOrCreateVlan().VlanId = oc.UnionUint16(vlanID)
		} else {
			s.GetOrCreateVlan().GetOrCreateMatch().GetOrCreateSingleTagged().VlanId = ygot.Uint16(vlanID)
		}
	}
	return s
}

// configInterfaceDUT configures the interface with the Addrs.
func configInterfaceDUT(i *oc.Interface, dutPort *attrs.Attributes) *oc.Interface {
	i.Description = ygot.String(dutPort.Desc)
	i.Type = oc.IETFInterfaces_InterfaceType_ethernetCsmacd
	i.AppendSubinterface(getSubInterface(dutPort, 0, 0))
	return i
}

// configureDUT configures the base configuration on the DUT.
func configureDUT(t *testing.T, dut *ondatra.DUTDevice) {
	d := gnmi.OC()

	p1 := dut.Port(t, "port1")
	i1 := &oc.Interface{Name: ygot.String(p1.Name())}
	gnmi.Replace(t, dut, d.Interface(p1.Name()).Config(), configInterfaceDUT(i1, &dutPort1))

	p2 := dut.Port(t, "port2")
	i2 := &oc.Interface{Name: ygot.String(p2.Name())}
	gnmi.Replace(t, dut, d.Interface(p2.Name()).Config(), configInterfaceDUT(i2, &dutPort2))

	if *deviations.ExplicitPortSpeed {
		fptest.SetPortSpeed(t, p1)
		fptest.SetPortSpeed(t, p2)
	}
	if *deviations.ExplicitInterfaceInDefaultVRF {
		fptest.AssignToNetworkInstance(t, dut, p1.Name(), *deviations.DefaultNetworkInstance, 0)
		fptest.AssignToNetworkInstance(t, dut, p2.Name(), *deviations.DefaultNetworkInstance, 0)
	}

	outpath := d.Interface(p2.Name())
	// create VRFs and VRF enabled subinterfaces
	configNetworkInstance(t, dut, "VRF10", p2.Name(), uint32(1))

	// configure IP addresses on subinterfaces
	gnmi.Update(t, dut, outpath.Subinterface(1).Config(), getSubInterface(&dutPort2Vlan10, 1, 10))

	configNetworkInstance(t, dut, "VRF20", p2.Name(), uint32(2))
	gnmi.Update(t, dut, outpath.Subinterface(2).Config(), getSubInterface(&dutPort2Vlan20, 2, 20))

	configNetworkInstance(t, dut, "VRF30", p2.Name(), uint32(3))
	gnmi.Update(t, dut, outpath.Subinterface(3).Config(), getSubInterface(&dutPort2Vlan30, 3, 30))
}

// getIPinIPFlow returns an IPv4inIPv4 *ondatra.Flow with provided DSCP value for a given set of endpoints.
func getIPinIPFlow(args *testArgs, src attrs.Attributes, dst attrs.Attributes, flowName string, dscp int32) gosnappi.Flow {

	flow := gosnappi.NewFlow().SetName(flowName)
	flow.Metrics().SetEnable(true)
	flow.TxRx().Device().SetTxNames([]string{src.Name + "." + args.iptype}).SetRxNames([]string{dst.Name + "." + args.iptype})
	ethHeader := flow.Packet().Add().Ethernet()
	ethHeader.Src().SetValue(src.MAC)
	outerIPHeader := flow.Packet().Add().Ipv4()
	outerIPHeader.Src().SetValue(src.IPv4)
	outerIPHeader.Dst().SetValue(dst.IPv4)
	outerIPHeader.Priority().Dscp().Phb().SetValue(dscp)
	innerIPHeader := flow.Packet().Add().Ipv4()
	innerIPHeader.Src().SetValue("198.51.100.1")
	innerIPHeader.Dst().Increment().SetStart("203.0.113.1").SetStep("0.0.0.1").SetCount(10000)

	flow.Size().SetFixed(1024)
	flow.Rate().SetPps(100)
	flow.Duration().FixedPackets().SetPackets(100)

	return flow
}

// testTrafficFlows verifies traffic for one or more flows.
func testTrafficFlows(t *testing.T, args *testArgs, expectPass bool, flows ...gosnappi.Flow) {

	args.top.Flows().Clear()
	for _, flow := range flows {
		args.top.Flows().Append(flow)
	}
	args.ate.OTG().PushConfig(t, args.top)
	args.ate.OTG().StartProtocols(t)

	t.Logf("*** Starting traffic ...")
	args.ate.OTG().StartTraffic(t)
	time.Sleep(trafficDuration)
	t.Logf("*** Stop traffic ...")
	args.ate.OTG().StopTraffic(t)

	if expectPass {
		t.Log("Expecting traffic to pass for the flows")
	} else {
		t.Log("Expecting traffic to fail for the flows")
	}

	top := args.ate.OTG().FetchConfig(t)
	otgutils.LogFlowMetrics(t, args.ate.OTG(), top)
	for _, flow := range flows {
		t.Run(flow.Name(), func(t *testing.T) {
			t.Logf("*** Verifying %v traffic on OTG ... ", flow.Name())
			outPkts := gnmi.Get(t, args.ate.OTG(), gnmi.OTG().Flow(flow.Name()).Counters().OutPkts().State())
			inPkts := gnmi.Get(t, args.ate.OTG(), gnmi.OTG().Flow(flow.Name()).Counters().InPkts().State())

			if outPkts == 0 {
				t.Fatalf("OutPkts == 0, want >0.")
			}

			lossPct := ((outPkts - inPkts) * 100) / outPkts

			// log stats
			t.Log("Flow LossPct: ", lossPct)
			t.Log("Flow InPkts  : ", inPkts)
			t.Log("Flow OutPkts : ", outPkts)

			if (expectPass == true) && (lossPct == 0) {
				t.Logf("Traffic for %v flow is passing as expected", flow.Name())
			} else if (expectPass == false) && (lossPct == 100) {
				t.Logf("Traffic for %v flow is failing as expected", flow.Name())
			} else {
				t.Fatalf("Traffic is not working as expected for flow: %v.", flow.Name())
			}
		})
	}
}

// getL3PBRRule returns an IPv4 or IPv6 policy-forwarding rule configuration populated with protocol and/or DSCPset information.
func getL3PBRRule(args *testArgs, networkInstance string, index uint32, dscpset []uint8) *oc.NetworkInstance_PolicyForwarding_Policy_Rule {
	r := oc.NetworkInstance_PolicyForwarding_Policy_Rule{}
	r.SequenceId = ygot.Uint32(index)
	r.Action = &oc.NetworkInstance_PolicyForwarding_Policy_Rule_Action{NetworkInstance: ygot.String(networkInstance)}
	if args.iptype == "ipv4" {
		r.Ipv4 = &oc.NetworkInstance_PolicyForwarding_Policy_Rule_Ipv4{
			Protocol: args.protocol,
		}
		if len(dscpset) > 0 {
			r.Ipv4.DscpSet = dscpset
		}
	} else if args.iptype == "ipv6" {
		r.Ipv6 = &oc.NetworkInstance_PolicyForwarding_Policy_Rule_Ipv6{
			Protocol: args.protocol,
		}
		if len(dscpset) > 0 {
			r.Ipv6.DscpSet = dscpset
		}
	} else {
		return nil
	}
	return &r

}

// getPBRPolicyForwarding returns pointer to policy-forwarding populated with pbr policy and rules
func getPBRPolicyForwarding(args *testArgs, rules ...*oc.NetworkInstance_PolicyForwarding_Policy_Rule) *oc.NetworkInstance_PolicyForwarding {
	pf := oc.NetworkInstance_PolicyForwarding{}
	p := pf.GetOrCreatePolicy(args.policyName)
	p.Type = oc.Policy_Type_VRF_SELECTION_POLICY
	for _, rule := range rules {
		p.AppendRule(rule)
	}
	return &pf
}

func TestPBR(t *testing.T) {
	t.Logf("Description: Test RT3.2 with multiple DSCP, IPinIP protocol rule based VRF selection")
	dut := ondatra.DUT(t, "dut")

	// configure DUT
	configureDUT(t, dut)

	// Configure ATE
	ate := ondatra.ATE(t, "ate")
	top := configureATE(t, ate)
	ate.OTG().PushConfig(t, top)
	ate.OTG().StartProtocols(t)

	args := &testArgs{
		dut:        dut,
		ate:        ate,
		top:        top,
		policyName: "L3",
		iptype:     "ipv4",
		protocol:   oc.PacketMatchTypes_IP_PROTOCOL_IP_IN_IP,
	}

	// dut ingress interface
	port1 := dut.Port(t, "port1")

	cases := []struct {
		name         string
		desc         string
		policy       *oc.NetworkInstance_PolicyForwarding
		passingFlows []gosnappi.Flow
		failingFlows []gosnappi.Flow
	}{
		{
			name: "RT3.2 Case1",
			desc: "Ensure matching IPinIP with DSCP (10 - VRF10, 20- VRF20, 30-VRF30) traffic reaches appropriate VLAN.",
			policy: getPBRPolicyForwarding(args,
				getL3PBRRule(args, "VRF10", 1, []uint8{10}),
				getL3PBRRule(args, "VRF20", 2, []uint8{20}),
				getL3PBRRule(args, "VRF30", 3, []uint8{30})),
			// use IPinIP DSCP10, DSCP20, DSCP30 flows for VLAN10, VLAN20 and VLAN30 respectively.
			passingFlows: []gosnappi.Flow{
				getIPinIPFlow(args, atePort1, atePort2Vlan10, "ipinipd10", 10),
				getIPinIPFlow(args, atePort1, atePort2Vlan20, "ipinipd20", 20),
				getIPinIPFlow(args, atePort1, atePort2Vlan30, "ipinipd30", 30)},
		},
		{
			name: "RT3.2 Case2",
			desc: "Ensure matching IPinIP with DSCP (10-12 - VRF10, 20-22- VRF20, 30-32-VRF30) traffic reaches appropriate VLAN.",
			policy: getPBRPolicyForwarding(args,
				getL3PBRRule(args, "VRF10", 1, []uint8{10, 11, 12}),
				getL3PBRRule(args, "VRF20", 2, []uint8{20, 21, 22}),
				getL3PBRRule(args, "VRF30", 3, []uint8{30, 31, 32})),
			// use IPinIP flows with DSCP10-12 for VLAN10, DSCP20-22 for VLAN20, DSCP30-32 for VLAN30.
			passingFlows: []gosnappi.Flow{
				getIPinIPFlow(args, atePort1, atePort2Vlan10, "ipinipd10", 10),
				getIPinIPFlow(args, atePort1, atePort2Vlan10, "ipinipd11", 11),
				getIPinIPFlow(args, atePort1, atePort2Vlan10, "ipinipd12", 12),

				getIPinIPFlow(args, atePort1, atePort2Vlan20, "ipinipd20", 20),
				getIPinIPFlow(args, atePort1, atePort2Vlan20, "ipinipd21", 21),
				getIPinIPFlow(args, atePort1, atePort2Vlan20, "ipinipd22", 22),

				getIPinIPFlow(args, atePort1, atePort2Vlan30, "ipinipd30", 30),
				getIPinIPFlow(args, atePort1, atePort2Vlan30, "ipinipd31", 31),
				getIPinIPFlow(args, atePort1, atePort2Vlan30, "ipinipd32", 32)},
		},
		{
			name: "RT3.2 Case3",
			desc: "Ensure first matching of IPinIP with DSCP (10-12 - VRF10, 10-12 - VRF20) rule takes precedence.",
			policy: getPBRPolicyForwarding(args,
				getL3PBRRule(args, "VRF10", 1, []uint8{10, 11, 12}),
				getL3PBRRule(args, "VRF20", 2, []uint8{10, 11, 12})),
			// use IPinIP DSCP10-12 flows for VLAN10 as well as VLAN20.
			passingFlows: []gosnappi.Flow{
				getIPinIPFlow(args, atePort1, atePort2Vlan10, "ipinipd10", 10),
				getIPinIPFlow(args, atePort1, atePort2Vlan10, "ipinipd11", 11),
				getIPinIPFlow(args, atePort1, atePort2Vlan10, "ipinipd12", 12)},
			failingFlows: []gosnappi.Flow{
				getIPinIPFlow(args, atePort1, atePort2Vlan20, "ipinipd10v20", 10),
				getIPinIPFlow(args, atePort1, atePort2Vlan20, "ipinipd11v20", 11),
				getIPinIPFlow(args, atePort1, atePort2Vlan20, "ipinipd12v20", 12)},
		},
		{
			name: "RT3.2 Case4",
			desc: "Ensure matching IPinIP to VRF10, IPinIP with DSCP20 to VRF20 causes unspecified DSCP IPinIP traffic to match VRF10.",
			policy: getPBRPolicyForwarding(args,
				getL3PBRRule(args, "VRF10", 1, []uint8{}),
				getL3PBRRule(args, "VRF20", 2, []uint8{20})),
			// use IPinIP DSCP10-12 flows to match IPinIP to VRF10
			// use IPinIP DSCP20 flow to match to VRF20
			// use IPinIP DSCP10-12 flows to match to VRF20 to show they fail for VRF20
			passingFlows: []gosnappi.Flow{
				getIPinIPFlow(args, atePort1, atePort2Vlan10, "ipinipd10", 10),
				getIPinIPFlow(args, atePort1, atePort2Vlan10, "ipinipd11", 11),
				getIPinIPFlow(args, atePort1, atePort2Vlan10, "ipinipd12", 12)},
			failingFlows: []gosnappi.Flow{
				getIPinIPFlow(args, atePort1, atePort2Vlan20, "ipinipd10v20", 10),
				getIPinIPFlow(args, atePort1, atePort2Vlan20, "ipinipd11v20", 11),
				getIPinIPFlow(args, atePort1, atePort2Vlan20, "ipinipd12v20", 12),
				getIPinIPFlow(args, atePort1, atePort2Vlan20, "ipinipd20", 20)},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Log(tc.desc)
			pfpath := gnmi.OC().NetworkInstance(*deviations.DefaultNetworkInstance).PolicyForwarding()

			//configure pbr policy-forwarding
			gnmi.Replace(t, dut, gnmi.OC().NetworkInstance(*deviations.DefaultNetworkInstance).PolicyForwarding().Config(), tc.policy)
			// defer cleaning policy-forwarding
			defer gnmi.Delete(t, args.dut, pfpath.Config())

			// apply pbr policy on ingress interface
			gnmi.Replace(t, args.dut, pfpath.Interface(port1.Name()).ApplyVrfSelectionPolicy().Config(), args.policyName)

			// defer deletion of policy from interface
			defer gnmi.Delete(t, args.dut, pfpath.Interface(port1.Name()).ApplyVrfSelectionPolicy().Config())

			// traffic should pass
			testTrafficFlows(t, args, true, tc.passingFlows...)

			if len(tc.failingFlows) > 0 {
				// traffic should fail
				testTrafficFlows(t, args, false, tc.failingFlows...)
			}
		})
	}
}
