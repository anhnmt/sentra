package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const BANNER = `
  _________              __                 
 /   _____/ ____   _____/  |_____________   
 \_____  \_/ __ \ /    \   __\_  __ \__  \  
 /        \  ___/|   |  \  |  |  | \// __ \_
/_______  /\___  >___|  /__|  |__|  (____  /
        \/     \/     \/                 \/ 
`

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sentra",
	Long:  BANNER,
	Short: "A brief description of your application",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		fmt.Print(BANNER)
	},
}

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

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.sentra.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
