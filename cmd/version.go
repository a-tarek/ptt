package cmd

// Version is set via ldflags at build time
var Version = "dev"

func init() {
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("ptt version {{.Version}}\n")
}
