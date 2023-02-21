package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/vladikamira/funda-exporter/scraper"
)

var (
	telegramToken          = flag.String("telegramToken", "", "Telegram token to use")
	telegramChatId         = flag.Int64("telegranChatId", 0, "Telegram Chat ID")
	checkIntervalSeconds   = flag.Int("checkInterval", 1800, "Check interval in Seconds")
	fileNameToStoreResults = flag.String("fileNameToStoreResutls", "results.gob", "File to store/persist results from the previous run")

	FakeUserAgent           = flag.String("fakeUserAgent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36", "A fake User-Agent")
	FundaSearchUrl          = flag.String("fundaSearchUrl", "https://www.funda.nl/koop/amstelveen/300000-440000/70+woonopp/2+slaapkamers/", "Funda search page with paramethers")
	ScrapeDelayMilliseconds = flag.Int("scrapeDelayMilliseconds", 1000, "Delay between scrapes. Let's not overload Funda :)")
	PostCodesString         = flag.String("postCodes", "1186", "Post Codes to limit area of search")
)

type Message struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

func SaveStructToFile(fileName string, data []scraper.House) {

	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Couldn't open file")
	}

	dataEncoder := gob.NewEncoder(file)
	dataEncoder.Encode(data)

	log.Info("Saved elements in the struct into file: ", len(data))
	file.Close()
}

func ReadStructFromFile(fileName string) []scraper.House {

	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Couldn't open file")
	}

	data := []scraper.House{}

	dataDecoder := gob.NewDecoder(file)
	err = dataDecoder.Decode(&data)

	if err != nil {
		fmt.Println(err)
		//		os.Exit(1)
	}

	file.Close()

	log.Info("Reastored elements in the struct from file: ", len(data))

	return data
}

func SendMessage(url string, message *Message) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}
	response, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer func(body io.ReadCloser) {
		if err := body.Close(); err != nil {
			log.Println("failed to close response body")
		}
	}(response.Body)
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send successful request. Status was %q", response.Status)
	}
	return nil
}

func main() {

	// parse flags
	flag.Parse()

	// Setup better logging
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	if *telegramToken == "" {
		fmt.Println("Token is configured, please set it up")
		os.Exit(1)
	}

	if *telegramChatId == 0 {
		fmt.Println("Telegram Chat ID is 0, which is weird. You should find your personal ChatID to use")
		os.Exit(1)
	}

	PostCodes := []string{}
	// convert String of PostCodes into array of strings
	if len(*PostCodesString) > 0 {
		PostCodes = strings.Split(*PostCodesString, ",")
	}

	// a loop with the Sleep
	for {
		newResults := []scraper.House{}
		oldResults := ReadStructFromFile(*fileNameToStoreResults)

		// run scraper
		scraper.RunScraper(&newResults, FakeUserAgent, FundaSearchUrl, ScrapeDelayMilliseconds, &PostCodes)

		// compare results only when there was some restored one
		if len(oldResults) > 0 {
			oldUrlMap := map[string]string{}
			diffResults := []scraper.House{}

			for _, house := range oldResults {
				oldUrlMap[house.Link] = "exist"
			}

			// compare current and the privios
			for _, r := range newResults {
				_, ok := oldUrlMap[r.Link]
				if !ok {
					diffResults = append(diffResults, r)
				}
			}

			// there are some updates. Let's send a message
			if len(diffResults) > 0 {
				log.Info("YEAY! we found some new stuff")

				text := "Found results: \n"
				for _, r := range diffResults {
					text += r.Link + "\n"
				}

				fundaMessage := Message{
					ChatID: *telegramChatId,
					Text:   text,
				}

				// send message to telegram
				SendMessage("https://api.telegram.org/bot"+*telegramToken+"/sendMessage", &fundaMessage)

			}

		}

		// save results
		SaveStructToFile(*fileNameToStoreResults, newResults)

		// wait
		log.Infof("Wait %v Seconds before next check", *checkIntervalSeconds)
		time.Sleep(time.Duration(*checkIntervalSeconds) * time.Second)
	}
}
