# Elton Project

- 次回MTG 5/10 11:00  
minutes/ 議事録
thesis/ 修論等

## What's Elton?

パブリッククラウド・プライベートクラウド間(疎結合マルチクラスタ間)を連携し，
ネットワークトラフィックを抑えながら効率よくデータ共有をしようプロジェクト．
現在はDockerをターゲットにしたファイルシステムインタフェースとCDN向けのHTTPインタフェースを提供しようと頑張ってる．

Eltonは，以下のサブプロジェクトで構成される．

- Elton Master
 - メタデータを管理・発行し，バックアップやプロキシ等のスケジューリング機能を持つEltonの中核
- Elton Slave
 - HTTPインタフェースを提供し，ファイル管理やキャッシュ機能を持つ
- Eltonfs
 - FUSEベースのファイルシステムインタフェースを提供する
- Docker Volume Plugin for Eltonfs
 - Docker Volume Plugin機能を用い，EltonfsをDockerで利用できるようにする

## Development

実装は全てGolang 1.4.2で行っている．
開発環境に必要なものは以下の様なものがあげられる．

- Golang 1.4.2
- Editor (go-modeがきちんとしているものがいい ex. Emacs，Vim，Atom)
  - godef，goimportsとかも入れると良い
- Godeps (依存管理ツール `go get -u github.com/kr/godep` でインストール)
- Docker 1.9≦
  - docker-compose
- make
- git

ビルド用のDockerfileを参考にCentOS7環境で開発環境を構築すると良い(OSX，Windowsだとかなり面倒)．

### File Tree

GOPATHにcloneする．
clone後に `godep restore -v` とかやるとgodefが効くようになって便利．

```
.
├── Dockerfile(ビルド用のDockerfile)
├── Godeps(依存関係を管理するGodeps用ディレクトリ)
│   ├── Godeps.json
│   └── Readme
├── Makefile
├── README.md
├── docker-compose.yml(簡易実行用のdocker-composeファイル)
├── docker-volume-elton(Docker Volume Plugin実装ディレクトリ)
│   ├── Makefile
│   ├── driver.go
│   └── main.go
├── eltonfs(Eltonfs実装ディレクトリ)
│   ├── README.md
│   ├── eltonfs
│   │   ├── Makefile
│   │   └── main.go
│   ├── files.go
│   ├── fs.go
│   ├── grpc.go
│   └── node.go
├── examples(簡易実行用のサンプルディレクトリ)
│   ├── master.tml
│   ├── slave.tml
│   └── start.sh
├── grpc(gRPC実装ディレクトリ)
│   ├── master.go
│   ├── proto
│   │   ├── Makefile
│   │   ├── elton_service.pb.go
│   │   └── elton_service.proto
│   └── slave.go
└── server(Elton Master・Slave実装ディレクトリ)
    ├── config.go
    ├── elton
    │   ├── Dockerfile
    │   ├── Makefile
    │   ├── commands.go
    │   ├── main.go
    │   └── version.go
    ├── fs.go
    └── registry.go
```

### Build

ビルドは基本的にmakeを用いてDockerを通じて行う．
実行後binディレクトリが作成され，その中にバイナリが作成される．

```bash
[root]
$ make binary
```

### Run

docker-composeを用いるとEltonを簡易実行できる．
docker-compose.ymlをよしなに変更するといろいろ楽しい．

```bash
[root]
$ make testall
```
