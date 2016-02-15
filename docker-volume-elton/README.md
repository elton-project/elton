# Docker Volume Plugin for Eltonfs
## Development
[本家ドキュメント](https://docs.docker.com/engine/extend/plugins_volume/)，[プラグインヘルパー](https://github.com/docker/go-plugins-helpers)を要チェック

## Usage
CentOS7環境を想定する．
Dockerはセットアップ済みであるとする．

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

### systemdのサービスファイルを登録する

```bash

# /usr/lib/systemd/system/docker-volume-elton.service

[Unit]
Description=docker-volume-elton

[Service]
EnvironmentFile=-/etc/sysconfig/docker-volume-elton
Environment=GOMAXPROCS=4
ExecStart=/usr/local/bin/docker-volume-elton $OPTION
KillSignal=SIGINT

[Install]
WantedBy=multi-user.target
```

### systemdのサービスファイルを反映させる

```bash
[root]
$ systemctl daemon-reload
```

### optionファイルを作成する

```bash
# /etc/sysconfig/docker-volume-elton

# OPTION="-root=[target directory] -hostname=[this hostname] [elton master]"
OPTION="-root=/home/elton -hostname=192.168.189.37 192.168.189.37:12345"
```

### 起動する

```bash
[root]
$ systemctl start docker-volume-elton
```

### volumeを作成する

```bash
[root]
$ docker volume create --driver=eltonfs --name=[volume name]
```

### マウントする

```bash
[root]
$ docker run -d -v [volume name]:[mount point] hogehoge
```

### よしなに遊ぶ


## Notes
Dockerはかなり開発の早いプロダクトです．
ライブラリ側のバージョンは一応固定してますが，そのうち動かなくなる未来が見えます．

あくまでDocker 1.9系で動くVolume Pluginくらいに考えてください．
