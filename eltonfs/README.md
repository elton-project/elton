# Eltonfs

## Development

[go-fuse](https://github.com/hanwen/go-fuse)を要チェック

## Usage

CentOS7環境を想定．
基本的にLinux環境で動くことしか考えてない．

```bash
$ eltonfs --help
Usage:
 eltonfs [OPTIONS] ELTON_HOST MOUNTPOINT

Application Options:
     --debug       print debbuging messages. (false)
 -o=               fuse options
     --host=       this host name (localhost)
 -p, --port=       grpc listen port (51823)
     --upperdir=   union mount to upper rw directory.
     --lowerdir=   union mount to lower ro directory
     --standalone  stand-alone mode (false)

Help Options:
 -h, --help        Show this help message
```

### 必要なもののインストール

```bash
[root]
$ yum -y install fuse fuse-devel
```

### FUSEの設定ファイルを変更

```bash
# /etc/fuse.conf

user_allow_other
```

### マウントする

マウントに必要なupper，lower，mountpointのディレクトリは予め作成しておく．

```bash
# eltonfs [elton master] --upperdir=[upper] --lowerdir=[lower] --host=[this hostname] MOUNTPOINT &
$ eltonfs 192.168.189.37:12345 --upperdir=/tmp/upper --lowerdir=/tmp/lower --host=192.168.189.37 /tmp/mountpoint &
```

### アンマウントする

```bash
$ fusermount -u /tmp/mountpoint
```

### コミット

.eltonfsディレクトリ内のCOMMITファイルに書き込みが発生するとコミットされる．

```bash
$ echo hoge > MOUNTPOINT/.eltonfs/COMMIT
```

### 共有

upperディレクトリ内のCONFIGファイルを同一にすることで共有できる．

```bash
# upper/CONFIG

{"object_id":"00abaffcd2c94cddae418f597b4e9e6a1e0276b9af19399003a8e65374acb548","version":1,"delegate":"192.168.189.37"}
```

## Notes

複数マウントするときはgRPCのListen Portを変えましょう．
