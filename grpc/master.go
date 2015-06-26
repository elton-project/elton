package grpc

import (
	"fmt"
	"log"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	elton "../elton"
	pb "./proto"
)

type EltonService struct {
	Conf     elton.Config
	Registry *elton.Registry
	Masters  map[string]string
}

func NewEltonServer(conf elton.Config) (*EltonService, error) {
	registry, err := elton.NewRegistry(conf)
	if err != nil {
		return nil, err
	}

	masters := make(map[string]string, len(conf.Masters))
	for _, master := range conf.Masters {
		masters[master.Name] = master.HostName
	}

	return &EltonService{Conf: conf, Registry: registry, Masters: masters}, nil
}

func (e *EltonService) Serve() {
	defer e.Registry.Close()

	lis, err := net.Listen("tcp", e.Conf.Elton.HostName)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	server := grpc.NewServer()
	pb.RegisterEltonServiceServer(server, e)
	log.Fatal(server.Serve(lis))
}

func (e *EltonService) GenerateObjectID(o *pb.ObjectName, stream pb.EltonService_GenerateObjectIDServer) error {
	list, err := e.Registry.GenerateObjectsInfo(o.Names)
	if err != nil {
		log.Println(err)
		return err
	}

	for _, info := range list {
		if err = stream.Send(
			&pb.ObjectInfo{
				ObjectId: info.ObjectID,
				Delegate: info.Delegate},
		); err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func (e *EltonService) CreateObjectInfo(o *pb.ObjectInfo, stream pb.EltonService_CreateObjectInfoServer) error {
	obj, err := e.createObjectInfo(o)
	if err != nil {
		log.Println(err)
		return err
	}

	if err = e.Registry.SetObjectInfo(obj, o.RequestHostname); err != nil {
		log.Println(err)
		return err
	}

	if err = stream.Send(
		&pb.ObjectInfo{
			ObjectId:        obj.ObjectID,
			Version:         obj.Version,
			Delegate:        obj.Delegate,
			RequestHostname: o.RequestHostname,
		},
	); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (e *EltonService) createObjectInfo(o *pb.ObjectInfo) (elton.ObjectInfo, error) {
	if o.Delegate == e.Conf.Elton.Name {
		obj, err := e.Registry.GetNewVersion(
			elton.ObjectInfo{
				ObjectID: o.ObjectId,
				Version:  o.Version,
				Delegate: o.Delegate,
			},
		)
		if err != nil {
			return elton.ObjectInfo{}, err
		}

		return obj, nil
	}

	obj, err := e.createObjectInfoByOtherMaster(o)
	if err != nil {
		return elton.ObjectInfo{}, err

	}
	return obj, nil
}

func (e *EltonService) createObjectInfoByOtherMaster(o *pb.ObjectInfo) (elton.ObjectInfo, error) {
	host := e.Masters[o.Delegate]
	if host == "" {
		return elton.ObjectInfo{}, fmt.Errorf("Unknown master %s: ", o.Delegate)
	}

	conn, err := grpc.Dial(host)
	if err != nil {
		return elton.ObjectInfo{}, err
	}
	defer conn.Close()

	client := pb.NewEltonServiceClient(conn)
	stream, err := client.CreateObjectInfo(context.Background(), o)
	if err != nil {
		return elton.ObjectInfo{}, err
	}

	obj, err := stream.Recv()
	if err != nil {
		return elton.ObjectInfo{}, err
	}

	return elton.ObjectInfo{ObjectID: obj.ObjectId, Version: obj.Version, Delegate: obj.Delegate}, nil
}

func (e *EltonService) GetObject(o *pb.ObjectInfo, stream pb.EltonService_GetObjectServer) error {
	host, err := e.Registry.GetObjectHost(o.ObjectId, o.Version)
	if err != nil {
		if err = e.getObjectFromOtherMaster(o, stream); err != nil {
			log.Println(err)
			return err
		}

		return nil
	}

	conn, err := grpc.Dial(host)
	if err != nil {
		log.Println(err)
		return err
	}

	client := pb.NewEltonServiceClient(conn)
	if err = e.getObject(o, stream, client); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (e *EltonService) getObjectFromOtherMaster(o *pb.ObjectInfo, stream pb.EltonService_GetObjectServer) error {
	host := e.Masters[o.Delegate]
	if host == "" {
		return fmt.Errorf("Unknown master %s: ", o.Delegate)
	}

	conn, err := grpc.Dial(host)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pb.NewEltonServiceClient(conn)

	return e.getObject(o, stream, client)
}

func (e *EltonService) getObject(o *pb.ObjectInfo, stream pb.EltonService_GetObjectServer, client pb.EltonServiceClient) error {
	s, err := client.GetObject(context.Background(), o)
	if err != nil {
		return err
	}

	obj, err := s.Recv()
	if err != nil {
		return err
	}

	if err = e.Registry.SetObjectInfo(
		elton.ObjectInfo{
			ObjectID: o.ObjectId,
			Version:  o.Version,
			Delegate: o.Delegate,
		},
		o.RequestHostname,
	); err != nil {
		log.Println(err)
		return err
	}

	if err = stream.Send(obj); err != nil {
		return err
	}

	return nil
}

func (e *EltonService) DeleteObject(c context.Context, o *pb.ObjectInfo) (*pb.ResponseType, error) {
	if err := e.Registry.DeleteObjectVersions(o.ObjectId); err != nil {
		log.Println(err)
		return new(pb.ResponseType), err
	}

	if o.Delegate == e.Conf.Elton.Name {
		if err := e.Registry.DeleteObjectInfo(o.ObjectId); err != nil {
			log.Println(err)
			return new(pb.ResponseType), err
		}
	} else {
		host := e.Masters[o.Delegate]
		if host == "" {
			err := fmt.Errorf("Unknown master %s: ", o.Delegate)
			log.Println(err)
			return new(pb.ResponseType), err
		}

		conn, err := grpc.Dial(host)
		if err != nil {
			return new(pb.ResponseType), err
		}
		defer conn.Close()

		client := pb.NewEltonServiceClient(conn)

		_, err = client.DeleteObject(context.Background(), o)
		return new(pb.ResponseType), err
	}

	return new(pb.ResponseType), nil
}
