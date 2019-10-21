package eltonfs_rpc

//StructID=1
type Setup1 struct {
	XXX_XDR_ID      struct{} `xdrid:"1"`
	ClientName      string   `xdr:"1"`
	VersionMajor    uint64   `xdr:"2"`
	VersionMinor    uint64   `xdr:"3"`
	VersionRevision uint64   `xdr:"4"`
}

//StructID=2
type Setup2 struct {
	XXX_XDR_ID      struct{} `xdrid:"2"`
	Error           uint64   `xdr:"1"`
	Reason          string   `xdr:"2"`
	ServerName      string   `xdr:"3"`
	VersionMajor    uint64   `xdr:"4"`
	VersionMinor    uint64   `xdr:"5"`
	VersionRevision uint64   `xdr:"6"`
}

//StructID=3
type Ping struct {
	XXX_XDR_ID struct{} `xdrid:"3"`
}
