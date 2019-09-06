package simple

import (
	"context"
	. "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
)

func newLocalMetaServer() *localMetaServer {
	return &localMetaServer{
		kvs: map[metaKey]*metaValue{},
	}
}

type localMetaServer struct {
	lock sync.RWMutex
	kvs  map[metaKey]*metaValue
}
type metaKey struct {
	Id string
}
type metaValue struct {
	Body         string
	AllowReplace bool
}

func (m *localMetaServer) GetMeta(ctx context.Context, req *GetMetaRequest) (*GetMetaResponse, error) {
	key := newMetaKey(req.GetKey())

	m.lock.RLock()
	defer m.lock.RUnlock()
	value := m.kvs[key]

	return &GetMetaResponse{
		Key:  req.GetKey(),
		Body: value.ToProperty(),
	}, nil
}
func (m *localMetaServer) SetMeta(ctx context.Context, req *SetMetaRequest) (*SetMetaResponse, error) {
	key := newMetaKey(req.GetKey())
	value := newMetaValue(req.GetBody())

	m.lock.Lock()
	defer m.lock.Unlock()

	if req.GetMustCreate() {
		if _, ok := m.kvs[key]; ok {
			return nil, status.Error(codes.AlreadyExists, "key is already exists")
		}
	}

	old := m.kvs[key]
	if old != nil && !old.AllowReplace {
		return nil, status.Errorf(codes.Unauthenticated, "replacement not allowed")
	}

	m.kvs[key] = value

	return &SetMetaResponse{
		Key:     req.GetKey(),
		OldBody: old.ToProperty(),
		Created: old != nil,
	}, nil
}

func newMetaKey(key *PropertyKey) metaKey {
	return metaKey{
		Id: key.GetId(),
	}
}
func newMetaValue(body *PropertyBody) *metaValue {
	return &metaValue{
		Body:         body.GetBody(),
		AllowReplace: body.GetAllowReplace(),
	}
}
func (v *metaValue) ToProperty() *PropertyBody {
	if v == nil {
		return nil
	}
	return &PropertyBody{
		Body:         v.Body,
		AllowReplace: v.AllowReplace,
	}
}
