package eltonfs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/net/context"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"

	pb "gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc/proto"
)

type eltonFile struct {
	nodefs.File
	node *eltonNode
}

func (f *eltonFile) InnerFile() nodefs.File {
	return f.File
}

func (f *eltonFile) SetInode(n *nodefs.Inode) {
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

func (f *eltonFile) Write(data []byte, off int64) (n uint32, code fuse.Status) {
	if f.node.key == ELTONFS_COMMIT_NAME {
		f.commit()
	}

	n, code = f.File.Write(data, off)
	f.Flush()

	if f.node.key == ELTONFS_CONFIG_NAME {
		f.reload()
	}

	return n, code
}

func (f *eltonFile) Release() {
	f.File.Release()
}

func (f *eltonFile) Fsync(flags int) (code fuse.Status) {
	return f.File.Fsync(flags)
}

func (f *eltonFile) Truncate(size uint64) fuse.Status {
	return f.File.Truncate(size)
}

func (f *eltonFile) Chmod(mode uint32) fuse.Status {
	return f.File.Chmod(mode)
}

func (f *eltonFile) Chown(uid uint32, gid uint32) fuse.Status {
	return f.File.Chown(uid, gid)
}

func (f *eltonFile) GetAttr(a *fuse.Attr) fuse.Status {
	return f.File.GetAttr(a)
}

func (f *eltonFile) Utimens(a *time.Time, m *time.Time) fuse.Status {
	return f.File.Utimens(a, m)
}

// ディレクトリツリーを再構築する。
// CONFIGを再読込してから、EltonTreeを再構築する。
// このメソッドは、configファイルが書き込まれた直後に呼び出される。
func (f *eltonFile) reload() {
	f.node.fs.mux.Lock()
	defer f.node.fs.mux.Unlock()
	f.node.fs.root = f.node.fs.newNode("", time.Now())
	f.node.fs.newEltonTree()
}

// ディレクトリツリーをコミットする。
// このメソッドは、COMMITファイルを書き換えようとしたときに呼び出される。
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
	ioutil.WriteFile(filepath.Join(f.node.fs.lower, obj.ObjectId[:2], fmt.Sprintf("%s-%d", obj.ObjectId[2:], obj.Version)), data, 0644)

	obj.RequestHostname = fmt.Sprintf("%s:%d", f.node.fs.options.HostName, f.node.fs.options.Port)
	_, err = client.CommitObjectInfo(context.Background(), obj)

	cdata, _ := json.Marshal(obj)
	ioutil.WriteFile(p, cdata, 0644)
	return err
}

// コミット時に呼び出されるメソッド。
// ファイルツリーをwalkｋして、FileInfoのリストを返す。
// このときに、変更済みのファイル(upperディレクトリに存在するファイル)はコミットされ、lowerディレクトリに移動される。
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
			Key:      n.key,
			Version:  n.version,
			Delegate: n.delegate,
			Size:     n.info.Size,
			Time:     n.info.ChangeTime(),
		}

		if n.basePath == n.fs.upper {
			// 未コミットのファイルをコミットする。
			// コミットに成功すると、nおよびfiのバージョン番号などのフィールドを更新する。
			if err := f.commitObject(client, n, &fi); err != nil {
				return files, err
			}
		}

		files = append(files, fi)
	}

	return files, nil
}

// nのコミット直後に呼び出される。
// nの実態をupperからlowerに移動される。
func (f *eltonFile) moveFile(n *eltonNode, obj *pb.ObjectInfo) error {
	p := filepath.Join(n.fs.lower, obj.ObjectId[:2], fmt.Sprintf("%s-%d", obj.ObjectId[2:], obj.Version))
	dir := filepath.Dir(p)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0744); err != nil {
			return err
		}
	}

	if err := os.Rename(
		n.filename(),
		p,
	); err != nil {
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
	n.key = obj.ObjectId
	n.version = obj.Version
	n.delegate = obj.Delegate

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
