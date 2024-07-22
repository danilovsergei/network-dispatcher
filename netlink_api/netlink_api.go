package netlink_api

import (
	"errors"
	"fmt"
	"log"
	"net"
	"network-dispatcher/config"
	"time"

	"github.com/vishvananda/netlink"
)

// Gets gateway through netlink
func GetGatewayFromTheSystem() *config.ConnectedGateway {
	gateway, err := ParseDefaultGateway()
	if err != nil {
		log.Printf("Error while  parseDefaultGateway using netlink : %v\n", err)
		return nil
	}
	if gateway == "" {
		log.Println("Netlink returned empty gateway. Will wait for dbus update")
		return nil
	}
	macAddress, err := GetGatewayMacAddress(gateway)
	if err != nil {
		log.Printf("Failed to parse gw %s mac address using netlink : %v\n", gateway, err)
		return nil
	}
	if macAddress == "" {
		log.Printf("Netlink returned empty macaddress for %s gateway. Will wait for dbus update\n", gateway)
		return nil
	}
	gatewayEntity := config.ConnectedGateway{Gateway: gateway, MacAddress: macAddress}
	log.Printf("Default gateway received early through netlink: %s\n", gatewayEntity)
	return &gatewayEntity
}

// parseDefaultGateway gets gateway using go binding for ip route show command
func ParseDefaultGateway() (string, error) {
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return "", fmt.Errorf("failed to parse default IPv4 gateway: %v", err)
	}
	// Get Ipv4 gateway
	for _, route := range routes {
		// equivalent to ip route show default
		if route.Dst == nil || route.Dst.String() == "0.0.0.0/0" || route.Dst.String() == "::/0" {
			return route.Gw.To4().String(), nil
		}
	}
	// Get Ipv6 gateway if noIpv4 were found
	ipv6Routes, err := netlink.RouteList(nil, netlink.FAMILY_V6)
	if err != nil {
		return "", fmt.Errorf("failed to parse default IPv6 gateway: %v", err)
	} else {
		for _, route := range ipv6Routes {
			if route.Gw != nil {
				return route.Gw.String(), nil
			}
		}
	}
	return "", errors.New("failed to find default gateway in the routes table")
}

func GetGatewayMacAddress(gateway string) (string, error) {
	getMacAddress := func() (string, error) {
		filterIP := net.ParseIP(gateway) // IP to filter for

		// equivalent to ip neigh show <gateway_ip_address>
		neighbors, err := netlink.NeighList(0, netlink.FAMILY_ALL)
		if err != nil {
			return "", fmt.Errorf("failed to receive macaddress for gateway %s using netlink: %v", gateway, err)
		}
		for _, neighbor := range neighbors {
			if neighbor.IP.Equal(filterIP) {
				return neighbor.HardwareAddr.String(), nil
			}
		}
		return "", fmt.Errorf("failed to receive macaddress for gateway %s using netlink", gateway)
	}
	return GetGatewayMacAddressWithRetries(getMacAddress)
}

func GetGatewayMacAddressWithRetries(macaddressFunc func() (string, error)) (string, error) {
	retries := 1
	// it's 5 seconds
	retries_count := 50
	var err error
	var address string

	for retries <= retries_count {
		address, err = macaddressFunc()
		if err != nil {
			continue
		}
		if address != "" {
			fmt.Printf("Received macaddress %s from %d attempt\n", address, retries)
			return address, nil
		}
		time.Sleep(100 * time.Millisecond)
		retries++
	}
	return "", fmt.Errorf("%v in %d attempts: %v", err, retries_count, err)
}
