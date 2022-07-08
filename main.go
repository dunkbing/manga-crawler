package main

import (
    "fmt"
    "github.com/gocolly/colly"
    "time"
)

type Chapter struct {
    Title  string
    Link   string
    Images []string
}

type Manga struct {
    Title       string
    Link        string
    Image       string
    Description string
    Authors     []string
    Status      string
    Genres      []string
    Chapters    []Chapter
    CrawledAt   time.Time
}

type MangaCrawler struct {
    crawler *colly.Collector
    Mangas  []Manga
}

func MakeCrawler() *MangaCrawler {
    return &MangaCrawler{
        crawler: colly.NewCollector(
            colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.149 Safari/537.36"),
        ),
        Mangas: make([]Manga, 0),
    }
}

func (m *MangaCrawler) Crawl() {
    m.crawler.OnHTML(".items", func(e *colly.HTMLElement) {})
}

func main() {
    c := colly.NewCollector(
        colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.149 Safari/537.36"),
    )
    c.OnRequest(func(r *colly.Request) {
        r.Headers.Set("X-Requested-With", "XMLHttpRequest")
        r.Headers.Set("Referer", "http://www.nettruyenco.com/")
        fmt.Printf("Visiting %s\n", r.URL)
    })
    c.OnError(func(r *colly.Response, err error) {
        fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "Error:", err)
    })
    c.OnResponse(func(r *colly.Response) {
        fmt.Printf("Received response %v\n", r.StatusCode)
    })
    c.OnHTML(".items", func(e *colly.HTMLElement) {
        e.ForEach(".row > .item", func(index int, e *colly.HTMLElement) {
            if index > 0 {
                return
            }
            mangaPage := c.Clone()
            mangaPage.OnRequest(func(r *colly.Request) {
                fmt.Printf("Visiting %s\n", r.URL)
            })
            mangaPage.OnHTML("#ctl00_divCenter", func(e *colly.HTMLElement) {
                manga := Manga{}
                manga.Title = e.ChildText(".title-detail")
                manga.Link = e.Request.URL.String()
                manga.Image = e.ChildAttr("img", "src")
                manga.Description = e.ChildText("p.shortened")
                authors := make([]string, 0)
                genres := make([]string, 0)
                e.ForEach("p.col-xs-8", func(pIndex int, e *colly.HTMLElement) {
                    switch pIndex {
                    case 0:
                        e.ForEach("a", func(aIndex int, e *colly.HTMLElement) {
                            authors = append(authors, e.Text)
                        })
                    case 1:
                        manga.Status = e.Text
                    case 2:
                        e.ForEach("a", func(aIndex int, e *colly.HTMLElement) {
                            genres = append(manga.Genres, e.Text)
                        })
                    }
                })
                manga.Authors = authors
                manga.Genres = genres
                manga.CrawledAt = time.Now()
                e.ForEach("nav > ul > li", func(index int, e *colly.HTMLElement) {
                    chapterPage := c.Clone()
                    chapterPage.OnRequest(func(r *colly.Request) {
                        fmt.Printf("Visiting chapter %s\n", r.URL)
                    })
                    chapterPage.OnHTML(".reading-detail", func(e *colly.HTMLElement) {
                        chapter := Chapter{}
                        chapter.Title = e.ChildText(".title-detail")
                        chapter.Link = e.Request.URL.String()
                        chapter.Images = make([]string, 0)
                        e.ForEach("img", func(index int, e *colly.HTMLElement) {
                            image := fmt.Sprintf("http:%s", e.Attr("data-original"))
                            chapter.Images = append(chapter.Images, image)
                            imagePage := c.Clone()
                            imagePage.OnError(func(r *colly.Response, err error) {
                                fmt.Println("Request URL:", r.Request.URL, "failed with response:", r.StatusCode, "Error:", err)
                            })
                            imagePage.OnResponse(func(r *colly.Response) {
                                // save image to disk
                                err := r.Save(fmt.Sprintf("%s.jpg", image))
                                if err != nil {
                                    fmt.Println(err)
                                    return
                                }
                            })
                            err := imagePage.Visit(image)
                            if err != nil {
                                fmt.Println(err)
                                return
                            }
                        })
                        manga.Chapters = append(manga.Chapters, chapter)
                    })
                    chapterLink := e.ChildAttr("a", "href")
                    err := chapterPage.Visit(chapterLink)
                    if err != nil {
                        return
                    }
                    chapterPage.Wait()
                })
                //fmt.Println("manga", manga)
            })
            mangaLink := e.ChildAttr("a", "href")
            err := mangaPage.Visit(mangaLink)
            if err != nil {
                fmt.Println(err)
            }
        })
    })
    c.OnScraped(func(r *colly.Response) {
        fmt.Println("Finished", r.Request.URL)
    })
    err := c.Visit("http://www.nettruyenco.com/")
    if err != nil {
        _ = fmt.Errorf("%s", err.Error())
    }
}
