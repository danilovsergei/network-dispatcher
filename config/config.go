package config

import (
	"fmt"
	"slices"
)

type Entity struct {
	IncludedMacAddresses []string `json:"Included_MacAddresses,omitempty"`
	ExcludedMacAddresses []string `json:"Excluded_MacAddresses,omitempty"`
	Script               string   `json:"Script"`
	// Supported events: connect, disconnect
	Event        string            `json:"Event"`
	EnvVariables map[string]string `json:"EnvVariables,omitempty"`
}

type Event struct {
	Gateway    string
	MacAddress string
	Event      string
}

// Represents currently connected gateway
type ConnectedGateway struct {
	Gateway    string
	MacAddress string
}

func (e *Entity) HasIncludedMacAddresses() bool {
	return len(e.IncludedMacAddresses) > 0
}

func (e *Entity) HasExcludedMacAddresses() bool {
	return len(e.ExcludedMacAddresses) > 0
}

func (e *Entity) ContainsIncludedMacAddress(address string) bool {
	return slices.Contains(e.IncludedMacAddresses, address)
}

func (e *Entity) ContainsExcludedMacAddress(address string) bool {
	return slices.Contains(e.ExcludedMacAddresses, address)
}

func (cg ConnectedGateway) String() string {
	return fmt.Sprintf("Gateway: %s    MacAddress: %s", cg.Gateway, cg.MacAddress)
}
