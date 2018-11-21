package p2p

import (
	pb "gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc/proto2"
)

type unsafeListenerStore struct {
	m map[pb.EventType]map[NodeID]*pb.EventListenerInfo
}

func (s *unsafeListenerStore) init() {
	s.m = map[pb.EventType]map[NodeID]*pb.EventListenerInfo{}
}
func (s *unsafeListenerStore) Clear() {
	s.m = nil
}
func (s *unsafeListenerStore) Add(info *pb.EventListenerInfo) {
	s.init()
	if s.m[info.Type] == nil {
		s.m[info.Type] = map[NodeID]*pb.EventListenerInfo{}
	}
	s.m[info.Type][NodeID(info.Node.Id)] = info
}
func (s *unsafeListenerStore) Remove(info *pb.EventListenerInfo) {
	s.init()
	if s.m[info.Type] == nil {
		return
	}
	delete(s.m[info.Type], NodeID(info.Node.Id))
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
	m map[NodeID]*pb.EventDelivererInfo
}

func (s *unsafeDelivererStore) init() {
	s.m = map[NodeID]*pb.EventDelivererInfo{}
}
func (s *unsafeDelivererStore) Clear() {
	s.m = nil
}
func (s *unsafeDelivererStore) Add(info *pb.EventDelivererInfo) {
	s.init()
	s.m[NodeID(info.Node.Id)] = info
}
func (s *unsafeDelivererStore) Remove(info *pb.EventDelivererInfo) {
	s.init()
	delete(s.m, NodeID(info.Node.Id))
}
func (s *unsafeDelivererStore) Foreach(fn func(info *pb.EventDelivererInfo) error) error {
	for _, info := range s.m {
		if err := fn(info); err != nil {
			return err
		}
	}
	return nil
}
