package proxy

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/ehazlett/stellar"
	applicationapi "github.com/ehazlett/stellar/api/services/application/v1"
	eventsapi "github.com/ehazlett/stellar/api/services/events/v1"
	api "github.com/ehazlett/stellar/api/services/proxy/v1"
	"github.com/ehazlett/stellar/client"
	"github.com/ehazlett/stellar/events"
	"github.com/ehazlett/stellar/services"
	appsvc "github.com/ehazlett/stellar/services/application"
	nssvc "github.com/ehazlett/stellar/services/nameserver"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
	"github.com/stellarproject/element"
	"google.golang.org/grpc"
)

const (
	serviceID         = "stellar.services.proxy.v1"
	dsProxyBucketName = "stellar." + stellar.APIVersion + ".services.proxy"
)

var (
	empty = &ptypes.Empty{}
)

type service struct {
	containerdAddr string
	namespace      string
	agent          *element.Agent
	config         *stellar.Config
	errCh          chan error
}

func New(cfg *stellar.Config, agent *element.Agent) (services.Service, error) {
	errCh := make(chan error)
	go func() {
		for {
			err := <-errCh
			logrus.Errorf("proxy: %s", err)
		}
	}()

	return &service{
		containerdAddr: cfg.ContainerdAddr,
		namespace:      cfg.Namespace,
		agent:          agent,
		config:         cfg,
		errCh:          errCh,
	}, nil
}

func (s *service) Register(server *grpc.Server) error {
	api.RegisterProxyServer(server, s)
	return nil
}

func (s *service) ID() string {
	return serviceID
}

func (s *service) Type() services.Type {
	return services.ProxyService
}

func (s *service) Requires() []services.Type {
	return []services.Type{
		services.DatastoreService,
		services.ApplicationService,
		services.NameserverService,
	}
}

func (s *service) Info(ctx context.Context, req *api.InfoRequest) (*api.InfoResponse, error) {
	return &api.InfoResponse{
		ID: serviceID,
	}, nil
}

func (s *service) Start() error {
	c, err := s.client(s.agent.Self().Address)
	if err != nil {
		return err
	}

	// start listener for application events
	go func() {
		defer c.Close()

		stream, err := c.EventsService().Subscribe(context.Background(), &eventsapi.SubscribeRequest{
			Subject: "stellar.services.>",
		})
		if err != nil {
			logrus.WithError(err).Error("error subscribing to application events")
			return
		}

		for {
			evt, err := stream.Recv()
			if err != nil {
				logrus.WithError(err).Error("error subscribing to application events")
				return
			}

			msg, err := events.UnmarshalEvent(evt)
			if err != nil {
				logrus.WithError(err).Error("error unmarshalling event")
				return
			}

			switch e := msg.(type) {
			case *appsvc.UpdateEvent, *nssvc.CreateEvent, *nssvc.DeleteEvent:
				logrus.WithFields(logrus.Fields{
					"event": fmt.Sprintf("%T", e),
				}).Debug("reloading proxy")
			default:
				logrus.Errorf("unknown event type: %+v", e)
			}
		}
	}()

	return nil
}

func (s *service) Stop() error {
	return nil
}

func (s *service) containerd() (*containerd.Client, error) {
	return stellar.DefaultContainerd(s.containerdAddr, s.namespace)
}

func (s *service) nodeName() string {
	return s.agent.Self().ID
}

func (s *service) client(address string) (*client.Client, error) {
	opts, err := client.DialOptionsFromConfig(s.config)
	if err != nil {
		return nil, err
	}
	return client.NewClient(address, opts...)
}

func (s *service) getApplications() ([]*applicationapi.App, error) {
	c, err := s.client(s.agent.Self().Address)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	return c.Application().List()
}
