package main

/* TODO:
 * Add RemoveDownload() method to Article
 */

import (
  "compress/gzip"
  "encoding/json"
  "errors"
  "io/ioutil"
  "net/http"
  "os"
  "path"
  "runtime"
  "strconv"
  "strings"
)

var (
  articlesDir string
)

func init() {
  _, thisFile, _, ok := runtime.Caller(0)
  if !ok {
    logger.Fatal("error getting articles directory")
  }
  articlesDir = path.Join(path.Dir(thisFile), "articles")
}

type Article struct {
  ID int `json:"id"`
  Name string `json:"name"`
  Link string `json:"link"`
  Read bool `json:"read"`
  Favorite bool `json:"favorite"`
  Tags []string `json:"tags"`
  Downloaded bool `json:"downloaded"`
}

func (a *Article) Download() error {
  // Must be a link present
  if a.Link == "" {
    return errors.New("must provide link")
  }
  // Get the article (page)
  resp, err := http.Get(a.Link)
  if err != nil {
    return err
  }
  // Read the contents of the response body
  contents, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return err
  }
  resp.Body.Close()
  // Create the article file
  filePath := path.Join(articlesDir, strconv.Itoa(a.ID)+".html.gz")
  f, err := os.Create(filePath)
  if err != nil {
    return err
  }
  defer f.Close()
  // Write the compressed contents to the file
  w, err := gzip.NewWriterLevel(f, gzip.BestCompression)
  if err != nil {
    return err
  }
  if _, err := w.Write(contents); err != nil {
    return err
  }
  a.Downloaded = true
  return w.Close()
}

func (a *Article) Retrieve() error {
  // Must be downloaded
  if !a.Downloaded {
    return errors.New("file not downloaded")
  }
  // Open the article zip file
  filePath := path.Join(articlesDir, strconv.Itoa(a.ID)+".html.gz")
  f, err := os.Open(filePath)
  if err != nil {
    if strings.Contains(err.Error(), "exists") {
      return errors.New("article not downloaded")
    }
    return err
  }
  defer f.Close()
  // Read the uncompressed article data and write it to the file
  r, err := gzip.NewReader(f)
  if err != nil {
    return err
  }
  defer r.Close()
  contents, err := ioutil.ReadAll(r)
  if err != nil {
    return err
  }
  return ioutil.WriteFile(a.Name+".html", contents, 0644)
}

func (a *Article) Print() {
  logger.Printf("Article %d: %s", a.ID, a.Name)
  if a.Link != "" {
    logger.Printf("\t%s", a.Link)
  }
  if a.Read {
    logger.Printf("\tRead")
  } else {
    logger.Printf("\tUnread")
  }
  if a.Favorite {
    logger.Printf("\tFavorited")
  }
  if len(a.Tags) != 0 {
    logger.Printf("\t%s", strings.Join(a.Tags, "|"))
  }
  if a.Downloaded {
    logger.Printf("\tDownloaded")
  }
}

type Articles struct {
  Articles []*Article `json:"articles"`
  // Holds all tags and which articles have each tag
  // map[tagName][]articleID
  ArticleTags map[string][]int `json:"articleTags"`
  // LastID is the previous article id
  LastID int `json:"lastID"`
}

func LoadArticles() (*Articles, error) {
  // Open the articles json file
  filePath := path.Join(articlesDir, "articles.json")
  f, err := os.Open(filePath)
  if err != nil {
    return nil, err
  }
  defer f.Close()
  // Read the articles data from the file and map it to the struct
  articles := new(Articles)
  if err := json.NewDecoder(f).Decode(articles); err != nil {
    return nil, err
  }
  return articles, nil
}

func (a *Articles) ExportArticles() error {
  // Open (create) the articles json file
  filePath := path.Join(articlesDir, "articles.json")
  f, err := os.Create(filePath)
  if err != nil {
    return err
  }
  defer f.Close()
  // Write the articles json data to the file
  if err := json.NewEncoder(f).Encode(a); err != nil {
    return err
  }
  return nil
}

func (a *Articles) AddArticle(article *Article) {
  // Increment the last id and set the new article's id to it
  a.LastID++
  article.ID = a.LastID
  // Add the new article to the list of articles
  a.Articles = append(a.Articles, article)
}

func (a *Articles) PrintArticles() {
  for _, article := range a.Articles {
    logger.Printf("Article %d: %s", article.ID, article.Name)
  }
}

/* TODO: Improve or remove since multiple articles can have the same name */
func (a *Articles) GetArticleByName(name string) *Article {
  for _, article := range a.Articles {
    if article.Name == name {
      return article
    }
  }
  return nil
}

func (a *Articles) GetArticleByLink(link string) *Article {
  for _, article := range a.Articles {
    if article.Link == link {
      return article
    }
  }
  return nil
}

func (a *Articles) GetArticleByID(id int) *Article {
  for _, article := range a.Articles {
    if article.ID == id {
      return article
    }
  }
  return nil
}

func (a *Articles) DeleteArticleByName(name string) (exists bool) {
  for i, article := range a.Articles {
    if article.Name == name {
      if i == 0 {
        a.Articles = a.Articles[1:]
      } else if l := len(a.Articles); i == l {
        a.Articles = a.Articles[:l-1]
      } else {
        a.Articles = append(a.Articles[:i], a.Articles[i+1:]...)
      }
      return true
    }
  }
  return false
}

func (a *Articles) DeleteArticleByID(id int) (exists bool) {
  for i, article := range a.Articles {
    if article.ID == id {
      if i == 0 {
        a.Articles = a.Articles[1:]
      } else if l := len(a.Articles); i == l {
        a.Articles = a.Articles[:l-1]
      } else {
        a.Articles = append(a.Articles[:i], a.Articles[i+1:]...)
      }
      return true
    }
  }
  return false
}
