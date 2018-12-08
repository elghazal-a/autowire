package util

import (
  "fmt"
  "net"
  "strings"
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

func IsTheSameAllowedips(a1 string, a2 string) (bool) {
  // a1=a11,a12,a13
  // a2=a21,a22,a23
  // Split a1 and a2
  // If they have different length => return false
  // Loop over a1
  // Each item of a1 must be in a2
  a1Array := strings.Split(a1, ",")
  a2Array := strings.Split(a2, ",")
  if(len(a1Array) != len(a2Array)) {
    return false
  }

  for _, a1i := range a1Array {
    if(!SliceContains(a2Array, a1i)) {
      return false
    }
  }
  return true
}
