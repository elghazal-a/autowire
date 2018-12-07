package ifconfig

import (
  "fmt"
  "net"

)



func GetIpOfIf(interfaceName string) (string, error) {
  ifaces, err := net.Interfaces()
  if err != nil {
    return "", err
  }

  if len(ifaces) <= 0 {
    return "", fmt.Errorf("No network interface found")
  }

  for _, i := range ifaces {
    if i.Name == interfaceName {
      addrs, err := i.Addrs()
      if err != nil {
        return "", err
      }
      if len(addrs) > 0 {
        return addrs[0].String(), err
      }
    }
  }
  return "", fmt.Errorf("The network interface %s not found or doesn't have any ip address", interfaceName)
}

func GetFirstIpOfFirtIf() (string, error) {
  ifaces, err := net.Interfaces()
  if err != nil {
    return "", err
  }

  if len(ifaces) <= 0 {
    return "", fmt.Errorf("No network interface found")
  }

  for _, iface := range ifaces {
    if(iface.Name == "lo") {
      continue
    }
    addrs, err := iface.Addrs()
    if err != nil {
      return "", err
    }
    if len(addrs) > 0 {
      return addrs[0].String(), err
    }
    
  }

  return "", fmt.Errorf("No network interface or ip address found")
}


func IsInterfaceStarted(interfaceName string) (bool, error){
  ifaces, err := net.Interfaces()
  if err != nil {
    return false, err
  }

  if len(ifaces) <= 0 {
    return false, nil
  }

  for _, i := range ifaces {
    if i.Name == interfaceName {
      return true, nil
    }
  }
  return false, nil
}

