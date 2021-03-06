package network

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/containerd/containerd/errdefs"
	api "github.com/ehazlett/stellar/api/services/network/v1"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
)

var (
	// ErrNoAvailableSubnets returns an error if there are no available subnets to allocate
	ErrNoAvailableSubnets = errors.New("no available subnets in network configuration")
	// ErrSubnetNotFound returns an error if no subnet is found for the node
	ErrSubnetNotFound = errors.New("subnet not found")
	// format: subnets.<node>
	dsSubnetsKey = "subnets.%s"
)

func (s *service) AllocateSubnet(ctx context.Context, req *api.AllocateSubnetRequest) (*api.AllocateSubnetResponse, error) {
	c, err := s.client(s.agent.Self().Address)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	logrus.Debug("service.network allocating subnet")
	// check for existing assigned subnet; if not, allocate
	localSubnetKey := fmt.Sprintf(dsSubnetsKey, req.Node)

	resp, err := s.Subnets(ctx, nil)
	if err != nil {
		return nil, err
	}

	subnets := resp.Subnets

	if len(subnets) == 0 {
		return nil, ErrNoAvailableSubnets
	}

	localSubnetBytes, err := c.Datastore().Get(dsNetworkBucketName, localSubnetKey)
	if err != nil {
		err = errdefs.FromGRPC(err)
		if !errdefs.IsNotFound(err) {
			return nil, err
		}
	}
	localSubnet := string(localSubnetBytes)

	logrus.WithFields(logrus.Fields{
		"subnet": string(localSubnet),
	}).Debug("local subnet from datastore")

	// check for subnet change in config
	// we convert it back to the original cluster subnet
	if localSubnet != "" {
		_, localSub, err := net.ParseCIDR(localSubnet)
		if err != nil {
			return nil, err
		}
		localSize, localBits := localSub.Mask.Size()
		clusterMask := net.CIDRMask(localSize-subnetMaskBits, localBits)
		m, _ := clusterMask.Size()
		clusterCIDR := fmt.Sprintf("%s/%d", localSub.IP.String(), m)
		_, clusterSubnet, err := net.ParseCIDR(clusterCIDR)
		if err != nil {
			return nil, err
		}
		configSubnetCIDR := s.config.Subnet.String()
		// local subnet from ds does not match config
		if clusterSubnet.String() != configSubnetCIDR {
			logrus.WithFields(logrus.Fields{
				"cluster_subnet": clusterSubnet.String(),
				"config_subnet":  configSubnetCIDR,
			}).Warn("config subnet updated")
			localSubnet = ""
		}
	}

	if localSubnet == "" {
		logrus.Debug("local subnet key not found; assigning new subnet")

		searchKey := fmt.Sprintf(dsSubnetsKey, "")
		existingSubnets, err := c.Datastore().Search(dsNetworkBucketName, searchKey)
		if err != nil {
			err = errdefs.FromGRPC(err)
			if !errdefs.IsNotFound(err) {
				return nil, err
			}
		}

		assigned := len(existingSubnets)
		if len(subnets) < assigned {
			return nil, fmt.Errorf("no available subnet for current node; need %d subnets", assigned)
		}

		cidr := subnets[assigned].CIDR
		logrus.Debugf("subnet for node: %s", cidr)

		localSubnet = cidr
		if err := c.Datastore().Set(dsNetworkBucketName, localSubnetKey, []byte(localSubnet), true); err != nil {
			return nil, err
		}
	}

	return &api.AllocateSubnetResponse{
		SubnetCIDR: string(localSubnet),
		Node:       req.Node,
	}, nil
}

func (s *service) GetSubnet(ctx context.Context, req *api.GetSubnetRequest) (*api.GetSubnetResponse, error) {
	c, err := s.client(s.agent.Self().Address)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	localSubnetKey := fmt.Sprintf(dsSubnetsKey, req.Node)

	localSubnet, err := c.Datastore().Get(dsNetworkBucketName, localSubnetKey)
	if err != nil {
		err = errdefs.FromGRPC(err)
		if !errdefs.IsNotFound(err) {
			return nil, ErrSubnetNotFound
		}
	}
	return &api.GetSubnetResponse{
		SubnetCIDR: string(localSubnet),
	}, nil
}

func (s *service) DeallocateSubnet(ctx context.Context, req *api.DeallocateSubnetRequest) (*ptypes.Empty, error) {
	// TODO: remove alias gateway address from local nic

	// TODO: notify peers of removed route

	return &ptypes.Empty{}, nil
}

func (s *service) Subnets(ctx context.Context, _ *ptypes.Empty) (*api.SubnetsResponse, error) {
	subs, err := divideSubnet(s.network, maxSubnets)
	if err != nil {
		return nil, err
	}

	var subnets []*api.Subnet
	for _, subnet := range subs {
		ip := subnet.IP
		gw := net.IPv4(ip[0], ip[1], ip[2], 1)
		subnets = append(subnets, &api.Subnet{
			CIDR:    subnet.String(),
			Gateway: gw.String(),
		})
	}

	return &api.SubnetsResponse{
		Subnets: subnets,
	}, nil
}
