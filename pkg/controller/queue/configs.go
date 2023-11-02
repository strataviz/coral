package queue

// TODO: potentially turn this into a template as we start needing
// more configuration options.
var NatsConfig = `
{
	"http_port": 8222,
	"lame_duck_duration": "30s",
	"lame_duck_grace_period": "10s",
	"pid_file": "/var/run/nats/nats.pid",
	"port": 4222,
	"server_name": $SERVER_NAME
}
`
