package main

import (
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{Use: "webpushy"}
	root.AddCommand(sendCmd())
	root.AddCommand(recvCmd())

	err := root.Execute()
	if err != nil {
		panic(err)
	}
}


