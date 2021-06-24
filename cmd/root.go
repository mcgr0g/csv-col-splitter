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
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile           string
	cfgFileDefault    string
	workDir           string
	sourcePattern     string
	colTartetPosition int
	colSeparator      string
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
		splitCmd()
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

	rootCmd.PersistentFlags().StringVarP(&workDir, "work-dir", "w", "", "full path to work dir; default: current directory")
	rootCmd.PersistentFlags().StringVarP(&sourcePattern, "source-pattern", "p", "*.csv", "file name pattern to process; (default is *.csv)")
	rootCmd.PersistentFlags().IntVarP(&colTartetPosition, "target-col", "t", 8, "column target-col, beginigs from 1")
	rootCmd.PersistentFlags().StringVar(&colSeparator, "col-separator", ";", "delimeter for main column")
	rootCmd.PersistentFlags().StringVarP(&subcolDelimeter, "subcol-delimeter", "d", "&", "delimeter for subcolumns in target-col column")
	rootCmd.PersistentFlags().StringVarP(&kvDelimeter, "keyvalue-delimeter", "k", "@", "delimeter for key-value in subcolumns")
	rootCmd.PersistentFlags().IntVarP(&appendPosition, "append-position", "a", 0, "position for append parsed columns, default 0 (after the last)")
	rootCmd.PersistentFlags().StringVarP(&resultFileSfx, "result-file-sfx", "r", "_splt", "suffix for processed file")

	viper.BindPFlag("work-dir", rootCmd.PersistentFlags().Lookup("work-dir"))
	viper.BindPFlag("source-pattern", rootCmd.PersistentFlags().Lookup("source-pattern"))
	viper.BindPFlag("target-col", rootCmd.PersistentFlags().Lookup("target-col"))
	viper.BindPFlag("col-separator", rootCmd.PersistentFlags().Lookup("col-separator"))
	viper.BindPFlag("subcol-delimeter", rootCmd.PersistentFlags().Lookup("subcol-delimeter"))
	viper.BindPFlag("keyvalue-delimeter", rootCmd.PersistentFlags().Lookup("keyvalue-delimeter"))
	viper.BindPFlag("append-position", rootCmd.PersistentFlags().Lookup("append-position"))
	viper.BindPFlag("result-file-sfx", rootCmd.PersistentFlags().Lookup("result-file-sfx"))

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != cfgFileDefault {
		// Use config file from the flag.
		fmt.Println("In initialization used config file from the flag")
		viper.SetConfigFile(cfgFile)
	} else {
		fmt.Println("In initialization try to find config file in curr dir")
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(cfgFileDefault)
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Print("Used dafault values because of: ")
		fmt.Println(err)
	}
}

func splitCmd() {
	// спиздить  группы ожидания https://github.com/vmware-tanzu/octant/blob/master/build.go#L395
	var wg sync.WaitGroup // create waitgroup (empty struct)
	var filesToProcess []string
	findCsv(&filesToProcess)

	inCh := make(chan []string)
	//TODO make multiple file processing https://stackoverflow.com/questions/47295259/concurrently-write-multiple-csv-files-from-one-splitting-on-a-partition-column
	wg.Add(1)
	go readCsv(&wg, filesToProcess[0], inCh)
	// wg.Wait() // blocks here
	fmt.Println("Processing: ", filesToProcess)

	wg.Add(1)
	go writeCsv(&wg, filesToProcess[0], inCh)
	wg.Wait() // blocks here
}

func findCsv(files *[]string) {
	fmt.Println("searching target files")
	matches, err := filepath.Glob(viper.GetString("work-dir") + (viper.GetString("source-pattern")))
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Print("Found: ")
		fmt.Println(matches)
		if len(matches) == 0 {
			fmt.Println("nothing found")
			os.Exit(1)
		} else {
			for i := range matches {
				// fmt.Println(matches[i])
				*files = append(*files, matches[i])
			}
		}
	}
}

func readCsv(wg *sync.WaitGroup, fileToProcess string, ch chan []string) {
	// load module  https://github.com/hlawrenz/csvmung
	fmt.Println("read csv called")
	var reader *csv.Reader
	content, err := os.Open(fileToProcess)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer content.Close()
	reader = csv.NewReader(content)
	// надежда на ширину колонк, как в 1й строке
	reader.FieldsPerRecord = 0
	r, _ := utf8.DecodeRuneInString(viper.GetString("col-separator"))
	reader.Comma = r
	// не ожидаем кавычек в значениях
	reader.LazyQuotes = false

	for {
		record, err := reader.Read()
		if err == io.EOF {
			close(ch)
			break
		} else if err != nil {
			fmt.Println("Error:", err)
			close(ch)
			break
		}
		fmt.Println("old record: ", record)
		//stuff there
		fmt.Println("new record: ", record)
		ch <- record
	}
	wg.Done() // decrement counter

}

func writeCsv(wg *sync.WaitGroup, fileToProcess string, ch chan []string) {
	fmt.Println("write csv called")
	var writer *csv.Writer
	outFileExt := filepath.Ext(fileToProcess)
	outputFile := strings.ReplaceAll(fileToProcess, outFileExt, "") + viper.GetString("result-file-sfx") + outFileExt
	fmt.Println("output file: ", outputFile)

	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer file.Close()
	writer = csv.NewWriter(file)

	separator, _ := utf8.DecodeRuneInString(viper.GetString("col-separator"))
	writer.Comma = separator

	for row := range ch {
		err := writer.Write(row)
		if err != nil {
			fmt.Println("Error:", err)
			close(ch)
			return
		}
	}
	writer.Flush()
	wg.Done() // decrement counter
}
