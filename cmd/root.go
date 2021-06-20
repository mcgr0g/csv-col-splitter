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
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile           string
	cfgFileDefault    string
	workDir           string
	dataframePattern  string
	colTartetPosition int
	subcolDelimeter   string
	kvDelimeter       string
	appendPosition    int
	resultFileSfx     string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "csv-col-splitter",
	Short: "split column in csv for a several columns",
	Long: `It is a CLI application for manipulating csv dataset.
This application take you file and split difined column.
New columns append to right after last.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		readFile()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cfgFileDefault = "csv-col-splitter.yaml"
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", cfgFileDefault, "config file (default is "+cfgFileDefault+")")

	rootCmd.PersistentFlags().StringVarP(&workDir, "workDir", "w", "", "full path to work dir; default: current directory")
	rootCmd.PersistentFlags().StringVarP(&dataframePattern, "dataframePattern", "p", ".csv", "file name pattern to process; (default is .csv)")
	rootCmd.PersistentFlags().IntVarP(&colTartetPosition, "target", "t", 4, "column target, beginigs from 1")
	rootCmd.PersistentFlags().StringVarP(&subcolDelimeter, "subcolDelimeter", "d", "&", "delimeter for subcolDelimeterumnms")
	rootCmd.PersistentFlags().StringVarP(&kvDelimeter, "kvDelimeter", "k", "@", "delimeter for key-value in subcolDelimeterumnm")
	rootCmd.PersistentFlags().IntVarP(&appendPosition, "appendPosition", "a", 0, "position for append parsed columns, default 0 (after the last)")
	rootCmd.PersistentFlags().StringVarP(&resultFileSfx, "resultFileSfx", "r", "_splt", "suffix for processed file")

	viper.BindPFlag("workDir", rootCmd.PersistentFlags().Lookup("workDir"))
	viper.BindPFlag("dataframePattern", rootCmd.PersistentFlags().Lookup("dataframePattern"))
	viper.BindPFlag("target", rootCmd.PersistentFlags().Lookup("target"))
	viper.BindPFlag("subcolDelimeter", rootCmd.PersistentFlags().Lookup("subcolDelimeter"))
	viper.BindPFlag("kvDelimeter", rootCmd.PersistentFlags().Lookup("kvDelimeter"))
	viper.BindPFlag("appendPosition", rootCmd.PersistentFlags().Lookup("appendPosition"))
	viper.BindPFlag("resultFileSfx", rootCmd.PersistentFlags().Lookup("resultFileSfx"))

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	//fmt.Println("cfgFile = ", cfgFile)
	if cfgFile != cfgFileDefault {
		// Use config file from the flag.
		fmt.Println("In initialization used config file from the flag")
		viper.SetConfigFile(cfgFile)
	} else {
		fmt.Println("In initialization used config file from curr dir")
		viper.AddConfigPath(".")
		viper.SetConfigName(cfgFileDefault)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func readFile() {
	// load module  https://github.com/hlawrenz/csvmung
	println("read csv called")

}
