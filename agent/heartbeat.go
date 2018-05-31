package agent

import (
	"context"
	"strings"

	"github.com/ehazlett/element"
	"github.com/sirupsen/logrus"
)

func (a *Agent) heartbeat() {
	peers, err := a.Peers()
	if err != nil {
		logrus.Errorf("error getting peers: %s", err)
		return
	}

	for _, peer := range peers {
		ac, err := element.NewClient(peer.Addr)
		if err != nil {
			logrus.Errorf("error communicating with peer: %s", err)
			return
		}
		defer ac.Close()

		health, err := ac.HealthService.Health(context.Background(), nil)
		if err != nil {
			logrus.Errorf("error communicating with peer: %s", err)
			return
		}

		logrus.WithFields(logrus.Fields{
			"peer":         peer.Name,
			"os_name":      health.OSName,
			"os_version":   health.OSVersion,
			"uptime":       health.Uptime,
			"cpus":         health.Cpus,
			"memory_total": health.MemoryTotal,
			"memory_free":  health.MemoryFree,
			"memory_used":  health.MemoryUsed,
		}).Debug("peer health")

		containers, err := ac.Containers()
		if err != nil {
			logrus.Errorf("error getting containers: %s", err)
			return
		}

		ids := []string{}
		for _, c := range containers {
			ids = append(ids, c.ID)
		}

		logrus.WithFields(logrus.Fields{
			"peer":       peer.Name,
			"containers": strings.Join(ids, ", "),
		}).Debug("containers")
	}
}
