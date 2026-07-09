module github.com/plexusone/omniskill

go 1.26.0

require (
	github.com/grokify/mogo v0.74.6
	github.com/modelcontextprotocol/go-sdk v1.6.1
	golang.ngrok.com/ngrok v1.13.0
	golang.org/x/mod v0.38.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/google/jsonschema-go v0.4.3 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/inconshreveable/log15 v3.0.0-testing.5+incompatible // indirect
	github.com/inconshreveable/log15/v3 v3.1.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	github.com/segmentio/encoding v0.5.4 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.ngrok.com/muxado/v2 v2.0.1 // indirect
	golang.org/x/exp v0.0.0-20260611194520-c48552f49976 // indirect
	golang.org/x/net v0.56.0 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/sync v0.21.0 // indirect
	golang.org/x/sys v0.46.0 // indirect
	golang.org/x/term v0.44.0 // indirect
	golang.org/x/text v0.38.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

// Force log15/v3 to version compatible with ngrok v1.12.0 (has ext.RandId)
replace github.com/inconshreveable/log15/v3 => github.com/inconshreveable/log15/v3 v3.0.0-testing.5
