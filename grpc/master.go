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

var opts []grpc.DialOption

type EltonMaster struct {
	Conf        elton.Config
	Registry    *elton.Registry
	Masters     map[string]string
	Connections map[string]*grpc.ClientConn
}

func NewEltonMaster(conf elton.Config) (*EltonMaster, error) {
	opts = append(opts, grpc.WithInsecure())
	registry, err := elton.NewRegistry(conf)
	if err != nil {
		return nil, err
	}

	masters := make(map[string]string)
	for _, master := range conf.Masters {
		masters[master.Name] = fmt.Sprintf("%s:%d", master.Name, master.Port)
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

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", e.Conf.Master.Port))
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	pb.RegisterEltonServiceServer(server, e)

	log.Fatal(server.Serve(lis))
	return nil
}

func (e *EltonMaster) GenerateObjectInfo(o *pb.ObjectInfo, stream pb.EltonService_GenerateObjectInfoServer) error {
	log.Printf("GenerateObjectInfo: %v", o)
	obj, err := e.generateObjectInfo(o)
	if err != nil {
		log.Println(err)
		return err
	}

	if err = stream.Send(
		&pb.ObjectInfo{
			ObjectId: obj.ObjectID,
			Version:  obj.Version,
			Delegate: obj.Delegate,
		},
	); err != nil {
		log.Println(err)
		return err
	}

	log.Printf("Return GenerateObjectInfo: %v", obj)
	return nil
}

func (e *EltonMaster) CommitObjectInfo(c context.Context, o *pb.ObjectInfo) (*pb.EmptyMessage, error) {
	log.Printf("CommitObjectInfo: %v", o)
	if err := e.Registry.SetObjectInfo(
		elton.ObjectInfo{
			ObjectID: o.ObjectId,
			Version:  o.Version,
			Delegate: o.Delegate,
		},
		o.RequestHostname,
	); err != nil {
		log.Println(err)
		return new(pb.EmptyMessage), err
	}

	// TODO: キューマネージャとかでやる方がいいと思う
	go func() {
		err := e.doBackup(o)
		log.Println(err)
		// if err != nil {
		// 	err = e.doBackup(o)
		// }
	}()

	return new(pb.EmptyMessage), nil
}

func (e *EltonMaster) doBackup(o *pb.ObjectInfo) error {
	log.Printf("doBackup(): %v", o)
	conn, err := grpc.Dial(o.RequestHostname, opts...)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pb.NewEltonServiceClient(conn)
	stream, err := client.GetObject(context.Background(), o)
	if err != nil {
		return err
	}

	obj, err := stream.Recv()

	log.Println(fmt.Sprintf("%s:%d", e.Conf.Backup.Name, e.Conf.Backup.Port))
	bconn, err := grpc.Dial(fmt.Sprintf("%s:%d", e.Conf.Backup.Name, e.Conf.Backup.Port), opts...)
	if err != nil {
		return err
	}
	defer bconn.Close()

	bclient := pb.NewEltonServiceClient(bconn)
	_, err = bclient.PutObject(context.Background(), obj)
	return err
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
	stream, err := client.GenerateObjectInfo(context.Background(), o)
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
	}, err
}

func (e *EltonMaster) getConnection(host string) (conn *grpc.ClientConn, err error) {
	conn = e.Connections[host]
	if conn == nil {
		conn, err = grpc.Dial(host, opts...)
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
		conn, err = grpc.Dial(fmt.Sprintf("%s:%d", e.Conf.Backup.Name, e.Conf.Backup.Port), opts...)
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
		conn, err = grpc.Dial(fmt.Sprintf("%s:%d", e.Conf.Backup.Name, e.Conf.Backup.Port), opts...)
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
		return err
	}

	if err = stream.Send(obj); err != nil {
		return err
	}

	return nil
}

func (e *EltonMaster) PutObject(c context.Context, o *pb.Object) (*pb.EmptyMessage, error) {
	return new(pb.EmptyMessage), nil
}

func (e *EltonMaster) DeleteObject(c context.Context, o *pb.ObjectInfo) (*pb.EmptyMessage, error) {
	log.Printf("DeleteObject: %v", o)
	if err := e.Registry.DeleteObjectVersions(o.ObjectId); err != nil {
		log.Println(err)
		return new(pb.EmptyMessage), err
	}

	if o.Delegate == e.Conf.Master.Name {
		if err := e.Registry.DeleteObjectInfo(o.ObjectId); err != nil {
			log.Println(err)
			return new(pb.EmptyMessage), err
		}
	} else {
		conn, err := e.getConnection(e.Masters[o.Delegate])
		if err != nil {
			log.Println(err)
			return new(pb.EmptyMessage), err
		}

		client := pb.NewEltonServiceClient(conn)
		_, err = client.DeleteObject(context.Background(), o)
	}

	return new(pb.EmptyMessage), nil
}
