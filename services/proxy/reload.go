package proxy

import (
	"context"

	api "github.com/ehazlett/stellar/api/services/proxy/v1"
	ptypes "github.com/gogo/protobuf/types"
)

func (s *service) Reload(ctx context.Context, req *api.ReloadRequest) (*ptypes.Empty, error) {
	return empty, nil
}
