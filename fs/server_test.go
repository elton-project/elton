package fs

import (
	"nashio-lab.info/elton/fs"
	"testing"
)

func init() {
	fs.ServerInitialize("./", "localhost:56789")
}

func TestServerMigration(t *testing.T) {
}
