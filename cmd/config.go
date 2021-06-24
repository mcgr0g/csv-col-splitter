package cmd

/*
Copyright © 2021 https://github.com/mcgr0g

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "show or freeze configs",
	Long: `Command to view configuration key-values in terminal 
or save it to file`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Inside subCmd PreRun with args: %v\n", args)
		if cmd.Flag("show").Changed {
			fmt.Println("config called")
			fmt.Println("cfgFile = ", cmd.Flag("config").Value)

			for ikey, ival := range viper.AllSettings() {
				fmt.Println("'"+ikey+"'", "setted to ", ival)
			}
		}
		if cmd.Flag("freeze").Changed {
			//fmt.Println("AllKeys = ", viper.AllSettings())
			if cmd.Flag("config").Changed {
				// viper подцепил файл конфигурации, так как прередан флагом
				err := viper.WriteConfig()
				if err != nil {
					println(err)
				}
			} else {
				// viper не подцепил файл конфигурации по дефолтному имени, так как его нет на ФС
				err := viper.WriteConfigAs(cmd.Flag("config").Value.String())
				if err != nil {
					println(err)
				}
			}
			fmt.Println("cfgFile = ", cmd.Flag("config").Value)
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.Flags().BoolP("show", "s", true, "show in terminal")
	configCmd.Flags().BoolP("freeze", "f", true, "freeze in file")
}
