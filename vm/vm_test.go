package vm

import (
	"testing"
)

func TestTemplateCreation(t *testing.T) {
	// 1. Create a new Config object
	cfg := new(Config)

	// 2. Create an XML configuration file
	file := cfg.CreateXMLConfig("test", 1024, 10, ALPINE_LINUX)
	path := "/home/xen/Desktop/code/virt/vm-manger/templates/" + file
	cfg.XmlConfig = path

	vm := new(VM)
	vm.NewVM(cfg)
	vm.Start()
}
