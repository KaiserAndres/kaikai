package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"
	"math"
	"regexp"
	"strconv"
	"strings"
)

const (
	version    float64 = 1.4
	startOrNeg string  = "(\\s|^)(-?)"
	end        string  = "(\\s|$)"
	dev string = "268908682266411009"
)

var (
	db            *sql.DB
	noFoundsError error          = errors.New("Not enough founds")
	celciusExp    *regexp.Regexp = regexp.MustCompile(magicRegexMaker("c"))
	fahrExp       *regexp.Regexp = regexp.MustCompile(magicRegexMaker("f"))
	metExp        *regexp.Regexp = regexp.MustCompile(magicRegexMaker("m"))
	ftExp         *regexp.Regexp = regexp.MustCompile(
		startOrNeg + "\\d+(\"|ft)(\\d*(')?)?" + end)
	conv map[string]bool = make(map[string]bool)
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

func magicRegexMaker(c string) string {
	return fmt.Sprintf(startOrNeg+"(\\d(.\\d)?)+(%s|%s)"+end, c,
		strings.ToTitle(c))
}

func command(s string, c string) bool {
	return strings.Contains(s, c) && s[0] == 'k'
}

func getCurrencyIdFromName(name string) (int64, error) {
	// returns the ID of the questioned currency
	var currencyID int64
	query := "SELECT id FROM currencies WHERE name=?"
	err := db.QueryRow(query, name).Scan(&currencyID)
	if err != nil {
		return -1, err
	} else {
		return currencyID, nil
	}
}

func hasWallet(uid int64, currency int64) bool {
	var useless int64
	query := "SELECT ammount FROM wallet WHERE owner=? AND currency=?"
	err := db.QueryRow(query, uid, currency).Scan(&useless)
	return err == nil
}

func createWallet(user int64, currId int64, ammount int64) {
	query := "INSERT INTO wallet VALUES (?,?,?)"
	_, err := db.Exec(query, user, currId, ammount)
	if err != nil {
		fmt.Print(err.Error())
		return
	}
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

//TODO add function to destroy currency from owner's wallet, fuck inflation

func sendMoney(source int64, target int64,
	currencyID int64, ammount int64) error {
	/* database interaction, returns an error if the transaction
	*  cannot happen, supposes both accounts have a wallet.
	 */
	var founds int64
	query := "SELECT ammount FROM wallet WHERE owner=? AND currency=?"
	err := db.QueryRow(query, source, currencyID).Scan(&founds)
	if founds < ammount {
		return noFoundsError
	} else {
		query = "UPDATE wallet SET ammount=ammount-? WHERE" +
			" owner=? AND currency=?"
		_, err = db.Exec(query, ammount, source, currencyID)
		if err != nil {
			fmt.Print(err.Error())
			return err
		}
		query = "UPDATE wallet SET ammount=ammount+? WHERE" +
			" owner=? AND currency=?"
		_, err = db.Exec(query, ammount, target, currencyID)
		return nil
	}
}

func addCirculation(ammount int64, currencyID int64) {
	query := "UPDATE currencies SET circulation=circulation+? WHERE id=?"
	_, err := db.Exec(query, ammount, currencyID)
	if err != nil {
		fmt.Print(err.Error())
	}
	return
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

func viewWallet(s *discordgo.Session, m *discordgo.MessageCreate) {
	if command(m.Content, "k!wallet") {
		message := fmt.Sprintf("Here's your wallet %s!\n", m.Author.Username)
		userID, err := strconv.ParseInt(m.Author.ID, 10, 64)
		if err != nil {
			fmt.Print(err.Error())
			return
		}
		query := "SELECT currency, ammount FROM wallet WHERE owner=?"
		rows, err := db.Query(query, userID)
		if err != nil {
			fmt.Print(err.Error())
			return
		}
		defer rows.Close()
		for rows.Next() {
			var (
				currencyID   int64
				currencyName string
				ammount      int64
			)
			if err := rows.Scan(&currencyID, &ammount); err != nil {
				fmt.Print(err.Error())
			}
			nameQ := "SELECT name FROM currencies WHERE id=?"
			err = db.QueryRow(nameQ, currencyID).Scan(&currencyName)
			if err != nil {
				fmt.Print(err.Error())
			}
			message += fmt.Sprintf("%s\t%d\n", currencyName, ammount)
		}
		_, err = s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			fmt.Print(err.Error())
		}
	}
}

func transferFounds(s *discordgo.Session, m *discordgo.MessageCreate) {
	if command(m.Content, "k!transfer") {
		// k!transfer kaiser kaiserBuck 99
		command := strings.Split(m.Content, " ")
		channy, _ := s.Channel(m.ChannelID)
		guild, _ := s.Guild(channy.GuildID)
		if len(command) < 4 {
			message := "Error: too few arguments"
			s.ChannelMessageSend(m.ChannelID, message)
			return
		}
		originalUser, err := strconv.ParseInt(m.Author.ID, 10, 64)
		if err != nil {
			fmt.Print(err.Error())
			return
		}
		targetUser := getClosest(guild.Members, strings.ToLower(command[1]), 0)
		target, err := strconv.ParseInt(targetUser.User.ID, 10, 64)
		if err != nil {
			fmt.Print(err.Error())
			return
		}
		currencyID, err := getCurrencyIdFromName(command[2])
		if err != nil {
			message := fmt.Sprintf("It appears there's"+
				" no currency called %s :(", command[2])
			_, err := s.ChannelMessageSend(m.ChannelID, message)
			if err != nil {
				fmt.Print(err.Error())
			}
			return
		}
		ammount, err := strconv.ParseInt(command[3], 10, 64)
		if err != nil {
			fmt.Print(err.Error())
			return
		}
		if ammount < 0 {
			ammount *= -1
		}
		if !hasWallet(originalUser, currencyID) {
			createWallet(originalUser, currencyID, 0)
		}
		if !hasWallet(target, currencyID) {
			createWallet(target, currencyID, 0)
		}
		err = sendMoney(originalUser, target, currencyID, ammount)
		if err == nil {
			message := fmt.Sprintf("%d %ss have been sent to %s! "+
				"Say thanks to %s!", ammount, command[2],
				targetUser.User.Username, m.Author.Username)
			_, err = s.ChannelMessageSend(m.ChannelID, message)
			if err != nil {
				fmt.Print(err.Error())
				return
			}
		} else if err == noFoundsError {
			message := "It appears you don't have enough moeny" +
				" fuck capitalism >:C"
			_, err = s.ChannelMessageSend(m.ChannelID, message)
			if err != nil {
				fmt.Print(err.Error())
			}
		} else {
			message := "I wasn't able to do the transfer," +
				" something went wrong! :("
			_, err = s.ChannelMessageSend(m.ChannelID, message)
			if err != nil {
				fmt.Print(err.Error())
				return
			}
		}
	}
}

func issueCurrency(s *discordgo.Session, m *discordgo.MessageCreate) {
	if command(m.Content, "k!mons") {
		command := strings.Split(m.Content, " ")
		user, err := strconv.ParseInt(m.Author.ID, 10, 64)
		if err != nil {
			fmt.Print(err.Error())
			return
		}
		if len(command) < 3 {
			message := "Error: not enough arguments"
			_, err := s.ChannelMessageSend(m.ChannelID, message)
			if err != nil {
				fmt.Print(err.Error())
				return
			}
			return
		}
		currName := command[1]
		ammount, err := strconv.ParseInt(command[2], 10, 64)
		if err != nil {
			message := "Error, " + command[2] + "is not a number"
			_, err := s.ChannelMessageSend(m.ChannelID, message)
			if err != nil {
				fmt.Print(err.Error())
				return
			}
			return
		}
		if ammount < 0 {
			ammount *= -1
		}
		currId, err := getCurrencyIdFromName(currName)
		if err == sql.ErrNoRows {
			message := "There's no currency called " + currName + " :C"
			_, err := s.ChannelMessageSend(m.ChannelID, message)
			if err != nil {
				fmt.Print(err.Error())
				return
			}
			return
		} else if err != nil {
			fmt.Print(err.Error())
			return
		}
		{
			// verify if the user is the woner
			var ownerId int64
			query := "SELECT creator FROM currencies WHERE id=?"
			err := db.QueryRow(query, currId).Scan(&ownerId)
			if err != nil {
				fmt.Print(err.Error())
				return
			}
			trueOwner, err := s.User(fmt.Sprintf("%d", ownerId))
			if err != nil {
				fmt.Print(err.Error())
				return
			}
			message := "Huh it appears the owner is " +
				trueOwner.Username +
				" better luck stealing next time ;)"
			if ownerId != user {
				s.ChannelMessageSend(m.ChannelID, message)
				return
			}
		}
		if !hasWallet(user, currId) {
			createWallet(user, currId, ammount)
			addCirculation(ammount, currId)
			message := "I created your wallet and added " +
				fmt.Sprintf("%d! :D", ammount)
			_, err = s.ChannelMessageSend(m.ChannelID, message)
			if err != nil {
				fmt.Print(err.Error())
				return
			}
			return
		}
		querry := "UPDATE wallet SET ammount=ammount+? WHERE owner=?" +
			" AND currency=?"
		_, err = db.Exec(querry, ammount, user, currId)
		addCirculation(ammount, currId)
		if err != nil {
			fmt.Print(err.Error())
			return
		}
		message := fmt.Sprintf("*Carefully adds %d", ammount) +
			" into your wallet* ;)"
		_, err = s.ChannelMessageSend(m.ChannelID, message)
	}
}

func registerCurrency(s *discordgo.Session, m *discordgo.MessageCreate) {
	if command(m.Content, "k!regCurr") {
		user := m.Author
		command := strings.Split(m.Content, " ")
		if len(command) == 1 {
			message := "Please give me a currency name!"
			_, err := s.ChannelMessageSend(m.ChannelID, message)
			if err != nil {
				fmt.Print(err.Error())
				return
			}
			return
		}
		currName := command[1]
		res, err := db.Query(
			"SELECT * FROM currencies")
		if err != nil {
			fmt.Print(err.Error())
			return
		}

		var (
			id          int
			name        string
			creator     int
			circulation int
		)

		defer res.Close()

		for res.Next() {
			err := res.Scan(&id, &name, &creator, &circulation)
			if err != nil {
				fmt.Print(err.Error())
				return
			}
			if name == currName {
				message := "The currency under the name <" +
					currName + "> already exists!"
				_, err := s.ChannelMessageSend(
					m.ChannelID, message)
				if err != nil {
					fmt.Print(err.Error())
					return
				}
				return
			}
		}
		db.Exec("INSERT INTO currencies VALUES (null,?,?,?)",
			currName, user.ID, 0)
		s.ChannelMessageSend(m.ChannelID, "Registering "+currName+"!")
	}
}

func versionCheck(s *discordgo.Session, m *discordgo.MessageCreate) {
	if command(m.Content, "k!version") {
		message := fmt.Sprintf("I am version %.1f :blush:", version)
		s.ChannelMessageSend(m.ChannelID, message)
	}
}

func intrusionSwitch(s *discordgo.Session, m *discordgo.MessageCreate) {
	if command(m.Content, "k!annoy") {
		if _, t := conv[m.ChannelID]; t == false {
			conv[m.ChannelID] = false
		}
		conv[m.ChannelID] = !conv[m.ChannelID]
		channel, err := s.Channel(m.ChannelID)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		message := fmt.Sprintf("Automatic unit conversion for "+
			"channel %s set to: %t", channel.Name, conv[m.ChannelID])
		s.ChannelMessageSend(m.ChannelID, message)
	}
}

func translate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot || !conv[m.ChannelID] {
		return
	}
	cFound := celciusExp.FindAllString(m.Content, -1)
	fFound := fahrExp.FindAllString(m.Content, -1)
	mFound := metExp.FindAllString(m.Content, -1)
	fAndiFound := ftExp.FindAllString(m.Content, -1)
	message := ""
	if len(cFound)+len(fFound) > 0 {
		message = "Hello, I'll convert this to other units :D\n"
	}

	if len(cFound) > 0 {
		for _, n := range cFound {
			cS := strings.Trim(n, " ")
			num, err := strconv.ParseFloat(string(cS[:len(cS)-1]), 64)
			if err != nil {
				fmt.Print(err.Error())
				return
			}
			message += fmt.Sprintf("%s translates to %.3fF\n", n, cTof(num))
		}
	}
	if len(fFound) > 0 {
		for _, n := range fFound {
			fS := strings.Trim(n, " ")
			num, err := strconv.ParseFloat(string(fS[:len(fS)-1]), 64)
			if err != nil {
				fmt.Print(err.Error())
				return
			}
			message += fmt.Sprintf("%s translates to %.3fC\n", n, fToc(num))
		}
	}
	if len(fAndiFound) > 0 {
		for _, n := range fAndiFound {
			var (
				ft float64 = 0.0
				in float64 = 0.0
			)
			parts := strings.Split(n, "\"")
			ft, err := strconv.ParseFloat(
				strings.Trim(parts[0], " "), 64)
			if err != nil {
				fmt.Print(err.Error())
				return
			}
			if len(parts) == 2 {
				inS := strings.Trim(parts[1], " ")
				if strings.Contains(inS, "'") {
					inS = inS[:len(inS)-1]
				}
				in, err = strconv.ParseFloat(inS, 64)
			}
			message += fmt.Sprintf(
				"%s translates to %.3f\n", n, fAndiTom(ft, in))
		}
	}
	if len(mFound) > 0 {
		for _, n := range mFound {
			mS := strings.Trim(n, " ")
			num, err := strconv.ParseFloat(string(mS[:len(mS)-1]), 64)
			if err != nil {
				fmt.Print(err.Error())
				return
			}
			feet := mTof(num)
			inch := mToi(num - float64(feet)*0.305)
			fs := ""
			if inch != 0 {
				fs = fmt.Sprintf("%d'", inch)
			}
			message += fmt.Sprintf(
				"%s translates to %d\""+fs+"\n", n, feet)
		}
	}
	if len(message) > 0 {
		s.ChannelMessageSend(m.ChannelID, message)
	}
}

func fAndiTom(feet, inch float64) float64 {
	return feet*0.305 + inch*0.025
}

func mToi(n float64) int {
	in := n * 39.37
	if d, f := math.Modf(in); f > 0.5 {
		return int(d) + 1
	} else {
		return int(d)
	}
}

func mTof(n float64) int {
	return int(n * 3.281)
}

func fToc(n float64) float64 {
	return (n - 32.0) * 5.0 / 9.0
}

func cTof(n float64) float64 {
	return (9.0/5.0)*n + 32.0
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
