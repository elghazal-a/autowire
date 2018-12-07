package wg_quick

import (
  "bytes"
  "fmt"
  "io"
  "os/exec"
)


func wg_quick(stdin io.Reader, arg ...string) ([]byte, error) {
  path, err := exec.LookPath("wg-quick")
  if err != nil {
    return nil, fmt.Errorf("the wg-quick command is not available in your PATH")
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


func StartInterface(wgInterfaceName string) ([]byte, error) {
  result, err := wg_quick(nil, "up", wgInterfaceName)
  if err != nil {
    return nil, fmt.Errorf("error bringing up the interface with wg-quick: %s", err.Error())
  }
  return result, nil
}

func StopInterface(wgInterfaceName string) ([]byte, error) {
  result, err := wg_quick(nil, "down", wgInterfaceName)
  if err != nil {
    return nil, fmt.Errorf("error bringing down the interface with wg-quick: %s", err.Error())
  }
  return result, nil
}
