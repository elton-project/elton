package grpc

import (
	"fmt"
	"log"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	elton "../api"
	pb "./proto"
)

type EltonMaster struct {
	Conf        elton.Config
	Registry    *elton.Registry
	Masters     map[string]string
	Connections map[string]*grpc.ClientConn
}

func NewEltonMaster(conf elton.Config) (*EltonMaster, error) {
	registry, err := elton.NewRegistry(conf)
	if err != nil {
		return nil, err
	}

	masters := make(map[string]string)
	for _, master := range conf.Masters {
		masters[master.Name] = master.HostName
	}

	return &EltonMaster{Conf: conf, Registry: registry, Masters: masters, Connections: make(map[string]*grpc.ClientConn)}, nil
}

func (e *EltonMaster) Serve() error {
	defer e.Registry.Close()
	defer func() {
		for _, c := range e.Connections {
			c.Close()
		}
	}()

	lis, err := net.Listen("tcp", e.Conf.Master.HostName)
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	pb.RegisterEltonServiceServer(server, e)

	log.Fatal(server.Serve(lis))
}

func (e *EltonMaster) GenerateObjectID(o *pb.ObjectName, stream pb.EltonService_GenerateObjectIDServer) error {
	log.Printf("GenerateObjectID: %v", o)
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

	log.Printf("Return GenerateObjectID: %v", list)
	return nil
}

func (e *EltonMaster) CreateObjectInfo(o *pb.ObjectInfo, stream pb.EltonService_CreateObjectInfoServer) error {
	log.Printf("CreateObjectInfo: %v", o)
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

	log.Printf("Return CreateObjectInfo: %v", obj)
	return nil
}

func (e *EltonMaster) createObjectInfo(o *pb.ObjectInfo) (obj elton.ObjectInfo, err error) {
	if o.Delegate == e.Conf.Master.Name {
		obj, err = e.Registry.GetNewVersion(
			elton.ObjectInfo{
				ObjectID: o.ObjectId,
				Version:  o.Version,
				Delegate: o.Delegate,
			},
		)

		if err != nil {
			return
		}

		return
	}

	obj, err = e.createObjectInfoByOtherMaster(o)
	if err != nil {
		return

	}

	return
}

func (e *EltonMaster) createObjectInfoByOtherMaster(o *pb.ObjectInfo) (object elton.ObjectInfo, err error) {
	conn, err := e.getConnection(e.Masters[o.Delegate])
	if err != nil {
		return
	}

	client := pb.NewEltonServiceClient(conn)
	stream, err := client.CreateObjectInfo(context.Background(), o)
	if err != nil {
		return
	}

	obj, err := stream.Recv()
	if err != nil {
		return
	}

	object = elton.ObjectInfo{
		ObjectID: obj.ObjectId,
		Version:  obj.Version,
		Delegate: obj.Delegate,
	}
	return
}

func (e *EltonMaster) getConnection(host string) (conn *grpc.ClientConn, err error) {
	conn = e.Connections[host]
	if conn == nil {
		conn, err = grpc.Dial(host)
		if err != nil {
			return conn, fmt.Errorf("Unknown host: %s", host)
		}

		e.Connections[host] = conn
	}

	return
}

func (e *EltonMaster) GetObject(o *pb.ObjectInfo, stream pb.EltonService_GetObjectServer) error {
	log.Printf("GetObject: %v", o)
	host, err := e.Registry.GetObjectHost(o.ObjectId, o.Version)
	if err != nil {
		if err = e.getObjectFromOtherMaster(o, stream); err != nil {
			log.Println(err)
			return err
		}

		return nil
	}

	conn, err := e.getConnection(host)
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

func (e *EltonMaster) getObjectFromOtherMaster(o *pb.ObjectInfo, stream pb.EltonService_GetObjectServer) error {
	conn, err := e.getConnection(e.Masters[o.Delegate])
	if err != nil {
		return err
	}

	client := pb.NewEltonServiceClient(conn)
	return e.getObject(o, stream, client)
}

func (e *EltonMaster) getObject(o *pb.ObjectInfo, stream pb.EltonService_GetObjectServer, client pb.EltonServiceClient) error {
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

func (e *EltonMaster) DeleteObject(c context.Context, o *pb.ObjectInfo) (r *pb.ResponseType, err error) {
	log.Printf("DeleteObject: %v", o)
	if err = e.Registry.DeleteObjectVersions(o.ObjectId); err != nil {
		log.Println(err)
		return
	}

	if o.Delegate == e.Conf.Master.Name {
		if err = e.Registry.DeleteObjectInfo(o.ObjectId); err != nil {
			log.Println(err)
			return
		}
	} else {
		conn, err := e.getConnection(e.Masters[o.Delegate])
		if err != nil {
			return r, err
		}

		client := pb.NewEltonServiceClient(conn)
		_, err = client.DeleteObject(context.Background(), o)
	}

	return
}
