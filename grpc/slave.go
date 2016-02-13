package grpc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/context"

	pb "git.t-lab.cs.teu.ac.jp/nashio/elton/grpc/proto"
	elton "git.t-lab.cs.teu.ac.jp/nashio/elton/server"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

const metadataHeaderPrefix = "Grpc-Metadata-"
const chunkSize int = 4096

type EltonSlave struct {
	FS     *elton.FileSystem
	Conf   elton.Config
	backup bool
}

func NewEltonSlave(conf elton.Config, backup bool) *EltonSlave {
	return &EltonSlave{FS: elton.NewFileSystem(conf.Slave.Dir, backup), Conf: conf, backup: backup}
}

func (e *EltonSlave) Serve() {
	if !e.backup {
		go func() {
			ctx := context.Background()
			ctx, cansel := context.WithCancel(ctx)
			defer cansel()

			router := mux.NewRouter()
			if err := e.RegisterEltonServiceHandlerFromEndpoint(ctx, router, fmt.Sprintf("%s:%d", e.Conf.Slave.MasterName, e.Conf.Slave.MasterPort)); err != nil {
				log.Fatal(err)
			}

			log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", e.Conf.Slave.HttpPort), router))
		}()
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", e.Conf.Slave.GrpcPort))
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer()
	pb.RegisterEltonServiceServer(server, e)

	log.Fatal(server.Serve(lis))
}

func (e *EltonSlave) RegisterEltonServiceHandlerFromEndpoint(ctx context.Context, router *mux.Router, endpoint string) (err error) {
	conn, err := grpc.Dial(endpoint, []grpc.DialOption{grpc.WithInsecure()}...)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				log.Printf("Failed to close conn to %s: %v", endpoint, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				log.Printf("Failed to close conn to %s: %v", endpoint, cerr)
			}
		}()
	}()

	return e.RegisterEltonServiceHandler(ctx, router, conn)
}

func (e *EltonSlave) RegisterEltonServiceHandler(ctx context.Context, router *mux.Router, conn *grpc.ClientConn) error {
	client := pb.NewEltonServiceClient(conn)

	router.HandleFunc(
		"/generate/object",
		func(w http.ResponseWriter, r *http.Request) {
			resp, err := e.requestGenerateObjectInfo(AnnotateContext(ctx, r), client, r)
			if err != nil {
				HTTPError(w, err)
				return
			}

			ForwardResponseStream(w, func() (proto.Message, error) {
				return resp.Recv()
			})
		},
	).Methods("PUT")
	router.HandleFunc(
		"/{delegate:(elton[1-9][0-9]*)/{object_id}",
		func(w http.ResponseWriter, r *http.Request) {
			resp, err := e.requestCommitObjectInfo(AnnotateContext(ctx, r), client, r)
			if err != nil {
				HTTPError(w, err)
				return
			}

			ForwardResponseMessage(ctx, w, resp)
		},
	).Methods("PUT")
	router.HandleFunc(
		"/{delegate:(elton[1-9][0-9]*)/{object_id}/{version:([1-9][0-9]*)}}",
		func(w http.ResponseWriter, r *http.Request) {
			resp, err := e.requestCommitObjectInfo(AnnotateContext(ctx, r), client, r)
			if err != nil {
				HTTPError(w, err)
				return
			}

			ForwardResponseMessage(ctx, w, resp)
		},
	).Methods("PUT")
	router.HandleFunc(
		"/{delegate:(elton[1-9][0-9]*)/{object_id}/{version:([1-9][0-9]*)}}",
		func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)

			oid := vars["object_id"]
			version, err := strconv.ParseUint(vars["version"], 10, 64)
			if err != nil {
				HTTPError(w, err)
				return
			}

			p, err := e.FS.Find(oid, version)
			if err != nil {
				resp, err := e.requestGetObject(AnnotateContext(ctx, r), client, r)
				if err != nil {
					HTTPError(w, err)
					return
				}

				fp, err := e.FS.Create(oid, version)
				if err != nil {
					HTTPError(w, err)
					return
				}
				defer fp.Close()

				writer := bufio.NewWriter(fp)
				for {
					obj, err := resp.Recv()
					if err == io.EOF {
						break
					}

					if err != nil {
						HTTPError(w, err)
						return
					}

					_, err = writer.Write(obj.Body)
					if err != nil {
						HTTPError(w, err)
						return
					}
					writer.Flush()
				}
			}

			http.ServeFile(w, r, p)
		},
	).Methods("GET")
	router.HandleFunc(
		"/{delegate:(elton[1-9][0-9]*)/{object_id}/{version:([1-9][0-9]*)}}",
		func(w http.ResponseWriter, r *http.Request) {
			resp, err := e.requestDeleteObject(AnnotateContext(ctx, r), client, r)
			if err != nil {
				HTTPError(w, err)
				return
			}

			ForwardResponseMessage(ctx, w, resp)
		},
	).Methods("DELETE")

	return nil
}

func (e *EltonSlave) requestGenerateObjectInfo(ctx context.Context, client pb.EltonServiceClient, r *http.Request) (pb.EltonService_GenerateObjectInfoClient, error) {
	var protoReq pb.ObjectInfo

	if err := json.NewDecoder(r.Body).Decode(&protoReq); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "%v", err)
	}

	return client.GenerateObjectInfo(ctx, &protoReq)
}

func (e *EltonSlave) requestCommitObjectInfo(ctx context.Context, client pb.EltonServiceClient, r *http.Request) (msg proto.Message, err error) {
	vars := mux.Vars(r)

	version, err := strconv.ParseUint(vars["version"], 10, 64)
	if err != nil {
		stream, err := client.GenerateObjectInfo(
			ctx,
			&pb.ObjectInfo{
				ObjectId: vars["object_id"],
				Delegate: vars["delegate"],
			},
		)
		if err != nil {
			return nil, err
		}

		obj, err := stream.Recv()
		if err != nil {
			return nil, err
		}

		version = obj.Version
	}

	protoReq := pb.ObjectInfo{
		ObjectId:        vars["object_id"],
		Version:         version,
		Delegate:        vars["delegate"],
		RequestHostname: fmt.Sprintf("%s:%d", e.Conf.Slave.Name, e.Conf.Slave.GrpcPort),
	}

	return client.CommitObjectInfo(ctx, &protoReq)
}

func (e *EltonSlave) requestGetObject(ctx context.Context, client pb.EltonServiceClient, r *http.Request) (pb.EltonService_GetObjectClient, error) {
	vars := mux.Vars(r)

	version, err := strconv.ParseUint(vars["version"], 10, 64)
	if err != nil {
		return nil, err
	}

	protoReq := &pb.ObjectInfo{
		ObjectId:        vars["object_id"],
		Version:         version,
		Delegate:        vars["delegate"],
		RequestHostname: fmt.Sprintf("%s:%d", e.Conf.Slave.Name, e.Conf.Slave.GrpcPort),
	}

	return client.GetObject(ctx, protoReq)
}

func (e *EltonSlave) requestDeleteObject(ctx context.Context, client pb.EltonServiceClient, r *http.Request) (msg proto.Message, err error) {
	vars := mux.Vars(r)

	version, err := strconv.ParseUint(vars["version"], 10, 64)
	if err != nil {
		return nil, err
	}

	protoReq := &pb.ObjectInfo{
		ObjectId:        vars["object_id"],
		Version:         version,
		Delegate:        vars["delegate"],
		RequestHostname: fmt.Sprintf("%s:%d", e.Conf.Slave.Name, e.Conf.Slave.GrpcPort),
	}

	return client.DeleteObject(ctx, protoReq)
}

func AnnotateContext(ctx context.Context, r *http.Request) context.Context {
	var pairs []string

	for k, v := range r.Header {
		if strings.HasPrefix(k, metadataHeaderPrefix) {
			pairs = append(pairs, k[len(metadataHeaderPrefix):], v[0])
		}
	}

	if len(pairs) != 0 {
		ctx = metadata.NewContext(ctx, metadata.Pairs(pairs...))
	}

	return ctx
}

type responseStreamChunk struct {
	Result proto.Message `json:"result,omitempty"`
	Error  string        `json:"error,omitempty"`
}

func ForwardResponseStream(w http.ResponseWriter, recv func() (proto.Message, error)) {
	f, ok := w.(http.Flusher)
	if !ok {
		log.Printf("Flush not supported in %T", w)
		http.Error(w, "unexpected type of web server", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	f.Flush()

	for {
		resp, err := recv()
		if err == io.EOF {
			return
		}

		if err != nil {
			buf, merr := json.Marshal(responseStreamChunk{Error: err.Error()})
			if merr != nil {
				log.Printf("Failed to marshal an error: %v", merr)
				return
			}

			if _, werr := fmt.Fprintf(w, "%s\n", buf); werr != nil {
				log.Printf("Failed to notify error to client: %v", werr)
				return
			}
			return
		}

		buf, err := json.Marshal(responseStreamChunk{Result: resp})
		if err != nil {
			log.Printf("Failed to marshal response chunk: %v", err)
			return
		}
		if _, err = fmt.Fprintf(w, "%s\n", buf); err != nil {
			log.Printf("Failed to send response chunk: %v", err)
			return
		}
		f.Flush()
	}
}

func ForwardResponseMessage(ctx context.Context, w http.ResponseWriter, r proto.Message) {
	buf, err := json.Marshal(r)
	if err != nil {
		log.Printf("Marshal error: %v", err)
		HTTPError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(buf); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

func HTTPStatusFromCode(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return http.StatusRequestTimeout
	case codes.Unknown:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusRequestTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusForbidden
	case codes.FailedPrecondition:
		return http.StatusPreconditionFailed
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	}

	log.Printf("Unknown gRPC error code: %v", code)
	return http.StatusInternalServerError
}

type errorBody struct {
	Error string `json:"error"`
}

func HTTPError(w http.ResponseWriter, err error) {
	const fallback = `{"error": "failed to marshal error message"}`

	w.Header().Set("Content-Type", "application/json")
	body := errorBody{Error: err.Error()}
	buf, merr := json.Marshal(body)
	if merr != nil {
		log.Printf("Failed to marshal error message %q: %v", body, merr)
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := io.WriteString(w, fallback); err != nil {
			log.Printf("Failed to write response: %v", err)
		}
		return
	}

	st := HTTPStatusFromCode(grpc.Code(err))
	w.WriteHeader(st)
	if _, err := w.Write(buf); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

func (e *EltonSlave) GenerateObjectInfo(o *pb.ObjectInfo, stream pb.EltonService_GenerateObjectInfoServer) error {
	return nil
}

func (e *EltonSlave) CommitObjectInfo(c context.Context, o *pb.ObjectInfo) (r *pb.EmptyMessage, err error) {
	return
}

func (e *EltonSlave) GetObject(o *pb.ObjectInfo, stream pb.EltonService_GetObjectServer) error {
	fp, err := e.FS.Open(o.ObjectId, o.Version)
	if err != nil {
		log.Println(err)
		return err
	}
	defer fp.Close()

	reader := bufio.NewReader(fp)
	buf := make([]byte, chunkSize)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if err = stream.Send(
			&pb.Object{
				ObjectId: o.ObjectId,
				Version:  o.Version,
				Body:     buf[:n],
			},
		); err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func (e *EltonSlave) PutObject(c context.Context, o *pb.Object) (r *pb.EmptyMessage, err error) {
	if err != nil {
		log.Println(err)
		return new(pb.EmptyMessage), err
	}

	err = e.FS.CreateFile(o.ObjectId, o.Version, []byte(o.Body))
	return new(pb.EmptyMessage), err
}

func (e *EltonSlave) DeleteObject(c context.Context, o *pb.ObjectInfo) (r *pb.EmptyMessage, err error) {
	return
}
