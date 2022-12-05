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

package backup_nhg_multiple_nh_test

import (
	"context"
	"testing"
	"time"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/openconfig/gribigo/fluent"
	"github.com/openconfig/ondatra"
	"github.com/openconfig/ondatra/telemetry"
	"github.com/openconfig/ygot/ygot"

	"github.com/openconfig/featureprofiles/internal/attrs"
	"github.com/openconfig/featureprofiles/internal/deviations"
	"github.com/openconfig/featureprofiles/internal/fptest"
	"github.com/openconfig/featureprofiles/internal/gribi"
	"github.com/openconfig/featureprofiles/internal/otgutils"
	"github.com/openconfig/ondatra/gnmi"
	otgtelemetry "github.com/openconfig/ondatra/telemetry/otg"
	"github.com/openconfig/ygnmi/ygnmi"
)

const (
	ipv4PrefixLen = 30
	ipv6PrefixLen = 126
	dstPfx        = "203.0.113.0/24"
	dstPfxMin     = "203.0.113.0"
	dstPfxMask    = "24"
)

// testArgs holds the objects needed by a test case.
type testArgs struct {
	dut    *ondatra.DUTDevice
	ate    *ondatra.ATEDevice
	top    gosnappi.Config
	ctx    context.Context
	client *gribi.Client
}

var (
	dutPort1 = attrs.Attributes{
		Desc:    "dutPort1",
		IPv4:    "192.0.2.1",
		IPv4Len: ipv4PrefixLen,
		IPv6:    "2001:0db8::192:0:2:1",
		IPv6Len: ipv6PrefixLen,
	}

	atePort1 = attrs.Attributes{
		Name:    "atePort1",
		MAC:     "02:00:01:01:01:01",
		IPv4:    "192.0.2.2",
		IPv4Len: ipv4PrefixLen,
		IPv6:    "2001:0db8::192:0:2:2",
		IPv6Len: ipv6PrefixLen,
	}

	dutPort2 = attrs.Attributes{
		Desc:    "dutPort2",
		IPv4:    "192.0.2.5",
		IPv4Len: ipv4PrefixLen,
		IPv6:    "2001:0db8::192:0:2:5",
		IPv6Len: ipv6PrefixLen,
	}

	atePort2 = attrs.Attributes{
		Name:    "atePort2",
		MAC:     "02:00:02:01:01:01",
		IPv4:    "192.0.2.6",
		IPv4Len: ipv4PrefixLen,
		IPv6:    "2001:0db8::192:0:2:6",
		IPv6Len: ipv6PrefixLen,
	}

	dutPort3 = attrs.Attributes{
		Desc:    "dutPort3",
		IPv4:    "192.0.2.9",
		IPv4Len: ipv4PrefixLen,
		IPv6:    "2001:0db8::192:0:2:9",
		IPv6Len: ipv6PrefixLen,
	}

	atePort3 = attrs.Attributes{
		Name:    "atePort3",
		MAC:     "02:00:03:01:01:01",
		IPv4:    "192.0.2.10",
		IPv4Len: ipv4PrefixLen,
		IPv6:    "2001:0db8::192:0:2:a",
		IPv6Len: ipv6PrefixLen,
	}

	dutPort4 = attrs.Attributes{
		Desc:    "dutPort4",
		IPv4:    "192.0.2.13",
		IPv4Len: ipv4PrefixLen,
		IPv6:    "2001:0db8::192:0:2:D",
		IPv6Len: ipv6PrefixLen,
	}

	atePort4 = attrs.Attributes{
		Name:    "atePort4",
		MAC:     "02:00:04:01:01:01",
		IPv4:    "192.0.2.14",
		IPv4Len: ipv4PrefixLen,
		IPv6:    "2001:0db8::192:0:2:E",
		IPv6Len: ipv6PrefixLen,
	}
)

func TestMain(m *testing.M) {
	fptest.RunTests(m)
}

// configureATE configures ports on the ATE.
func configureATE(t *testing.T, ate *ondatra.ATEDevice) gosnappi.Config {
	top := ate.OTG().NewConfig(t)

	p1 := ate.Port(t, "port1")
	p2 := ate.Port(t, "port2")
	p3 := ate.Port(t, "port3")
	p4 := ate.Port(t, "port4")

	atePort1.AddToOTG(top, p1, &dutPort1)
	atePort2.AddToOTG(top, p2, &dutPort2)
	atePort3.AddToOTG(top, p3, &dutPort3)
	atePort4.AddToOTG(top, p4, &dutPort4)

	return top
}

// configureDUT configures port1, port2, port3 and port4 on the DUT.
func configureDUT(t *testing.T, dut *ondatra.DUTDevice) {
	d := gnmi.OC()

	p1 := dut.Port(t, "port1")
	gnmi.Replace(t, dut, d.Interface(p1.Name()).Config(), dutPort1.NewOCInterface(p1.Name()))

	p2 := dut.Port(t, "port2")
	gnmi.Replace(t, dut, d.Interface(p2.Name()).Config(), dutPort2.NewOCInterface(p2.Name()))

	p3 := dut.Port(t, "port3")
	gnmi.Replace(t, dut, d.Interface(p3.Name()).Config(), dutPort3.NewOCInterface(p3.Name()))

	p4 := dut.Port(t, "port4")
	gnmi.Replace(t, dut, d.Interface(p4.Name()).Config(), dutPort4.NewOCInterface(p4.Name()))
}

func TestBackup(t *testing.T) {
	ctx := context.Background()
	dut := ondatra.DUT(t, "dut")

	//configure DUT
	configureDUT(t, dut)

	// Configure ATE
	ate := ondatra.ATE(t, "ate")
	top := configureATE(t, ate)
	ate.OTG().PushConfig(t, top)
	ate.OTG().StartProtocols(t)
	waitOTGARPEntry(t)

	t.Run("IPv4BackUpSwitch", func(t *testing.T) {
		t.Logf("Name: IPv4BackUpSwitch")
		t.Logf("Description: Set primary and backup path with gribi and shutdown the primary path validating traffic switching over backup path")

		// Configure the gRIBI client clientA
		client := gribi.Client{
			DUT:         dut,
			FIBACK:      false,
			Persistence: true,
		}
		defer client.Close(t)

		// Flush all entries after the test
		defer client.FlushAll(t)

		if err := client.Start(t); err != nil {
			t.Fatalf("gRIBI Connection can not be established")
		}
		// Client becomes leader
		client.BecomeLeader(t)

		// Flush past entries before running the tc
		client.FlushAll(t)

		tcArgs := &testArgs{
			ctx:    ctx,
			dut:    dut,
			client: &client,
			ate:    ate,
			top:    top,
		}
		testIPv4BackUpSwitch(ctx, t, tcArgs)
	})
}

// testIPv4BackUpSwitch Ensure that backup NHGs are honoured with NextHopGroup entries containing >1 NH
//
// Setup Steps
//   - Connect ATE port-1 to DUT port-1.
//   - Connect ATE port-2 to DUT port-2.
//   - Connect ATE port-3 to DUT port-3.
//   - Connect ATE port-4 to DUT port-4.
//   - Connect a gRIBI client to the DUT and inject an IPv4Entry for 203.0.113.0/24 pointing to a NextHopGroup containing:
//   - Two primary next-hops:
//   - 2: to ATE port-2
//   - 3: to ATE port-3.
//   - A backup NHG containing a single next-hop:
//   - 4: to ATE port-4.
//   - Ensure that traffic forwarded to a destination in 203.0.113.0/24 is received at ATE port-2 and port-3.
//   - Disable ATE port-2. Ensure that traffic for a destination in 203.0.113.0/24 is received at ATE port-3.
//   - Disable ATE port-3. Ensure that traffic for a destination in 203.0.113.0/24 is received at ATE port-4.
//
// Validation Steps
//   - Verify AFT telemetry after shutting each port
//   - Verify traffic switches to the right ports
func testIPv4BackUpSwitch(ctx context.Context, t *testing.T, args *testArgs) {

	const (
		// Next hop group adjacency identifier.
		NHGID = 100
		// Backup next hop group ID that the dstPfx will forward to.
		BackupNHGID = 200

		NH1ID, NH2ID, NH3ID = 1001, 1002, 1003
	)
	t.Logf("Program a backup pointing to ATE port-4 via gRIBI")
	args.client.AddNH(t, NH3ID, atePort4.IPv4, *deviations.DefaultNetworkInstance, fluent.InstalledInRIB)
	args.client.AddNHG(t, BackupNHGID, map[uint64]uint64{NH3ID: 10}, *deviations.DefaultNetworkInstance, fluent.InstalledInRIB)

	t.Logf("an IPv4Entry for %s pointing to ATE port-2 and port-3 via gRIBI", dstPfx)
	args.client.AddNH(t, NH1ID, atePort2.IPv4, *deviations.DefaultNetworkInstance, fluent.InstalledInRIB)
	args.client.AddNH(t, NH2ID, atePort3.IPv4, *deviations.DefaultNetworkInstance, fluent.InstalledInRIB)
	args.client.AddNHG(t, NHGID, map[uint64]uint64{NH1ID: 80, NH2ID: 20}, *deviations.DefaultNetworkInstance, fluent.InstalledInRIB, &gribi.NHGOptions{BackupNHG: BackupNHGID})
	args.client.AddIPv4(t, dstPfx, NHGID, *deviations.DefaultNetworkInstance, *deviations.DefaultNetworkInstance, fluent.InstalledInRIB)

	// create flow
	dstMac := args.ate.OTG().Telemetry().Interface(atePort1.Name + ".Eth").Ipv4Neighbor(dutPort1.IPv4).LinkLayerAddress().Get(t)
	BaseFlow := createFlow(t, args.ate, args.top, "BaseFlow", dstMac)

	// validate programming using AFT
	aftCheck(t, args.dut, dstPfx, []string{"192.0.2.6", "192.0.2.10"})
	// Validate traffic over primary path port2, port3
	validateTrafficFlows(t, args.ate, args.top, BaseFlow, false)

	//shutdown port2
	flapinterface(t, args.dut, "port2", false)
	defer flapinterface(t, args.dut, "port2", true)
	// validate programming using AFT
	aftCheck(t, args.dut, dstPfx, []string{"192.0.2.10"})
	// Validate traffic over primary path port3
	validateTrafficFlows(t, args.ate, args.top, BaseFlow, false)

	//shutdown port3
	flapinterface(t, args.dut, "port3", false)
	defer flapinterface(t, args.dut, "port3", true)
	// validate programming using AFT
	aftCheck(t, args.dut, dstPfx, []string{"192.0.2.14"})
	// validate traffic over backup
	validateTrafficFlows(t, args.ate, args.top, BaseFlow, false)
}

// createFlow returns a flow from atePort1 to the dstPfx
func createFlow(t *testing.T, ate *ondatra.ATEDevice, top gosnappi.Config, name, dstMac string) string {

	flow := top.Flows().Add().SetName(name)
	flow.Metrics().SetEnable(true)
	e1 := flow.Packet().Add().Ethernet()
	e1.Src().SetValue(atePort1.MAC)
	flow.TxRx().Port().SetTxName(atePort1.Name)
	e1.Dst().SetChoice("value").SetValue(dstMac)
	v4 := flow.Packet().Add().Ipv4()
	v4.Src().SetValue(atePort1.IPv4)
	v4.Dst().SetValue(dstPfxMin)
	ate.OTG().PushConfig(t, top)
	// StartProtocols required for running on hardware
	ate.OTG().StartProtocols(t)
	return name

}

// validateTrafficFlows verifies that the flow on ATE and check interface counters on DUT
func validateTrafficFlows(t *testing.T, ate *ondatra.ATEDevice, ateTop gosnappi.Config, flow string, drop bool) {
	ate.OTG().StartTraffic(t)
	time.Sleep(60 * time.Second)
	ate.OTG().StopTraffic(t)
	otgutils.LogFlowMetrics(t, ate.OTG(), ateTop)
	otgutils.LogPortMetrics(t, ate.OTG(), ateTop)
	flowPath := ate.OTG().Telemetry().Flow(flow)
	got := flowPath.LossPct().Get(t)
	if drop {
		if got != 100 {
			t.Fatalf("Traffic passing for flow %s got %f, want 100 percent loss", flow, got)
		}
	} else {
		if got > 0 {
			t.Fatalf("LossPct for flow %s got %f, want 0", flow, got)
		}
	}
}

// flapinterface shut/unshut interface, action true bringsup the interface and false brings it down
func flapinterface(t *testing.T, dut *ondatra.DUTDevice, port string, action bool) {
	// Currently, setting the OTG port down has no effect on kne and thus the corresponding dut port will be used
	dutP := dut.Port(t, port)
	dc := dut.Config()
	i := &telemetry.Interface{}
	i.Enabled = ygot.Bool(action)
	dc.Interface(dutP.Name()).Update(t, i)
}

// aftCheck does ipv4, NHG and NH aft check
func aftCheck(t testing.TB, dut *ondatra.DUTDevice, prefix string, expectedNH []string) {
	// check prefix and get NHG ID
	aftPfxNHG := gnmi.OC().NetworkInstance(*deviations.DefaultNetworkInstance).Afts().Ipv4Entry(prefix).NextHopGroup()
	aftPfxNHGVal, found := gnmi.Watch(t, dut, aftPfxNHG.State(), 10*time.Second, func(val *ygnmi.Value[uint64]) bool {
		return val.IsPresent()
	}).Await(t)
	if !found {
		t.Fatalf("Could not find prefix %s in telemetry AFT", dstPfx)
	}
	nhg, _ := aftPfxNHGVal.Val()

	// using NHG ID validate NH
	aftNHG := gnmi.Get(t, dut, gnmi.OC().NetworkInstance(*deviations.DefaultNetworkInstance).Afts().NextHopGroup(nhg).State())
	if got := len(aftNHG.NextHop); got < 1 && aftNHG.BackupNextHopGroup == nil {
		t.Fatalf("Prefix %s reachability didn't switch to backup path", prefix)
	}
	if len(aftNHG.NextHop) != 0 {
		for k := range aftNHG.NextHop {
			aftnh := gnmi.Get(t, dut, gnmi.OC().NetworkInstance(*deviations.DefaultNetworkInstance).Afts().NextHop(k).State())
			totalIPs := len(expectedNH)
			for _, ip := range expectedNH {
				if ip == aftnh.GetIpAddress() {
					break
				}
				totalIPs--
			}
			if totalIPs == 0 {
				t.Fatalf("No matching NH found")
			}
		}
	}
}

// Waits for at least one ARP entry on the tx OTG interface
func waitOTGARPEntry(t *testing.T) {
	t.Helper()
	ate := ondatra.ATE(t, "ate")
	ate.OTG().Telemetry().Interface(atePort1.Name+".Eth").Ipv4NeighborAny().LinkLayerAddress().Watch(
		t, time.Minute, func(val *otgtelemetry.QualifiedString) bool {
			return val.IsPresent()
		}).Await(t)
}