package main

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/glassechidna/webpushy"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type senderConfig struct {
	Subscriber string
	Public     string
	Private    string
}

func senderConfigPath() string {
	path, err := homedir.Expand("~/.pushy/keys.json")
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll(filepath.Dir(path), 0700)
	if err != nil {
		panic(err)
	}

	return path
}

func sendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "",
		Run:   send,
	}

	cmd.Flags().String("endpoint", "", "")
	cmd.Flags().String("public", "", "")
	cmd.Flags().String("private", "", "")
	cmd.Flags().String("payload", "", "")
	cmd.Flags().Int("ttl", 0, "")

	cmd.AddCommand(sendInitCmd())
	return cmd
}

func send(cmd *cobra.Command, args []string) {
	endpoint, _ := cmd.Flags().GetString("endpoint")
	payload, _ := cmd.Flags().GetString("payload")
	ttl, _ := cmd.Flags().GetInt("ttl")
	public, _ := cmd.Flags().GetString("public")
	private, _ := cmd.Flags().GetString("private")
	subscriber, _ := cmd.Flags().GetString("subscriber")

	sc := senderConfig{}
	scbody, _ := ioutil.ReadFile(senderConfigPath())
	_ = json.Unmarshal(scbody, &sc)

	if len(public) == 0 {
		public = sc.Public
	}

	if len(private) == 0 {
		private = sc.Private
	}

	if len(subscriber) == 0 {
		subscriber = sc.Subscriber
	}

	sender := webpushy.NewSender(&webpushy.SenderOptions{
		Identifier: subscriber,
		Keys: webpushy.SenderKeys{
			Public:  public,
			Private: private,
		},
		Serializer: func(input interface{}) ([]byte, error) {
			return input.([]byte), nil
		},
	})

	if len(payload) > 0 {
		err := sender.Send(endpoint, []byte(payload), time.Duration(ttl)*time.Second)
		if err != nil {
			panic(err)
		}
		return
	}

	scan := bufio.NewScanner(os.Stdin)
	for scan.Scan() {
		err := sender.Send(endpoint, scan.Bytes(), time.Duration(ttl)*time.Second)
		if err != nil {
			panic(err)
		}
	}
}

func sendInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "",
		Run:   sendInit,
	}

	cmd.Flags().String("subscriber", "", "")

	return cmd
}

func sendInit(cmd *cobra.Command, args []string) {
	subscriber, _ := cmd.Flags().GetString("subscriber")
	if len(subscriber) < 3 {
		fmt.Fprintln(os.Stderr, "Please provide an email address in the --subscriber flag")
		os.Exit(1)
	}

	keys, err := webpushy.GenerateSenderKeys(rand.Reader)
	if err != nil {
		panic(err)
	}

	sc := senderConfig{Public: keys.Public, Private: keys.Private, Subscriber: subscriber}
	scbody, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		panic(err)
	}

	path := senderConfigPath()
	err = ioutil.WriteFile(path, scbody, 0600)
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(os.Stderr, "Generated and wrote the following to %s:\n\n", path)
	fmt.Println(string(scbody))
}
