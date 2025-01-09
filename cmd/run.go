package cmd

import (
	"fmt"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

type resultError struct {
	commandsList
	Error string
}

type result struct {
	errors  []resultError
	success []commandsList
}

func run(readOnly bool, rewriteSnap bool) {

	_ = ensureDirectoryExists(path.Join(currentDir(), SNAP_DIR))

	currentCommands, err := loadList()

	if err != nil {
		fmt.Println(err)
		return

	}
	result := result{}
	var wg sync.WaitGroup

	for _, cmmd := range currentCommands.Commands {
		if readOnly {
			if !cmmd.ReadOnly {
				continue
			}
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			output := []byte{}
			if cmmd.SecToWaitBefore > 0 {
				time.Sleep(time.Duration(cmmd.SecToWaitBefore) * time.Second)
			}
			var varMap map[string]string
			varMap = map[string]string{}

			originalCommand := cmmd.Command
			if strings.Contains(cmmd.Command, "{{") {
				varName := extractVarName(cmmd.Command)
				if found, ok := varMap[varName]; ok {
					cmmd.Command = replaceVariable(cmmd.Command, varName, found)
				}
			}

			cmmd.Command = replaceRandom(cmmd.Command)
			strCommand := path.Join(currentCommands.BinaryPath, cmmd.Command)
			saveCommand := []byte("Command: " + originalCommand + "\nOutput:\n")
			output = append(output, saveCommand...)

			cmd := exec.Command("sh", "-c", strCommand)
			cmdOutput, err := cmd.CombinedOutput()

			output = append(output, cmdOutput...)

			if cmmd.OutputTarget != "" && cmmd.OutputVariable != "" {
				outputValue := matchPattern(cmmd.OutputTarget, string(output))
				varMap[cmmd.OutputVariable] = outputValue
			}

			if exitError, ok := err.(*exec.ExitError); ok {
				if exitError.ExitCode() != int(cmmd.ExitCode) {
					result.errors = append(result.errors, resultError{
						commandsList: cmmd,
						Error:        err.Error(),
					})
					return
				}
			} else if err != nil {
				result.errors = append(result.errors, resultError{
					commandsList: cmmd,
					Error:        err.Error(),
				})
				return
			}

			if cmmd.SecToWaitAfter > 0 {
				time.Sleep(time.Duration(cmmd.SecToWaitAfter) * time.Second)
			}

			for _, scmmd := range cmmd.SubCommands {

				if scmmd.SecToWaitBefore > 0 {
					time.Sleep(time.Duration(scmmd.SecToWaitBefore) * time.Second)
				}

				soutput := []byte{}
				originalSubCommand := scmmd.Command
				if strings.Contains(scmmd.Command, "{{") {
					varName := extractVarName(scmmd.Command)
					if found, ok := varMap[varName]; ok {
						scmmd.Command = replaceVariable(scmmd.Command, varName, found)
					}
				}

				scmmd.Command = replaceRandom(scmmd.Command)
				sstrCommand := path.Join(currentCommands.BinaryPath, scmmd.Command)

				var scmdOut string
				var err error

				for {
					scmd := exec.Command("sh", "-c", sstrCommand)
					scmdOutput, errcmd := scmd.CombinedOutput()
					err = errcmd

					if strings.Contains(string(scmdOutput), "409 Conflict") {
						time.Sleep(10 * time.Second)
					} else {
						scmdOut = string(scmdOutput)
						break
					}
				}

				ssaveCommand := []byte("Sub Command: " + originalSubCommand + "\nOutput:\n")
				soutput = append(soutput, ssaveCommand...)
				soutput = append(soutput, scmdOut...)

				if scmmd.OutputTarget != "" && scmmd.OutputVariable != "" {
					outputValue := matchPattern(scmmd.OutputTarget, scmdOut)
					varMap[scmmd.OutputVariable] = outputValue
				}

				if exitError, ok := err.(*exec.ExitError); ok {
					if exitError.ExitCode() != int(scmmd.ExitCode) {
						result.errors = append(result.errors, resultError{
							commandsList: cmmd,
							Error:        err.Error(),
						})
					}
				} else if err != nil {
					result.errors = append(result.errors, resultError{
						commandsList: cmmd,
						Error:        err.Error(),
					})
				}

				breakLine := []byte("\n")
				output = append(output, breakLine...)
				output = append(output, soutput...)
				if scmmd.SecToWaitAfter > 0 {
					time.Sleep(time.Duration(scmmd.SecToWaitAfter) * time.Second)
				}
			}

			snapshotFile := normalizeCommandToFile(cmmd.Command)
			if rewriteSnap {
				_ = writeSnapshot(output, SNAP_DIR, snapshotFile)
			}

			if !rewriteSnap {
				err = compareSnapshot(output, SNAP_DIR, snapshotFile)
				if err != nil {
					result.errors = append(result.errors, resultError{
						commandsList: cmmd,
						Error:        err.Error(),
					})
					return
				}
			}

			result.success = append(result.success, cmmd)
		}()
	}

	wg.Wait()

	//TODO: Fazer um output bonitinho =)
	if len(result.errors) == 0 {
		fmt.Println("Sucesso! Todos os comandos executados sem alterações.")
		return
	}

	fmt.Print("\nErros encontrados:\n\n")
	for _, er := range result.errors {

		fmt.Println("Command: ", er.Command)
		fmt.Println("Error: ", er.Error)
	}
}

func RunCommand() *cobra.Command {

	var rewriteSnap bool
	var readOnly bool
	var runTestsCmd = &cobra.Command{
		Use:    "run",
		Short:  "Run all available tests",
		Hidden: false,
		Run: func(cmd *cobra.Command, args []string) {
			run(readOnly, rewriteSnap)
		},
	}

	runTestsCmd.Flags().BoolVarP(&rewriteSnap, "rewrite-snapshots", "s", false, "Rewrite all snapshots")
	runTestsCmd.Flags().BoolVarP(&readOnly, "read-only", "r", true, "Run only commands setted as a read-only")

	// Marca a flag command como obrigatória
	runTestsCmd.MarkFlagRequired("rewrite-snapshots")
	runTestsCmd.MarkFlagRequired("read-only")

	return runTestsCmd
}
