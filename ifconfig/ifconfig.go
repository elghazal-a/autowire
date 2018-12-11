package ifconfig

import (
  "fmt"
  "net"
  "io"
  "os/exec"
  "bytes"
)


func ip(stdin io.Reader, arg ...string) ([]byte, error) {
  path, err := exec.LookPath("ip")
  if err != nil {
    return nil, fmt.Errorf("ip command is not available in your PATH")
  }

  cmd := exec.Command(path, arg...)

  cmd.Stdin = stdin
  var buf bytes.Buffer
  cmd.Stderr = &buf
  output, err := cmd.Output()

  if err != nil {
    return nil, fmt.Errorf("%s - %s", err.Error(), buf.String())
  }
  return output, nil

}

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


func StartWGInterface(wgInterfaceName string, wgInterfaceAddr string) ([]byte, error) {
  result, err := ip(nil, "link", "add", "dev", wgInterfaceName, "type", "wireguard")
  if err != nil {
    return nil, fmt.Errorf("error adding wireguard interface: %s", err.Error())
  }
  result, err = ip(nil, "address", "add", "dev", wgInterfaceName, wgInterfaceAddr)
  if err != nil {
    return nil, fmt.Errorf("error adding ip address to wireguard interface: %s", err.Error())
  }
  result, err = ip(nil, "link", "set", "up", "dev", wgInterfaceName)
  if err != nil {
    return nil, fmt.Errorf("error bringing up wireguard interface: %s", err.Error())
  }
  return result, nil
}

func StopWGInterface(wgInterfaceName string) ([]byte, error) {
  result, err := ip(nil, "link", "delete", "dev", wgInterfaceName)
  if err != nil {
    return nil, fmt.Errorf("error bringing down wireguard interface: %s", err.Error())
  }
  return result, nil
}


