package wireguard

const wgConfigTemplate = `[Interface]
ListenPort = {{ .ListenPort  }}
PrivateKey = {{ .PrivateKey }}
`
