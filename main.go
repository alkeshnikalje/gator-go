package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
	"html"

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

func handlerReset(s *state,cmd command) error {
	ctx := context.Background()

	err := s.db.DeleteUsers(ctx)
	if err != nil {
		log.Fatal("error deleting users",err)
	}
	fmt.Println("users table has been reset successfully")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	ctx := context.Background()
	
	users,err := s.db.GetUsers(ctx)
	if err != nil {
		log.Fatal("error getting users",err)
	}

	if len(users) == 0 {
		fmt.Println("No users found")
		os.Exit(1)
	}

	for _,user := range users {
		if user.Name == s.cfg.CurrentUserName {
			fmt.Println("*", user.Name,"(current)")
		}else{
			fmt.Println("*",user.Name)
		}
	}
	return nil
}

func handlerAgg(s *state,cmd command) error {
	url := "https://www.wagslane.dev/index.xml"
	ctx := context.Background()
	rssFeed,err := fetchFeed(ctx,url)
	if err != nil {
		return err
	}
	fmt.Println(rssFeed)
	return nil
}

func handlerAddFeed(s *state,cmd command) error {
	ctx := context.Background()
	user,err := s.db.GetUser(ctx,s.cfg.CurrentUserName)		
	if err != nil {
		return err
	}

	if len(cmd.args) < 2 {
		log.Fatal("not enough arguments were provided.")
	}
	
	// Create params for new feed
    arg := database.CreateFeedParams{
        ID:        uuid.New(),
    	Name:      cmd.args[0],
		Url:       cmd.args[1],
		UserID:    user.ID,
    }

	feed,err := s.db.CreateFeed(ctx,arg)
	if err != nil {
		fmt.Println("error creating a feed",err)
		return err
	}
	fmt.Println(feed)
	return nil

}

func handlerFeeds(s *state,cmd command) error {
	if len(cmd.args) > 0 {
		log.Fatal("feeds command needs to be used without any args.")
		return nil
	}
	ctx := context.Background()

	feeds,err := s.db.GetFeeds(ctx)
	if err != nil {
		return err
	}
	for _,feed := range feeds {
		fmt.Println(feed)
	}
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

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed,error) {
	req,err := http.NewRequestWithContext(ctx,"GET",feedURL,nil)
	if err != nil {
		fmt.Println("error creating request",err)
		return &RSSFeed{},err
	}
	req.Header.Set("User-Agent", "gator")
	client := &http.Client{}
	resp,err := client.Do(req)
	if err != nil {
		fmt.Println("error making a request",err)
		return &RSSFeed{},err
	}
	defer resp.Body.Close()
	
	data,err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error reading response body",err) 
		return &RSSFeed{},err
	}

	var rss RSSFeed
	err = xml.Unmarshal(data,&rss)
	if err != nil {
		fmt.Println("error unmarshaling xml",err)
		return &RSSFeed{},err
	}

	// Decode HTML entities in channel title & description
	rss.Channel.Title = html.UnescapeString(rss.Channel.Title)
	rss.Channel.Description = html.UnescapeString(rss.Channel.Description)		

	// Decode HTML entities in each item
	for i := range rss.Channel.Item {
    	rss.Channel.Item[i].Title = html.UnescapeString(rss.Channel.Item[i].Title)
    	rss.Channel.Item[i].Description = html.UnescapeString(rss.Channel.Item[i].Description)
	}

	return &rss,nil
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
	cmds.register("users",handlerUsers)
	cmds.register("reset",handlerReset)
	cmds.register("agg",handlerAgg)
	cmds.register("addfeed",handlerAddFeed)
	cmds.register("feeds",handlerFeeds)
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



















