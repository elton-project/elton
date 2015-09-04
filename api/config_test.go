package api

import "testing"

func TestLoad(t *testing.T) {
	conf, err := Load("../examples/template.tml")
	if err != nil {
		t.Fatalf("Error by Load(). %v", err)
	}

	t.Logf("%v", conf)
}
