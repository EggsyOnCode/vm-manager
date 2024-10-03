package vm

import (
	"testing"
)

func TestTemplateCreation(t *testing.T) {
	// 1. Create a new Config object
	cfg := &Config{
		Name:          "hello",
		Mem:           1024,
		PathToBootImg: "test.img",
		Storage:       1,
	}

	// 2. Create an XML configuration file
	cfg.CreateXMLConfig("test", 1024, 10, "test.img")
}
