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
const NM_DEVICE_TYPE_WIFI = 2

var conn *dbus.Conn

type NetworkAdapter struct {
	object dbus.BusObject
}

type Ip4Config struct {
	object dbus.BusObject
	Path   dbus.ObjectPath
}

type Ip6Config struct {
	object dbus.BusObject
	Path   dbus.ObjectPath
}

func Connect() error {
	if conn != nil {
		return nil
	}
	var err error
	conn, err = dbus.SystemBus()
	return err
}

func MonitorNetworkCardStateChanged(onConnected func(*dbus.Signal), onDisconnected func(*dbus.Signal)) {
	if conn == nil {
		if err := Connect(); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
			os.Exit(1)
		}
	}

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

func GetDeviceByInterfaceName(name string) (*NetworkAdapter, error) {
	var path dbus.ObjectPath
	err := conn.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager").Call("org.freedesktop.NetworkManager.GetDeviceByIpIface", 0, name).Store(&path)
	if err != nil {
		return nil, err
	}
	return NewNetworkAdapter(path), nil
}

func (n *NetworkAdapter) Ip4Config() (*Ip4Config, error) {
	var path dbus.ObjectPath
	err := n.object.Call("org.freedesktop.DBus.Properties.Get", 0,
		"org.freedesktop.NetworkManager.Device", "Ip4Config").Store(&path)

	if err != nil {
		return nil, err
	}
	return newIp4Config(path), nil
}

func (n *NetworkAdapter) Ip6Config() (*Ip6Config, error) {
	var path dbus.ObjectPath
	err := n.object.Call("org.freedesktop.DBus.Properties.Get", 0,
		"org.freedesktop.NetworkManager.Device", "Ip6Config").Store(&path)

	if err != nil {
		return nil, err
	}
	return newIp6Config(path), nil
}

func (n *NetworkAdapter) GetState() (uint32, error) {
	var state uint32
	err := n.object.Call("org.freedesktop.DBus.Properties.Get", 0,
		"org.freedesktop.NetworkManager.Device", "State").Store(&state)
	return state, err
}

func (n *NetworkAdapter) GetDeviceType() (uint32, error) {
	var deviceType uint32
	err := n.object.Call("org.freedesktop.DBus.Properties.Get", 0,
		"org.freedesktop.NetworkManager.Device", "DeviceType").Store(&deviceType)
	return deviceType, err
}

func (n *NetworkAdapter) GetInterfaceName() (string, error) {
	var name string
	err := n.object.Call("org.freedesktop.DBus.Properties.Get", 0,
		"org.freedesktop.NetworkManager.Device", "Interface").Store(&name)
	return name, err
}

func newIp4Config(path dbus.ObjectPath) *Ip4Config {
	return &Ip4Config{object: conn.Object("org.freedesktop.NetworkManager", path), Path: path}
}

func newIp6Config(path dbus.ObjectPath) *Ip6Config {
	return &Ip6Config{object: conn.Object("org.freedesktop.NetworkManager", path), Path: path}
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
