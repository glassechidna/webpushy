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
	path, err := homedir.Expand(fmt.Sprintf("~/.pushy/%s.json", name))
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
		Short: "",
		Run:   recv,
	}

	cmd.Flags().String("name", "", "")
	cmd.Flags().Int("limit", 0, "")
	cmd.Flags().Int("timeout", 0, "")

	cmd.AddCommand(recvInitCmd())
	return cmd
}

func recv(cmd *cobra.Command, args []string) {
	name, _ := cmd.Flags().GetString("name")
	limit, _ := cmd.Flags().GetInt("limit")
	timeout, _ := cmd.Flags().GetInt("timeout")

	ctx := context.Background()
	if timeout > 0 {
		ctx, _ = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	}

	rc := receiverConfig{}
	rcbody, err := ioutil.ReadFile(receiverConfigPath(name))
	if err != nil {
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
		Short: "",
		Run:   recvInit,
	}

	cmd.Flags().String("name", "", "")
	cmd.Flags().String("public", "", "")

	return cmd
}

func recvInit(cmd *cobra.Command, args []string) {
	name, _ := cmd.Flags().GetString("name")
	public, _ := cmd.Flags().GetString("public")

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
