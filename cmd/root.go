// Copyright © 2017 Ricardo Aravena <raravena@branch.io>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/raravena80/sshrunner/exec"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile  string
	command  string
	user     string
	key      string
	port     string
	machines []string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "sshrunner",
	Short: "Sshrunner runs ssh commands across multiple servers",
	Long:  `Sshrunner runs ssh commands across multiple servers`,
	// Bare app run
	Run: func(cmd *cobra.Command, args []string) {
		var options []func(*exec.Options)
		options = append(options,
			exec.Machines(viper.GetStringSlice("sshrunner.machines")))
		options = append(options,
			exec.User(viper.GetString("sshrunner.user")))
		options = append(options,
			exec.Port(viper.GetString("sshrunner.port")))
		options = append(options,
			exec.Cmd(viper.GetString("sshrunner.command")))
		options = append(options,
			exec.Key(viper.GetString("sshrunner.key")))
		options = append(options,
			exec.UseAgent(viper.GetBool("sshrunner.useagent")))
		exec.Run(options...)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	curUser := os.Getenv("LOGNAME")
	sshKey := os.Getenv("HOME") + "/.ssh/id_rsa"

	// Persistent flags
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.sshrunner.yaml)")

	// Local flags
	RootCmd.Flags().StringArrayVarP(&machines, "machines", "m", []string{}, "Hosts to run command on")
	viper.BindPFlag("sshrunner.machines", RootCmd.Flags().Lookup("machines"))
	RootCmd.Flags().StringVarP(&port, "port", "p", "22", "Ssh port to connect to")
	viper.BindPFlag("sshrunner.port", RootCmd.Flags().Lookup("port"))
	RootCmd.Flags().StringVarP(&command, "command", "c", "", "Command to run")
	viper.BindPFlag("sshrunner.command", RootCmd.Flags().Lookup("command"))
	RootCmd.Flags().StringVarP(&user, "user", "u", curUser, "User to run the command as")
	viper.BindPFlag("sshrunner.user", RootCmd.Flags().Lookup("user"))
	RootCmd.Flags().StringVarP(&key, "key", "k", sshKey, "Ssh key to use for authentication, full path")
	viper.BindPFlag("sshrunner.key", RootCmd.Flags().Lookup("key"))
	RootCmd.Flags().BoolP("useagent", "a", false, "Use agent for authentication")
	viper.BindPFlag("sshrunner.useagent", RootCmd.Flags().Lookup("useagent"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".sshrunner" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".sshrunner")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
