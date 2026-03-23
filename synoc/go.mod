module github.com/ishua/a3bot6/synoc

go 1.23.1

require (
	github.com/cristalhq/aconfig v0.18.6
	github.com/cristalhq/aconfig/aconfigyaml v0.17.1
	github.com/ishua/a3bot6/mcore v0.0.0
)

require (
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/ishua/a3bot6/mcore => ../mcore
