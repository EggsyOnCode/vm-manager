package vm

import (
	"testing"
)

func TestTemplateCreation(t *testing.T) {
	// 1. Create a new Config object
	cfg := new(Config)

	// 2. Create an XML configuration file
	cfg.CreateXMLConfig("test", 1024, 10, ALPINE_LINUX)
}
