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
require (
	github.com/hashicorp/hcl v1.0.0
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

	replace github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.9.3
	replace github.com/maliceio/malice => ./
	replace github.com/docker/docker => github.com/docker/docker v24.0.7+incompatible
