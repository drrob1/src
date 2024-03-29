/*
Copyright © 2022 Robert Solomon <rob@drrob1.com>

*/

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pscan",
	Short: "pscan is a port scanner",
	Long: `pScan - short for Port Scanner - executes TCP port scan
on a list of hosts.

pScan allows you to add list, and delete hosts from the list.

pScan executes a port scan on specified TCP ports.  You can customize the
target ports using a command line flag.`,
	Version: "0.1",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

var cfgFile = "$HOME/.pscan.yaml"

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.pscan.yaml)")

	// first param is long name, next is short name, next is default value, and finally usage.
	// THis uses the package pflag which replaces the std flag package and supports POSIX flags.  Cobra automatically imports it.
	rootCmd.PersistentFlags().StringP("hosts-file", "f", "pScan.hosts", "pScan hosts file")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")   This is a sample flag which is not needed by this pgm.

	versionTemplate := `{{printf "%s: %s - version %s\n"  .Name  .Short  .Version}}`
	rootCmd.SetVersionTemplate(versionTemplate)
}

func initConfig() {
	// the book says this is automatically defined, but for me, it wasn't.  I don't yet know what I'm going to do about that.
}
