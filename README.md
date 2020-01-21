# Elton
Elton is a distributed transactional file system to improve read performance.

## Instllation
```bash
$ git clone --depth=1 https://gitlab.t-lab.cs.teu.ac.jp/yuuki/elton
$ cd elton
$ make build
$ sudo make install
```

## Usage
```bash
$ modprobe elton && sleep 2
$ elton volume create foo
$ mount -t elton -o vol=foo dummy /mnt
```

## Development
開発環境は下記の環境を想定する。

- Ubuntu 19.04 (Disco Dingo)
- Go 1.13
- gcc
- make
- git
- clang-format
- docker
- protobuf-compiler (protoc)
- protobuf-compiler-grpc
- protoc-gen-go

```bash
$ make generate fmt
$ make -j8 build
$ make test
``` 
