package vm

import (
	"fmt"
	"log"
	"os"
	"text/template"
	"time"

	"github.com/google/uuid"
	libvirt "gitlab.com/libvirt/libvirt-go"
)

const (
	vCPUS = 2
)

type Config struct {
	Name          string
	Mem           int
	PathToBootImg string
	vCPUs         int
	Storage       int    // how much storage in GB
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

func (cfg *Config) CreateXMLConfig(n string, mem int, store int, img string) {
	uuid, _ := uuid.NewUUID()
	cfg = &Config{
		Name:          n,
		Mem:           mem,
		PathToBootImg: img,
		Storage:       store,
		vCPUs:         vCPUS,
	}

	// 2. Parse the XML template
	tmpl, err := template.ParseFiles("templates/template.xml")
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	// 3. Create a file to write the generated XML
	name := fmt.Sprintf("templates/%s.xml", uuid)
	outputFile, err := os.Create(name)
	if err != nil {
		log.Fatalf("Error creating output file: %v", err)
	}
	defer outputFile.Close()

	// 4. Execute the template with the data
	err = tmpl.Execute(outputFile, cfg)
	if err != nil {
		log.Fatalf("Error executing template: %v", err)
	}
}

func (vm *VM) NewVM(cfg *Config) error {
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		log.Fatalf("Failed to connect to qemu:///system: %v", err)
	}
	vm.Config = cfg
	vm.Conn = conn
	dom, err := vm.Conn.DomainDefineXML(vm.Config.XmlConfig)
	if err != nil {
		return fmt.Errorf("failed to define domain: %v", err)
	}
	vm.Domain = dom
	return nil
}

func (vm *VM) Start() error {
	err := vm.Domain.Create()
	if err != nil {
		return fmt.Errorf("failed to start domain: %v", err)
	}
	return nil
}

func (vm *VM) Stop() error {
	err := stopDomain(vm.Domain)
	if err != nil {
		return fmt.Errorf("failed to stop domain: %v", err)
	}
	return nil
}

func (vm *VM) Delete() error {
	err := vm.Domain.Destroy()
	if err != nil {
		return fmt.Errorf("failed to delete domain: %v", err)
	}
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
