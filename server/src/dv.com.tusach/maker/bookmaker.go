package maker

import (
	"bytes"
	"dv.com.tusach/util"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func CreateBook(c chan string, book Book, parser string) {
	var numPagesLoaded = 0
	var aborted = false
	var errorMsg = ""

	go func() {
		log.Printf("start monitoring book: %d-%s\n", book.ID, book.Title)
		for {
			msg, more := <-c
			log.Printf("Received message: %s for book: %d, more:%v\n", msg, book.ID, more)
			if msg == "abort" {
				aborted = true
			}
			if !more {
				break
			}
		}
		log.Printf("stop monitoring book: %d-%s\n", book.ID, book.Title)
	}()

	url := book.CurrentPageUrl
	if url == "" {
		url = book.StartPageUrl
	}
	loader := PageLoader{}

	// TODO set loader configuration

	for {
		if aborted || book.MaxNumPages > 0 && book.MaxNumPages <= book.CurrentPageNo {
			break
		}

		// load page
		log.Println("Loading page: ", url)
		rawHtml, err := loader.executeRequest(url)
		if err != nil {
			errorMsg = "Failed to load from url: " + url + ". " + err.Error()
			break
		}

		newChapter := Chapter{BookId: book.ID, ChapterNo: book.CurrentPageNo + 1}
		nextPageUrl, err := parse(parser, rawHtml, &newChapter)
		if err != nil {
			errorMsg = err.Error()
			break
		}
		log.Printf("completed chapter: %d:%s, nextPageUrl:%s\n", newChapter.ChapterNo, newChapter.Title, nextPageUrl)
		// save the chapter
		if book.CurrentPageUrl == "" || book.CurrentPageUrl != url {
			book.CurrentPageNo++
			book.CurrentPageUrl = url
			err := SaveChapter(newChapter)
			if err != nil {
				errorMsg = err.Error()
				break
			}
			numPagesLoaded++
			// check for no more pages
			if nextPageUrl == "" {
				log.Println("No more next page url found.")
				break
			} else {
				_, err := SaveBook(book)
				if err != nil {
					errorMsg = err.Error()
					break
				}
			}
		}
		if nextPageUrl == url {
			log.Println("Internal error. next page url is same as current page url: ", url)
			break
		}
		url = nextPageUrl
	}

	book.ErrorMsg = errorMsg
	if book.ErrorMsg != "" {
		book.Status = STATUS_ERROR
		log.Println(book.ErrorMsg)
	} else if aborted {
		book.Status = STATUS_ABORTED
	} else {
		book.Status = STATUS_COMPLETED
	}

	chapters, err := LoadChapters(book.ID)
	if err != nil {
		errorMsg = fmt.Sprintf("Error loading chapters for book: %d. %s", book.ID, err.Error())
		log.Println(errorMsg)
		book.ErrorMsg = errorMsg
		book.Status = STATUS_ERROR
	} else {
		err = makeEpub(book, chapters)
		if err != nil {
			errorMsg = fmt.Sprintf("Error creating epub for book: %d. %s", book.ID, err.Error())
			log.Println(errorMsg)
			book.ErrorMsg = errorMsg
			book.Status = STATUS_ERROR
			book.EpubCreated = false
		} else {
			book.EpubCreated = true
		}
	}

	SaveBook(book)
}

func GetParserName(url string) string {
	return "tangthuvien"
}

func makeEpub(book Book, chapters []Chapter) error {
	if book.CurrentPageNo == 0 {
		return errors.New("Book has no chapters")
	}
	// validate chapter html files
	for i := 0; i < len(chapters); i++ {
		chapter := chapters[i]
		filename := util.GetChapterFilename(book.ID, chapter.ChapterNo)
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			return errors.New("Missing chapter file: " + filename)
		}
	}

	// update toc.ncx file
	tocFile := util.GetBookPath(book.ID) + "/OEBPS/toc.ncx"
	data, err := ioutil.ReadFile(tocFile)
	if err != nil {
		return errors.New("Failed to open file: " + tocFile + ". " + err.Error())
	}
	str := string(data)
	index := strings.Index(str, "</head>")
	if index == -1 {
		return errors.New("Cannot find </head> in " + tocFile + ". ")
	}

	var buffer bytes.Buffer
	buffer.WriteString(str[0 : index+len("</head>")])
	// add <docTitle>
	buffer.WriteString("<docTitle><text>" + book.Title + "</text></docTitle>\n")
	// add <docAuthor>
	buffer.WriteString("<docAuthor><text>" + book.Author + "</text></docAuthor>\n")
	buffer.WriteString("<navMap>\n")
	// add all <navPoint>
	for i := 0; i < len(chapters); i++ {
		chapter := chapters[i]
		filepath := util.GetChapterFilename(book.ID, chapter.ChapterNo)
		index := strings.LastIndex(filepath, "/")
		filename := filepath[index+1:]
		buffer.WriteString(fmt.Sprintf("<navPoint id=\"navPoint-%d\" playOrder=\"%d\" class=\"chapter\">\n", chapter.ChapterNo, chapter.ChapterNo))

		buffer.WriteString(fmt.Sprintf("<navLabel><text>%s</text></navLabel>\n", chapter.Title))
		buffer.WriteString(fmt.Sprintf("<content src=\"%s\"/>\n", filename))

		buffer.WriteString("</navPoint>\n")
	}
	buffer.WriteString("</navMap></ncx>")
	err = util.SaveFile(tocFile, buffer.Bytes())
	if err != nil {
		return err
	}

	// update content.opf
	contentFile := util.GetBookPath(book.ID) + "/OEBPS/content.opf"
	data, err = ioutil.ReadFile(contentFile)
	if err != nil {
		return errors.New("Failed to open file: " + contentFile + ". " + err.Error())
	}
	str = string(data)
	buffer.Reset()
	index = strings.Index(str, "<opf:item id=\"chapter")
	if index == -1 {
		index = strings.Index(str, "</opf:manifest>")
	}
	if index == -1 {
		return errors.New("Cannot find </opf:manifest> in " + contentFile + ". ")
	}
	str = str[0:index]
	// replace title
	re, _ := regexp.Compile("(<dc:title>.+</dc:title>)")
	str = re.ReplaceAllString(str, "<dc:title>"+book.Title+"</dc:title>")
	// replace creator
	re, _ = regexp.Compile("(<dc:creator.+</dc:creator>)")
	str = re.ReplaceAllString(str, "<dc:creator opf:role=\"aut\">"+book.Author+"</dc:creator>")

	buffer.WriteString(str)

	// write opf:items
	for i := 0; i < len(chapters); i++ {
		chapter := chapters[i]
		filepath := util.GetChapterFilename(book.ID, chapter.ChapterNo)
		index := strings.LastIndex(filepath, "/")
		filename := filepath[index+1:]
		buffer.WriteString(fmt.Sprintf("<opf:item id=\"%s\" href=\"%s\" media-type=\"application/xhtml+xml\" />\n", filename, filename))
	}
	buffer.WriteString("</opf:manifest>\n")
	buffer.WriteString("<opf:spine toc=\"ncx\">\n")
	for i := 0; i < len(chapters); i++ {
		chapter := chapters[i]
		filepath := util.GetChapterFilename(book.ID, chapter.ChapterNo)
		index := strings.LastIndex(filepath, "/")
		filename := filepath[index+1:]
		buffer.WriteString(fmt.Sprintf("<opf:itemref idref=\"%s\" />\n", filename))
	}
	buffer.WriteString("</opf:spine>\n</opf:package>")

	err = util.SaveFile(contentFile, buffer.Bytes())
	if err != nil {
		return err
	}

	epubFile := util.GetConfiguration().LibraryPath + "/" + strings.Replace(book.Title, " ", "-", -1) + ".epub"
	// delete existing epub
	os.Remove(epubFile)

	// zip the epub file
	cmd := exec.Command(util.GetConfiguration().LibraryPath+"/make-epub.sh", epubFile, util.GetBookPath(book.ID))
	out, err := cmd.CombinedOutput()
	str = string(out)
	log.Println("epub command output: ", str)
	if err != nil {
		return errors.New("Error creating epub file. " + err.Error())
	}
	str = string(out)
	log.Println("epub command output: ", str)

	if _, err := os.Stat(epubFile); os.IsNotExist(err) {
		return errors.New("Error creating epub file: " + epubFile)
	} else {
		log.Println("Created epub file: " + epubFile)
	}

	return nil
}

func parse(parser string, rawHtml []byte, chapter *Chapter) (string, error) {
	var nextPageUrl = ""

	// write to raw html file
	rawFilename := util.GetRawChapterFilename(chapter.BookId, chapter.ChapterNo)
	err := util.SaveFile(rawFilename, rawHtml)
	if err != nil {
		return "", err
	}

	filename := util.GetChapterFilename(chapter.BookId, chapter.ChapterNo)

	// call parser to parse chapter html
	log.Println("executing parser command: ", parser)
	cmd := exec.Command(util.GetParserPath()+"/"+parser+"/"+parser,
		"-configFile="+util.GetConfigFile(), "-op=p",
		"-inputFile="+rawFilename, "-outputFile="+filename)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.New("Error parsing html file. " + err.Error())
	}
	str := string(out)
	log.Println("parser command output: ", str)

	found := false
	// extract the nextPageUrl & chapterTitle
	lines := strings.Split(str, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "***nextPageUrl=") {
			nextPageUrl = line[len("***nextPageUrl="):]
			found = true
		}
		if strings.HasPrefix(line, "***chapterTitle=") {
			chapter.Title = line[len("***chapterTitle="):]
		}
	}
	if !found {
		return "", errors.New(str)
	}

	os.Remove(rawFilename)

	return nextPageUrl, nil
}
