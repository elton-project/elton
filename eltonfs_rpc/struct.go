package eltonfs_rpc

const Setup1StructID = 1

type Setup1 struct {
	XXX_XDR_ID      struct{} `xdrid:"1"`
	ClientName      string   `xdr:"1"`
	VersionMajor    uint64   `xdr:"2"`
	VersionMinor    uint64   `xdr:"3"`
	VersionRevision uint64   `xdr:"4"`
}

const Setup2StructID = 2

type Setup2 struct {
	XXX_XDR_ID      struct{} `xdrid:"2"`
	Error           uint64   `xdr:"1"`
	Reason          string   `xdr:"2"`
	ServerName      string   `xdr:"3"`
	VersionMajor    uint64   `xdr:"4"`
	VersionMinor    uint64   `xdr:"5"`
	VersionRevision uint64   `xdr:"6"`
}

const PingStructID = 3

type Ping struct {
	XXX_XDR_ID struct{} `xdrid:"3"`
}

const MaxStructID = 3
