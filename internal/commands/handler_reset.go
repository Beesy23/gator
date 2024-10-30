package commands

import (
	"context"
	"fmt"
)

func HandlerReset(s *State, cmd Command) error {
	err := s.Db.Reset(context.Background())
	if err != nil {
		return err
	}
	fmt.Println("Database successfully reset")
	return nil
}
