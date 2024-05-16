package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/vishvananda/netlink"
)

// based on the https://gist.github.com/mkol5222/d6bb9660ee369a040ea59370c391322e

// dbus-monitor --system "type='signal',sender='org.freedesktop.NetworkManager',interface='org.freedesktop.NetworkManager'"
const ApplicationName = "network-dispatcher"
const ConfigFileName = "config.json"
const ConnectedGatewayFileName = "connected_gateway.json"
const (
	Connected    string = "connected"
	Disconnected string = "disconnected"
)

// Supported environment variables passed to the dispatched scripts
const DISPATCHER_GATEWAY = "DISPATCHER_GATEWAY"
const DISPATCHER_GATEWAY_MACADDRESS = "DISPATCHER_GATEWAY_MACADDRESS"

// end of supported variables

type configuration struct {
	Entities []Entity
}

type Entity struct {
	MacAddress string `json:"MacAddress,omitempty"`
	Script     string `json:"Script"`
	// Supported events: connect, disconnect
	Event        string            `json:"Event"`
	EnvVariables map[string]string `json:"EnvVariables,omitempty"`
}

type Event struct {
	Gateway    string
	MacAddress string
	Event      string
}

type ExecScriptOut struct {
	ScriptName string
	Err        string
	Out        string
	Combined   string
	ErrOut     string
}

// Represents currently connected gateway
type ConnectedGateway struct {
	Gateway    string
	MacAddress string
}

// True if ActiveEndpoint changes signal received and we need to wait for the assigned gateway
var waitingForGateway bool

func main() {
	config := readConfigurationFile(getConfigFilePath())

	// Mitigate the case when user starts service for the first time
	//On connect event didn't run yet and there is no macaddress/gateway saved yet
	saveNetworkStateOnStartup()

	conn, err := dbus.SystemBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
		os.Exit(1)
	}

	call := conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
		"type='signal',interface='org.freedesktop.DBus.Properties',sender='org.freedesktop.NetworkManager',member='PropertiesChanged'")

	log.Printf("\n The call object is \n %#v \n", call)
	if call.Err != nil {
		log.Printf("\n Dbus connection error: %s \n", call.Err)
	}

	c := make(chan *dbus.Signal, 10)
	conn.Signal(c)
	for signal := range c {
		if signal.Body[0] == "org.freedesktop.NetworkManager.IP4Config" {
			onIp4ConfigChange(config, signal)
		}
		if signal.Body[0] == "org.freedesktop.NetworkManager.IP6Config" {
			onIp4ConfigChange(config, signal)
		}
		// Handle only PropertiesChanged for the wireless connection
		if signal.Body[0] == "org.freedesktop.NetworkManager.Device.Wireless" {
			onWirelessConfigurationChange(config, signal, conn)
		}
	}
}

func saveNetworkStateOnStartup() {
	startupGateway, err := parseDefaultGateway()
	if startupGateway != "" {
		macAddress, err := getGatewayMacAddress(startupGateway)
		if err != nil {
			log.Println(err)
		} else {
			gatewayEntity := ConnectedGateway{Gateway: startupGateway, MacAddress: macAddress}
			saveConnectedGateway(getConnectedGatewayFilePath(), &gatewayEntity)
		}
	} else {
		// Just do nothing when machine is offline and there is no gateway
		// In that case gateway will be updated on first onConnect event
		log.Println(err)
	}
}

// This event is fired on network changes , including receiving Gateway and IP address
func onIp4ConfigChange(config *configuration, signal *dbus.Signal) {
	//  We are interested to execute this event only after Wifi connected event happen
	if !waitingForGateway {
		return
	}
	propMap, isok := signal.Body[1].(map[string]dbus.Variant)
	if !isok {
		return
	}
	addressData, keyExists := propMap["AddressData"]
	if !keyExists {
		return
	}
	addresses, _ := addressData.Value().([]map[string]dbus.Variant)
	if len(addresses) > 0 {
		gatewayVariant, keyExists := propMap["Gateway"]
		// ignoring not relevant events which do not contain gateway
		if !keyExists {
			return
		}
		gateway := strings.ReplaceAll(gatewayVariant.String(), "\"", "")

		macAddress, err := getGatewayMacAddress(gateway)
		log.Printf("Gateway: '%s'\n", gateway)
		log.Printf("Mac address: '%s'\n", macAddress)
		log.Println("Wifi connected")

		// IP addressed not used anywhere yet.
		// addressVariant := addresses[0]["address"]
		// ipAddress := strings.ReplaceAll(addressVariant.String(), "\"", "")

		if err != nil {
			log.Println(err)
			return
		}
		gatewayEntity := ConnectedGateway{Gateway: gateway, MacAddress: macAddress}
		saveConnectedGateway(getConnectedGatewayFilePath(), &gatewayEntity)
		executeEntityScripts(config, Event{Gateway: gateway, MacAddress: gatewayEntity.MacAddress, Event: Connected})
	}
}

// This event is fired on Wifi connect disconnect
func onWirelessConfigurationChange(config *configuration, signal *dbus.Signal, conn *dbus.Conn) {
	propertiesMap, isok := signal.Body[1].(map[string]dbus.Variant)
	if !isok {
		return
	}
	pointPath, ok := propertiesMap["ActiveAccessPoint"]
	// Avoid not relevant keys such as AccessPoints or LastScan objects
	if !ok {
		return
	}

	if isAccessPointConnected(&pointPath, conn) {
		log.Printf("Access point connected. Waiting for gateway\n")
		waitingForGateway = true
	} else {
		gatewayEntity := getConnectedGateway(getConnectedGatewayFilePath())
		if gatewayEntity.MacAddress == "" {
			log.Println("Active gateway is not detected. Please re-connect your network to trigger onConnect event")
			return
		}
		log.Println("Wifi disconnected")
		log.Println("Default gateway: " + gatewayEntity.Gateway)
		executeEntityScripts(config, Event{Gateway: gatewayEntity.Gateway, MacAddress: gatewayEntity.MacAddress, Event: Disconnected})
	}
}

func isAccessPointConnected(accessPointPath *dbus.Variant, conn *dbus.Conn) bool {
	activeobj := conn.Object("org.freedesktop.NetworkManager", accessPointPath.Value().(dbus.ObjectPath))
	_, err := activeobj.GetProperty("org.freedesktop.NetworkManager.AccessPoint.Ssid")

	if err != nil {
		return false
	} else {
		return true
		// fmt.Printf("\nAP Name = %s", name.Value().([]byte))
	}
}

func parseDefaultGateway() (string, error) {
	routes, err := netlink.RouteList(nil, netlink.FAMILY_ALL)
	if err != nil {
		return "", fmt.Errorf("failed to parse default gateway: %v", err)
	}
	for _, route := range routes {
		// equivalent to ip route show default
		if route.Dst == nil || route.Dst.String() == "0.0.0.0/0" || route.Dst.String() == "::/0" {
			if route.Gw.To4() == nil {
				return "", errors.New("failed to get default gateway IP , it's empty")
			}
			return route.Gw.To4().String(), nil
		}
	}
	return "", errors.New("failed to find default gateway in the routes table")
}

func parseGatewayMacAddress(gateway string) (string, error) {
	filterIP := net.ParseIP(gateway) // IP to filter for

	// equivalent to ip neigh show <gateway_ip_address>
	neighbors, err := netlink.NeighList(0, netlink.FAMILY_ALL)
	if err != nil {
		return "", err
	}
	for _, neighbor := range neighbors {
		if neighbor.IP.Equal(filterIP) {
			return neighbor.HardwareAddr.String(), nil
		}
	}
	return "", fmt.Errorf("failed to find mac address for %s", gateway)
}

func getGatewayMacAddress(gateway string) (string, error) {
	retries := 1
	retries_count := 50
	var err error
	var address string
	// Mac address is empty right after gateway ip address received
	// Need to wait for it to appear
	for retries <= retries_count {
		address, err = parseGatewayMacAddress(gateway)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
		retries++
	}
	if address == "" {
		return "", err
	}
	return address, nil
}

func executeScript(command string, envVars map[string]string, args ...string) *ExecScriptOut {
	cmd := exec.Command(command, args...)
	cmd.Env = os.Environ()

	for key, value := range envVars {
		keyvalue := fmt.Sprintf("%s=%s", key, value)
		cmd.Env = append(cmd.Env, keyvalue)
	}

	// Set output to Byte Buffers
	if cmd.Stdout != nil || cmd.Stderr != nil {
		return &ExecScriptOut{
			ScriptName: filepath.Base(command),
			Err:        "Stdout/StdErr already set"}
	}

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err := cmd.Run()
	errString := ""
	if err != nil {
		errString = err.Error()
		// Add more information from error output in case of critical error
		if errb.String() != "" {
			errString = errString + "\n" + errb.String()
		}
	}
	return &ExecScriptOut{
		ScriptName: filepath.Base(command),
		Out:        outb.String(),
		ErrOut:     errb.String(),
		Combined:   outb.String() + "\n" + errb.String(),
		Err:        errString}
}

func readConfigurationFile(jsonPath string) *configuration {
	content, err := os.ReadFile(jsonPath)
	// file does not exist is expected behavior and just use empty configuration
	config := configuration{}
	if err == nil {
		err = json.Unmarshal(content, &config)
		if err != nil {
			log.Fatalf("failed to parse config.json: %v", err)
		}
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("failed to parse config.json: %v", err)
	}
	return &config
}

func saveConnectedGateway(jsonPath string, gateway *ConnectedGateway) {
	content, err := json.MarshalIndent(gateway, "", " ")
	if err != nil {
		log.Println(err)
		return
	}
	err = os.WriteFile(jsonPath, content, 0644)
	if err != nil {
		log.Println(err)
		return
	}
}

func getConnectedGateway(jsonPath string) *ConnectedGateway {
	content, err := os.ReadFile(jsonPath)
	gatewayEntity := ConnectedGateway{}
	if err == nil {
		err = json.Unmarshal(content, &gatewayEntity)
		if err != nil {
			log.Fatal(err)
		}
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatal(err)
	}
	return &gatewayEntity
}

func printJson() {
	config := configuration{}
	entityt1 := Entity{MacAddress: "8c:de:f9:21:6c:e4", Script: "/bon/connect", Event: Connected}
	entityt2 := Entity{MacAddress: "8c:de:f9:21:6c:e4", Script: "/bon/disconnnect", Event: Disconnected}
	config.Entities = []Entity{entityt1, entityt2}
	printable, _ := json.Marshal(config)
	fmt.Println(string(printable))
}

func getConfigFilePath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Join(configDir, ApplicationName, ConfigFileName)
}

func getConnectedGatewayFilePath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Join(configDir, ApplicationName, ConnectedGatewayFileName)
}

func executeEntityScripts(config *configuration, event Event) {
	var entities []Entity
	for _, entity := range config.Entities {
		if strings.ToLower(entity.Event) != event.Event {
			continue
		}
		// empty entity.MacAddress applies script on all networks
		if entity.MacAddress == "" || entity.MacAddress == event.MacAddress {
			entities = append(entities, entity)
		}
	}
	for _, entity := range entities {
		if _, err := os.Stat(entity.Script); err != nil {
			fmt.Printf("Failed to execute %s. Script does not exist\n", entity.Script)
			continue
		}
		envVars := make(map[string]string)
		envVars[DISPATCHER_GATEWAY] = event.Gateway
		envVars[DISPATCHER_GATEWAY_MACADDRESS] = event.MacAddress

		for key, value := range entity.EnvVariables {
			envVars[key] = value
		}

		log.Println("Execute dispatch script " + entity.Script)
		execOut := executeScript(entity.Script, envVars)
		if execOut.Err != "" {
			log.Printf("Failed to execute %s", execOut.ScriptName)
			logMultilineScriptOutput(execOut.Err, execOut.ScriptName)
			continue
		}
		logMultilineScriptOutput(execOut.Combined, execOut.ScriptName)
	}
}

func logMultilineScriptOutput(out string, script string) {
	if out != "" {
		for _, line := range strings.Split(out, "\n") {
			if line == "" {
				continue
			}
			log.Printf("[%s] %s\n", script, line)
		}
	}
}
