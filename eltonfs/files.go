package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/net/context"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"

	pb "../grpc/proto"
)

type eltonFile struct {
	nodefs.File
	node *eltonNode

	key      string
	version  uint64
	delegate string
}

func (f *eltonFile) InnerFile() nodefs.File {
	return f.File
}

func (f *eltonFile) Flush() fuse.Status {
	code := f.File.Flush()
	if !code.Ok() {
		return code
	}

	st := syscall.Stat_t{}
	err := syscall.Stat(f.node.filename(), &st)
	f.node.info.Size = uint64(st.Size)
	f.node.info.Blocks = uint64(st.Blocks)
	return fuse.ToStatus(err)
}

func (f *eltonFile) Read(buf []byte, off int64) (res fuse.ReadResult, code fuse.Status) {
	return f.File.Read(buf, off)
}

func (f *eltonFile) Write(data []byte, off int64) (uint32, fuse.Status) {
	if f.key == ELTONFS_COMMIT_NAME {
		f.commit()
	}

	return f.File.Write(data, off)
}

func (f *eltonFile) commit() error {
	f.node.fs.mux.Lock()
	defer f.node.fs.mux.Unlock()
	files, err := f.getFileTree("", f.node.fs.Root())
	if err != nil {
		return err
	}

	p := filepath.Join(f.node.fs.upper, ELTONFS_CONFIG_NAME)
	buf, err := ioutil.ReadFile(p)
	if err != nil {
		return err
	}

	obj := new(pb.ObjectInfo)
	if err = json.Unmarshal(buf, obj); err != nil {
		return err
	}

	client := pb.NewEltonServiceClient(f.node.fs.connection)
	stream, err := client.GenerateObjectInfo(context.Background(), obj)
	if err != nil {
		return err
	}
	obj, err = stream.Recv()
	if err != nil {
		return err
	}

	data, _ := json.Marshal(files)
	ioutil.WriteFile(filepath.Join(f.node.fs.lower, obj.ObjectId[:2], fmt.Sprintf("%s-%d", obj.ObjectId[2:], obj.Version)), data, 0600)

	obj.RequestHostname = fmt.Sprintf("%s:%d", f.node.fs.options.HostName, f.node.fs.options.Port)
	_, err = client.CommitObjectInfo(context.Background(), obj)

	cdata, _ := json.Marshal(obj)
	ioutil.WriteFile(p, cdata, 0600)
	return err
}

func (f *eltonFile) getFileTree(prefix string, root nodefs.Node) ([]FileInfo, error) {
	if prefix != "" {
		prefix += "/"
	}

	files := make([]FileInfo, 0)
	client := pb.NewEltonServiceClient(f.node.fs.connection)
	for k, v := range root.Inode().Children() {
		if k == ELTONFS_CONFIG_DIR {
			continue
		}

		p := fmt.Sprintf("%s%s", prefix, k)
		if v.IsDir() {
			fis, err := f.getFileTree(p, v.Node())
			if err != nil {
				return files, err
			}

			files = append(files, fis...)
			continue
		}

		n := v.Node().(*eltonNode)
		fi := FileInfo{
			Name:     p,
			Key:      n.file.key,
			Version:  n.file.version,
			Delegate: n.file.delegate,
			Size:     n.info.Size,
			Time:     n.info.ChangeTime(),
		}

		if n.basePath == n.fs.upper {
			if err := f.commitObject(client, n, &fi); err != nil {
				return files, err
			}
		}

		files = append(files, fi)
	}

	return files, nil
}

func (f *eltonFile) moveFile(n *eltonNode, obj *pb.ObjectInfo) error {
	p := filepath.Join(n.fs.lower, obj.ObjectId[:2], fmt.Sprintf("%s-%d", obj.ObjectId[2:], obj.Version))
	dir := filepath.Dir(p)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0700); err != nil {
			log.Println(err)
			return err
		}
	}

	if err := os.Rename(
		n.filename(),
		p,
	); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (f *eltonFile) commitObject(client pb.EltonServiceClient, n *eltonNode, fi *FileInfo) error {
	stream, err := client.GenerateObjectInfo(
		context.Background(),
		&pb.ObjectInfo{
			ObjectId: fi.Key,
			Delegate: fi.Delegate,
		})
	if err != nil {
		return err
	}

	obj, err := stream.Recv()
	if err != nil {
		return err
	}

	if err = f.moveFile(n, obj); err != nil {
		return err
	}

	n.basePath = n.fs.lower
	n.file.key = obj.ObjectId
	n.file.version = obj.Version
	n.file.delegate = obj.Delegate

	fi.Key = obj.ObjectId
	fi.Version = obj.Version
	fi.Delegate = obj.Delegate

	_, err = client.CommitObjectInfo(context.Background(), &pb.ObjectInfo{
		ObjectId:        fi.Key,
		Version:         fi.Version,
		Delegate:        fi.Delegate,
		RequestHostname: fmt.Sprintf("%s:%d", f.node.fs.options.HostName, f.node.fs.options.Port),
	})

	return err
}
