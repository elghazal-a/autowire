package wireguard

const wgConfigTemplate = `[Interface]
SaveConfig = false
Address = {{ .Address  }}
ListenPort = {{ .ListenPort  }}
PrivateKey = {{ .PrivateKey }}
`
