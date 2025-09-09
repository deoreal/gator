package main

import (
	"fmt"
	"os"

	"github.com/deoreal/gator/internal/config"
)

const configFileName = ".gatorconfig.json"

type state struct {
	conf *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	cmds map[string]func(*state, command) error
}

func getConfigFilePath() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to read homedir: %s", err)
	}
	return homedir + "/" + configFileName, nil
}

func handlerLogins(st *state, cmd command) error {
	if cmd.args == nil {
		return fmt.Errorf("command is missing")
	}

	switch cmd.name {
	case "login":
		cmd.args[0] = st.conf.CurrentUserName
		fmt.Println("user has been set to: ", cmd.args[0])
	}

	return nil
}

func (c *commands) run(s *state, cmd command) error {
	return nil
}

func (c *commands) register(name string, f func(s *state, cmd command) error) {
	c.cmds[name] = f
}

func login(s *state, cmd command) error {
	return nil
}

func main() {
	c := &commands{}
	s := &state{}
	var err error

	s.conf, err = config.ReadConfig()
	if err != nil {
		fmt.Printf("failed to read config: %s\n", err)
	}

	c.register("login", login)

	if len(os.Args) < 2 {
		fmt.Println("not enough arguments")
		os.Exit(1)
	}

	kommando := command{name: os.Args[0], args: os.Args[1:]}

	/*
	   err = config.WriteConfig(*cfg)

	   	if err != nil {
	   		fmt.Printf("failed to write config: %s", err)
	   	}

	   cfg, err = config.ReadConfig()

	   	if err != nil {
	   		fmt.Printf("failed to read config: %s\n", err)
	   	}

	   fmt.Println(cfg.DBURL, cfg.CurrentUserName)
	*/
}
