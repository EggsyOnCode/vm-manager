package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	libvirt "gitlab.com/libvirt/libvirt-go"
)

func main() {
	// 1. Connect to the Libvirt daemon
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		log.Fatalf("Failed to connect to qemu:///system: %v", err)
	}
	defer conn.Close()
	fmt.Println("Connected to Libvirt.")

	// 2. Read the VM XML configuration from a file
	xmlFilePath := "eg.xml" 
	xmlConfig, err := ioutil.ReadFile(xmlFilePath)
	if err != nil {
		log.Fatalf("Failed to read XML file: %v", err)
	}
	fmt.Println("VM XML configuration loaded.")

	// 3. Define the domain (VM) using the XML configuration
	dom, err := conn.DomainDefineXML(string(xmlConfig))
	if err != nil {
		log.Fatalf("Failed to define domain: %v", err)
	}
	name, _ := dom.GetName()
	fmt.Printf("Domain '%s' defined.\n", name)

	// 4. Optionally, set the domain to autostart
	err = dom.SetAutostart(true)
	if err != nil {
		log.Printf("Failed to set autostart for domain '%s': %v", name, err)
	} else {
		fmt.Printf("Domain '%s' set to autostart.\n", name)
	}

	// 5. Start the domain (VM)
	err = dom.Create()
	if err != nil {
		log.Fatalf("Failed to start domain: %v", err)
	}
	fmt.Printf("Domain '%s' started.\n", name)

	// 6. Optionally, retrieve and display the domain's state
	state, reason, err := dom.GetState()
	if err != nil {
		log.Printf("Failed to get domain state: %v", err)
	} else {
		fmt.Printf("Domain '%s' state: %s (reason: %v)\n", name, stateToString(state), (reason))
	}

	// 7. Wait for 3 minutes before stopping the VM
	time.Sleep(2 * time.Minute)
	fmt.Println("Stopping the VM after 3 minutes...")
	err = stopDomain(dom)
	if err != nil {
		log.Fatalf("Failed to stop domain: %v", err)
	}
	fmt.Println("Domain stopped successfully.")
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

// Helper function to convert domain reason to string
// func reasonToString(reason libvirt.DomainStateReason) string {
// 	switch reason {
// 	case libvirt.DOMAIN_NONE:
// 		return "No reason"
// 	case libvirt.DOMAIN_RUNNING:
// 		return "Running"
// 	case libvirt.DOMAIN_BLOCKED:
// 		return "Blocked by resource"
// 	case libvirt.DOMAIN_PAUSED:
// 		return "Paused by user"
// 	case libvirt.DOMAIN_SHUTDOWN:
// 		return "Shutdown by guest"
// 	case libvirt.DOMAIN_SHUTOFF:
// 		return "Shut off by user"
// 	case libvirt.DOMAIN_CRASHED:
// 		return "Crashed"
// 	case libvirt.DOMAIN_PMSUSPENDED:
// 		return "Power management suspended"
// 	default:
// 		return "Unknown reason"
// 	}
// }

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
