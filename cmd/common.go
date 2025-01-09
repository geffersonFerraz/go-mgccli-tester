package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type configsToTest struct {
	Commands   []commandsList
	BinaryPath string `yaml:"binary_path"`
}

type commandsList struct {
	Module          string `yaml:"module"`
	Command         string `yaml:"command"`
	ReadOnly        bool   `yaml:"readonly"`
	ExitCode        int    `yaml:"exitcode"`
	OutputVariable  string `yaml:"output_variable"`
	OutputTarget    string `yaml:"output_target"`
	SecToWaitBefore int    `yaml:"sec_wait_before_run"`
	SecToWaitAfter  int    `yaml:"sec_wait_after_run"`
	SubCommands     []subCommandsList
}

type subCommandsList struct {
	Command         string `yaml:"subcommand"`
	ExitCode        int    `yaml:"exitcode"`
	OutputVariable  string `yaml:"output_variable"`
	OutputTarget    string `yaml:"output_target"`
	SecToWaitBefore int    `yaml:"sec_wait_before_run"`
	SecToWaitAfter  int    `yaml:"sec_wait_after_run"`
}

func interfaceToMap(i interface{}) (map[string]interface{}, bool) {
	mapa, ok := i.(map[string]interface{})
	if !ok {
		fmt.Println("A interface não é um mapa ou mapa de interfaces.")
		return nil, false
	}
	return mapa, true
}

func loadList() (configsToTest, error) {
	var configs configsToTest
	configs.Commands = []commandsList{}

	configs.BinaryPath = viper.GetString("binary_path")

	config := viper.Get("commands")

	if config != nil {
		for _, v := range config.([]interface{}) {
			vv, ok := interfaceToMap(v)
			if !ok {
				return configsToTest{}, fmt.Errorf("fail to load current config")
			}

			outputVariable := ""
			if outVar, ok := vv["output_variable"]; outVar != nil && ok {
				outputVariable = outVar.(string)
			}

			outputTarget := ""
			if outTarget, ok := vv["output_target"]; outTarget != nil && ok {
				outputTarget = outTarget.(string)
			}

			secBefore := 0
			if secWaitBefore, ok := vv["sec_wait_before_run"]; secWaitBefore != nil && ok {
				secBefore = secWaitBefore.(int)
			}

			secAfter := 0
			if secWaitAfter, ok := vv["sec_wait_after_run"]; secWaitAfter != nil && ok {
				secAfter = secWaitAfter.(int)
			}

			cmd := commandsList{
				Module:          vv["module"].(string),
				Command:         vv["command"].(string),
				ReadOnly:        vv["readonly"].(bool),
				ExitCode:        vv["exitcode"].(int),
				OutputVariable:  outputVariable,
				OutputTarget:    outputTarget,
				SecToWaitBefore: secBefore,
				SecToWaitAfter:  secAfter,
			}

			// Check for subcommands
			if subcommands, exists := vv["subcommands"]; exists {
				for _, sub := range subcommands.([]interface{}) {
					subCmd, ok := interfaceToMap(sub)
					if !ok {
						continue
					}

					outputVariable := ""
					if outVar, ok := subCmd["output_variable"]; outVar != nil && ok {
						outputVariable = outVar.(string)
					}

					outputTarget := ""
					if outTarget, ok := subCmd["output_target"]; outTarget != nil && ok {
						outputTarget = outTarget.(string)
					}

					secBefore := 0
					if secWaitBefore, ok := subCmd["sec_wait_before_run"]; secWaitBefore != nil && ok {
						secBefore = secWaitBefore.(int)
					}

					secAfter := 0
					if secWaitAfter, ok := subCmd["sec_wait_after_run"]; secWaitAfter != nil && ok {
						secAfter = secWaitAfter.(int)
					}

					cmd.SubCommands = append(cmd.SubCommands, subCommandsList{
						Command:         subCmd["command"].(string),
						ExitCode:        subCmd["exitcode"].(int),
						OutputVariable:  outputVariable,
						OutputTarget:    outputTarget,
						SecToWaitBefore: secBefore,
						SecToWaitAfter:  secAfter,
					})
				}
			}

			configs.Commands = append(configs.Commands, cmd)
		}

	}
	return configs, nil
}

func ensureDirectoryExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, 0755)
	}
	return nil
}

func createFile(content []byte, dir, filePath string) error {
	return os.WriteFile(filepath.Join(dir, filePath), content, 0644)
}

func loadFile(dir, filePath string) ([]byte, error) {
	return os.ReadFile(filepath.Join(dir, filePath))
}

func writeSnapshot(output []byte, dir string, id string) error {
	_ = createFile(output, dir, fmt.Sprintf("%s.cli", id))
	return nil
}

// func compareBytes(expected, actual []byte, ignoreDateUUID bool) error {
// 	if bytes.Equal(expected, actual) {
// 		return nil
// 	}

// 	allEqual := true

// 	expectedLines := strings.Split(string(expected), "\n")
// 	actualLines := strings.Split(string(actual), "\n")

// 	var diff strings.Builder
// 	diff.WriteString("\nDiferenças encontradas:\n")

// 	dateRegex := regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`)
// 	uuidRegex := regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)

// 	i, j := 0, 0
// 	for i < len(expectedLines) && j < len(actualLines) {
// 		expectedLine := expectedLines[i]
// 		actualLine := actualLines[j]

// 		if ignoreDateUUID {
// 			expectedLine = dateRegex.ReplaceAllString(expectedLine, "DATE")
// 			expectedLine = uuidRegex.ReplaceAllString(expectedLine, "UUID")
// 			actualLine = dateRegex.ReplaceAllString(actualLine, "DATE")
// 			actualLine = uuidRegex.ReplaceAllString(actualLine, "UUID")
// 		}

// 		if expectedLine == actualLine {
// 			diff.WriteString("  " + expectedLines[i] + "\n")
// 			i++
// 			j++
// 		} else {
// 			allEqual = false
// 			diff.WriteString("- " + expectedLines[i] + "\n")
// 			diff.WriteString("+ " + actualLines[j] + "\n")
// 			i++
// 			j++
// 		}
// 	}

// 	for ; i < len(expectedLines); i++ {
// 		diff.WriteString("- " + expectedLines[i] + "\n")
// 	}
// 	for ; j < len(actualLines); j++ {
// 		diff.WriteString("+ " + actualLines[j] + "\n")
// 	}

// 	if allEqual {
// 		return nil
// 	}

//		return fmt.Errorf("%s", diff.String())
//	}
func compareBytes(expected, actual []byte, ignoreDateUUID bool) error {
	if bytes.Equal(expected, actual) {
		return nil
	}

	expectedLines := strings.Split(string(expected), "\n")
	actualLines := strings.Split(string(actual), "\n")
	var diff strings.Builder
	diff.WriteString("\nDiferenças encontradas:\n")

	dateRegex := regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`)
	uuidRegex := regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)

	allEqual := true
	i, j := 0, 0

	for i < len(expectedLines) && j < len(actualLines) {
		expectedLine := expectedLines[i]
		actualLine := actualLines[j]

		if ignoreDateUUID {
			expectedLine = dateRegex.ReplaceAllString(expectedLine, "DATE")
			expectedLine = uuidRegex.ReplaceAllString(expectedLine, "UUID")
			actualLine = dateRegex.ReplaceAllString(actualLine, "DATE")
			actualLine = uuidRegex.ReplaceAllString(actualLine, "UUID")
		}

		// Compare flag lists if the lines contain flags
		if strings.Contains(expectedLine, "missing required flags:") {
			if compareFlagLists(expectedLine, actualLine) {
				diff.WriteString(" " + expectedLines[i] + "\n")
				i++
				j++
				continue
			}
		}

		if expectedLine == actualLine {
			diff.WriteString(" " + expectedLines[i] + "\n")
			i++
			j++
		} else {
			allEqual = false
			diff.WriteString("- " + expectedLines[i] + "\n")
			diff.WriteString("+ " + actualLines[j] + "\n")
			i++
			j++
		}
	}

	for ; i < len(expectedLines); i++ {
		diff.WriteString("- " + expectedLines[i] + "\n")
	}
	for ; j < len(actualLines); j++ {
		diff.WriteString("+ " + actualLines[j] + "\n")
	}

	if allEqual {
		return nil
	}
	return fmt.Errorf("%s", diff.String())
}
func compareFlagLists(expected, actual string) bool {
	flagsRegex := regexp.MustCompile(`Error: missing required flags: (.+)`)

	expectedMatch := flagsRegex.FindStringSubmatch(expected)
	actualMatch := flagsRegex.FindStringSubmatch(actual)

	if len(expectedMatch) < 2 || len(actualMatch) < 2 {
		return false
	}

	expectedFlags := strings.Split(expectedMatch[1], ", ")
	actualFlags := strings.Split(actualMatch[1], ", ")

	if len(expectedFlags) != len(actualFlags) {
		return false
	}

	sort.Strings(expectedFlags)
	sort.Strings(actualFlags)

	return reflect.DeepEqual(expectedFlags, actualFlags)
}

func compareSnapshot(output []byte, dir string, id string) error {
	snapContent, err := loadFile(dir, fmt.Sprintf("%s.cli", id))
	if err == nil {
		return compareBytes(snapContent, output, true)
	}

	if errors.Is(err, os.ErrNotExist) {
		_ = writeSnapshot(output, dir, id)
		return nil
	}

	return fmt.Errorf("Diosmio")
}

func normalizeCommandToFile(input string) string {
	words := strings.Fields(input)
	var filteredWords []string
	for _, word := range words {
		if !strings.HasPrefix(word, "--") {
			filteredWords = append(filteredWords, word)
		}
	}
	result := strings.Join(filteredWords, "-")
	return result
}

func replaceVariable(target, variable, value string) string {
	variable = "{{" + variable + "}}"
	return strings.ReplaceAll(target, variable, value)
}

func extractVarName(target string) string {
	re := regexp.MustCompile(`\{\{(.*?)\}\}`)
	match := re.FindStringSubmatch(target)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func matchPattern(target, input string) string {
	var pattern strings.Builder
	pattern.WriteString(target)

	re := regexp.MustCompile(pattern.String())
	str := strings.Trim(input, " \n")
	return re.FindString(str)
}

func replaceRandom(command string) string {
	if !strings.Contains(command, "{{random}}") {
		return command
	}

	random := strconv.FormatInt(time.Now().Unix(), 10)
	return strings.ReplaceAll(command, "{{random}}", random)
}
