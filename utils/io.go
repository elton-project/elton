package utils

type MustWriter interface {
	MustWrite([]byte)
}
type MustReader interface {
	MustRead([]byte) (n int)
	MustReadAll([]byte)
}
