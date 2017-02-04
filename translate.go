package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"math"
	"regexp"
	"strconv"
	"strings"
)

const (
	startOrNeg string = "(\\s|^)(-?)"
	end        string = "(\\s|$)"
)

var (
	celciusExp *regexp.Regexp = regexp.MustCompile(magicRegexMaker("c"))
	fahrExp    *regexp.Regexp = regexp.MustCompile(magicRegexMaker("f"))
	metExp     *regexp.Regexp = regexp.MustCompile(magicRegexMaker("m"))
	ftExp      *regexp.Regexp = regexp.MustCompile(
		startOrNeg + "\\d+(\"|ft)(\\d*(')?)?" + end)
	conv map[string]bool = make(map[string]bool)
)

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

func magicRegexMaker(c string) string {
	return fmt.Sprintf(startOrNeg+"(\\d(.\\d)?)+(%s|%s)"+end, c,
		strings.ToTitle(c))
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
