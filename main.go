package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/deoreal/gator/internal/config"
	"github.com/deoreal/gator/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

const configFileName = ".gatorconfig.json"

type state struct {
	conf *config.Config
	db   *database.Queries
}

type command struct {
	name string
	args []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

func getConfigFilePath() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to read homedir: %s", err)
	}
	return homedir + "/" + configFileName, nil
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		fmt.Println("username is required")
		os.Exit(1)
	}

	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	if slices.Contains(users, cmd.args[0]) {

		s.conf.CurrentUserName = cmd.args[0]
		config.WriteConfig(*s.conf)
		fmt.Println("user has been set to: ", cmd.args[0])
	} else {
		return fmt.Errorf("unknown user")
	}
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range users {
		if user == s.conf.CurrentUserName {
			fmt.Println(user, "(current)")
		} else {
			fmt.Println(user)
		}
	}

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		fmt.Println("username is required")
		os.Exit(1)

	}

	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	if slices.Contains(users, cmd.args[0]) {
		return fmt.Errorf("user %s already in database", cmd.args[0])
	}
	//	fmt.Println(cmd.args)
	err = s.db.CreateUser(context.Background(), database.CreateUserParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: cmd.args[0]})
	if err != nil {
		return err
	}
	fmt.Printf("user %s has been added to the database\n", cmd.args[0])

	s.conf.CurrentUserName = cmd.args[0]

	s.conf.SetUser(cmd.args[0])
	config.WriteConfig(*s.conf)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.Reset(context.Background())
	if err != nil {
		return fmt.Errorf("failed to truncate users table: %s", err)
	}
	fmt.Printf("truncated the users table: %v\n", cmd)
	return nil
}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.handlers[cmd.name]
	if !ok {
		return fmt.Errorf("command not registered")
	}

	return handler(s, cmd)
}

func (c *commands) register(name string, f func(s *state, cmd command) error) {
	c.handlers[name] = f
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("not enough arguments")
		os.Exit(1)
	}

	c := &commands{
		handlers: make(map[string]func(*state, command) error),
	}
	s := &state{}
	var err error

	s.conf, err = config.ReadConfig()
	if err != nil {
		fmt.Printf("failed to read config: %s\n", err)
	}
	c.register("login", handlerLogin)
	c.register("register", handlerRegister)
	c.register("reset", handlerReset)
	c.register("users", handlerUsers)

	cmd := command{
		name: os.Args[1],
		args: os.Args[2:],
	}
	db, err := sql.Open("postgres", s.conf.DBURL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	dbQueries := database.New(db)

	s.db = dbQueries

	if err = c.run(s, cmd); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
