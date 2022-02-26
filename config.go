package main

import (
  "flag"
  "fmt"
  "github.com/geniousphp/autowire/ifconfig"
  "log"
  "net"
  "os"
  "strings"
)


type Config struct {
  AW_Version bool
  AW_LogLevel string
  AW_Backend string
  CS_Address string
  CS_KVPrefix string
  CS_FullKVPrefix string
  WG_ConfigFolder string
  WG_Relay bool
  WG_InterfaceConfigFolder string
  WG_InterfaceName string
  WG_Range string
  WG_IP string
  WG_Port int
  WG_EndpointInterfaceName string
  WG_EndpointIP string
  WG_AllowedIPs string
  WG_PostUp string
  WG_PostDown string
  GC_Enable bool
}

var config Config

func init() {
  flag.BoolVar(&config.AW_Version, "version", false, "Show version")
  flag.StringVar(&config.AW_LogLevel, "log-level", "INFO", "Log Level")
  flag.StringVar(&config.AW_Backend, "backend", "consul", "Storage backend. Available backends: [consul]")
  flag.StringVar(&config.CS_Address, "cs-ip", "127.0.0.1:8500", "Consul IP")
  flag.StringVar(&config.CS_KVPrefix, "cs-kv-prefix", "autowire", "Prefix in KV store where configurations will be stored")
  flag.StringVar(&config.WG_ConfigFolder, "wg-config-folder", "/etc/wireguard", "Wireguard config folder")
  flag.BoolVar(&config.WG_Relay, "wg-relay", false, "Will the server act as relay? Must set the WG_Range in this case")
  flag.StringVar(&config.WG_InterfaceName, "wg-if-name", "wg0", "Wireguard interface name")
  flag.StringVar(&config.WG_Range, "wg-range", "192.168.10.0/24", "Wireguard CIDR, ignored when WG_IP is set")
  flag.StringVar(&config.WG_IP, "wg-ip", "", "Wireguard IP address, optional. When not set, Autowire will pick unused IP from WG_Range")
  flag.IntVar(&config.WG_Port, "wg-port", 51820, "Wireguard Port")
  flag.StringVar(&config.WG_EndpointInterfaceName, "wg-endpoint-if-name", "", "The network interface name whose ip address will be used for wireguard endpoint. Ignored when wg-endpoint-ip is specified")
  flag.StringVar(&config.WG_EndpointIP, "wg-endpoint-ip", "", "The IP address will be used for wireguard endpoint. Will be fetched from wg-endpoint-if-name if not specified")
  flag.StringVar(&config.WG_AllowedIPs, "wg-allowed-ips", "", "Wireguard Allowed IPs (comma separated cidr)")
  flag.StringVar(&config.WG_PostUp, "wg-post-up", "", "steps to be run after the wireguard interface is up")
  flag.StringVar(&config.WG_PostDown, "wg-post-down", "", "steps to be run after the wireguard interface is down")
  flag.BoolVar(&config.GC_Enable, "gc-enable", false, "Enable peers garbage collection. Consul is the only node discovery supported. When disabled, you have to manually remove the left peers from the k/s")
}


func initConfig() error {

  if(config.AW_Version){
    fmt.Println("Autowire v0.2.2")
    os.Exit(0)
  }

  config.CS_FullKVPrefix = config.CS_KVPrefix + "/" + config.WG_InterfaceName
  config.WG_InterfaceConfigFolder = config.WG_ConfigFolder + "/" + config.WG_InterfaceName + "/"


  if config.WG_IP == "" && config.WG_Range == "" {
    return fmt.Errorf("-wg-range and -wg-ip are not set. Need to set at least one of them")
  }

  if config.WG_Relay && config.WG_Range == "" {
    return fmt.Errorf("You must set the -wg-range in relay mode")
  }

  if config.WG_EndpointIP == "" {
    log.Print("INFO: -wg-endpoint-ip is not set, will try to fetch the endpoint IP from -wg-endpoint-if-name")
    if config.WG_EndpointInterfaceName != "" {
      log.Print("INFO: Autowire will try to fetch the Endpoint IP from -wg-endpoint-if-name=", config.WG_EndpointInterfaceName)

      //get the interface IP related to WG_EndpointInterfaceName
      inet, err := ifconfig.GetIpOfIf(config.WG_EndpointInterfaceName)
      if err != nil {
        return err
      }

      ipAddr, _, err := net.ParseCIDR(inet)
      if err != nil {
        return err
      }
      config.WG_EndpointIP = ipAddr.String()
      log.Printf("INFO: Autowire will use %s as Endpoint IP", config.WG_EndpointIP)
    } else{
      //client behind a NAT are not yet supported
      return fmt.Errorf("You must set either -wg-endpoint-if-name or -wg-endpoint-ip")
    }
  }

  if config.WG_IP == "" {
    isWGInterfaceStarted, err := ifconfig.IsInterfaceStarted(config.WG_InterfaceName)
    if err != nil {
      return err
    }
    if isWGInterfaceStarted {
      actualWGInterfaceAddress, _ := ifconfig.GetIpOfIf(config.WG_InterfaceName)
      config.WG_IP = strings.Split(actualWGInterfaceAddress, "/")[0]
    }
  }



  return nil
}