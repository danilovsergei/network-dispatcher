package dbusapi

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/godbus/dbus/v5"
)

const NM_DEVICE_STATE_ACTIVATED = 100
const NM_DEVICE_STATE_DISCONNECTED = 30

var conn *dbus.Conn

type NetworkAdapter struct {
	object dbus.BusObject
}

type Ip4Config struct {
	object dbus.BusObject
}

type Ip6Config struct {
	object dbus.BusObject
}

func MonitorNetworkCardStateChanged(onConnected func(*dbus.Signal), onDisconnected func(*dbus.Signal)) {
	dbusconn, err := dbus.SystemBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
		os.Exit(1)
	}
	conn = dbusconn

	call := conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
		"interface='org.freedesktop.NetworkManager.Device'")
	if call.Err != nil {
		log.Printf("\n Dbus connection error: %s \n", call.Err)
	}

	c := make(chan *dbus.Signal, 10)
	conn.Signal(c)
	for signal := range c {
		if signal.Name == "org.freedesktop.NetworkManager.Device.StateChanged" {
			if len(signal.Body) != 3 {
				log.Printf("Incorrect signal body. Expected 3 arguments , got %+v\n", signal.Body)
				continue
			}
			state := signal.Body[0].(uint32)
			if state == NM_DEVICE_STATE_ACTIVATED {
				go onConnected(signal)
			}
			if state == NM_DEVICE_STATE_DISCONNECTED {
				go onDisconnected(signal)
			}
		}
	}
}

func NewNetworkAdapter(path dbus.ObjectPath) *NetworkAdapter {
	return &NetworkAdapter{object: conn.Object("org.freedesktop.NetworkManager", path)}
}

func (n *NetworkAdapter) Ip4Config() (*Ip4Config, error) {
	ip4Config, err := n.object.GetProperty("org.freedesktop.NetworkManager.Device.Ip4Config")

	if err != nil {
		return nil, err
	}
	return newIp4Config(ip4Config.Value().(dbus.ObjectPath)), nil
}

func (n *NetworkAdapter) Ip6Config() (*Ip6Config, error) {
	ip6Config, err := n.object.GetProperty("org.freedesktop.NetworkManager.Device.Ip6Config")

	if err != nil {
		return nil, err
	}
	return newIp6Config(ip6Config.Value().(dbus.ObjectPath)), nil
}

func newIp4Config(path dbus.ObjectPath) *Ip4Config {
	return &Ip4Config{object: conn.Object("org.freedesktop.NetworkManager", path)}
}

func newIp6Config(path dbus.ObjectPath) *Ip6Config {
	return &Ip6Config{object: conn.Object("org.freedesktop.NetworkManager", path)}
}

func (c *Ip4Config) Gateway() (string, error) {
	gateway, err := c.object.GetProperty("org.freedesktop.NetworkManager.IP4Config.Gateway")
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(gateway.String(), "\"", ""), nil
}

func (c *Ip6Config) Gateway() (string, error) {
	gateway, err := c.object.GetProperty("org.freedesktop.NetworkManager.IP6Config.Gateway")
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(gateway.String(), "\"", ""), nil
}
