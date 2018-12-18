This project is at an early stage development and is not production ready even though we're running it in our production. Run it at your own risk.

# Autowire

**Run WireGuard VPN with zero configuration.**

This project provides a convenient way to automatically configure WireGuard. If you're running a Consul cluster and willing to configure WireGuard as VPN solution, you'll find this project very helpful.

`Autowire` picks an IP Address from the available pool of addresses, configures the local WireGuard interface as well as the peers and starts it. `Autowire` leverages distributed locking of Consul to ensure that picked IP address is not used by any other WireGuard peer. This method is described in the [leader election](https://www.consul.io/docs/guides/leader-election.html) guide.

`Autowire` also takes advantage of Consul blocking queries to watch nodes and KV, this allows Autowire to automatically reconfigure WireGuard Peers when nodes join or leave the Consul cluster.

`Autowire` uses Consul KV to store WireGuard interface and Peers configurations. This makes WireGuard config distributed and available to all nodes of the cluster.


## Installation

Autowire doesn't install WireGuard. It's expected to be installed and available in the `$PATH`
Autowire is meant to be installed on every node of the cluster where WireGuard is need to be configured. It's better to schedule it as system daemon in all cluster nodes.

1. Download a pre-compiled release from the release page.
1. Extract the binary.
1. Run it with `./autowire`.

## Configuration

Example usage:

* if-name: Network interface whose IP Address will be used for WireGuard endpoints
* wg-range: IP Address range. Autowire will pick address within this range
* wg-config-folder: Folder where WireGuard configurations will be stored
* wg-port: WireGuard Port

Find out the updated list of configurations in `config.go`

```
autowire --if-name enp0s2 --wg-range 192.168.10.0/24 --wg-config-folder /etc/wireguard --wg-port 51820
```

## ToDo

* Code Refactoring and cleaning and enhance logging
* Write automated tests
* Support more backends (etcd, zookeeper,...)
* Support IPv6
* And a lot more coming

