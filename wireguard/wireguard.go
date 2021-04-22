package wireguard

import (
  "bytes"
  "fmt"
  "github.com/geniousphp/autowire/ifconfig"
  "io"
  "io/ioutil"
  "log"
  "os"
  "os/exec"
  "strings"
  "strconv"
  "text/template"
)



func wg(stdin io.Reader, arg ...string) ([]byte, error) {
  path, err := exec.LookPath("wg")
  if err != nil {
    return nil, fmt.Errorf("the wireguard (wg) command is not available in your PATH")
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

func wgQuick(stdin io.Reader, arg ...string) ([]byte, error) {
  path, err := exec.LookPath("wg-quick")
  if err != nil {
    return nil, fmt.Errorf("wg-quick command is not available in your PATH")
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

func genkey() ([]byte, error) {
  result, err := wg(nil, "genkey")
  if err != nil {
    return nil, fmt.Errorf("error generating the private key for wireguard: %s", err.Error())
  }
  return result, nil
}

func extractPubKey(privateKey []byte) ([]byte, error) {
  stdin := bytes.NewReader(privateKey)
  result, err := wg(stdin, "pubkey")
  if err != nil {
    return nil, fmt.Errorf("error extracting the public key: %s", err.Error())
  }
  return result, nil
}

func InitWgKeys(wgInterfaceConfigFolder string) (string, string, error) {
  if _, err := os.Stat(wgInterfaceConfigFolder + "/private"); os.IsNotExist(err) {
    err := os.MkdirAll(wgInterfaceConfigFolder, 0700)

    if err != nil {
      return "", "", err
    }

    privKey, err := genkey()
    if err != nil {
      return "", "", err
    }

    err = ioutil.WriteFile(wgInterfaceConfigFolder + "/private", privKey, 0600)
    if err != nil {
      return "", "", err
    }

  }

  privKey, err := ioutil.ReadFile(wgInterfaceConfigFolder + "/private")
  if err != nil {
    return "", "", err
  }

  pubKey, err := extractPubKey(privKey)
  if err != nil {
    return "", "", err
  }

  return strings.TrimSuffix(string(privKey[:]), "\n"), strings.TrimSuffix(string(pubKey[:]), "\n"), nil
}


func ConfigureWireguard(wgConfig Configuration) ([]byte, error) {
  isWGInterfaceStarted, err := ifconfig.IsInterfaceStarted(wgConfig.Interface.Name)
  if err != nil {
    return nil, err
  }

  if isWGInterfaceStarted {
    log.Printf("INFO: WireGuard=%s interface is already up", wgConfig.Interface.Name)
    if isInterfaceAlreadyConfigured(wgConfig.Interface) {
      log.Print("INFO: WireGuard interface is already configured, skipping reconfiguration")
    } else {
      log.Print("INFO: Bringing Down WireGuard Interface because of configuration inconsistencies")
      _, err := wgQuick(nil, "down", wgConfig.Interface.Name)
      if err != nil {
        return nil, fmt.Errorf("error Bringing Down the Wireguard Interface: %s", err.Error())
      }
    }
  }


  configFile, err := os.Create("/etc/wireguard/" + wgConfig.Interface.Name + ".conf")
  if err != nil {
    return nil, err
  }
  t := template.Must(template.New("config").Parse(WgConfigTemplate))
  err = t.Execute(configFile, wgConfig)
  if err != nil {
    return nil, err
  }
  configFile.Chmod(0600)

  isWGInterfaceStarted, err = ifconfig.IsInterfaceStarted(wgConfig.Interface.Name)
  if err != nil {
    return nil, err
  }

  if isWGInterfaceStarted {
    log.Print("INFO: Syncing WireGuard Peers")
    strip, _ := wgQuick(nil, "strip", wgConfig.Interface.Name)
    stdin := bytes.NewReader(strip)
    result, err := wg(stdin, "syncconf", wgConfig.Interface.Name, "/dev/stdin")
    if err != nil {
      return nil, fmt.Errorf("Error Syncing WireGuard Peers: %s", err.Error())
    }
    return result, nil

  } else{
    log.Print("INFO: Bringing Up WireGuard Interface")
    result, err := wgQuick(nil, "up", wgConfig.Interface.Name)
    if err != nil {
      return nil, fmt.Errorf("Error Bringing Up the Wireguard Interface: %s", err.Error())
    }
    return result, nil
  }
}

func isInterfaceAlreadyConfigured(wgInterface Interface) (bool) {
  // Check consistency with .Address and "ip addr show dev wg0"
  actualWGInterfaceAddress, _ := ifconfig.GetIpOfIf(wgInterface.Name)
  if(actualWGInterfaceAddress != wgInterface.Address){
    return false
  }

  // Check consistency with "wg show wg0"
  result, _ := wg(nil, "show", wgInterface.Name, "dump")
  currentWgConfigString := strings.Split(string(result[:]), "\n")[0]
  currentWgConfig := strings.Split(currentWgConfigString, "\t")

  if(currentWgConfig[0] != wgInterface.PrivateKey){
    return false
  }

  if(currentWgConfig[1] != wgInterface.PublicKey){
    return false
  }

  currentWgPort, _ := strconv.Atoi(currentWgConfig[2])
  if(currentWgPort != wgInterface.ListenPort){
    return false
  }

  return true
}


