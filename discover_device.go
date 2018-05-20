package ble

import (
//	"fmt"
	"log"
	"time"

	"github.com/godbus/dbus"
)


// Discover puts the adapter in discovery mode,
// waits for the specified timeout to discover one of the given name,
// and then stops discovery mode.
func (adapter *blob) DiscoverByDevice(timeout time.Duration, name string) error {
	conn := adapter.conn
	signals := make(chan *dbus.Signal)
	defer close(signals)
	conn.bus.Signal(signals)
	defer conn.bus.RemoveSignal(signals)
	rule := "type='signal',interface='org.freedesktop.DBus.ObjectManager',member='InterfacesAdded'"
	err := adapter.conn.addMatch(rule)
	if err != nil {
		return err
	}
	defer func() { _ = conn.removeMatch(rule) }()
	err = adapter.StartDiscovery()
	if err != nil {
		return err
	}
	defer func() { _ = adapter.StopDiscovery() }()
	var t <-chan time.Time
	if timeout != 0 {
		t = time.After(timeout)
	}
	return adapter.discoverByDeviceLoop(name, signals, t)
}

func (adapter *blob) discoverByDeviceLoop(name string, signals <-chan *dbus.Signal, timeout <-chan time.Time) error {
	for {
		select {
		case s := <-signals:
			switch s.Name {
			case interfacesAdded:
				if adapter.discoveryByDeviceComplete(s, name) {
					return nil
				}
			default:
				log.Printf("%s: unexpected signal %s\n", adapter.Name(), s.Name)
			}
		case <-timeout:
				log.Printf("whatever")
//			return DiscoveryTimeoutError(name)
		}
	}
}

func (adapter *blob) discoveryByDeviceComplete(s *dbus.Signal, nameFilter string) bool {
	props := interfaceProperties(s)
	if props == nil {
		log.Printf("%s: skipping signal %+v with no device interface\n", adapter.Name(), s)
		return false
	}
	name, ok := props["Name"].Value().(string)
	if !ok {
		name = "[unknown]"
	}
        if name != nameFilter {
		log.Printf("%s: skipping signal %+v where name(%s) doesn't match nameFilter(%s)\n", adapter.Name(), s, name, nameFilter)
		return false
	} else {
		log.Printf("%s: found signal %+v where name(%s) does match nameFilter(%s)\n", adapter.Name(), s, name, nameFilter)
	   	log.Printf("%s: discovered %s", adapter.Name(), name)
		return true
       }
}

// Discover initiates discovery for a LE peripheral with the given device name.
// It waits for at most the specified timeout, or indefinitely if timeout = 0.
func (conn *Connection) DiscoverByDevice(timeout time.Duration, name string) (Device, error) {
	adapter, err := conn.GetAdapter()
	if err != nil {
		return nil, err
	}
	err = adapter.DiscoverByDevice(timeout, name)
	if err != nil {
		return nil, err
	}
	err = conn.Update()
	if err != nil {
		return nil, err
	}
	return conn.GetDeviceByName(name)
}
