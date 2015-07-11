package grpc

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/context"

	elton "../api"
	pb "./proto"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

const metadataHeaderPrefix = "Grpc-Metadata-"

type EltonSlave struct {
	FS   elton.FileSystem
	Conf elton.Config
}

func NewEltonSlave(conf elton.Config) *EltonSlave {
	return &EltonSlave{FS: elton.NewFileSystem(conf.Slave.Dir), Conf: conf}
}

func (e *EltonSlave) Serve() {
	defer e.Connection.Close()

	go func() {
		ctx := context.Background()
		ctx, cansel := context.WithCancel(ctx)
		defer cansel()

		router := mux.NewRouter()
		if err := RegisterEltonServiceHandlerFromEndpoint(ctx, router, e.Conf.Slave.MasterHostName); err != nil {
			log.Fatal(err)
		}

		log.Fatal(http.ListenAndServe(e.Conf.Slave.HttpHostName, router))
	}()

	lis, err := net.Listen("tcp", e.Conf.Slave.GrpcHostName)
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer()
	pb.RegisterEltonServiceServer(server, e)

	log.Fatal(server.Serve(lis))
}

func RegisterEltonServiceHandlerFromEndpoint(ctx context.Context, router *mux.Router, endpoint string) (err error) {
	conn, err := grpc.Dial(endpoint)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				log.Errorf("Failed to close conn to %s: %v", endpoint, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				log.Errorf("Failed to close conn to %s: %v", endpoint, cerr)
			}
		}()
	}()

	return RegisterEltonServiceHandler(ctx, router, conn)
}

func RegisterEltonServiceHandler(ctx context.Context, router *mux.Router, conn *grpc.ClientConn) error {
	client := pb.NewEltonServiceClient(conn)

	router.HandleFunc(
		"/generate/objectid",
		func(w http.ResponseWriter, r *http.Request) {
			resp, err := e.requestGenerateObjectID(AnnotateContext(ctx, r), client, r)
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
		"/{delegate:(elton[1-9][0-9]*)/{object_id}/{version:([1-9][0-9]*)}}",
		func(w http.ResponseWriter, r *http.Request) {
			resp, err := e.requestCreateObjectInfo(AnnotateContext(ctx, r), client, r)
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
		"/{delegate:(elton[1-9][0-9]*)/{object_id}/{version:([1-9][0-9]*)}}",
		func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)

			oid := vars["object_id"]
			version, err := strconv.ParseUint(vars["version"], 10, 64)
			if err != nil {
				return nil, err
			}
			delegate := vars["delegate"]
			p, err := e.fs.Find(oid, version)
			if err != nil {
				resp, err := e.requestGetObject(AnnotateContext(ctx, r), client, r)
				if err != nil {
					HTTPError(w, err)
					return
				}

				ForwardResponseStream(w, func() (proto.Message, error) {
					return resp.Recv()
				})
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

			ForwardResponseMesage(w, resp)
		},
	).Methods("DELETE")

	return nil
}

func (e *EltonSlave) requestGenerateObjectID(ctx context.Context, client pb.EltonServiceClient, r *http.Request) (pb.EltonService_GenerateObjectIDClient, error) {
	var protoReq pb.ObjectName

	if err = json.NewDecoder(r.Body).Decode(&protoReq); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "%v", err)
	}

	return client.GenerateObjectID(ctx, &protoReq)
}

func (e *EltonSlave) requestCreateObjectInfo(ctx context.Context, client pb.EltonServiceClient, r *http.Request) (pb.EltonService_CreateObjectInfoClient, error) {
	vars := mux.Vars(r)

	version, err := strconv.ParseUint(vars["version"], 10, 64)
	if err != nil {
		return nil, err
	}

	protoReq := &pb.ObjectInfo{
		ObjectId:        vars["object_id"],
		Version:         version,
		Delegate:        vars["delegate"],
		RequestHostname: e.Conf.Slave.GrpcHostName,
	}

	return client.CreateObjectInfo(ctx, protoReq)
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
		RequestHostname: e.Conf.Slave.GrpcHostName,
	}
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
		RequestHostname: e.Conf.Slave.GrpcHostName,
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

func ForwardResponseStream(w http.ResponseWriter, resv func() (proto.Message, error)) {
	f, ok := w.(http.Flusher)
	if !ok {
		log.Errorf("Flush not supported in %T", w)
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
				log.Errorf("Failed to marshal an error: %v", merr)
				return
			}

			if _, werr := fmt.Fprintf(w, "%s\n", buf); werr != nil {
				log.Errorf("Failed to notify error to client: %v", werr)
				return
			}
			return
		}

		buf, err := json.Marshal(responseStreamChunk{Result: resp})
		if err != nil {
			log.Errorf("Failed to marshal response chunk: %v", err)
			return
		}
		if _, err = fmt.Fprintf(w, "%s\n", buf); err != nil {
			log.Errorf("Failed to send response chunk: %v", err)
			return
		}
		f.Flush()
	}
}

func ForwardResponseMessage(ctx context.Context, w http.ResponseWriter, r proto.Message) {
	buf, err := json.Marshal(resp)
	if err != nil {
		log.Errorf("Marshal error: %v", err)
		HTTPError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(buf); err != nil {
		log.Errorf("Failed to write response: %v", err)
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

	log.Errorf("Unknown gRPC error code: %v", code)
	return http.StatusInternalServerError
}

type errorBody struct {
	Error string `json:"error"`
}

func HTTPError(ctx context.Context, w http.ResponseWriter, err error) {
	const fallback = `{"error": "failed to marshal error message"}`

	w.Header().Set("Content-Type", "application/json")
	body := errorBody{Error: err.Error()}
	buf, merr := json.Marshal(body)
	if merr != nil {
		log.Errorf("Failed to marshal error message %q: %v", body, merr)
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := io.WriteString(w, fallback); err != nil {
			log.Errorf("Failed to write response: %v", err)
		}
		return
	}

	st := HTTPStatusFromCode(grpc.Code(err))
	w.WriteHeader(st)
	if _, err := w.Write(buf); err != nil {
		log.Errorf("Failed to write response: %v", err)
	}
}
