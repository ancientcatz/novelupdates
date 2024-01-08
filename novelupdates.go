package novelupdates

import (
	"encoding/json"
	"fmt"
	// "log"
	"net/http"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/imroc/req/v3"
)

type SearchResult struct {
	Title        string              `json:"title"`
	ID           string              `json:"id"`
	URL          string              `json:"url"`
	Image        string              `json:"image"`
	SearchRating string              `json:"search_rating"`
	Description  string              `json:"description"`
	Releases     string              `json:"releases"`
	UpdateFreq   string              `json:"update_freq"`
	NuReaders    string              `json:"nu_readers"`
	NuReviews    string              `json:"nu_reviews"`
	LastUpdated  string              `json:"last_updated"`
	Genres       []map[string]string `json:"genres"`
}

type SeriesResult struct {
	Title                string              `json:"title"`
	Image                string              `json:"image"`
	Type                 map[string]string   `json:"type"`
	Genre                []map[string]string `json:"genre"`
	Rating               []map[string]string `json:"rating"`
	Language             map[string]string   `json:"language"`
	Authors              []map[string]string `json:"authors"`
	Artists              []map[string]string `json:"artists"`
	Year                 string              `json:"year"`
	Status               string              `json:"status"`
	Licensed             string              `json:"licensed"`
	CompletelyTranslated string              `json:"completely_translated"`
	OriginalPublisher    map[string]string   `json:"original_publisher"`
	EnglishPublisher     map[string]string   `json:"english_publisher"`
	ReleaseFreq          string              `json:"release_freq"`
	Description          string              `json:"description"`
	AssociatedNames      []string            `json:"associated_names"`
	Groups               []map[string]string `json:"groups"`
	Tags                 []map[string]string `json:"tags"`
}

// SearchSeries performs a search for a series on NovelUpdates
// It takes a series name as a parameter and returns the parsed results.
func SearchSeries(seriesName string) ([]*SearchResult, error) {
	// Construct the URL for searching a series
	searchURL := "https://novelupdates.com/?s=" + seriesName
	resp, err := makeGetRequest(searchURL)
	if err != nil {
		return []*SearchResult{}, err
	}

	return parseSearch(resp)
}

// SearchSeriesJSON performs a search for a series on NovelUpdates
// It takes a series name as a parameter and returns the search results in JSON format as a byte slice.
func SearchSeriesJSON(seriesName string) ([]byte, error) {
	// Construct the URL for searching a series
	searchURL := "https://novelupdates.com/?s=" + seriesName
	resp, err := makeGetRequest(searchURL)
	if err != nil {
		return nil, err
	}

	// Parse and return the search results in JSON format
	results, err := parseSearch(resp)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

// ToJSON converts SearchResult to JSON format with customizable indentation.
// If the indent parameter is provided and is a positive integer within a reasonable range,
// it determines the number of spaces for indentation; otherwise, it defaults to 2 spaces.
func (searchResults *SearchResult) ToJSON(indent ...int) ([]byte, error) {
	customIndent := 2
	if len(indent) > 0 {
		requestedIndent := indent[0]
		if requestedIndent > 0 && requestedIndent <= 8 {
			customIndent = requestedIndent
		}
	}

	jsonOutput, err := json.MarshalIndent(searchResults, "", strings.Repeat(" ", customIndent))
	if err != nil {
		return nil, err
	}
	return jsonOutput, nil
}

// GetSeriesInfo retrieves detailed information about a series on NovelUpdates
// It takes a series ID as a parameter and returns the parsed results.
func GetSeriesInfo(seriesID string) (*SeriesResult, error) {
	// Construct the URL for getting information about a series
	seriesURL := "https://novelupdates.com/series/" + seriesID
	resp, err := makeGetRequest(seriesURL)
	if err != nil {
		return &SeriesResult{}, err
	}

	return parseSeries(resp)
}

// ToJSON converts SeriesResult to JSON format with customizable indentation.
// If the indent parameter is provided and is a positive integer within a reasonable range,
// it determines the number of spaces for indentation; otherwise, it defaults to 2 spaces.
func (seriesResults *SeriesResult) ToJSON(indent ...int) ([]byte, error) {
	customIndent := 2
	if len(indent) > 0 {
		requestedIndent := indent[0]
		if requestedIndent > 0 && requestedIndent <= 8 {
			customIndent = requestedIndent
		}
	}

	jsonOutput, err := json.MarshalIndent(seriesResults, "", strings.Repeat(" ", customIndent))
	if err != nil {
		return nil, err
	}
	return jsonOutput, nil
}

func parseSearch(req *http.Response) ([]*SearchResult, error) {
	doc, err := htmlquery.Parse(req.Body)
	if err != nil {
		return []*SearchResult{}, err
	}

	results := []*SearchResult{}

	for _, result := range htmlquery.Find(doc, "//div[@class='search_main_box_nu']") {
		body := htmlquery.FindOne(result, ".//div[@class='search_body_nu']")
		imageBody := htmlquery.FindOne(result, ".//div[@class='search_img_nu']")

		title := strings.TrimSpace(htmlquery.InnerText(htmlquery.FindOne(body, ".//div[@class='search_title']/a")))
		url := htmlquery.SelectAttr(htmlquery.FindOne(body, ".//div[@class='search_title']/a"), "href")

		image := htmlquery.SelectAttr(htmlquery.FindOne(imageBody, ".//img"), "src")
		if strings.HasSuffix(image, "noimagemid.jpg") {
			image = ""
		}

		imageBodyRatings := htmlquery.FindOne(imageBody, ".//div[@class='search_ratings']")
		ratingSpan := htmlquery.FindOne(imageBodyRatings, ".//span")
		if ratingSpan != nil {
			imageBodyRatings.RemoveChild(ratingSpan)
		}
		// searchRating := strings.TrimSpace(strings.Trim(htmlquery.InnerText(imageBodyRatings), "()"))
		searchRating := regexp.MustCompile(`[\s()]`).ReplaceAllString(htmlquery.InnerText(imageBodyRatings), "")

		ogDescription := strings.TrimSpace(htmlquery.InnerText(htmlquery.FindOne(body, "./text()")))
		moreDescription := htmlquery.FindOne(body, ".//span[@class='testhide']")
		unwantedElement := htmlquery.FindOne(moreDescription, ".//span[@class='morelink list']")
		if unwantedElement != nil {
			moreDescription.RemoveChild(unwantedElement)
		}
		pElements := htmlquery.Find(moreDescription, ".//p[@style='margin-top:-5px;']")
		for _, p := range pElements {
			moreDescription.RemoveChild(p)
		}
		moreURL := htmlquery.FindOne(moreDescription, ".//span[@class='moreurl list']")
		if moreURL != nil {
			moreDescription.RemoveChild(moreURL)
		}
		description := ogDescription + "\n" + strings.TrimSpace(htmlquery.InnerText(moreDescription))

		stats := htmlquery.Find(body, ".//div[@class='search_stats']/span[@class='ss_desk']")
		releases := strings.TrimSpace(htmlquery.InnerText(stats[0]))
		updateFreq := strings.TrimSpace(htmlquery.InnerText(stats[1]))
		nuReaders := strings.TrimSpace(htmlquery.InnerText(stats[2]))
		nuReviews := strings.TrimSpace(htmlquery.InnerText(stats[3]))
		lastUpdated := strings.TrimSpace(htmlquery.InnerText(stats[4]))

		genres := []map[string]string{}
		for _, genre := range htmlquery.Find(body, ".//div[@class='search_genre']/a") {
			genreName := strings.TrimSpace(htmlquery.InnerText(genre))
			genreURL := htmlquery.SelectAttr(genre, "href")
			genres = append(genres, map[string]string{
				"name": genreName,
				"url":  genreURL,
			})
		}

		resultStruct := &SearchResult{
			Title:        title,
			ID:           extractIDFromURL(url),
			URL:          url,
			Image:        image,
			SearchRating: searchRating,
			Description:  strings.TrimRight(description, "\n"),
			Releases:     releases,
			UpdateFreq:   updateFreq,
			NuReaders:    nuReaders,
			NuReviews:    nuReviews,
			LastUpdated:  lastUpdated,
			Genres:       genres,
		}

		results = append(results, resultStruct)
	}

	return results, nil
}

func extractIDFromURL(url string) string {
	re := regexp.MustCompile(`/series/([a-z0-9-]+)/$`)
	match := re.FindStringSubmatch(url)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func parseSeries(req *http.Response) (*SeriesResult, error) {
	doc, err := htmlquery.Parse(req.Body)
	if err != nil {
		return &SeriesResult{}, err
	}

	page := htmlquery.FindOne(doc, "//div[@class='w-blog-content']")
	body := htmlquery.FindOne(page, "//div[@class='g-cols wpb_row offset_default']")
	ot := htmlquery.FindOne(body, "//div[@class='one-third']/div[contains(@class, 'wpb_text_column')]/div[@class='wpb_wrapper']")
	tt := htmlquery.FindOne(body, "//div[@class='two-thirds']/div[contains(@class, 'wpb_text_column')]/div[@class='wpb_wrapper']")

	titleNode := htmlquery.FindOne(page, "//div[@class='seriestitlenu']")
	title := htmlquery.InnerText(titleNode)

	imageNode := htmlquery.FindOne(ot, "//div[@class='seriesimg']//img")
	image := htmlquery.SelectAttr(imageNode, "src")

	typeRaw := htmlquery.FindOne(ot, "//div[@id='showtype']")
	typeText := htmlquery.InnerText(htmlquery.FindOne(typeRaw, "//a")) + " " + htmlquery.InnerText(htmlquery.FindOne(typeRaw, "//span"))
	typeURL := htmlquery.SelectAttr(htmlquery.FindOne(typeRaw, "//a"), "href")
	resultType := map[string]string{"name": typeText, "url": typeURL}

	genre := make([]map[string]string, 0)
	for _, g := range htmlquery.Find(ot, "//div[@id='seriesgenre']//a") {
		name := htmlquery.InnerText(g)
		url := htmlquery.SelectAttr(g, "href")
		description := htmlquery.SelectAttr(g, "title")
		genre = append(genre, map[string]string{"name": name, "url": url, "description": description})
	}

	tags := make([]map[string]string, 0)
	for _, t := range htmlquery.Find(ot, "//div[@id='showtags']//a") {
		name := htmlquery.InnerText(t)
		url := htmlquery.SelectAttr(t, "href")
		description := htmlquery.SelectAttr(t, "title")
		tags = append(tags, map[string]string{"name": name, "url": url, "description": description})
	}

	rating := []map[string]string{}
	tempR := 5

	overallRatingNode := htmlquery.FindOne(doc, "//h5[@class='seriesother']/span[@class='uvotes']")
	overallRating := overallRatingNode.FirstChild.Data
	overallRating = regexp.MustCompile(`[()]`).ReplaceAllString(overallRating, "")
	rating = append(rating, map[string]string{"name": "Overall", "rating": overallRating})

	trNodes := htmlquery.Find(doc, "//table[@id='myrates']/tbody/tr")
	for _, tr := range trNodes {
		tdNodes := htmlquery.Find(tr, "td")
		rating = append(rating, map[string]string{"name": fmt.Sprint(tempR), "rating": htmlquery.InnerText(tdNodes[1])})
		tempR--
	}

	languageNode := htmlquery.FindOne(ot, "//div[@id='showlang']//a")
	languageName := htmlquery.InnerText(languageNode)
	languageURL := htmlquery.SelectAttr(languageNode, "href")
	language := map[string]string{"name": languageName, "url": languageURL}

	authors := make([]map[string]string, 0)
	for _, author := range htmlquery.Find(ot, "//div[@id='showauthors']//a") {
		name := htmlquery.InnerText(author)
		url := htmlquery.SelectAttr(author, "href")
		authors = append(authors, map[string]string{"name": name, "url": url})
	}

	artists := make([]map[string]string, 0)
	for _, artist := range htmlquery.Find(ot, "//div[@id='showartists']//a") {
		name := htmlquery.InnerText(artist)
		url := htmlquery.SelectAttr(artist, "href")
		artists = append(artists, map[string]string{"name": name, "url": url})
	}

	// year := strings.TrimPrefix(htmlquery.InnerText(ot.FirstChild.NextSibling.NextSibling.NextSibling), " ")[1:]
	year := strings.TrimSpace(htmlquery.InnerText(htmlquery.FindOne(ot, "//div[@id='edityear']")))

	statusRaw := htmlquery.FindOne(ot, "//div[@id='editstatus']")
	status := strings.TrimPrefix(htmlquery.InnerText(statusRaw), " ")[1:]
	if regexp.MustCompile(`\n$`).MatchString(statusRaw.FirstChild.Data) {
		status = regexp.MustCompile(`\n$`).ReplaceAllString(status, "")
	}

	licensed := strings.TrimPrefix(htmlquery.InnerText(htmlquery.FindOne(ot, "//div[@id='showlicensed']")), " ")[1:]
	completelyTranslated := strings.TrimPrefix(htmlquery.InnerText(htmlquery.FindOne(ot, "//div[@id='showtranslated']")), " ")[1:]

	originalPublisherNode := htmlquery.FindOne(ot, "//div[@id='showopublisher']//a")
	var originalPublisher map[string]string
	if originalPublisherNode != nil {
		originalPublisherName := htmlquery.InnerText(originalPublisherNode)
		originalPublisherURL := htmlquery.SelectAttr(originalPublisherNode, "href")
		originalPublisher = map[string]string{"name": originalPublisherName, "url": originalPublisherURL}
	}

	englishPublisherNode := htmlquery.FindOne(ot, "//div[@id='showepublisher']//a")
	var englishPublisher map[string]string
	if englishPublisherNode != nil {
		englishPublisherName := htmlquery.InnerText(englishPublisherNode)
		englishPublisherURL := htmlquery.SelectAttr(englishPublisherNode, "href")
		englishPublisher = map[string]string{"name": englishPublisherName, "url": englishPublisherURL}
	}

	releaseFreqNodes := htmlquery.Find(ot, "//h5[@class='seriesother'][contains(text(),'Release Frequency')]/following-sibling::text()[1]")
	var releaseFreq string
	for _, node := range releaseFreqNodes {
		releaseFreq += strings.TrimSpace(htmlquery.InnerText(node))
	}

	descriptionRaw := htmlquery.FindOne(tt, "//div[@id='editdescription']")
	description := strings.TrimSuffix(htmlquery.InnerText(descriptionRaw), "\n")

	associatedNames := []string{}
	divNode := htmlquery.FindOne(doc, "//div[@id='editassociated']")

	for child := divNode.FirstChild; child != nil; child = child.NextSibling {
		if htmlquery.OutputHTML(child, true) != "<br/>" {
			associatedNames = append(associatedNames, htmlquery.InnerText(child))
		}
	}

	var groups []map[string]string
	groupsNode := htmlquery.FindOne(doc, "//ol[@class='sp_grouptable']")
	if groupsNode != nil {
		for _, g := range htmlquery.Find(groupsNode, "//li") {
			name := htmlquery.SelectAttr(htmlquery.FindOne(g, "//span[@style='padding-left:20px;']"), "title")
			temp := strings.ReplaceAll(strings.ToLower(strings.ReplaceAll(name, " ", "-")), "-", "")
			url := fmt.Sprintf("https://www.novelupdates.com/group/%s", regexp.MustCompile(`[^\w\s]`).ReplaceAllString(temp, ""))
			groups = append(groups, map[string]string{"name": name, "url": url})
		}
	}

	result := &SeriesResult{
		Title:                title,
		Image:                image,
		Type:                 resultType,
		Genre:                genre,
		Tags:                 tags,
		Rating:               rating,
		Language:             language,
		Authors:              authors,
		Artists:              artists,
		Year:                 year,
		Status:               status,
		Licensed:             licensed,
		CompletelyTranslated: completelyTranslated,
		OriginalPublisher:    originalPublisher,
		EnglishPublisher:     englishPublisher,
		ReleaseFreq:          releaseFreq,
		Description:          description,
		AssociatedNames:      associatedNames,
		Groups:               groups,
	}

	return result, nil
}

func makeGetRequest(url string) (*http.Response, error) {
	client := req.C().ImpersonateChrome()
	resp, err := client.R().Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Response, nil
}
