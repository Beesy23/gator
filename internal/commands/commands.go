package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Beesy23/gator/internal/config"
	"github.com/Beesy23/gator/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type State struct {
	Cfg *config.Config
	Db  *database.Queries
}

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	handlers map[string]func(*State, Command) error
}

func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}

	_, err := s.Db.GetUser(context.Background(), cmd.Args[0])
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if err := s.Cfg.SetUser(cmd.Args[0]); err != nil {
		return err
	}

	fmt.Printf("Logged in as user: %s\n", cmd.Args[0])
	return nil
}

func HandlerRegister(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}

	params := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      cmd.Args[0],
	}

	result, err := s.Db.CreateUser(context.Background(), params)
	if err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == "23505" {
			fmt.Println("User already exists")
			os.Exit(1)
		}
		return err
	}

	if err := s.Cfg.SetUser(cmd.Args[0]); err != nil {
		return err
	}
	fmt.Printf("User created: %s.\n%v: %v\n", cmd.Args[0], result.Name, result.ID)
	return nil
}

func HandlerUsers(s *State, cmd Command) error {
	users, err := s.Db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, name := range users {
		if name == s.Cfg.CurrentUserName {
			fmt.Printf("%s (current)\n", name)
		} else {
			fmt.Println(name)
		}

	}
	return nil
}

func NewCommands() *Commands {
	return &Commands{
		handlers: make(map[string]func(*State, Command) error),
	}
}

func (c *Commands) Register(name string, f func(*State, Command) error) {
	c.handlers[name] = f
}

func (c *Commands) Run(s *State, cmd Command) error {
	handler, exists := c.handlers[cmd.Name]
	if !exists {
		return fmt.Errorf("command doesn't exist")
	}
	err := handler(s, cmd)
	return err
}
