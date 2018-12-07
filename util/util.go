package util

import (
  "fmt"
  "net"
)

func IncIp(ip net.IP) {
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

func PrintPeersMap(peers map[string]map[string]string) {
  fmt.Println("===========Peers==========")
  for physicalIpAddrKey, peer := range peers {
    for key, value := range peer {
      fmt.Println(physicalIpAddrKey, key, value)
    }
  }
  fmt.Println("==========================")
}
