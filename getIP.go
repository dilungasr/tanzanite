package tz

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

// GetIP is for getting the IP address from the client request
func GetIP(r *http.Request) (ip string, err error) {
	// get ip from the X-REAL-IP
	ip = r.Header.Get("X-REAL-IP")
	netIP := net.ParseIP(ip)
	if netIP != nil {
		return ip, nil
	}

	//get ip from the X-FORWARDED-FOR
	ips := r.Header.Get("X-FORWARDED-FOR")
	ipsSlice := strings.Split(ips, ",")

	// parse the ip to check if the ip is assigned
	for _, ip := range ipsSlice {
		netIP = net.ParseIP(ip)

		//  return the ip address if ip is valid
		if netIP != nil {
			return ip, nil
		}
	}

	// get ip address from the RemoteAddr
	ip, _, err = net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}

	netIP = net.ParseIP(ip)
	if netIP != nil {
		return ip, nil
	}

	// if no valid ip found
	return "", fmt.Errorf("No valid IP Address found ")
}
