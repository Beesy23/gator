package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/Beesy23/gator/internal/commands"
	"github.com/Beesy23/gator/internal/config"
	"github.com/Beesy23/gator/internal/database"
	"github.com/Beesy23/gator/internal/feeds"
	"github.com/Beesy23/gator/internal/middleware"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("Error reading config:", err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		fmt.Println("Error opening database:", err)
		os.Exit(1)
	}
	dbQueries := database.New(db)

	s := commands.State{
		Db:  dbQueries,
		Cfg: &cfg,
	}

	cmds := commands.NewCommands()

	cmds.Register("login", commands.HandlerLogin)
	cmds.Register("register", commands.HandlerRegister)
	cmds.Register("reset", commands.HandlerReset)
	cmds.Register("users", commands.HandlerUsers)
	cmds.Register("agg", feeds.HandlerAgg)
	cmds.Register("addfeed", middleware.MiddlewareLoggedIn(feeds.HandlerAddFeed))
	cmds.Register("feeds", middleware.MiddlewareLoggedIn(feeds.HandlerFeeds))
	cmds.Register("follow", middleware.MiddlewareLoggedIn(feeds.HandlerFollow))
	cmds.Register("following", middleware.MiddlewareLoggedIn(feeds.HandlerFollowing))
	cmds.Register("unfollow", middleware.MiddlewareLoggedIn(feeds.HandlerUnfollow))

	if len(os.Args) < 2 {
		fmt.Println("Not enough arguments given")
		os.Exit(1)
	}
	cmd := commands.Command{}
	cmd.Name, cmd.Args = os.Args[1], os.Args[2:]
	err = cmds.Run(&s, cmd)
	if err != nil {
		fmt.Printf("Error running %s command: %v\n", cmd.Name, err)
		os.Exit(1)
	}
}
