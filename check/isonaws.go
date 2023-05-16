package check

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/jreisinger/checkip"
)

type awsIpRanges struct {
	IsOn bool
	Info struct {
		IpPrefix           string   `json:"ip_prefix"`
		Region             string   `json:"region"`
		Services           []string `json:"services"`
		NetworkBorderGroup string   `json:"network_border_group"`
	}
}

// Json implements checkip.Info
func (a awsIpRanges) Json() ([]byte, error) {
	return json.Marshal(&a)
}

// Summary implements checkip.Info
func (a awsIpRanges) Summary() string {
	if a.IsOn {
		return fmt.Sprintf("%t, prefix: %s, region: %s, sevices: %v",
			a.IsOn, a.Info.IpPrefix, a.Info.Region, a.Info.Services)
	}
	return fmt.Sprintf("%t", a.IsOn)
}

// IsOnAWS checks if ipaddr belongs to AWS. If so it provides info about the IP
// address. It gets the info from https://ip-ranges.amazonaws.com/ip-ranges.json
func IsOnAWS(ipaddr net.IP) (checkip.Result, error) {
	result := checkip.Result{
		Name: "is on AWS",
	}
	resp := struct {
		Prefixes []struct {
			IpPrefix           string `json:"ip_prefix"`
			Region             string `json:"region"`
			Service            string `json:"service"`
			NetworkBorderGroup string `json:"network_border_group"`
		} `json:"prefixes"`
	}{}
	apiUrl := "https://ip-ranges.amazonaws.com/ip-ranges.json"
	if err := defaultHttpClient.GetJson(apiUrl, map[string]string{}, map[string]string{}, &resp); err != nil {
		return result, newCheckError(err)
	}
	var a awsIpRanges
	for _, prefix := range resp.Prefixes {
		_, network, err := net.ParseCIDR(prefix.IpPrefix)
		if err != nil {
			return result, fmt.Errorf("parse CIDR %q: %v", prefix.IpPrefix, err)
		}
		if network.Contains(ipaddr) {
			a.IsOn = true
			a.Info.IpPrefix = prefix.IpPrefix
			a.Info.NetworkBorderGroup = prefix.NetworkBorderGroup
			a.Info.Region = prefix.Region
			a.Info.Services = append(a.Info.Services, prefix.Service)
		}

	}
	result.Info = a
	return result, nil
}
