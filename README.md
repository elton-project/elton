# Elton
Elton is a distributed file system for accelerate data sharing through the WAN.


## What's Elton?

パブリッククラウド・プライベートクラウド間(疎結合マルチクラスタ間)を連携し，
ネットワークトラフィックを抑えながら効率よくデータ共有をしようプロジェクト．
現在はDockerをターゲットにしたファイルシステムインタフェースとCDN向けのHTTPインタフェースを提供しようと頑張ってる．


## Development

- Golang 1.12
- make
- git

その他の依存しているツールやライブラリは、`make build-deps`でインストールできる。
ビルド環境は、AMD64のDebian busterを想定。


### Build

```bash
$ make binary
```

### Test

### Run
