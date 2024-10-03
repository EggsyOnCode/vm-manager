package vm

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"text/template"
	"time"

	"github.com/google/uuid"
	libvirt "gitlab.com/libvirt/libvirt-go"
)

const (
	vCPUS = 2
)

type OSTYPE string

var ALPINE_LINUX OSTYPE

type Config struct {
	Name          string
	Mem           int
	OsType        OSTYPE
	PathToBootImg string
	UUID          string
	VCPUs         int
	Storage       int    // how much storage in MB
	XmlConfig     string // path to xml config on server
}

type VmState struct {
	State  string
	Config *Config
}

type VM struct {
	Conn   *libvirt.Connect
	Domain *libvirt.Domain
	Config *Config
}

func (cfg *Config) CreateXMLConfig(n string, mem int, store int, osType string) string {
	uuid, _ := uuid.NewUUID()
	cfg = &Config{
		Name:    n,
		Mem:     mem,
		Storage: store,
		VCPUs:   vCPUS,
		UUID:    uuid.String(),
	}

	if osType == "alpine_linux" {
		pathToAlpineLinux := os.Getenv("DEFAULT_OS_PATH")
		cfg.PathToBootImg = string(pathToAlpineLinux)
	}

	log.Printf("%v \n", cfg)

	cwd, _ := os.Getwd()
	basePath := cwd + "/templates"
	os.Chdir(basePath)

	curDir, _ := os.Getwd()
	log.Printf("curr dir is %v", curDir)
	tmpl, err := template.ParseFiles("template.xml")
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	// 3. Create a file to write the generated XML
	name := fmt.Sprintf("%s.xml", uuid)
	outputFile, err := os.Create(name)
	if err != nil {
		log.Fatalf("Error creating output file: %v", err)
	}
	defer outputFile.Close()

	// Create an XML encoder
	encoder := xml.NewEncoder(outputFile)
	encoder.Indent("", "  ") // Indent the XML for readability

	// 4. Execute the template with the data
	err = tmpl.Execute(outputFile, cfg)
	if err != nil {
		log.Fatalf("Error executing template: %v", err)
	}

	// Ensure all data is written to the file
	if err := encoder.Flush(); err != nil {
		log.Fatalf("Error flushing encoder: %v", err)
	}

	log.Printf("Created XML config file: %s", name)

	return name
}

func (vm *VM) NewVM(cfg *Config) error {
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		log.Fatalf("Failed to connect to qemu:///system: %v", err)
	}
	vm.Config = cfg
	vm.Conn = conn
	log.Printf("config ffile is at %v", vm.Config.XmlConfig)
	// reading the file
	xmlFile, _ := ioutil.ReadFile(vm.Config.XmlConfig)

	// creating storage file at predefined path
	err1 := createDiskImage(cfg)
	if err1 != nil {
		return fmt.Errorf("failed to create storage %s", err1)
	}

	dom, err := conn.DomainDefineXML(string(xmlFile))
	if err != nil {
		return fmt.Errorf("failed to define domain: %v", err)
	}
	vm.Domain = dom
	err = dom.SetAutostart(true)
	if err != nil {
		log.Printf("Failed to set autostart for domain '%s': %v", cfg.Name, err)
	} else {
		fmt.Printf("Domain '%s' set to autostart.\n", cfg.Name)
	}

	log.Printf(
		"Created new VM with name: %s, memory: %d, storage: %d, osType: %s",
		cfg.Name,
		cfg.Mem,
		cfg.Storage,
		cfg.OsType,
	)
	return nil
}

func (vm *VM) Start() error {
	n, _ := vm.Domain.GetName()
	log.Printf("name is %v", n)
	err := vm.Domain.Create()
	if err != nil {
		return fmt.Errorf("failed to start domain: %v", err)
	}

	currState, _, _ := vm.Domain.GetState()
	log.Printf("Current state of the vm %s is %s", vm.Config.Name, stateToString(currState))

	log.Printf("Started VM %s", vm.Config.Name)

	return nil
}

func (vm *VM) Stop() error {
	err := stopDomain(vm.Domain)
	if err != nil {
		return fmt.Errorf("failed to stop domain: %v", err)
	}

	log.Printf("Stopped VM %s", vm.Config.Name)

	return nil
}

func (vm *VM) Delete() error {
	err := vm.Domain.Destroy()
	if err != nil {
		return fmt.Errorf("failed to delete domain: %v", err)
	}

	log.Printf("Deleted VM %s", vm.Config.Name)
	return nil
}

func (vm *VM) GetInfo() VmState {
	state, _, err := vm.Domain.GetState()
	if err != nil {
		log.Printf("Failed to get domain state: %v", err)
	}

	return VmState{
		State:  stateToString(state),
		Config: vm.Config,
	}
}

// Helper function to convert domain state to string
func stateToString(state libvirt.DomainState) string {
	switch state {
	case libvirt.DOMAIN_NOSTATE:
		return "No state"
	case libvirt.DOMAIN_RUNNING:
		return "Running"
	case libvirt.DOMAIN_BLOCKED:
		return "Blocked"
	case libvirt.DOMAIN_PAUSED:
		return "Paused"
	case libvirt.DOMAIN_SHUTDOWN:
		return "Shutdown"
	case libvirt.DOMAIN_SHUTOFF:
		return "Shut off"
	case libvirt.DOMAIN_CRASHED:
		return "Crashed"
	case libvirt.DOMAIN_PMSUSPENDED:
		return "Power management suspended"
	default:
		return "Unknown"
	}
}

// stopDomain stops the specified domain gracefully
func stopDomain(dom *libvirt.Domain) error {
	// Attempt to shut down the domain gracefully
	err := dom.Shutdown()
	if err != nil {
		return fmt.Errorf("failed to shutdown domain gracefully: %v", err)
	}

	// Optionally, wait for the domain to shut down completely
	state, _, err := dom.GetState()
	for state == libvirt.DOMAIN_RUNNING {
		time.Sleep(1 * time.Second) // Wait before checking again
		state, _, err = dom.GetState()
		if err != nil {
			return fmt.Errorf("failed to get domain state: %v", err)
		}
	}

	return nil
}

func createDiskImage(cfg *Config) error {
	// defualt path to storage
	pathToStorage := fmt.Sprintf("/var/lib/libvirt/images/%s.qcow2", cfg.Name)
	cmd := exec.Command("qemu-img", "create", "-f", "qcow2", pathToStorage, fmt.Sprintf("%dM", cfg.Storage))
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to create disk image: %v", err)
	}
	return nil
}
