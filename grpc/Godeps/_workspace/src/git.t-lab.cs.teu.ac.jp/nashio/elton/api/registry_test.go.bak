package api

import "testing"

var reg *Registry

func init() {
	reg, _ = NewRegistry(Config{Database: DBConfig{DBPath: "/tmp/test.db"}})
}

func TestGenerateObjectsInfo(t *testing.T) {
	objs, err := reg.GenerateObjectsInfo(FileName{Name: []string{"a.txt", "b.txt"}})
	if err != nil {
		t.Fatalf("Error by GenerateObjectsInfo(). %v", err)
	}

	if len(objs) != 2 {
		t.Fatalf("Expected 2 objects Got %d", len(objs))
	}

	for _, v := range objs {
		if v.ObjectID == "" {
			t.Fatalf("Expected object id Got %s", v.ObjectID)
		}
	}
}
