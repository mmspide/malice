module github.com/mmspide/malice

go 1.21

require (
	github.com/BurntSushi/toml v1.3.2
	github.com/dustin/go-jsonpointer v0.0.0-20160814072949-ba0abeacc3dc
	github.com/fatih/structs v1.1.0
	github.com/gorilla/mux v1.8.1
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.7.0
	github.com/spf13/viper v1.17.0
	github.com/docker/docker v24.0.7+incompatible
	github.com/docker/go-connections v0.5.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

replace github.com/BurntSushi/toml => ./vendor/github.com/BurntSushi/toml
replace github.com/dustin/go-jsonpointer => ./vendor/github.com/dustin/go-jsonpointer
replace github.com/fatih/structs => ./vendor/github.com/fatih/structs
replace github.com/gorilla/mux => ./vendor/github.com/gorilla/mux
replace github.com/sirupsen/logrus => ./vendor/github.com/Sirupsen/logrus
replace github.com/spf13/cobra => ./vendor/github.com/spf13/cobra
replace github.com/spf13/viper => ./vendor/github.com/spf13/viper
replace github.com/docker/docker => ./vendor/github.com/docker/docker
replace github.com/docker/go-connections => ./vendor/github.com/docker/go-connections
replace gopkg.in/natefinch/lumberjack.v2 => ./vendor/gopkg.in/natefinch/lumberjack.v2

replace github.com/hashicorp/hcl => ./vendor/github.com/hashicorp/hcl
replace github.com/inconshreveable/log15 => ./vendor/github.com/inconshreveable/log15
replace github.com/inconshreveable/log15/v3 => ./vendor/github.com/inconshreveable/log15/v3
replace github.com/magiconair/properties => ./vendor/github.com/magiconair/properties
replace github.com/mattn/go-colorable => ./vendor/github.com/mattn/go-colorable
replace github.com/mattn/go-isatty => ./vendor/github.com/mattn/go-isatty
replace github.com/mitchellh/mapstructure => ./vendor/github.com/mitchellh/mapstructure
replace github.com/pelletier/go-toml/v2 => ./vendor/github.com/pelletier/go-toml/v2
replace github.com/spf13/afero => ./vendor/github.com/spf13/afero
replace github.com/spf13/cast => ./vendor/github.com/spf13/cast
replace github.com/spf13/jwalterweatherman => ./vendor/github.com/spf13/jwalterweatherman
replace github.com/spf13/pflag => ./vendor/github.com/spf13/pflag
replace github.com/subosito/gotenv => ./vendor/github.com/subosito/gotenv
replace golang.org/x/sys => ./vendor/golang.org/x/sys
replace golang.org/x/text => ./vendor/golang.org/x/text
replace gopkg.in/yaml.v2 => ./vendor/gopkg.in/yaml.v2
replace github.com/docker/go-units => ./vendor/github.com/docker/go-units
replace github.com/docker/distribution => ./vendor/github.com/docker/distribution
replace github.com/moby/term => ./vendor/github.com/moby/term
replace github.com/opencontainers/go-digest => ./vendor/github.com/opencontainers/go-digest
replace github.com/opencontainers/image-spec => ./vendor/github.com/opencontainers/image-spec


require (
	github.com/hashicorp/hcl v1.0.0
	github.com/inconshreveable/log15 v2.0.0+incompatible
	github.com/inconshreveable/log15/v3 v3.0.0-testing.5+incompatible
	github.com/magiconair/properties v1.8.7
	github.com/mattn/go-colorable v0.1.13
	github.com/mattn/go-isatty v0.0.20
	github.com/mitchellh/mapstructure v1.5.0
	github.com/pelletier/go-toml/v2 v2.1.0
	github.com/spf13/afero v1.10.0
	github.com/spf13/cast v1.6.0
	github.com/spf13/jwalterweatherman v1.1.0
	github.com/spf13/pflag v1.0.5
	github.com/subosito/gotenv v1.6.0
	golang.org/x/sys v0.15.0
	golang.org/x/text v0.14.0
	gopkg.in/yaml.v2 v2.4.0
	github.com/docker/go-units v0.5.0
	github.com/docker/distribution v2.8.3+incompatible
	github.com/moby/term v0.5.0
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.1.0-rc5
)
