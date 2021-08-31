package main

import (
  "flag"
  "log"
  "net/url"
  "os"
  "strings"
)

var (
  logger = log.New(os.Stderr, "", 0)
  articles *Articles
)

func main() {
  args := os.Args[1:]
  if len(args) == 0 {
    flag.Usage()
    os.Exit(1)
  }

  // Load articles
  var err error
  if articles, err = LoadArticles(); err != nil {
    logger.Fatalf("error getting articles: %v", err)
  }
  defer func() {
    if err := articles.ExportArticles(); err != nil {
      logger.Fatalf("error saving articles: %v", err)
    }
  }()

  cmds := map[string]Command{
    "add": NewAddCommand(),
    "del": NewDelCommand(),
    "get": NewGetCommand(),
  }
  if cmd := cmds[args[0]]; cmd == nil {
    logger.Printf("invalid subcommand: %s", args[0])
    flag.Usage()
    os.Exit(1)
  } else {
    cmd.Run(args[1:])
  }
}

type LinkValue struct {
  LinkPtr *string
}

func (v LinkValue) String() string {
  return *(v.LinkPtr)
}

func (v LinkValue) Set(s string) error {
  if u, err := url.Parse(s); err != nil {
    return err
  } else {
    *(v.LinkPtr) = u.String()
  }
  return nil
}

type TagsValue struct {
  TagsPtr *[]string
}

func (v TagsValue) String() string {
  if len(*(v.TagsPtr)) != 0 {
    return strings.Join(*(v.TagsPtr), "|")
  }
  return ""
}

func (v TagsValue) Set(s string) error {
  if s != "" {
    *(v.TagsPtr) = strings.Split(s, "|")
  }
  return nil
}

type Command interface {
  Run(args []string)
}

type AddCommand struct {
  fs *flag.FlagSet
}

func NewAddCommand() *AddCommand {
  return &AddCommand{
    fs: flag.NewFlagSet("add", flag.ExitOnError),
  }
}

func (ac *AddCommand) Run(args []string) {
  article := new(Article)
  ac.fs.StringVar(&article.Name, "name", "", "Article name")
  ac.fs.Var(&LinkValue{&article.Link}, "link", "Link (url) to article")
  ac.fs.BoolVar(&article.Read, "read", false, "Article has been read")
  ac.fs.BoolVar(&article.Favorite, "fav", false, "Favorite article")
  ac.fs.Var(&TagsValue{&article.Tags}, "tags", "Comma-separated list of tags for the article")
  download := ac.fs.Bool("download", false, "Download article")
  ac.fs.Parse(args)

  if article.Name == "" {
    logger.Fatal("must provide name")
  }
  articles.AddArticle(article)
  if *download {
    if article.Link != "" {
      if err := article.Download(); err != nil {
        logger.Printf("error downloading article: %v", err)
      }
    }
  }
  logger.Printf(`adding article %d: %s`, article.ID, article.Name)
}

type GetCommand struct {
  fs *flag.FlagSet
}

func NewGetCommand() *GetCommand {
  return &GetCommand{
    fs: flag.NewFlagSet("get", flag.ExitOnError),
  }
}

/* TODO: try to download article if article retrieval fails (possibly only if download doesn't exist) */
func (gc *GetCommand) Run(args []string) {
  article := new(Article)
  gc.fs.StringVar(&article.Name, "name", "", "Article name")
  gc.fs.IntVar(&article.ID, "id", -1, "Article ID")
  download := gc.fs.Bool("download", false, "Retrieve article download")
  gc.fs.Parse(args)

  if article.ID != -1 {
    if a := articles.GetArticleByID(article.ID); a != nil {
      if *download {
        if err := a.Retrieve(); err != nil {
          logger.Fatalf("error retrieving article: %v", err)
        }
        logger.Printf("retrieved article %d: %s", a.ID, a.Name)
      } else {
        a.Print()
      }
      return
    }
    logger.Printf("article id %d not found", article.ID)
  }
  if article.Name != "" {
    if a := articles.GetArticleByName(article.Name); a != nil {
      if *download {
        if err := a.Retrieve(); err != nil {
          logger.Fatalf("error retrieving article: %v", err)
        }
        logger.Printf("retrieved article %d: %s", a.ID, a.Name)
      } else {
        a.Print()
      }
      return
    }
    logger.Printf(`article "%s" not found`, article.Name)
  }
  if article.Name == "" && article.ID == -1 {
    articles.PrintArticles()
  }
}

type UpdateCommand struct {
  fs *flag.FlagSet
}

func NewUpdateCommand() *UpdateCommand {
  return &UpdateCommand{
    fs: flag.NewFlagSet("update", flag.ExitOnError),
  }
}

func (uc *UpdateCommand) Run(args []string) {
  article := new(Article)
  uc.fs.IntVar(&article.ID, "id", -1, "Article ID")
  uc.fs.StringVar(&article.Name, "name", "", "Article Name")
  uc.fs.Var(&LinkValue{&article.Link}, "link", "Link (url) to article")
  uc.fs.Var(&TagsValue{&article.Tags}, "tags", "Comma-separated list of tags for the article")
  uc.fs.Parse(args)

  if article.ID == -1 {
    logger.Fatal("must provide article id")
  }
  a := articles.GetArticleByID(article.ID)
  if a == nil {
    logger.Fatal("article id %d not found", article.ID)
  }
  if article.Name != "" {
    logger.Printf(`updated article name from "%s" to "%s"`, a.Name, article.Name)
    a.Name = article.Name
  }
  if article.Link != "" {
    logger.Printf(`updated article link from %s to %s`, a.Link, article.Link)
    a.Link = article.Link
  }
  if len(article.Tags) != 0 {
    logger.Printf(
      `updated article tags from %s to %s`,
      strings.Join(a.Tags, "|"), strings.Join(article.Tags, "|"))
    a.Tags = article.Tags
  }
}

type DelCommand struct {
  fs *flag.FlagSet
}

func NewDelCommand() *DelCommand {
  return &DelCommand{
    fs: flag.NewFlagSet("del", flag.ExitOnError),
  }
}

func (dc *DelCommand) Run(args []string) {
  article := new(Article)
  dc.fs.StringVar(&article.Name, "name", "", "Article name")
  dc.fs.IntVar(&article.ID, "id", -1, "Article ID")
  dc.fs.Parse(args)

  if article.ID != -1 {
    if articles.DeleteArticleByID(article.ID) {
      logger.Printf("deleted article %d: %s", article.ID, article.Name)
      return
    }
    logger.Printf("article id %d not found", article.ID)
  }
  if article.Name != "" {
    if articles.DeleteArticleByName(article.Name) {
      logger.Printf("deleted article %d: %s", article.ID, article.Name)
      return
    }
    logger.Printf(`article "%d" not found`, article.Name)
  }
  if article.Name == "" && article.ID == -1 {
    logger.Fatal("must provide article id or name")
  }
}
