# Elton
![](https://img.shields.io/github/license/elton-project/elton)
![](https://tokei.rs/b1/github/elton-project/elton?category=code)
![](https://img.shields.io/github/commit-activity/m/elton-project/elton)
![](https://img.shields.io/github/v/tag/elton-project/elton)

Elton is a distributed transactional file system to improve read performance.

## Instllation
```bash
$ git clone --depth=1 https://github.com/elton-project/elton.git
$ cd elton
$ make build
$ sudo make install
```

## Usage
```bash
$ elton volume create foo
$ mount -t elton -o vol=foo dummy /mnt
```

## Development
開発環境は下記の環境を想定する。

- Ubuntu 19.10 (Eoan Ermine)
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
