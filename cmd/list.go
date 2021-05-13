package cmd

import (
	"log"

	"github.com/spf13/cobra"
	driver "gitlab.com/gomidi/rtmididrv"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List the available MIDI devices",
	Run:   runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) {
	drv, err := driver.New()
	if err != nil {
		log.Fatal(err)
	}
	defer drv.Close()

	log.Println()
	log.Print("available input devices:")
	inputs, err := drv.Ins()
	if err != nil {
		log.Fatal(err)
	}
	for _, port := range inputs {
		log.Printf("%2d: %s", port.Number(), port.String())
	}

	log.Println()
	log.Print("available output devices:")
	outputs, err := drv.Ins()
	if err != nil {
		log.Fatal(err)
	}
	for _, port := range outputs {
		log.Printf("%2d: %s", port.Number(), port.String())
	}
}
