package client

import (
	"context"

	appapi "github.com/ehazlett/stellar/api/services/application/v1"
	clusterapi "github.com/ehazlett/stellar/api/services/cluster/v1"
	schedulerapi "github.com/ehazlett/stellar/api/services/scheduler/v1"
)

type scheduler struct {
	client schedulerapi.SchedulerClient
}

func (s *scheduler) Schedule(service *appapi.Service, nodes []*clusterapi.Node) ([]*clusterapi.Node, error) {
	ctx := context.Background()
	resp, err := s.client.Schedule(ctx, &schedulerapi.ScheduleRequest{
		Service:        service,
		AvailableNodes: nodes,
	})
	if err != nil {
		return nil, err
	}
	return resp.Nodes, nil
}
