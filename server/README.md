# Elton

## Usage

CentOS7環境を想定する．

```bash
$ elton --help
NAME:
   elton -

USAGE:
   elton [global options] command [command options] [arguments...]

VERSION:
   0.0.1

AUTHOR(S):
   Taku MIZUNO <dev@nashio-lab.info>

COMMANDS:
   master
   slave
   help, h	Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h		show help
   --version, -v	print the version


$ elton master --help
NAME:
   elton master -

USAGE:
   elton master [command options] [arguments...]

OPTIONS:
   --file, -f "config.tml"	config file


$ elton slave --help
NAME:
   elton slave -

USAGE:
   elton slave [command options] [arguments...]

OPTIONS:
   --file, -f "config.tml"	config file
   --backup			backup flag
```

### 必要なもののインストール

```bash
[root]
$ yum -y install make gcc-c++ curl tar gzip
$ curl -kL -O https://github.com/google/protobuf/releases/download/v3.0.0-beta-2/protobuf-cpp-3.0.0-beta-2.tar.gz
$ tar zxvf protobuf-cpp-3.0.0-beta-2.tar.gz
$ cd protobuf-3.0.0-beta-2 && ./configure && make && make install && cd ../
$ rm -rf protobuf*
```

### systemdのサービスファイルを登録する(Elton Master)

```
### /usr/lib/systemd/system/elton.service
[Unit]
Description=elton

[Service]
Environment=GOMAXPROCS=4
ExecStart=/usr/local/bin/elton master -f /etc/elton/master.tml
KillSignal=SIGINT

[Install]
WantedBy=multi-user.target
```

### systemdのサービスファイルを登録する(Elton Slave)

```
### /usr/lib/systemd/system/elton-slave.service
[Unit]
Description=elton-slave

[Service]
EnvironmentFile=-/etc/sysconfig/elton-slave
Environment=GOMAXPROCS=4
ExecStart=/usr/local/bin/elton slave $OPTIONS -f /etc/elton/slave.tml
KillSignal=SIGINT

[Install]
WantedBy=multi-user.target
```

### systemdのサービスファイルを反映させる

```bash
[root]
$ systemctl daemon-reload
```

### Elton Slaveのoptionファイルを作成する

バックアップ用途の場合はbackupオプションを設定

```
### /etc/sysconfig/elton-slave
# OPTIONS="--backup"
OPTIONS=
```

### 設定ファイルを作成する(Elton Master)

設定ファイルは[TOML記法](http://qiita.com/b4b4r07/items/77c327742fc2256d6cbe)で書きます．

各種パラメータを適宜書き換えます．

```
### /etc/elton/master.tml
[master]
name = "192.168.189.37"
port = 12345

[[masters]]
name = "192.168.189.38"
port = 12345

[[masters]]
name = "192.168.189.39"
prot = 12345

[backup]
name = "192.168.189.37"
port = 34567

[database]
dbpath = "/mnt/elton/elton.db"
```

### 設定ファイルを作成する(Elton Slave)

設定ファイルは[TOML記法](http://qiita.com/b4b4r07/items/77c327742fc2256d6cbe)で書きます．

各種パラメータを適宜書き換えます．

```
### /etc/elton/slave.tml
[slave]
name = "192.168.189.37"
grpc_port = 34567
http_port = 23456
master_name = "192.168.189.37"
master_port = 12345
dir = "/mnt/elton"
```

### 起動する(Elton Master)

```bash
[root]
$ systemctl start elton
```

### 起動する(Elton Slave)

```bash
[root]
$ systemctl start elton-slave
```

## HTTPインタフェース
Eltonを操作するためのHTTPインタフェースの使い方です．

### PUT /generate/object
オブジェクトのIDをジェネレートするためのAPIです．
新しいオブジェクトを作成する際にまず実行するAPIです．

#### Request
Request BodyでJSONを送ります．
`object_id`に適当なお名前を入れます．

```bash
{
    "object_id":"elton.tar.gz"
}
```

#### Response
オブジェクトが作成されると以下のようなResponseが返ってきます．

```bash
{
    "result": {
        "object_id": "3509eebf71fa7ebaa86a8a2bab69847b1b4351f7d9b056a18239cff562aed8f0",
        "version": 1,
        "delegate": "192.168.189.75"
    }
}
```

#### Sample

```bash
$ curl -s -XPUT -d'{"object_id":"elton.tar.gz"}' http://slave.elton.internal.t-lab.cs.teu.ac.jp:23456/generate/object | jq .
{
    "result": {
        "object_id": "3509eebf71fa7ebaa86a8a2bab69847b1b4351f7d9b056a18239cff562aed8f0",
        "version": 1,
        "delegate": "192.168.189.75"
    }
}
```

### PUT /{delegate}/{object_id}
新しいオブジェクトを作成するAPIです．
ジェネレートしたオブジェクトIDの最新バージョン(自動)に対してファイルを送信します．

#### Request

`file=アップロードファイル`

#### Response
うまく行けば`{}`が返ってきます．

#### Sample

```bash
$ curl -s -XPUT -F file=@nashio_elton-bad1072cac599853bd9c1e40fb91e9ebb4bd5099.tar.gz http://slave.elton.internal.t-lab.cs.teu.ac.jp:23456/192.168.189.75/9e5ed6043d4b80054fc5a0ea83eebda2a37637f35a2b028cb0554d86968ffb90
{}
```

### PUT /{delegate}/{object_id}/{version:([1-9][0-9]*)}
新しいオブジェクトをバージョン指定で作成するAPIです．
ジェネレートしたオブジェクトID，バージョンに対してファイルを送信します．
あんまり使いません...

#### Request

#### Response

#### Sample

### GET /{delegate}/{object_id}/{version:([1-9][0-9]*)}

#### Request

#### Response

#### Sample

### DELETE /{delegate}/{object_id}/{version:([1-9][0-9]*)}

#### Request

#### Response

#### Sample
