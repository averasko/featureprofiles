// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package qos_policy_config_test

import (
	"math"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/openconfig/featureprofiles/internal/fptest"
	"github.com/openconfig/ondatra"
	"github.com/openconfig/ondatra/gnmi"
	"github.com/openconfig/ondatra/gnmi/oc"
)

func TestMain(m *testing.M) {
	fptest.RunTests(m)
}

// QoS policy OC config:
//  - classifiers:
//    - /qos/classifiers/classifier/config/name
//    - /qos/classifiers/classifier/config/type
//    - /qos/classifiers/classifier/terms/term/actions/config/target-group
//    - /qos/classifiers/classifier/terms/term/conditions/ipv4/config/dscp-set
//    - /qos/classifiers/classifier/terms/term/conditions/ipv6/config/dscp-set
//    - /qos/classifiers/classifier/terms/term/config/id
//  - classifiers on input interface:
//    - /qos/interfaces/interface/interface-id,
//    - /qos/interfaces/interface/config/interface-id
//    - /qos/interfaces/interface/input/classifiers/classifier/type
//    - /qos/interfaces/interface/input/classifiers/classifier/config/type
//    - /qos/interfaces/interface/input/classifiers/classifier/config/name
//  - forwarding-groups:
//    - /qos/forwarding-groups/forwarding-group/config/name
//    - /qos/forwarding-groups/forwarding-group/config/output-queue
//    - /qos/queues/queue/config/name
//  - scheduler-policies:
//    - /qos/scheduler-policies/scheduler-policy/config/name
//    - /qos/scheduler-policies/scheduler-policy/schedulers/scheduler/config/priority
//    - /qos/scheduler-policies/scheduler-policy/schedulers/scheduler/config/sequence
//    - /qos/scheduler-policies/scheduler-policy/schedulers/scheduler/inputs/input/config/id
//    - /qos/scheduler-policies/scheduler-policy/schedulers/scheduler/inputs/input/config/input-type
//    - /qos/scheduler-policies/scheduler-policy/schedulers/scheduler/inputs/input/config/queue
//    - /qos/scheduler-policies/scheduler-policy/schedulers/scheduler/inputs/input/config/weight
//  - ECN queue-management-profiles:
//    - /qos/queue-management-profiles/queue-management-profile/wred/uniform/config/min-threshold
//    - /qos/queue-management-profiles/queue-management-profile/wred/uniform/config/max-threshold
//    - /qos/queue-management-profiles/queue-management-profile/wred/uniform/config/enable-ecn
//    - /qos/queue-management-profiles/queue-management-profile/wred/uniform/config/weight
//    - /qos/queue-management-profiles/queue-management-profile/wred/uniform/config/drop
//    - /qos/queue-management-profiles/queue-management-profile/wred/uniform/config/max-drop-probability-percent
//  - Output interface queue and scheduler-policy:
//    - /qos/interfaces/interface/output/queues/queue/config/queue-management-profile
//    - /qos/interfaces/interface/output/queues/queue/config/name
//    - /qos/interfaces/interface/output/scheduler-policy/config/name
//
//
// Topology:
//   ate:port1 <--> port1:dut:port2 <--> ate:port2
//
// Test notes:
//
//  Sample CLI command to get telemetry using gmic:
//   - gnmic -a ipaddr:10162 -u username -p password --skip-verify get \
//      --path /components/component --format flat
//   - gnmic tool info:
//     - https://github.com/karimra/gnmic/blob/main/README.md
//

func TestQoSClassifierConfig(t *testing.T) {
	dut := ondatra.DUT(t, "dut")
	d := &oc.Root{}
	q := d.GetOrCreateQos()

	cases := []struct {
		desc         string
		name         string
		classType    oc.E_Qos_Classifier_Type
		termID       string
		targetGrpoup string
		dscpSet      []uint8
	}{{
		desc:         "classifier_ipv4_be1",
		name:         "dscp_based_classifier_ipv4",
		classType:    oc.Qos_Classifier_Type_IPV4,
		termID:       "0",
		targetGrpoup: "target-group-BE1",
		dscpSet:      []uint8{0, 1, 2, 3},
	}, {
		desc:         "classifier_ipv4_be0",
		name:         "dscp_based_classifier_ipv4",
		classType:    oc.Qos_Classifier_Type_IPV4,
		termID:       "1",
		targetGrpoup: "target-group-BE0",
		dscpSet:      []uint8{4, 5, 6, 7},
	}, {
		desc:         "classifier_ipv4_af1",
		name:         "dscp_based_classifier_ipv4",
		classType:    oc.Qos_Classifier_Type_IPV4,
		termID:       "2",
		targetGrpoup: "target-group-AF1",
		dscpSet:      []uint8{8, 9, 10, 11},
	}, {
		desc:         "classifier_ipv4_af2",
		name:         "dscp_based_classifier_ipv4",
		classType:    oc.Qos_Classifier_Type_IPV4,
		termID:       "3",
		targetGrpoup: "target-group-AF2",
		dscpSet:      []uint8{16, 17, 18, 19},
	}, {
		desc:         "classifier_ipv4_af3",
		name:         "dscp_based_classifier_ipv4",
		classType:    oc.Qos_Classifier_Type_IPV4,
		termID:       "4",
		targetGrpoup: "target-group-AF3",
		dscpSet:      []uint8{24, 25, 26, 27},
	}, {
		desc:         "classifier_ipv4_af4",
		name:         "dscp_based_classifier_ipv4",
		classType:    oc.Qos_Classifier_Type_IPV4,
		termID:       "5",
		targetGrpoup: "target-group-AF4",
		dscpSet:      []uint8{32, 33, 34, 35},
	}, {
		desc:         "classifier_ipv4_nc1",
		name:         "dscp_based_classifier_ipv4",
		classType:    oc.Qos_Classifier_Type_IPV4,
		termID:       "6",
		targetGrpoup: "target-group-NC1",
		dscpSet:      []uint8{48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59},
	}, {
		desc:         "classifier_ipv6_be1",
		name:         "dscp_based_classifier_ipv6",
		classType:    oc.Qos_Classifier_Type_IPV6,
		termID:       "0",
		targetGrpoup: "target-group-BE1",
		dscpSet:      []uint8{0, 1, 2, 3},
	}, {
		desc:         "classifier_ipv6_be0",
		name:         "dscp_based_classifier_ipv6",
		classType:    oc.Qos_Classifier_Type_IPV6,
		termID:       "1",
		targetGrpoup: "target-group-BE0",
		dscpSet:      []uint8{4, 5, 6, 7},
	}, {
		desc:         "classifier_ipv6_af1",
		name:         "dscp_based_classifier_ipv6",
		classType:    oc.Qos_Classifier_Type_IPV6,
		termID:       "2",
		targetGrpoup: "target-group-AF1",
		dscpSet:      []uint8{8, 9, 10, 11},
	}, {
		desc:         "classifier_ipv6_af2",
		name:         "dscp_based_classifier_ipv6",
		classType:    oc.Qos_Classifier_Type_IPV6,
		termID:       "3",
		targetGrpoup: "target-group-AF2",
		dscpSet:      []uint8{16, 17, 18, 19},
	}, {
		desc:         "classifier_ipv6_af3",
		name:         "dscp_based_classifier_ipv6",
		classType:    oc.Qos_Classifier_Type_IPV6,
		termID:       "4",
		targetGrpoup: "target-group-AF3",
		dscpSet:      []uint8{24, 25, 26, 27},
	}, {
		desc:         "classifier_ipv6_af4",
		name:         "dscp_based_classifier_ipv6",
		classType:    oc.Qos_Classifier_Type_IPV6,
		termID:       "5",
		targetGrpoup: "target-group-AF4",
		dscpSet:      []uint8{32, 33, 34, 35},
	}, {
		desc:         "classifier_ipv6_nc1",
		name:         "dscp_based_classifier_ipv6",
		classType:    oc.Qos_Classifier_Type_IPV6,
		termID:       "6",
		targetGrpoup: "target-group-NC1",
		dscpSet:      []uint8{48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59},
	}}

	t.Logf("qos Classifiers config cases: %v", cases)
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			classifier := q.GetOrCreateClassifier(tc.name)
			classifier.SetName(tc.name)
			classifier.SetType(tc.classType)
			term, err := classifier.NewTerm(tc.termID)
			if err != nil {
				t.Fatalf("Failed to create classifier.NewTerm(): %v", err)
			}

			term.SetId(tc.termID)
			action := term.GetOrCreateActions()
			action.SetTargetGroup(tc.targetGrpoup)
			condition := term.GetOrCreateConditions()
			if tc.name == "dscp_based_classifier_ipv4" {
				condition.GetOrCreateIpv4().SetDscpSet(tc.dscpSet)
			} else if tc.name == "dscp_based_classifier_ipv6" {
				condition.GetOrCreateIpv6().SetDscpSet(tc.dscpSet)
			}
			gnmi.Replace(t, dut, gnmi.OC().Qos().Config(), q)
		})

		// TODO: Remove the following t.Skipf() after the config verification code has been tested.
		t.Skipf("Skip the QoS config verification until it is tested against a DUT.")

		// Verify the Classifier is applied by checking the telemetry path state values.
		classifier := gnmi.OC().Qos().Classifier(tc.name)
		term := classifier.Term(tc.termID)
		action := term.Actions()
		condition := term.Conditions()

		cmp.Equal([]uint8{1, 2, 3}, []uint8{1, 2, 3})
		cmp.Equal([]uint8{1, 2, 3}, []uint8{1, 3, 2})

		if got, want := gnmi.Get(t, dut, classifier.Name().State()), tc.name; got != want {
			t.Errorf("classifier.Name().State(): got %v, want %v", got, want)
		}
		if got, want := gnmi.Get(t, dut, classifier.Type().State()), tc.classType; got != want {
			t.Errorf("classifier.Name().Type(): got %v, want %v", got, want)
		}
		if got, want := gnmi.Get(t, dut, term.Id().State()), tc.termID; got != want {
			t.Errorf("term.Id().State(): got %v, want %v", got, want)
		}
		if got, want := gnmi.Get(t, dut, action.TargetGroup().State()), tc.targetGrpoup; got != want {
			t.Errorf("action.TargetGroup().State(): got %v, want %v", got, want)
		}

		// This Transformer sorts a []uint8.
		trans := cmp.Transformer("Sort", func(in []uint8) []uint8 {
			out := append([]uint8(nil), in...)
			sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
			return out
		})

		if tc.name == "dscp_based_classifier_ipv4" {
			if equal := cmp.Equal(condition.Ipv4().DscpSet().State(), tc.dscpSet, trans); !equal {
				t.Errorf("condition.Ipv4().DscpSet().State(): got %v, want %v", condition.Ipv4().DscpSet().State(), tc.dscpSet)
			}
		} else if tc.name == "dscp_based_classifier_ipv6" {
			if equal := cmp.Equal(condition.Ipv6().DscpSet().State(), tc.dscpSet, trans); !equal {
				t.Errorf("condition.Ipv4().DscpSet().State(): got %v, want %v", condition.Ipv6().DscpSet().State(), tc.dscpSet)
			}
		}
	}
}

func TestQoSInputIntfClassifierConfig(t *testing.T) {
	dut := ondatra.DUT(t, "dut")
	dp := dut.Port(t, "port1")

	cases := []struct {
		desc                string
		inputClassifierType oc.E_Input_Classifier_Type
		classifier          string
	}{{
		desc:                "Input Classifier Type IPV4",
		inputClassifierType: oc.Input_Classifier_Type_IPV4,
		classifier:          "dscp_based_classifier_ipv4",
	}, {
		desc:                "Input Classifier Type IPV6",
		inputClassifierType: oc.Input_Classifier_Type_IPV6,
		classifier:          "dscp_based_classifier_ipv6",
	}}

	d := &oc.Root{}
	q := d.GetOrCreateQos()
	i := q.GetOrCreateInterface(dp.Name())
	i.SetInterfaceId(dp.Name())

	t.Logf("qos input classifier config cases: %v", cases)
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			c := i.GetOrCreateInput().GetOrCreateClassifier(tc.inputClassifierType)
			c.SetType(tc.inputClassifierType)
			c.SetName(tc.classifier)
			gnmi.Replace(t, dut, gnmi.OC().Qos().Config(), q)
		})

		// TODO: Remove the following t.Skipf() after the config verification code has been tested.
		t.Skipf("Skip the QoS config verification until it is tested against a DUT.")

		// Verify the Classifier is applied on interface by checking the telemetry path state values.
		classifier := gnmi.OC().Qos().Interface(dp.Name()).Input().Classifier(tc.inputClassifierType)
		if got, want := gnmi.Get(t, dut, classifier.Name().State()), tc.classifier; got != want {
			t.Errorf("classifier.Name().State(): got %v, want %v", got, want)
		}
		if got, want := gnmi.Get(t, dut, classifier.Type().State()), tc.inputClassifierType; got != want {
			t.Errorf("classifier.Name().State(): got %v, want %v", got, want)
		}
	}
}

func TestQoSForwadingGroupsConfig(t *testing.T) {
	dut := ondatra.DUT(t, "dut")
	d := &oc.Root{}
	q := d.GetOrCreateQos()

	cases := []struct {
		desc         string
		queueName    string
		targetGrpoup string
	}{{
		desc:         "forwarding-group-BE1",
		queueName:    "BE1",
		targetGrpoup: "target-group-BE1",
	}, {
		desc:         "forwarding-group-BE0",
		queueName:    "BE0",
		targetGrpoup: "target-group-BE0",
	}, {
		desc:         "forwarding-group-AF1",
		queueName:    "AF1",
		targetGrpoup: "target-group-AF1",
	}, {
		desc:         "forwarding-group-AF2",
		queueName:    "AF2",
		targetGrpoup: "target-group-AF2",
	}, {
		desc:         "forwarding-group-AF3",
		queueName:    "AF3",
		targetGrpoup: "target-group-AF3",
	}, {
		desc:         "forwarding-group-AF4",
		queueName:    "AF4",
		targetGrpoup: "target-group-AF4",
	}, {
		desc:         "forwarding-group-NC1",
		queueName:    "NC1",
		targetGrpoup: "target-group-NC1",
	}}

	t.Logf("qos forwarding groups config cases: %v", cases)
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			fwdGroup := q.GetOrCreateForwardingGroup(tc.targetGrpoup)
			fwdGroup.SetName(tc.targetGrpoup)
			fwdGroup.SetOutputQueue(tc.queueName)
			queue := q.GetOrCreateQueue(tc.queueName)
			queue.SetName(tc.queueName)
			gnmi.Replace(t, dut, gnmi.OC().Qos().Config(), q)
		})

		// TODO: Remove the following t.Skipf() after the config verification code has been tested.
		t.Skipf("Skip the QoS config verification until it is tested against a DUT.")

		// Verify the ForwardingGroup is applied by checking the telemetry path state values.
		forwardingGroup := gnmi.OC().Qos().ForwardingGroup(tc.targetGrpoup)
		if got, want := gnmi.Get(t, dut, forwardingGroup.Name().State()), tc.targetGrpoup; got != want {
			t.Errorf("forwardingGroup.Name().State(): got %v, want %v", got, want)
		}
		if got, want := gnmi.Get(t, dut, forwardingGroup.OutputQueue().State()), tc.queueName; got != want {
			t.Errorf("forwardingGroup.OutputQueue().State(): got %v, want %v", got, want)
		}
	}
}

func TestSchedulerPoliciesConfig(t *testing.T) {
	dut := ondatra.DUT(t, "dut")
	d := &oc.Root{}
	q := d.GetOrCreateQos()

	cases := []struct {
		desc         string
		sequence     uint32
		priority     oc.E_Scheduler_Priority
		inputID      string
		inputType    oc.E_Input_InputType
		weight       uint64
		queueName    string
		targetGrpoup string
	}{{
		desc:         "scheduler-policy-BE1",
		sequence:     uint32(1),
		priority:     oc.Scheduler_Priority_UNSET,
		inputID:      "BE1",
		inputType:    oc.Input_InputType_QUEUE,
		weight:       uint64(1),
		queueName:    "BE1",
		targetGrpoup: "target-group-BE1",
	}, {
		desc:         "scheduler-policy-BE0",
		sequence:     uint32(1),
		priority:     oc.Scheduler_Priority_UNSET,
		inputID:      "BE0",
		inputType:    oc.Input_InputType_QUEUE,
		weight:       uint64(2),
		queueName:    "BE0",
		targetGrpoup: "target-group-BE0",
	}, {
		desc:         "scheduler-policy-AF1",
		sequence:     uint32(1),
		priority:     oc.Scheduler_Priority_UNSET,
		inputID:      "AF1",
		inputType:    oc.Input_InputType_QUEUE,
		weight:       uint64(4),
		queueName:    "AF1",
		targetGrpoup: "target-group-AF1",
	}, {
		desc:         "scheduler-policy-AF2",
		sequence:     uint32(1),
		priority:     oc.Scheduler_Priority_UNSET,
		inputID:      "AF2",
		inputType:    oc.Input_InputType_QUEUE,
		weight:       uint64(8),
		queueName:    "AF2",
		targetGrpoup: "target-group-AF2",
	}, {
		desc:         "scheduler-policy-AF3",
		sequence:     uint32(1),
		priority:     oc.Scheduler_Priority_UNSET,
		inputID:      "AF3",
		inputType:    oc.Input_InputType_QUEUE,
		weight:       uint64(16),
		queueName:    "AF3",
		targetGrpoup: "target-group-AF3",
	}, {
		desc:         "scheduler-policy-AF4",
		sequence:     uint32(0),
		priority:     oc.Scheduler_Priority_STRICT,
		inputID:      "AF4",
		inputType:    oc.Input_InputType_QUEUE,
		weight:       uint64(100),
		queueName:    "AF4",
		targetGrpoup: "target-group-AF4",
	}, {
		desc:         "scheduler-policy-NC1",
		sequence:     uint32(0),
		priority:     oc.Scheduler_Priority_STRICT,
		inputID:      "NC1",
		inputType:    oc.Input_InputType_QUEUE,
		weight:       uint64(200),
		queueName:    "NC1",
		targetGrpoup: "target-group-NC1",
	}}

	schedulerPolicy := q.GetOrCreateSchedulerPolicy("scheduler")
	schedulerPolicy.SetName("scheduler")
	t.Logf("qos scheduler policies config cases: %v", cases)
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			s := schedulerPolicy.GetOrCreateScheduler(tc.sequence)
			s.SetSequence(tc.sequence)
			s.SetPriority(tc.priority)
			input := s.GetOrCreateInput(tc.inputID)
			input.SetId(tc.inputID)
			input.SetInputType(tc.inputType)
			input.SetQueue(tc.queueName)
			input.SetWeight(tc.weight)
			gnmi.Replace(t, dut, gnmi.OC().Qos().Config(), q)
		})

		// TODO: Remove the following t.Skipf() after the config verification code has been tested.
		t.Skipf("Skip the QoS config verification until it is tested against a DUT.")

		// Verify the SchedulerPolicy is applied by checking the telemetry path state values.
		scheduler := gnmi.OC().Qos().SchedulerPolicy("scheduler").Scheduler(tc.sequence)
		input := scheduler.Input(tc.inputID)

		if got, want := gnmi.Get(t, dut, scheduler.Sequence().State()), tc.sequence; got != want {
			t.Errorf("scheduler.Sequence().State(): got %v, want %v", got, want)
		}
		if got, want := gnmi.Get(t, dut, scheduler.Priority().State()), tc.priority; got != want {
			t.Errorf("scheduler.Priority().State(): got %v, want %v", got, want)
		}
		if got, want := gnmi.Get(t, dut, input.Id().State()), tc.inputID; got != want {
			t.Errorf("input.Id().State(): got %v, want %v", got, want)
		}
		if got, want := gnmi.Get(t, dut, input.InputType().State()), tc.inputType; got != want {
			t.Errorf("input.InputType().State(): got %v, want %v", got, want)
		}
		if got, want := gnmi.Get(t, dut, input.Weight().State()), tc.weight; got != want {
			t.Errorf("input.Weight().State(): got %v, want %v", got, want)
		}
		if got, want := gnmi.Get(t, dut, input.Queue().State()), tc.queueName; got != want {
			t.Errorf("input.Queue().State(): got %v, want %v", got, want)
		}
	}
}

func TestECNConfig(t *testing.T) {
	dut := ondatra.DUT(t, "dut")
	d := &oc.Root{}
	q := d.GetOrCreateQos()

	ecnConfig := struct {
		ecnEnabled                bool
		dropEnabled               bool
		minThreshold              uint64
		maxThreshold              uint64
		maxDropProbabilityPercent uint8
		weight                    uint32
	}{
		ecnEnabled:                true,
		dropEnabled:               false,
		minThreshold:              uint64(80000),
		maxThreshold:              math.MaxUint64,
		maxDropProbabilityPercent: uint8(1),
		weight:                    uint32(0),
	}

	queueMgmtProfile := q.GetOrCreateQueueManagementProfile("DropProfile")
	queueMgmtProfile.SetName("DropProfile")
	wred := queueMgmtProfile.GetOrCreateWred()
	uniform := wred.GetOrCreateUniform()
	uniform.SetEnableEcn(ecnConfig.ecnEnabled)
	uniform.SetDrop(ecnConfig.dropEnabled)
	uniform.SetMinThreshold(ecnConfig.minThreshold)
	uniform.SetMaxThreshold(ecnConfig.maxThreshold)
	// TODO: uncomment the following config after it is supported.
	// uniform.SetMaxDropProbabilityPercent(ecnConfig.maxDropProbabilityPercent)
	// uniform.SetWeight(ecnConfig.weight)

	t.Logf("qos ECN QueueManagementProfile config cases: %v", ecnConfig)
	gnmi.Replace(t, dut, gnmi.OC().Qos().Config(), q)

	// TODO: Remove the following t.Skipf() after the config verification code has been tested.
	t.Skipf("Skip the QoS config verification until it is tested against a DUT.")

	// Verify the QueueManagementProfile is applied by checking the telemetry path state values.
	wredUniform := gnmi.OC().Qos().QueueManagementProfile("DropProfile").Wred().Uniform()
	if got, want := gnmi.Get(t, dut, wredUniform.EnableEcn().State()), ecnConfig.ecnEnabled; got != want {
		t.Errorf("wredUniform.EnableEcn().State(): got %v, want %v", got, want)
	}
	if got, want := gnmi.Get(t, dut, wredUniform.Drop().State()), ecnConfig.dropEnabled; got != want {
		t.Errorf("wredUniform.Drop().State(): got %v, want %v", got, want)
	}
	if got, want := gnmi.Get(t, dut, wredUniform.MinThreshold().State()), ecnConfig.minThreshold; got != want {
		t.Errorf("wredUniform.MinThreshold().State(): got %v, want %v", got, want)
	}
	if got, want := gnmi.Get(t, dut, wredUniform.MaxThreshold().State()), ecnConfig.maxThreshold; got != want {
		t.Errorf("wredUniform.MaxThreshold().State(): got %v, want %v", got, want)
	}
	if got, want := gnmi.Get(t, dut, wredUniform.MaxDropProbabilityPercent().State()), ecnConfig.maxDropProbabilityPercent; got != want {
		t.Errorf("wredUniform.MaxDropProbabilityPercent().State(): got %v, want %v", got, want)
	}
	if got, want := gnmi.Get(t, dut, wredUniform.Weight().State()), ecnConfig.weight; got != want {
		t.Errorf("wredUniform.Weight().State(): got %v, want %v", got, want)
	}
}

func TestQoSOutputIntfConfig(t *testing.T) {
	dut := ondatra.DUT(t, "dut")
	dp := dut.Port(t, "port2")

	cases := []struct {
		desc       string
		queueName  string
		ecnProfile string
		scheduler  string
	}{{
		desc:       "output-interface-BE1",
		queueName:  "BE1",
		ecnProfile: "DropProfile",
		scheduler:  "scheduler",
	}, {
		desc:       "output-interface-BE0",
		queueName:  "BE0",
		ecnProfile: "DropProfile",
		scheduler:  "scheduler",
	}, {
		desc:       "output-interface-AF1",
		queueName:  "AF1",
		ecnProfile: "DropProfile",
		scheduler:  "scheduler",
	}, {
		desc:       "output-interface-AF2",
		queueName:  "AF2",
		ecnProfile: "DropProfile",
		scheduler:  "scheduler",
	}, {
		desc:       "output-interface-AF3",
		queueName:  "AF3",
		ecnProfile: "DropProfile",
		scheduler:  "scheduler",
	}, {
		desc:       "output-interface-AF4",
		queueName:  "AF4",
		ecnProfile: "DropProfile",
		scheduler:  "scheduler",
	}, {
		desc:       "output-interface-NC1",
		queueName:  "NC1",
		ecnProfile: "DropProfile",
		scheduler:  "scheduler",
	}}

	d := &oc.Root{}
	q := d.GetOrCreateQos()
	i := q.GetOrCreateInterface(dp.Name())
	i.SetInterfaceId(dp.Name())

	t.Logf("qos output interface config cases: %v", cases)
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			output := i.GetOrCreateOutput()
			schedulerPolicy := output.GetOrCreateSchedulerPolicy()
			schedulerPolicy.SetName(tc.scheduler)
			queue := output.GetOrCreateQueue(tc.queueName)
			queue.SetQueueManagementProfile(tc.ecnProfile)
			queue.SetName(tc.queueName)
			gnmi.Replace(t, dut, gnmi.OC().Qos().Config(), q)
		})

		// TODO: Remove the following t.Skipf() after the config verification code has been tested.
		t.Skipf("Skip the QoS config verification until it is tested against a DUT.")

		// Verify the policy is applied by checking the telemetry path state values.
		policy := gnmi.OC().Qos().Interface(dp.Name()).Output().SchedulerPolicy()
		outQueue := gnmi.OC().Qos().Interface(dp.Name()).Output().Queue(tc.queueName)
		if got, want := gnmi.Get(t, dut, policy.Name().State()), tc.scheduler; got != want {
			t.Errorf("policy.Name().State(): got %v, want %v", got, want)
		}
		if got, want := gnmi.Get(t, dut, outQueue.Name().State()), tc.queueName; got != want {
			t.Errorf("outQueue.Name().State(): got %v, want %v", got, want)
		}
		if got, want := gnmi.Get(t, dut, outQueue.QueueManagementProfile().State()), tc.ecnProfile; got != want {
			t.Errorf("outQueue.QueueManagementProfile().State(): got %v, want %v", got, want)
		}
	}
}
