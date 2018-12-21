package p2p

import (
	"fmt"
	pb "gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc/proto2"
	"strings"
)

type ServerID string

func toServerID(info *pb.ServerInfo) ServerID {
	return ServerID(info.Guid)
}

type serverTypeKey struct {
	subsystem pb.SubsystemType
	service   pb.ServiceType
}

func toServerTypeKey(info *pb.ServerInfo) serverTypeKey {
	return serverTypeKey{
		subsystem: info.SubsystemType,
		service:   info.ServiceType,
	}
}

type unsafeListenerStore struct {
	m map[pb.EventType]map[ServerID]*pb.EventListenerInfo
}

func (s *unsafeListenerStore) init() {
	if s.m == nil {
		s.m = map[pb.EventType]map[ServerID]*pb.EventListenerInfo{}
	}
}
func (s *unsafeListenerStore) Clear() {
	s.m = nil
}
func (s *unsafeListenerStore) Add(info *pb.EventListenerInfo) {
	s.init()
	if s.m[info.Type] == nil {
		s.m[info.Type] = map[ServerID]*pb.EventListenerInfo{}
	}
	s.m[info.Type][toServerID(info.ServerInfo)] = info
}
func (s *unsafeListenerStore) Remove(info *pb.EventListenerInfo) {
	s.init()
	if s.m[info.Type] == nil {
		return
	}
	delete(s.m[info.Type], toServerID(info.ServerInfo))
}
func (s *unsafeListenerStore) Foreach(eventType pb.EventType, fn func(info *pb.EventListenerInfo) error) error {
	for _, info := range s.m[eventType] {
		if err := fn(info); err != nil {
			return err
		}
	}
	return nil
}
func (s *unsafeListenerStore) ListListeners(eventType pb.EventType) []*pb.EventListenerInfo {
	var list []*pb.EventListenerInfo
	for _, info := range s.m[eventType] {
		list = append(list, info)
	}
	return list
}

type unsafeDelivererStore struct {
	m map[ServerID]*pb.EventDelivererInfo
}

func (s *unsafeDelivererStore) init() {
	if s.m == nil {
		s.m = map[ServerID]*pb.EventDelivererInfo{}
	}
}
func (s *unsafeDelivererStore) Clear() {
	s.m = nil
}
func (s *unsafeDelivererStore) Add(info *pb.EventDelivererInfo) {
	s.init()
	s.m[toServerID(info.ServerInfo)] = info
}
func (s *unsafeDelivererStore) Remove(info *pb.EventDelivererInfo) {
	s.init()
	delete(s.m, toServerID(info.ServerInfo))
}
func (s *unsafeDelivererStore) Foreach(fn func(info *pb.EventDelivererInfo) error) error {
	for _, info := range s.m {
		if err := fn(info); err != nil {
			return err
		}
	}
	return nil
}

type unsafeServerStore struct {
	m map[serverTypeKey]*pb.ServerInfo
}

func (s *unsafeServerStore) init() {
	if s.m == nil {
		s.m = map[serverTypeKey]*pb.ServerInfo{}
	}
}
func (s *unsafeServerStore) Clear() {
	s.m = nil
}
func (s *unsafeServerStore) Add(info *pb.ServerInfo) {
	s.init()
	s.m[toServerTypeKey(info)] = info
}
func (s *unsafeServerStore) Remove(info *pb.ServerInfo) {
	s.init()
	delete(s.m, toServerTypeKey(info))
}
func (s *unsafeServerStore) Search(query *pb.ServerQuery, fn func(info *pb.ServerInfo) error) error {
	for _, info := range s.m {
		if err := fn(info); err != nil {
			return err
		}
	}
	return nil
}
func (s *unsafeServerStore) Dump() string {
	b := strings.Builder{}
	fmt.Fprintln(&b, "<unsafeServerStore>")
	for key, server := range s.m {
		fmt.Fprintf(&b, "  %s/%s: id=%s, address=%s\n",
			key.subsystem,
			key.service,
			server.Guid,
			server.Address,
		)
	}
	return b.String()
}
