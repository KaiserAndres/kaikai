package main

import (
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"
	"strings"
)

const (
	version float64 = 1.4
	dev     string  = "268908682266411009"
)

var (
	db *sql.DB
)

func main() {
	token := "MjA5MDcwNjk5MDY5ODMzMjE3.Cznkrw.WpgrVava7-U8Cg_-0CIkdVj3wMI"
	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Print(err.Error())
		return
	}

	db, err = sql.Open("sqlite3", "deb.db")
	if err != nil {
		fmt.Print(err.Error())
	}

	fmt.Print(db.Ping())

	discord.AddHandler(sayFuckU)
	discord.AddHandler(registerCurrency)
	discord.AddHandler(issueCurrency)
	discord.AddHandler(transferFounds)
	discord.AddHandler(viewWallet)
	discord.AddHandler(helpMe)
	discord.AddHandler(translate)
	discord.AddHandler(intrusionSwitch)
	discord.AddHandler(versionCheck)
	discord.AddHandler(report)
	err = discord.Open()
	if err != nil {
		fmt.Print(err.Error())
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	// Simple way to keep program running until CTRL-C is pressed.
	<-make(chan struct{})
	return
}

func command(s string, c string) bool {
	return strings.Contains(s, c) && s[0] == 'k'
}

func getClosest(memberList []*discordgo.Member,
	search string, i int) *discordgo.Member {
	//search both usernames and nicks in current server
	var acList []*discordgo.Member
	//makes it's size 0 to know when nothing was found
	for _, person := range memberList {
		uName := strings.ToLower(person.User.Username)
		nick := strings.ToLower(person.Nick)

		if len(search) <= len(uName) && uName[i] == search[i] {
			acList = append(acList, person)
		} else if len(search) <= len(nick) && nick[i] == search[i] {
			acList = append(acList, person)
		}
	}
	if len(acList) == 1 {
		//element found
		return acList[0]
	} else if len(acList) == 0 {
		//nothing fully matched, give closest
		return memberList[0]
	} else if i == len(search)-1 {
		//ran out of things to check with
		return memberList[0]
	} else {
		return getClosest(acList, search, i+1)
	}
}

func helpMe(s *discordgo.Session, m *discordgo.MessageCreate) {
	if command(m.Content, "k!help") {
		message := "Hello I am the kaisebot! Here are my commands and" +
			"how to use me! :D\n" +
			"```" +
			"k!help: gives you this message\n" +
			"k!wallet: tells you your current founds for each currency!\n" +
			"k!transfer <user> <currency> <ammount>: sends some of your" +
			" money to the user you want\n" +
			"k!mons <currency> <ammount> issues some of your currency" +
			" beware of inflation!\n" +
			"k!regCurr <currency>: Registers your very own currency!\n" +
			"k!annoy: will turn on or off automatic unit conversion!\n" +
			"k!version: will display the current bot version! \n" +
			"k!report <screenshot>: will tell my developer there " +
			"was a problem with me, please send them any issue " +
			"there is, no matter how small!\n" +
			"```"
		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			fmt.Print(err.Error())
		}
	}
}

func versionCheck(s *discordgo.Session, m *discordgo.MessageCreate) {
	if command(m.Content, "k!version") {
		message := fmt.Sprintf("I am version %.1f :blush:", version)
		s.ChannelMessageSend(m.ChannelID, message)
	}
}

func report(s *discordgo.Session, m *discordgo.MessageCreate) {
	if command(m.Content, "k!report") {
		content := m.Attachments
		if len(content) != 1 {
			_, err := s.ChannelMessageSend(m.ChannelID,
				"Please provide a screenshot.")
			if err != nil {
				fmt.Print(err.Error())
			}
			return
		}

		_, err := s.ChannelMessageSend(dev, content[0].URL)
		if err != nil {
			fmt.Print(err.Error())
			return
		}
		_, err = s.ChannelMessageSend(m.ChannelID,
			"I have notified Kaiser of your issue")
		if err != nil {
			fmt.Print(err.Error())
			return
		}
	}
}

func sayFuckU(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Content == "k!fuck" {
		source := m.ChannelID
		content := "FUCK YOU"
		_, err := s.ChannelMessageSend(source, content)
		if err != nil {
			fmt.Print(err.Error())
			return
		}
	}
}
