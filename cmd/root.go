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
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	log "github.com/mcgr0g/csv-col-splitter/pkg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile           string
	cfgFileDefault    string
	workDir           string
	hasHeaders        bool
	sourcePattern     string
	colTartetPosition int
	colSeparator      string
	subcolDelimeter   string
	kvDelimeter       string
	subkeyPosition    int
	subValuePosition  int
	resultFileSfx     string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "csv-col-splitter",
	Short: "split column in csv for a several columns",
	Long: `It is a CLI application for manipulating csv dataset.
This application take you file and split difined column.
New columns append to right after last.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := InBetween(subkeyPosition, 0, 1)
		if err != nil {
			log.Error(err.Error())
		} else {
			log.Info("позиция ключей в паре: " + strconv.Itoa(subkeyPosition))
			subValuePosition = 1 - subkeyPosition // key goes second, than value is first and vice versa. No other options.
			log.Info("позиция значений в паре: " + strconv.Itoa(subValuePosition))
			SplitCmd()
		}
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
	rootCmd.PersistentFlags().BoolVar(&hasHeaders, "with-headers", true, "use first line as headers")
	rootCmd.PersistentFlags().IntVarP(&colTartetPosition, "target-col", "t", 8, "column target-col, beginigs from 1")
	rootCmd.PersistentFlags().StringVar(&colSeparator, "col-separator", ";", "delimeter for main column")
	rootCmd.PersistentFlags().StringVarP(&subcolDelimeter, "subcol-delimeter", "d", "&", "delimeter for subcolumns in target-col column")
	rootCmd.PersistentFlags().StringVarP(&kvDelimeter, "keyvalue-delimeter", "k", "@", "delimeter for key-value in subcolumns")
	rootCmd.PersistentFlags().IntVar(&subkeyPosition, "key-position", 1, "key position in key-value pair, starts from 0. Default 1 - after the value. Other values are impossible")
	rootCmd.PersistentFlags().StringVarP(&resultFileSfx, "result-file-sfx", "r", "_splt", "suffix for processed file")

	viper.BindPFlag("work-dir", rootCmd.PersistentFlags().Lookup("work-dir"))
	viper.BindPFlag("source-pattern", rootCmd.PersistentFlags().Lookup("source-pattern"))
	viper.BindPFlag("with-headers", rootCmd.PersistentFlags().Lookup("with-headers"))
	viper.BindPFlag("target-col", rootCmd.PersistentFlags().Lookup("target-col"))
	viper.BindPFlag("col-separator", rootCmd.PersistentFlags().Lookup("col-separator"))
	viper.BindPFlag("subcol-delimeter", rootCmd.PersistentFlags().Lookup("subcol-delimeter"))
	viper.BindPFlag("key-position", rootCmd.PersistentFlags().Lookup("key-position"))
	viper.BindPFlag("keyvalue-delimeter", rootCmd.PersistentFlags().Lookup("keyvalue-delimeter"))
	viper.BindPFlag("result-file-sfx", rootCmd.PersistentFlags().Lookup("result-file-sfx"))

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != cfgFileDefault {
		// Use config file from the flag.
		log.Info("In initialization used config file from the flag")
		viper.SetConfigFile(cfgFile)
	} else {
		log.Info("In initialization try to find config file in curr dir")
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(cfgFileDefault)
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Info("Using config file: " + viper.ConfigFileUsed())
		hasHeaders = viper.GetBool("with-headers")
		colSeparator = viper.GetString("col-separator")
		colTartetPosition = viper.GetInt("target-col")
		subcolDelimeter = viper.GetString("subcol-delimeter")
		kvDelimeter = viper.GetString("keyvalue-delimeter")
		subkeyPosition = viper.GetInt("key-position")
		resultFileSfx = viper.GetString("result-file-sfx")

	} else {
		log.Info("Used dafault values because of: " + err.Error())
	}
}

func InBetween(i, min, max int) error {
	if (i >= min) && (i <= max) {
		return nil
	} else {
		return fmt.Errorf("var %d not in range [%d - %d]", i, min, max)
	}
}

func ValuePositionInSlice(sl []string, val string) (int, bool) {
	var match bool = false
	var position int = 0
	for k, v := range sl {
		if v == val {
			position = k
			match = true
		}
	}
	log.Debug("k= " + strconv.Itoa(position))
	log.Debug("match " + val + " = " + strconv.FormatBool(match))
	return position, match
}

func SplitCmd() {
	var wg sync.WaitGroup
	log.Info("sub main called")
	var filesToProcess []string
	findCsv(&filesToProcess)

	for _, workFile := range filesToProcess {
		headersSlice, headerMainsCount, fileContent := ScanCsvAndSubHeaders(workFile)
		outCh := make(chan []string)
		outFileExt := filepath.Ext(workFile)
		outputFile := strings.ReplaceAll(workFile, outFileExt, "") + resultFileSfx + outFileExt
		log.Info("Splitting to: " + outputFile)

		wg.Add(1)
		go RowSplitter(&wg, outCh, headersSlice, headerMainsCount, fileContent)

		wg.Add(1)
		go RowWriter(&wg, outputFile, outCh)
		wg.Wait()
	}
}

func findCsv(files *[]string) {
	log.Info("searching target files")
	matches, err := filepath.Glob(viper.GetString("work-dir") + (viper.GetString("source-pattern")))
	if err != nil {
		log.Error(err.Error())
	} else {
		logMsg := "Found: "
		if len(matches) == 0 {
			log.Warn("nothing found")
			os.Exit(1)
		} else {
			for i, v := range matches {
				if strings.Contains(v, viper.GetString("result-file-sfx")) {
					log.Warn("ignorig result file: " + v)
				} else {
					logMsg += matches[i] + " "
					*files = append(*files, matches[i])
				}
			}
			log.Info(logMsg)
		}
	}
}

func ScanCsvAndSubHeaders(fileToProcess string) ([]string, int, [][]string) {
	log.Info("scan csv headers called")
	var reader *csv.Reader
	content, err := os.Open(fileToProcess)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	defer content.Close()
	reader = csv.NewReader(content)
	reader.FieldsPerRecord = 0 // надежда на ширину колонк, как в 1й строке
	r, _ := utf8.DecodeRuneInString(colSeparator)
	reader.Comma = r
	reader.LazyQuotes = false // no quotes in cells

	csvContent, err := reader.ReadAll()
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	var unsortedSubHeaders []string
	var rowCnt int = 0
	var foundMainHeaders int = 0
	var headSlice []string
	for i, record := range csvContent {
		if i == 0 && len(record) < colTartetPosition { // lets check only first row
			errStr := "Sorry, you have " + strconv.Itoa(len(record)) + " columns but target colunn position is " + strconv.Itoa(colTartetPosition)
			log.Error(errStr)
			os.Exit(1)
		}

		rowCnt += 1
		if rowCnt == 1 && hasHeaders { // preocessing current headers in first row
			log.Info("len of headers: " + strconv.Itoa(len(record)))
			for _, value := range record {
				headSlice = append(headSlice, value)
				foundMainHeaders += 1
			}
			log.Info("saved headers with length: " + strconv.Itoa(len(headSlice)))
		} else { //scan for a new headers
			for key := range FindKVinColumn(record) {
				if _, isVisValueInHeader := ValuePositionInSlice(unsortedSubHeaders, key); !isVisValueInHeader {
					unsortedSubHeaders = append(unsortedSubHeaders, key)
					log.Info("new headers found in row = " + strconv.Itoa(rowCnt))
				}
			}
		}
	}

	if cap(unsortedSubHeaders) > 0 {
		sort.Strings(unsortedSubHeaders)
		log.Info("appended new value to headers = " + strings.Join(unsortedSubHeaders, " "))
		headSlice = append(headSlice, unsortedSubHeaders...)
	}

	// a little bit statistics
	totalHeadersCount := len(headSlice)
	foundSubHeaders := totalHeadersCount - foundMainHeaders
	log.Info("rows scanned: " + strconv.Itoa(rowCnt))
	log.Info("total headers count: " + strconv.Itoa(totalHeadersCount))
	log.Info("main headers count: " + strconv.Itoa(foundMainHeaders))
	log.Info("new headers count: " + strconv.Itoa(foundSubHeaders))
	if foundSubHeaders == 0 {
		log.Error("Sorry, you have nothig to split in target columns")
		os.Exit(1)
	}
	return headSlice, foundMainHeaders, csvContent
}

func RowSplitter(wg *sync.WaitGroup, ch chan []string, headSlice []string, headersMainCnt int, fileContent [][]string) {
	log.Info("row splitter called")

	var rowCnt int = 0
	subHeaderCount := len(headSlice) - headersMainCnt
	log.Info("subHeaderCount: " + strconv.Itoa(subHeaderCount))
	for _, record := range fileContent {
		rowCnt += 1
		log.Debug("splitting row: " + strconv.Itoa(rowCnt))
		if hasHeaders && rowCnt == 1 {
			log.Debug("processing original headers")
			log.Debug("on row " + strconv.Itoa(rowCnt) + " splitting :" + strings.Join(headSlice, " "))
			ch <- headSlice
		} else {
			log.Debug("scan row for subheaders")
			for pos, subHeaderToFind := range headSlice {
				if pos >= headersMainCnt { // searching only for subheaders;
					var foundSubHeaders int = 0
					for subHeaderInRecord, subValueToWrite := range FindKVinColumn(record) {
						if subHeaderInRecord == subHeaderToFind { // found the sought subheader
							foundSubHeaders += 1
							log.Debug("found sub-value = " + subValueToWrite)
							record = append(record, subValueToWrite)
						}
					}
					// not found the sought subheader, so  но надо дописать 1 разделитель, что бы ширина csv была однородной во всех строках
					log.Debug("count of processoed sub-header in row = " + strconv.Itoa(foundSubHeaders))
					if foundSubHeaders == 0 {
						record = append(record, "")
					}
				}
			}
			log.Debug("on row " + strconv.Itoa(rowCnt) + " splint result: " + strings.Join(record, ";"))
			ch <- record
		}
	}
	close(ch)
	log.Info("total rows scanned: " + strconv.Itoa(rowCnt))
	wg.Done() // decrement counter

}

func FindKVinColumn(record []string) map[string]string {
	resultMap := make(map[string]string)

	if strings.Contains(record[colTartetPosition], subcolDelimeter) {
		subs := strings.Split(record[colTartetPosition], subcolDelimeter)
		for _, subColumn := range subs {
			keyAndValueInSubColumn := strings.Split(subColumn, kvDelimeter)
			if cap(keyAndValueInSubColumn) > 1 {
				resultMap[keyAndValueInSubColumn[subkeyPosition]] = keyAndValueInSubColumn[subValuePosition]
			}
		}
	}
	return resultMap
}

func RowWriter(wg *sync.WaitGroup, fileToProcess string, ch chan []string) {
	log.Info("row writer called")

	var writer *csv.Writer
	log.Info("output file: " + fileToProcess)

	file, err := os.Create(fileToProcess)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	defer file.Close()
	writer = csv.NewWriter(file)

	separator, _ := utf8.DecodeRuneInString(colSeparator)
	writer.Comma = separator
	var rowCnt int = 0

	for row := range ch {
		rowCnt += 1
		log.Debug("writing output file row: " + strconv.Itoa(rowCnt))
		log.Debug("writing output file row value: " + strings.Join(row, " "))
		err := writer.Write(row)
		if err != nil {
			log.Error(err.Error())
			close(ch)
			return
		}
	}
	writer.Flush()
	wg.Done() // decrement counter
}
