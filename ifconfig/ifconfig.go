package ifconfig

import (
  // "fmt"
  "net"

)



func GetIpOfIf(interfaceName string) (string, error) {
  ifaces, err := net.Interfaces()
  if err != nil {
    return "", err
  }

  if len(ifaces) <= 0 {
    return "", nil
  }

  // TODO: If interface doesn't exist or exist and no ip found return error

  for _, i := range ifaces {
    // fmt.Println(i.Name)
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
  return "", nil
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

