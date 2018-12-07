package wireguard

import (
  "bytes"
  "fmt"
  "io"
  "strings"
  "strconv"
  "os"
  "os/exec"
  "text/template"
  "github.com/geniousphp/autowire/ifconfig"
)

type Interface struct {
  Name string
  Address string
  ListenPort int
  PrivateKey string
}

type Peer struct {
  PublicKey  string
  Ip         string
  AllowedIPs string
  Endpoint   string
  Port       int
}

type WGConfig struct {
  Interface Interface
  Peers     map[string]Peer
}

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

func Genkey() ([]byte, error) {
  result, err := wg(nil, "genkey")
  if err != nil {
    return nil, fmt.Errorf("error generating the private key for wireguard: %s", err.Error())
  }
  return result, nil
}

func ExtractPubKey(privateKey []byte) ([]byte, error) {
  stdin := bytes.NewReader(privateKey)
  result, err := wg(stdin, "pubkey")
  if err != nil {
    return nil, fmt.Errorf("error extracting the public key: %s", err.Error())
  }
  return result, nil
}


func ConfigureInterface(wgConfig WGConfig) (error) {
  configFile, err := os.Create("/etc/wireguard/" + wgConfig.Interface.Name + ".conf")
  if err != nil {
    return err
  }

  t := template.Must(template.New("config").Parse(wgConfigTemplate))

  err = t.Execute(configFile, wgConfig.Interface)
  if err != nil {
    return err
  }

  configFile.Chmod(0600)
  return nil
}

func IsWgInterfaceWellConfigured(wgConfig WGConfig) (bool) {
  // Check consistency with ip addr show dev wg0 (IP Address)
  actualIpAddr, _ := ifconfig.GetIpOfIf(wgConfig.Interface.Name)
  if(actualIpAddr != wgConfig.Interface.Address){
    return false
  }

  // Check consistency with wg show wg0 (Port and Private Key)
  result, _ := wg(nil, "show", wgConfig.Interface.Name, "dump")
  currentWgConfigString := strings.Split(string(result[:]), "\n")[0]
  // fmt.Println(currentWgConfigString)
  currentWgConfig := strings.Split(currentWgConfigString, "\t")

  if(currentWgConfig[0] != wgConfig.Interface.PrivateKey){
    return false
  }

  currentWgPort, _ := strconv.Atoi(currentWgConfig[2])
  if(currentWgPort != wgConfig.Interface.ListenPort){
    return false
  }

  return true
}

func GetPeers(wgInterfaceName string) (map[string]map[string]string, error) {
  result, err := wg(nil, "show", wgInterfaceName, "dump")
  if err != nil {
    return nil, fmt.Errorf("error getting peers list for wireguard: %s", err.Error())
  }

  peers := make(map[string]map[string]string)
  wgPeersString := strings.Split(string(result[:]), "\n")
  for i, wgPeerString := range wgPeersString {
    if(i == 0) {
      //Skip the fist line config which is interfacz config
      continue
    }
    if(wgPeerString == "") {
      //empty line, skip it
      continue
    }
    currentWgConfig := strings.Split(wgPeerString, "\t")
    physicalIpAddr := strings.Split(currentWgConfig[2], ":")[0]
    peers[physicalIpAddr] = make(map[string]string)
    peers[physicalIpAddr]["endpoint"] = physicalIpAddr
    peers[physicalIpAddr]["port"] = strings.Split(currentWgConfig[2], ":")[1]
    peers[physicalIpAddr]["pubkey"] = currentWgConfig[0]
    peers[physicalIpAddr]["allowedips"] = currentWgConfig[3]
  }

  return peers, nil
}

func ConfigurePeer(wgInterfaceName string, peer map[string]string) ([]byte, error) {
  result, err := wg(nil, "set", wgInterfaceName, "peer", peer["pubkey"], "endpoint", peer["endpoint"] + ":" + peer["port"], "allowed-ips", peer["allowedips"])
  if err != nil {
    return nil, fmt.Errorf("error configuring wg peer: %s", err.Error())
  }
  return result, nil
}

func RemovePeer(wgInterfaceName string, pubKey string) ([]byte, error) {
  result, err := wg(nil, "set", wgInterfaceName, "peer", pubKey, "remove")
  if err != nil {
    return nil, fmt.Errorf("error removing wg peer: %s", err.Error())
  }
  return result, nil
}


