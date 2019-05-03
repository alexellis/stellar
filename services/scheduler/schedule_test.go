package scheduler

import (
	"testing"

	appapi "github.com/ehazlett/stellar/api/services/application/v1"
	clusterapi "github.com/ehazlett/stellar/api/services/cluster/v1"
)

func TestScheduleNoPreference(t *testing.T) {
	expected := map[string]struct{}{
		"node-00": struct{}{},
	}
	availableNodes := []*clusterapi.Node{
		{
			ID:      "node-00",
			Address: "127.0.0.1:9000",
		},
		{
			ID:      "node-01",
			Address: "127.0.0.1:9001",
		},
		{
			ID:      "node-02",
			Address: "127.0.0.1:9002",
		},
	}

	appService := &appapi.Service{
		Name:                "test-service",
		PlacementPreference: nil,
	}

	svc := &service{}
	nodes, err := svc.schedule(appService, availableNodes)
	if err != nil {
		t.Fatal(err)
	}

	if len(nodes) != len(expected) {
		t.Fatalf("expected %d nodes; received %d", len(availableNodes), len(nodes))
	}

	for _, node := range nodes {
		if _, ok := expected[node.ID]; !ok {
			t.Fatalf("unexpected node %s", node.ID)
		}
	}
}

func TestScheduleNodesByIDEmptyReplica(t *testing.T) {
	expected := map[string]struct{}{
		"node-00": struct{}{},
	}
	availableNodes := []*clusterapi.Node{
		{
			ID:      "node-00",
			Address: "127.0.0.1:9000",
		},
		{
			ID:      "node-01",
			Address: "127.0.0.1:9001",
		},
		{
			ID:      "node-02",
			Address: "127.0.0.1:9002",
		},
	}

	nodeIDs := []string{"node-00", "node-01"}

	appService := &appapi.Service{
		Name: "test-service",
		PlacementPreference: &appapi.PlacementPreference{
			NodeIDs: nodeIDs,
		},
	}

	svc := &service{}
	nodes, err := svc.schedule(appService, availableNodes)
	if err != nil {
		t.Fatal(err)
	}

	if len(nodes) != len(expected) {
		t.Fatalf("expected %d nodes; received %d", len(expected), len(nodes))
	}

	for _, node := range nodes {
		if _, ok := expected[node.ID]; !ok {
			t.Fatalf("unexpected node %s", node.ID)
		}
	}
}

func TestScheduleNodesByIDSingleReplica(t *testing.T) {
	expected := map[string]struct{}{
		"node-00": struct{}{},
	}
	availableNodes := []*clusterapi.Node{
		{
			ID:      "node-00",
			Address: "127.0.0.1:9000",
		},
		{
			ID:      "node-01",
			Address: "127.0.0.1:9001",
		},
		{
			ID:      "node-02",
			Address: "127.0.0.1:9002",
		},
	}

	nodeIDs := []string{"node-00", "node-01"}

	appService := &appapi.Service{
		Name: "test-service",
		PlacementPreference: &appapi.PlacementPreference{
			NodeIDs: nodeIDs,
		},
		Replicas: uint64(1),
	}

	svc := &service{}
	nodes, err := svc.schedule(appService, availableNodes)
	if err != nil {
		t.Fatal(err)
	}

	if len(nodes) != len(expected) {
		t.Fatalf("expected %d nodes; received %d", len(expected), len(nodes))
	}

	for _, node := range nodes {
		if _, ok := expected[node.ID]; !ok {
			t.Fatalf("unexpected node %s", node.ID)
		}
	}
}
func TestScheduleNodesByIDWithReplicas(t *testing.T) {
	expected := map[string]struct{}{
		"node-00": struct{}{},
		"node-01": struct{}{},
	}
	availableNodes := []*clusterapi.Node{
		{
			ID:      "node-00",
			Address: "127.0.0.1:9000",
		},
		{
			ID:      "node-01",
			Address: "127.0.0.1:9001",
		},
		{
			ID:      "node-02",
			Address: "127.0.0.1:9002",
		},
	}

	nodeIDs := []string{"node-00", "node-01"}

	appService := &appapi.Service{
		Name: "test-service",
		PlacementPreference: &appapi.PlacementPreference{
			NodeIDs: nodeIDs,
		},
		Replicas: uint64(2),
	}

	svc := &service{}
	nodes, err := svc.schedule(appService, availableNodes)
	if err != nil {
		t.Fatal(err)
	}

	if len(nodes) != len(expected) {
		t.Fatalf("expected %d nodes; received %d", len(expected), len(nodes))
	}

	for _, node := range nodes {
		if _, ok := expected[node.ID]; !ok {
			t.Fatalf("unexpected node %s", node.ID)
		}
	}
}

func TestScheduleNodesByLabelWithReplicas(t *testing.T) {
	expected := map[string]struct{}{
		"node-00": struct{}{},
		"node-02": struct{}{},
	}

	availableNodes := []*clusterapi.Node{
		{
			ID:      "node-00",
			Address: "127.0.0.1:9000",
			Labels: map[string]string{
				"env": "prod",
			},
		},
		{
			ID:      "node-01",
			Address: "127.0.0.1:9001",
			Labels: map[string]string{
				"env": "qa",
			},
		},
		{
			ID:      "node-02",
			Address: "127.0.0.1:9002",
			Labels: map[string]string{
				"env": "prod",
			},
		},
	}

	labels := map[string]string{
		"env": "prod",
	}

	appService := &appapi.Service{
		Name: "test-service",
		PlacementPreference: &appapi.PlacementPreference{
			Labels: labels,
		},
		Replicas: uint64(2),
	}

	svc := &service{}
	nodes, err := svc.schedule(appService, availableNodes)
	if err != nil {
		t.Fatal(err)
	}

	if len(nodes) != len(expected) {
		t.Fatalf("expected %d nodes; received %d", len(expected), len(nodes))
	}

	for _, node := range nodes {
		if _, ok := expected[node.ID]; !ok {
			t.Fatalf("unexpected node %s", node.ID)
		}
	}
}

func TestScheduleNodesByMultipleLabelsWithReplicas(t *testing.T) {
	expected := []string{"node-00", "node-00"}

	availableNodes := []*clusterapi.Node{
		{
			ID:      "node-00",
			Address: "127.0.0.1:9000",
			Labels: map[string]string{
				"env":    "prod",
				"region": "east",
			},
		},
		{
			ID:      "node-01",
			Address: "127.0.0.1:9001",
			Labels: map[string]string{
				"env":    "qa",
				"region": "west",
			},
		},
		{
			ID:      "node-02",
			Address: "127.0.0.1:9002",
			Labels: map[string]string{
				"env":    "prod",
				"region": "south",
			},
		},
	}

	labels := map[string]string{
		"env":    "prod",
		"region": "east",
	}

	appService := &appapi.Service{
		Name: "test-service",
		PlacementPreference: &appapi.PlacementPreference{
			Labels: labels,
		},
		Replicas: uint64(2),
	}

	svc := &service{}
	nodes, err := svc.schedule(appService, availableNodes)
	if err != nil {
		t.Fatal(err)
	}

	if len(nodes) != len(expected) {
		t.Fatalf("expected %d nodes; received %d", len(expected), len(nodes))
	}

	for i, node := range nodes {
		if v := expected[i]; v != node.ID {
			t.Fatalf("unexpected node %s", node.ID)
		}
	}
}

func TestScheduleNodesByLabelMultiWithEmptyWithReplicas(t *testing.T) {
	expected := map[string]struct{}{}

	availableNodes := []*clusterapi.Node{
		{
			ID:      "node-00",
			Address: "127.0.0.1:9000",
			Labels: map[string]string{
				"env":    "prod",
				"region": "east",
			},
		},
		{
			ID:      "node-01",
			Address: "127.0.0.1:9001",
			Labels: map[string]string{
				"env":    "qa",
				"region": "west",
			},
		},
		{
			ID:      "node-02",
			Address: "127.0.0.1:9002",
			Labels: map[string]string{
				"env":    "prod",
				"region": "south",
			},
		},
	}

	labels := map[string]string{
		"env":    "prod",
		"region": "",
	}

	appService := &appapi.Service{
		Name: "test-service",
		PlacementPreference: &appapi.PlacementPreference{
			Labels: labels,
		},
		Replicas: uint64(2),
	}

	svc := &service{}
	nodes, err := svc.schedule(appService, availableNodes)
	if err != nil {
		t.Fatal(err)
	}

	if len(nodes) != len(expected) {
		t.Fatalf("expected %d nodes; received %d", len(expected), len(nodes))
	}

	for _, node := range nodes {
		if _, ok := expected[node.ID]; !ok {
			t.Fatalf("unexpected node %s", node.ID)
		}
	}
}
