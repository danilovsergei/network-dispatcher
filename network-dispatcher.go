package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"network-dispatcher/config"
	dbusapi "network-dispatcher/dbus_api"
	"network-dispatcher/netlink_api"
	"network-dispatcher/shell"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
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
	Entities []config.Entity
}

var configFilePath string

func main() {
	flag.StringVar(&configFilePath, "config", getConfigFilePath(), "Path to the configuration file")
	flag.Parse()

	// Make sure there are no leftovers of of the saved old gateway
	deleteGatewayFilePathIfPresent()

	// Mitigate the case when user starts service for the first time
	//On connect event didn't run yet and there is no macaddress/gateway saved yet
	saveNetworkStateOnStartup()

	dbusapi.MonitorNetworkCardStateChanged(
		onConnected,
		onDisconnected)
}

func onConnected(signal *dbus.Signal) {
	fmt.Println("Dbus network connected event")
	gateway, err := getGatewayFromDbus(signal)
	if err != nil {
		log.Printf("Failed to receive gateway on wifi connected: %s\n", gateway)
		return
	}
	gatewayEntity, err := getGatewayEntity(gateway)
	if err != nil {
		log.Printf("Failed to create gateway entity: %v\n", err)
		return
	}
	log.Println(gatewayEntity)
	log.Println("Wifi connected")

	saveLastConnectedGatewayToConfig(getConnectedGatewayFilePath(), gatewayEntity)
	executeEntityScripts(config.Event{Gateway: gatewayEntity.Gateway, MacAddress: gatewayEntity.MacAddress, Event: Connected})

}

func getGatewayFromDbus(signal *dbus.Signal) (string, error) {
	getGateway := func() (string, error) {
		netCard := dbusapi.NewNetworkAdapter(signal.Path)
		//ip4
		ip4, err := netCard.Ip4Config()
		if err != nil {
			return "", fmt.Errorf("failed to get Ip4Config from dbus: %v", err)
		}
		gateway, err := ip4.Gateway()
		if err != nil {
			return "", fmt.Errorf("failed to get Ip4 gateway from dbus: %v", err)
		}

		if gateway != "" {
			return gateway, nil
		}
		//ip6
		ip6, err := netCard.Ip6Config()
		if err != nil {
			return "", fmt.Errorf("failed to get Ip6Config from dbus: %v", err)
		}
		gateway, err = ip6.Gateway()
		if err != nil {
			return "", fmt.Errorf("failed to get Ip6 gateway from dbus: %v", err)
		}
		if gateway != "" {
			return gateway, nil
		}
		return "", errors.New("both IP4 and IP6 gateways are empty")
	}
	return getGatewayFromDbusWithRetries(getGateway)
}

func getGatewayFromDbusWithRetries(gatewayFunc func() (string, error)) (string, error) {
	retries := 1
	// it's 5 seconds
	retries_count := 50
	var err error
	var gateway string

	for retries <= retries_count {
		gateway, err = gatewayFunc()
		if err != nil {
			continue
		}
		if gateway != "" {
			fmt.Printf("Received dbus gateway %s from %d attempt\n", gateway, retries)
			return gateway, nil
		}
		time.Sleep(100 * time.Millisecond)
		retries++
	}
	return "", fmt.Errorf("%v in %d attempts: %v", err, retries_count, err)
}

func onDisconnected(signal *dbus.Signal) {
	fmt.Println("Dbus network disconnected event")
	gatewayEntity := getLastConnectedGatewayFromConfig(getConnectedGatewayFilePath())
	log.Println("Wifi disconnected")
	if gatewayEntity.MacAddress == "" {
		log.Printf("Last active gateway macaddress is not detected %v. Gateway specific disconnect events will not run\n", gatewayEntity)
	} else {
		log.Println(gatewayEntity)
		executeEntityScripts(config.Event{Gateway: gatewayEntity.Gateway, MacAddress: gatewayEntity.MacAddress, Event: Disconnected})
	}
	// cleanup gateway config file to avoid stale gateway information
	deleteGatewayFilePathIfPresent()
}

func deleteGatewayFilePathIfPresent() {
	gwPath := getConnectedGatewayFilePath()
	if _, err := os.Stat(gwPath); err == nil {
		err := os.Remove(gwPath)
		if err != nil {
			fmt.Printf("Error deleting gateway config file '%s': %v\n", gwPath, err)
		} else {
			log.Println("Gateway config file deleted")
		}
	} else if !os.IsNotExist(err) {
		fmt.Printf("Error checking file stat'%s': %v\n", gwPath, err)
	}
}

func saveNetworkStateOnStartup() {
	startupGateway, err := netlink_api.ParseDefaultGateway()
	if err != nil {
		fmt.Printf("Failed to receive gateway on startup. Gateway dependant scripts will not run: %v\n", err)
		return
	}
	if startupGateway == "" {
		fmt.Println("Failed to receive gateway on startup. Gateway dependant scripts will not run")
		return
	}

	macAddress, err := netlink_api.GetGatewayMacAddress(startupGateway)
	if err != nil {
		log.Printf("Failed to receive gateway %s macaddress on startup. Gateway dependant scripts will not run: %v\n", startupGateway, err)
		return
	}
	if macAddress == "" {
		fmt.Printf("Failed to receive gateway %s macaddress on startup. Gateway dependant scripts will not run\n", startupGateway)
		return
	}
	gatewayEntity := config.ConnectedGateway{Gateway: startupGateway, MacAddress: macAddress}
	fmt.Printf("Found gateway on startup: %s\n", gatewayEntity)
	saveLastConnectedGatewayToConfig(getConnectedGatewayFilePath(), &gatewayEntity)
}

// Parses gateway from dbus event and fetches macaddress for it using netlink
func getGatewayEntity(gateway string) (*config.ConnectedGateway, error) {
	if gateway == "" {
		return nil, errors.New("failed to parse gateway from dbus: it's empty")
	}
	macAddress, err := netlink_api.GetGatewayMacAddress(gateway)
	if err != nil {
		return nil, fmt.Errorf("failed to parse macaddress for gateway %s from dbus: %v", gateway, err)
	}
	if macAddress == "" {
		return nil, fmt.Errorf("failed to parse macaddress for gateway %s from dbus: it's empty", gateway)
	}
	return &config.ConnectedGateway{Gateway: gateway, MacAddress: macAddress}, nil
}

func readConfigurationFile(jsonPath string) (*configuration, error) {
	fmt.Printf("Read config file from %s\n", jsonPath)
	content, err := os.ReadFile(jsonPath)
	// file does not exist is expected behavior and just use empty configuration
	config := configuration{}
	if err == nil {
		err = json.Unmarshal(content, &config)
		if err != nil {
			log.Fatalf("failed to parse %s: %v", jsonPath, err)
		}
	} else if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("failed to parse config.json: %v", err)
		} else {
			return nil, fmt.Errorf("config file not found at %s", jsonPath)
		}
	}
	return &config, nil
}

func saveLastConnectedGatewayToConfig(jsonPath string, gateway *config.ConnectedGateway) {
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

func getLastConnectedGatewayFromConfig(jsonPath string) *config.ConnectedGateway {
	content, err := os.ReadFile(jsonPath)
	gatewayEntity := config.ConnectedGateway{}
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

func executeEntityScripts(event config.Event) {
	var entities []config.Entity
	config, err := readConfigurationFile(configFilePath)
	if err != nil {
		log.Println(err)
		log.Println("Script execution will be skipped")
		return
	}

	for _, entity := range config.Entities {
		if strings.ToLower(entity.Event) != event.Event {
			continue
		}
		// Skip excluded mac addresses
		if entity.ContainsExcludedMacAddress(event.MacAddress) {
			continue
		}

		// empty entity.MacAddress applies script on all networks
		if !entity.HasIncludedMacAddresses() || entity.ContainsIncludedMacAddress(event.MacAddress) {
			entities = append(entities, entity)
		}
	}
	for _, entity := range entities {
		script := os.ExpandEnv(entity.Script)
		if _, err := os.Stat(script); err != nil {
			fmt.Printf("Failed to execute %s. Script does not exist\n", script)
			continue
		}
		envVars := make(map[string]string)
		envVars[DISPATCHER_GATEWAY] = event.Gateway
		envVars[DISPATCHER_GATEWAY_MACADDRESS] = event.MacAddress

		for key, value := range entity.EnvVariables {
			// allow to have variables like $HOME in EnvVariables values.
			envVars[key] = os.ExpandEnv(value)
		}
		execOut := shell.ExecuteScript(script, envVars)
		if execOut.Err != "" {
			log.Printf("Failed to execute %s", execOut.ScriptName)
			logMultilineScriptOutput(
				execOut.Err+"\n"+execOut.ErrOut,
				execOut.ScriptName)
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
