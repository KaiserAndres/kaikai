package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"
	"strconv"
	"strings"
)

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


func addCirculation(ammount int64, currencyID int64) {
	query := "UPDATE currencies SET circulation=circulation+? WHERE id=?"
	_, err := db.Exec(query, ammount, currencyID)
	if err != nil {
		fmt.Print(err.Error())
	}
	return
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

