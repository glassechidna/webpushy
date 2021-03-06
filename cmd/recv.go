package main

import (
	"context"
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

type receiverConfig struct {
	Name     string
	Id       string
	Endpoint string
	Public   string
}

func receiverConfigPath(name string) string {
	path, err := homedir.Expand(fmt.Sprintf("~/.webpushy/%s.json", name))
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll(filepath.Dir(path), 0700)
	if err != nil {
		panic(err)
	}

	return path
}

func recvCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recv",
		Run:   recv,
		Short: "Listen for messages from sender",
		Long: `
Listens for messages from sender using a pre-defined profile.

Profile should be first created using 'webpushy recv init --name ...' and
then the same name should be used to invoke 'webpushy recv --name ...'.
`,
	}

	cmd.Flags().String("name", "", "Name of profile already created by recv init")
	cmd.Flags().Int("limit", 0, "Optional. Maximum number of messages to receive before exiting")
	cmd.Flags().Int("timeout", 0, "Optional. Maximum number of seconds to run before exiting")

	cmd.AddCommand(recvInitCmd())
	return cmd
}

func recv(cmd *cobra.Command, args []string) {
	name, _ := cmd.Flags().GetString("name")
	limit, _ := cmd.Flags().GetInt("limit")
	timeout, _ := cmd.Flags().GetInt("timeout")

	if len(name) == 0 {
		cmd.Help()
		os.Exit(1)
	}

	ctx := context.Background()
	if timeout > 0 {
		ctx, _ = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	}

	rc := receiverConfig{}
	rcbody, err := ioutil.ReadFile(receiverConfigPath(name))
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "You must first create a profile using 'webpushy recv init'")
			os.Exit(1)
		}
		panic(err)
	}

	err = json.Unmarshal(rcbody, &rc)
	if err != nil {
		panic(err)
	}

	opts := &webpushy.ReceiverOptions{
		Id: webpushy.ReceiverId{
			Id:       rc.Id,
			Endpoint: rc.Endpoint,
		},
		PublicKey: rc.Public,
		Deserializer: func(bytes []byte) (interface{}, error) {
			return bytes, nil
		},
	}
	recv, err := webpushy.NewReceiver(opts)
	if err != nil {
		panic(err)
	}

	go func() {
		err := recv.Run(ctx)
		if err != nil {
			panic(err)
		}
	}()

	count := 0
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-recv.Receive():
			bytes := msg.([]byte)
			fmt.Println(string(bytes))
			count++
			if limit > 0 && count >= limit {
				return
			}
		}
	}
}

func recvInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Run:   recvInit,
		Short: "Create new listening profile",
		Long: `
Prepares a new profile for usage by 'webpushy send'.

The --public KEY should be a key generated by 'webpushy send init'. This controls
who is able to send messages to this profile.
`,
	}

	cmd.Flags().String("name", "", "Name to associate with this endpoint URL")
	cmd.Flags().String("public", "", "Public key allowed to send messages")

	return cmd
}

func recvInit(cmd *cobra.Command, args []string) {
	name, _ := cmd.Flags().GetString("name")
	public, _ := cmd.Flags().GetString("public")

	if len(name) == 0 || len(public) == 0 {
		cmd.Help()
		os.Exit(1)
	}

	opts := &webpushy.ReceiverOptions{
		PublicKey: public,
		Deserializer: func(bytes []byte) (interface{}, error) {
			return bytes, nil
		},
	}

	_, err := webpushy.NewReceiver(opts)
	if err != nil {
		panic(err)
	}

	rc := receiverConfig{Name: name, Id: opts.Id.Id, Endpoint: opts.Id.Endpoint, Public: public}
	rcbody, err := json.MarshalIndent(rc, "", "  ")
	if err != nil {
		panic(err)
	}

	path := receiverConfigPath(name)
	err = ioutil.WriteFile(path, rcbody, 0600)
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(os.Stderr, "Generated and wrote the following to %s:\n\n", path)
	fmt.Println(string(rcbody))
}
