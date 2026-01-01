package commands

// AppHelpTemplate custom app help template compatible with urfave/cli v2
var AppHelpTemplate = `Usage: {{.HelpName}} {{if .VisibleFlags}}[OPTIONS] {{end}}COMMAND [arg...]

{{.Usage}}

Version: {{.Version}}{{if .Authors}}

Author:{{range .Authors}}
  {{.}}{{end}}{{end}}
{{if .VisibleFlags}}
Options:
  {{range .VisibleFlags}}{{.}}
  {{end}}{{end}}
Commands:
  {{range .Commands}}{{.Name}}{{ "\t" }}{{.Usage}}
  {{end}}
Run '{{.Name}} COMMAND --help' for more information on a command.
`

// CommandHelpTemplate custom command help template
var CommandHelpTemplate = `Usage: {{.HelpName}}{{if .VisibleFlags}} [OPTIONS]{{end}} [arg...]
{{.Usage}}{{if .Description}}

Description:
   {{.Description}}{{end}}{{if .Flags}}

Options:
   {{range .VisibleFlags}}
   {{.}}{{end}}{{ end }}
`
