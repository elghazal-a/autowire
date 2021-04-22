package util

import (
  "github.com/geniousphp/autowire/wireguard"
  "net"
)

func IsIPUsed(peers []wireguard.Peer, wgip string) int {
  for index, peer := range peers {
    if peer.IP.Equal(net.ParseIP(wgip)) {
      return index
    }
  }
  return -1
}

func IsEndpointIPExist(peers []wireguard.Peer, wgEndpointIP string) int {
  for index, peer := range peers {
    if peer.EndpointIP.Equal(net.ParseIP(wgEndpointIP)) {
      return index
    }
  }
  return -1
}


func IncIP(ip net.IP) {
  for j := len(ip) - 1; j >= 0; j-- {
    ip[j]++
    if ip[j] > 0 {
      break
    }
  }
}

func SliceContains(a []string, x string) bool {
  for _, n := range a {
    if x == n {
      return true
    }
  }
  return false
}


