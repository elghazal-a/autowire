package wireguard

type Configuration struct {
	Interface Interface
	Peers     []Peer
}

const WgConfigTemplate = `[Interface]
Address 		= {{ .Interface.Address  }}
ListenPort 	= {{ .Interface.ListenPort  }}
PrivateKey 	= {{ .Interface.PrivateKey }}
{{ if .Interface.PostUp }}PostUp 	= {{ .Interface.PostUp }}{{ end }}
{{ if .Interface.PostDown }}PostDown 	= {{ .Interface.PostDown }}{{ end }}

# Add servers and relay servers peers
{{ range .Peers}}
	{{ if and (not .IsClient) (ne .PublicKey $.Interface.PublicKey) }}
[Peer]
PublicKey = {{ .PublicKey }}
AllowedIPs = {{ .AllowedIPs }}
Endpoint = {{ .EndpointIP.String }}:{{.EndpointPort}}
	{{ end }}
{{ end }}

# Add clients peers
{{if .Interface.IsRelay}} 
	{{ range .Peers }}
		{{ if .IsClient}}
[Peer]
PublicKey = {{ .PublicKey }}
AllowedIPs = {{ .AllowedIPs }}
		{{ end }}
	{{ end }}
{{end}}
`
