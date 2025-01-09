package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func add(command string, readOnly bool, exitcode int) {
	var toSave []commandsList

	module := strings.Split(command, " ")[1]
	if module == "" {
		fmt.Println(`fail! Command syntax eg.: "mgc auth login"`)
		return
	}
	toAdd := commandsList{Command: command, Module: module, ReadOnly: readOnly, ExitCode: exitcode}

	currentConfig, err := loadList()
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, x := range currentConfig.Commands {
		if x.Command == toAdd.Command {
			continue
		}

		toSave = append(toSave, x)
	}

	toSave = append(toSave, toAdd)

	viper.Set("commands", toSave)
	err = viper.WriteConfigAs(VIPER_FILE)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("done")
}

func AddCommandCmd() *cobra.Command {
	var command string
	var readOnly bool
	var exitcode int

	addCommand := &cobra.Command{
		Use:     "add",
		Short:   "Add new command",
		Example: "specs add 'mgc auth login'",
		Hidden:  false,
		Run: func(cmd *cobra.Command, args []string) {
			add(command, readOnly, exitcode)
		},
	}

	addCommand.Flags().StringVarP(&command, "command", "c", "", "Command to be added")
	addCommand.Flags().IntVarP(&exitcode, "exit-code", "e", 0, "Exit code to pass")
	addCommand.Flags().BoolVarP(&readOnly, "read-only", "r", false, "Set command as read-only")

	// Marca a flag command como obrigatória
	addCommand.MarkFlagRequired("command")
	addCommand.MarkFlagRequired("exit-code")
	addCommand.MarkFlagRequired("read-only")

	return addCommand
}
