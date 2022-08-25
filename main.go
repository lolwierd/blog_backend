// TODO: find a way to store page ids with blog posts or anywhere to solve problrm of renamed posts not being deleted
// TODO: figure out logging
// TODO: figure out executable / deployment & wrangler
// TODO: get good code/file structure

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/jomei/notionapi"
)

type PostMetaData struct {
	postId      string
	title       string
	date        string
	tags        []string
	slug        string
	draft       string
	description string
}

// [postId] slug
type Posts map[string]string

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	notion_secret := os.Getenv("NOTION_SECRET")
	db_id := os.Getenv("DB_ID")
	blog_dir := os.Getenv("BLOG_DIR")
	currPosts := parseCurrentPosts()
	updatedPosts := currPosts
	client := notionapi.NewClient(notionapi.Token(notion_secret))
	db, dberr := client.Database.Query(context.Background(), notionapi.DatabaseID(db_id), &notionapi.DatabaseQueryRequest{Filter: notionapi.PropertyFilter{Property: "updated", Checkbox: &notionapi.CheckboxFilterCondition{Equals: true}}})
	if dberr != nil {
		log.Fatal("DB ID wrong || API KEY wrong")
	}
	for _, page := range db.Results {
		pMetadata := getPostMetadata(&page)
		slug := currPosts[pMetadata.postId]
		if slug == "" {
			updatedPosts[pMetadata.postId] = pMetadata.slug
			fmt.Println("Post Id not found!!", pMetadata.postId)
		} else {
			if slug != pMetadata.slug {
				fmt.Println("Reanamed post", pMetadata.postId)
				err := os.Remove(fmt.Sprintf("%s/%s.md", blog_dir, slug))
				if err != nil {
					fmt.Println("things went bad like mah bich ", err)
				} else {
					updatedPosts[pMetadata.postId] = pMetadata.slug
				}
			}
		}
		writeToFile(fmt.Sprintf("%s/%s.md", blog_dir, pMetadata.slug), getPostHead(pMetadata)+getPostContent(page.ID.String(), client))
	}
	var posts string
	for pageId, slug := range updatedPosts {
		posts += fmt.Sprintf("%s %s\n", pageId, slug)
	}
	writeToFile("posts", posts)
	deploy(blog_dir)
}

func getPostContent(postId string, client *notionapi.Client) (content string) {
	blocks, APIerr := client.Block.GetChildren(context.Background(), notionapi.BlockID(postId), nil)
	if APIerr != nil {
		log.Fatal("BAD page_id", APIerr)
	}
	for _, block := range blocks.Results {
		switch v := block.(type) {
		case *notionapi.ParagraphBlock:
			content += "\n"
			for _, r := range v.Paragraph.RichText {
				content += getBlockContent(&r)
			}

		case *notionapi.Heading1Block:
			content += "\n"
			content += "# "
			for _, r := range v.Heading1.RichText {
				content += getBlockContent(&r)
			}

		case *notionapi.Heading2Block:
			content += "\n"
			content += "## "
			for _, r := range v.Heading2.RichText {
				content += getBlockContent(&r)
			}

		case *notionapi.Heading3Block:
			content += "\n"
			content += "### "
			for _, r := range v.Heading3.RichText {
				content += getBlockContent(&r)
			}

		case *notionapi.NumberedListItemBlock:
			content += "\n"
			content += "1. "
			for _, r := range v.NumberedListItem.RichText {
				content += getBlockContent(&r)
			}

		case *notionapi.BulletedListItemBlock:
			content += "\n"
			content += "- "
			for _, r := range v.BulletedListItem.RichText {
				content += getBlockContent(&r)
			}

		case *notionapi.QuoteBlock:
			content += "\n"
			content += "> "
			for _, r := range v.Quote.RichText {
				content += getBlockContent(&r)
			}

		case *notionapi.DividerBlock:
			content += "\n"
			content += "---"

		case *notionapi.CodeBlock:
			content += "\n"
			content += "```"
			content += v.Code.Language
			content += "\n"
			for _, r := range v.Code.RichText {
				content += getBlockContent(&r)
			}
			content += "\n"
			content += "```"

		case *notionapi.ImageBlock:
			content += "\n"
			content += "{{< figure src=\""
			var title string
			for _, r := range v.Image.Caption {
				title += fmt.Sprintf("%s", r.Text.Content)
			}
			content += v.Image.GetURL()
			content += "\" title=\""
			content += title
			content += "\" >}}"

		case *notionapi.ToDoBlock:
			content += "\n"
			if v.ToDo.Checked {
				content += "- [x] "
			} else {
				content += "- [] "
			}
			for _, r := range v.ToDo.RichText {
				content += getBlockContent(&r)
			}

		default:
			log.Printf("Shit stinks %T", block)
		}
	}
	return
}

func parseCurrentPosts() (currPosts Posts) {
	currPosts = make(Posts)
	content, err := ioutil.ReadFile("posts")
	if err != nil {
		fmt.Println("Some error in opening file!!")
	}
	for _, s := range strings.Split(string(content), "\n") {
		post := strings.Split(s, " ")
		if len(post) == 2 {
			currPosts[post[0]] = post[1]
		}
	}
	return
}

func deploy(blog_dir string) {
	os.Chdir(blog_dir)
	_, err := exec.Command("git", "add", "-A").Output()
	if err != nil {
		fmt.Println("Error!!", err)
	}
	cmd := exec.Command("git", "commit", "-m", "New Post!!")
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error!!", err)
	}
}

func writeToFile(fileName, content string) {
	data := []byte(content)

	err := ioutil.WriteFile(fileName, data, 0644)

	if err != nil {
		log.Fatal(err)
	}
}

func getBlockContent(r *notionapi.RichText) (content string) {
	if r.Text.Link != nil {
		content += fmt.Sprintf("[%s](%s)", r.Text.Content, r.Text.Link.Url)
	} else {
		content += fmt.Sprintf("%s", annotateText(*r))
	}
	return
}

func getPostHead(postMetadata PostMetaData) (postHead string) {
	postHead += "---"
	postHead += "\n"
	postHead += fmt.Sprintf("title: \"%s\"", postMetadata.title)
	postHead += "\n"
	postHead += fmt.Sprintf("date: \"%s\"", postMetadata.date)
	postHead += "\n"
	postHead += fmt.Sprintf("description: \"%s\"", postMetadata.description)
	postHead += "\n"
	postHead += "tags: ["
	for _, tag := range postMetadata.tags {
		postHead += fmt.Sprintf("\"%s\", ", tag)
	}
	postHead += "]"
	postHead += "\n"
	postHead += fmt.Sprintf("draft: \"%s\"", postMetadata.draft)
	postHead += "\n"
	postHead += "---"
	postHead += "\n"
	return
}

func annotateText(r notionapi.RichText) (annotatedText string) {
	annotatedText = r.Text.Content
	if r.Annotations.Bold {
		annotatedText = fmt.Sprintf("**%s**", annotatedText)
	}
	if r.Annotations.Italic {
		annotatedText = fmt.Sprintf("*%s*", annotatedText)
	}
	if r.Annotations.Strikethrough {
		annotatedText = fmt.Sprintf("~~%s~~", annotatedText)
	}
	if r.Annotations.Code {
		annotatedText = fmt.Sprintf("`%s`", annotatedText)
	}
	return
}

func getPostMetadata(page *notionapi.Page) (pMetadata PostMetaData) {
	pMetadata = PostMetaData{date: page.LastEditedTime.Format("2006-01-02T15:04:05"), postId: page.ID.String()}
	for k, p := range page.Properties {
		switch v := p.(type) {
		case *notionapi.TitleProperty:
			var title string
			for _, s := range v.Title {
				title += s.Text.Content
			}
			pMetadata.title = title
		case *notionapi.RichTextProperty:
			var data string
			for _, s := range v.RichText {
				data += s.Text.Content
			}
			if k == "description" {
				pMetadata.description = data
			} else if k == "slug" {

				pMetadata.slug = data
			} else {
				fmt.Println("No supposed to be here are you??")
			}
		case *notionapi.MultiSelectProperty:
			var tags []string
			for _, tag := range v.MultiSelect {
				tags = append(tags, tag.Name)
			}
			pMetadata.tags = tags
		case *notionapi.CheckboxProperty:
			if k == "draft" {
				pMetadata.draft = strconv.FormatBool(v.Checkbox)
			}
		default:
			fmt.Printf("error somewhere somehow pata lagao daya %T\n", v)

		}
	}
	return
}
