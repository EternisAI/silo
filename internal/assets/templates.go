package assets

import _ "embed"

//go:embed docker-compose.yml.tmpl
var DockerComposeTemplate string

//go:embed config.yml.tmpl
var ConfigTemplate string
