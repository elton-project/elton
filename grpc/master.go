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
	return nil
}

func (e *EltonMaster) GenerateObjectsInfo(objects *pb.ObjectsInfo, stream pb.EltonService_GenerateObjectsInfoServer) error {
	log.Printf("GenerateObjectsInfo: %v", objects)
	for _, obj := range objects.GetObjects() {
		o, err := e.generateObjectInfo(obj)
		if err != nil {
			log.Println(err)
			return err
		}

		log.Printf("Return GenerateObjectsInfo: %v", o)
		if err = stream.Send(
			&pb.ObjectInfo{
				ObjectId: o.ObjectID,
				Version:  o.Version,
				Delegate: o.Delegate,
			},
		); err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func (e *EltonMaster) CommitObjectsInfo(c context.Context, o *pb.ObjectsInfo) (r *pb.EmptyMessage, err error) {
	log.Printf("CommitObjectsInfo: %v", o)
	for _, obj := range o.GetObjects() {
		if err = e.Registry.SetObjectInfo(
			elton.ObjectInfo{
				ObjectID: obj.ObjectId,
				Version:  obj.Version,
				Delegate: obj.Delegate,
			},
			obj.RequestHostname,
		); err != nil {
			log.Println(err)
			return
		}
	}

	return
}

func (e *EltonMaster) generateObjectInfo(o *pb.ObjectInfo) (elton.ObjectInfo, error) {
	if o.Version == 0 {
		return e.Registry.GenerateObjectInfo(o.ObjectId)
	}

	if o.Delegate == e.Conf.Master.Name {
		return e.Registry.GetNewVersion(
			elton.ObjectInfo{
				ObjectID: o.ObjectId,
				Version:  o.Version,
				Delegate: o.Delegate,
			},
		)
	}

	return e.generateObjectInfoByOtherMaster(o)
}

func (e *EltonMaster) generateObjectInfoByOtherMaster(o *pb.ObjectInfo) (object elton.ObjectInfo, err error) {
	conn, err := e.getConnection(e.Masters[o.Delegate])
	if err != nil {
		return
	}

	client := pb.NewEltonServiceClient(conn)
	stream, err := client.GenerateObjectsInfo(context.Background(), &pb.ObjectsInfo{[]*pb.ObjectInfo{o}})
	if err != nil {
		return
	}

	obj, err := stream.Recv()
	if err != nil {
		return
	}

	return elton.ObjectInfo{
		ObjectID: obj.ObjectId,
		Version:  obj.Version,
		Delegate: obj.Delegate,
	}, nil
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

func (e *EltonMaster) PutObject(c context.Context, o *pb.Object) (r *pb.EmptyMessage, err error) {
	return
}

func (e *EltonMaster) DeleteObject(c context.Context, o *pb.ObjectInfo) (r *pb.EmptyMessage, err error) {
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
