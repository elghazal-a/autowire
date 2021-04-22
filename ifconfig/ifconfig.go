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
  return "", fmt.Errorf("The network interface %s doesn't exist or doesn't have any IP address", interfaceName)
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



