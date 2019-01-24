package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/joho/godotenv"
	mastodon "github.com/mattn/go-mastodon"
)

const postingInterval = 180 * 60 * 1000 * time.Millisecond // once every 3 hours
const markovLines string = "./markov-results.txt"
const postLength int = 500
const timInRuislip = "https://www.youtube.com/watch?v=6GT18lYRRDQ"
const lastPostFile = ".lastmasto"

// We do have some global state, because this is a one-file dealio.
var malcolm []string
var images []string
var client *mastodon.Client

// Knuth shuffle in place.
func shuffle(slice []string) {
	n := len(slice)
	for i := n - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}

func readGibberish() []string {
	content, err := ioutil.ReadFile(markovLines)
	if err != nil {
		log.Fatalf("Can't find %s", markovLines)
	}
	lines := strings.Split(string(content), "\n")
	// Trim empty final line.
	lines = lines[0 : len(lines)-1]
	shuffle(lines)
	return lines
}

func readImages() []string {
	var images []string

	files, err := ioutil.ReadDir("./images")
	if err != nil {
		// We're fine if we have no images to use.
		return images
	}

	for _, file := range files {
		name := file.Name()
		if isImage(name) {
			// TODO really should use a path lib
			images = append(images, fmt.Sprintf("./images/%s", name))
		}
	}

	shuffle(images)
	return images
}

func isImage(input string) bool {
	matches, _ := regexp.MatchString(".(gif|png|jpg)$", input)
	return matches
}

func ellipsize(input string, goal int) string {
	// TODO something less awful; split on spaces? this is unicode-aware, at least.
	runes := []rune(input)
	if len(runes) <= goal {
		return input
	}

	// oh my god
	var strs []string
	strs = append(strs, string(runes[:goal-1]))
	strs = append(strs, "…")
	return strings.Join(strs, "")
}

func chooseLine(goal int) string {
	if goal == 0 {
		goal = postLength
	}
	text := malcolm[0]
	malcolm = malcolm[1:]
	if len(malcolm) == 0 {
		malcolm = readGibberish()
	}
	return ellipsize(text, goal)
}

func postImage() {
	img := images[0]
	images = images[1:]
	if len(images) == 0 {
		// Upside of this approach is that we can drop new images in while the
		// program is running, and it should run forever. That's my excuse
		// and I'm sticking to it.
		images = readImages()
	}

	// upload media
	attachment, err := client.UploadMedia(context.Background(), img)
	if err != nil {
		log.Printf("failed to upload image media; skipping post; err=%s", err)
		return
	}
	var media []mastodon.ID
	media = append(media, attachment.ID)

	postToot(mastodon.Toot{
		Status:   chooseLine(postLength),
		MediaIDs: media,
	})
}

func recordTimestamp() {
	// write post time for process restart bookkeeping
	t := time.Now()
	ioutil.WriteFile(lastPostFile, []byte(t.Format("2006-01-02T15:04:05-0700")), 0777)
}

func shouldPostNow() bool {
	data, err := ioutil.ReadFile(lastPostFile)
	then, err := time.Parse("2006-01-02T15:04:05-0700", string(data))
	if err != nil {
		return true
	}
	return time.Since(then) > time.Duration(postingInterval/2)
}

func postToot(toot mastodon.Toot) (*mastodon.Status, error) {
	toot.SpoilerText = "Tuckerisms"
	status, err := client.PostStatus(context.Background(), &toot)
	if err != nil {
		log.Printf("failed to post status; err=%s", err)
	} else {
		log.Printf("posted: %s", toot.Status)
		recordTimestamp()
	}
	return status, err
}

func postPeriodically() {

	if rand.Intn(100) < 15 {
		postImage()
		return
	}

	line := chooseLine(postLength)
	timIsInRuislip, err := regexp.MatchString("Ruislip", line)
	if err == nil && timIsInRuislip {
		line = fmt.Sprintf("%s %s", line, timInRuislip)
	}
	postToot(mastodon.Toot{Status: line})
}

func main() {

	log.Println("---- Mastodon Markov Malcolm coming online")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var ok bool
	serverURL, ok := os.LookupEnv("MASTO_URL")
	if !ok {
		log.Fatal("You must provide a server url in MASTO_URL")
	}
	accessToken, ok := os.LookupEnv("MASTO_TOKEN")
	if !ok {
		log.Fatal("You must provide an access token in MASTO_TOKEN")
	}

	client = mastodon.NewClient(&mastodon.Config{
		Server:      serverURL,
		AccessToken: accessToken,
	})

	timeline, err := client.GetTimelineHome(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%d timeline entries read; env is okay", len(timeline))

	// Read in markov-generated lines & shuffle
	rand.Seed(time.Now().UnixNano())
	malcolm = readGibberish()
	log.Printf("Found %d lines to emit", len(malcolm))
	images = readImages()
	log.Printf("Found %d images to use", len(images))

	// Decide if we should post right away or hold off, based on time of last post.
	if shouldPostNow() {
		postPeriodically()
	}

	timerChannel := time.Tick(time.Duration(postingInterval))
	for range timerChannel {
		postPeriodically()
	}

	// TODO Read notifications & respond.
}
