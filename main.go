package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var auctions map[string]auctionChar

type auctionChar struct {
	charName   string
	level      string
	voc        string
	gender     string
	world      string
	start      string
	end        string
	winningBid string
	stats      charStats
}

type charStats struct {
	axe      int
	club     int
	distance int
	fishing  int
	fist     int
	magic    int
	shield   int
	sword    int
	creation string
	gold     string
	points   string
}

func getHTML(url string) *goquery.Document {

	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	return doc
}

func main() {
	auctions = make(map[string]auctionChar)
	//Load first page
	doc := getHTML("https://www.tibia.com/charactertrade/?subtopic=pastcharactertrades")

	//Get total number of pages
	pages, err := strconv.Atoi(getPages(doc))
	if err != nil {
		fmt.Println("Error: ", err)
	}

	//Load each page of auctions, oldest first
	for i := pages; i > 0; i-- {
		url := "https://www.tibia.com/charactertrade/?subtopic=pastcharactertrades&currentpage=" + fmt.Sprint(i)
		fmt.Println("Loading: ", url)
		doc = getHTML(url)
		fmt.Println("Parsing...")
		getData(doc)
		fmt.Println("Finished.")
	}
	fmt.Println("Saving to CSV.")
	//Save results to CSV
	f, err := os.Create("history.csv")
	w := csv.NewWriter(f)
	for _, r := range auctions {
		s := []string{r.charName, r.level, r.voc, r.gender, r.world, r.start, r.end, fmt.Sprint(strings.Replace(r.winningBid, ",", "", -1))}
		w.Write(s)
	}
	w.Flush()
}

func getData(doc *goquery.Document) {
	doc.Find(".Auction").Each(func(i int, s *goquery.Selection) {
		c := auctionChar{}
		nametmp := s.Find(".AuctionCharacterName")
		c.charName = nametmp.Find("a").Text()
		auctionheader := s.Find(".AuctionHeader")
		c.level = strings.TrimSpace(strings.Trim(strings.SplitAfterN(strings.SplitAfterN(auctionheader.Text(), ":", 4)[1], "|", 2)[0], "|"))
		c.voc = strings.TrimSpace(strings.Trim(strings.SplitAfterN(strings.SplitAfterN(auctionheader.Text(), ":", 4)[2], "|", 2)[0], "|"))
		c.gender = strings.TrimSpace(strings.Trim(strings.SplitAfterN(strings.SplitAfterN(auctionheader.Text(), ":", 4)[2], "|", 3)[1], "|"))
		c.world = strings.TrimSpace(strings.SplitAfterN(strings.SplitAfterN(auctionheader.Text(), ":", 4)[3], "|", 2)[0])
		auctionstartendbid := s.Find(".ShortAuctionDataValue").Text()
		c.start = strings.SplitAfterN(auctionstartendbid, "CET", 3)[0]
		c.end = strings.SplitAfterN(auctionstartendbid, "CET", 3)[1]
		c.winningBid = strings.SplitAfterN(auctionstartendbid, "CET", 3)[2]
		//fmt.Println(c.charName + ", " + c.level + ", " + c.voc + ", " + c.gender + ", " + c.world + ", " + auctionstart + ", " + auctionend + ", " + c.winningBid)
		auctions[c.charName] = c
	})
}

func getPages(doc *goquery.Document) string {
	var splits []string
	doc.Find(".PageLink").Each(func(i int, s *goquery.Selection) {
		url, ok := s.Find("a").Attr("href")
		if ok == true {
			splits = strings.SplitAfterN(url, "=", 3)
		}
	})
	return splits[2]
}
