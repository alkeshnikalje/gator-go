package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/alkeshnikalje/gator-go/internal/config"
	"github.com/alkeshnikalje/gator-go/internal/database"
	_ "github.com/lib/pq"
)

type state struct {
	db    *database.Queries
	cfg   *config.Config
}

type command struct {
	name 	string
	args 	[]string
} 

func handlerRegister(s *state,cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("username not provided")
	}
	
	ctx := context.Background()

	// check if user already exists
	existing, err := s.db.GetUser(ctx,cmd.args[0])
	if err != nil {
		if err == sql.ErrNoRows {
	    	// Create params for new user
    		arg := database.CreateUserParams{
        		ID:        uuid.New(),
        		CreatedAt: time.Now(),
        		UpdatedAt: time.Now(),
        		Name:      cmd.args[0],
    		}	

			user, err := s.db.CreateUser(ctx, arg)
    		if err != nil {
        		log.Fatal("failed to create user:", err)
    		}	
			s.cfg.SetUser(user.Name)
			fmt.Println("user was created successfully",user)
			return nil
		}else {
			log.Fatal("error checking user",err)
		}
	}
	log.Fatal("user already exists",existing)
	return nil	
}

func handlerLogin(s *state,cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("username not provided")
	}

	ctx := context.Background()

	user,err := s.db.GetUser(ctx,cmd.args[0])
	if err != nil {
		if err == sql.ErrNoRows {
			log.Fatal("user does not exist, please register")
		}
	}

	err = s.cfg.SetUser(user.Name) 
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("current user has been set to",user.Name)
	return nil
}

type commands struct {
	commandHandlers 	map[string]func (*state,command) error
}

func (c *commands) run(s *state,cmd command) error {
	commandHandler,exists := c.commandHandlers[cmd.name]
	if exists {
		return commandHandler(s,cmd)
	}
	
	return errors.New("command not found")
}

func (c *commands) register(name string, f func(*state,command) error) {
	_, exists := c.commandHandlers[name]
	if !exists {
		c.commandHandlers[name] = f
	}
}

func main () {
	cfg := config.Read()
		
	db,err := sql.Open("postgres",cfg.DbUrl)
	if err != nil {
		fmt.Println("error connecting to the db",err)
		return
	}

	dbQueries := database.New(db)
		
	s := state{
		db: dbQueries,
		cfg: &cfg,
	}

	cmds := commands{
		commandHandlers: make(map[string]func(*state,command) error),
	}

	cmds.register("login",handlerLogin)
	cmds.register("register",handlerRegister)
    args := os.Args
    if len(args) < 2 {
        fmt.Println("No arguments provided")
		os.Exit(1)
    }
	cmdName := args[1]
	cmdArgs := args[2:]

	cmd := command {
		name: cmdName,
		args: cmdArgs,
	}

	err = cmds.run(&s,cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}



















