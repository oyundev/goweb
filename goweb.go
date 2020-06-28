package main

import (
	"bytes"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=windows-1254;")

	/*
		html := `<doctype html><html><head><title>Hello goweb.go</title></head>
		<body>
		<h1>Golang web sample</h1>
		<b>Hello goweb.go</b>
		</body></html>`
	*/

	blogTitles, err := GetLatestBlogTitles(" ~~~~~~~~~~ redacted-url-1 ~~~~~~~~~~ ")
	if err != nil {
		log.Println(err)
	}

	/*t := http.Response{
		Body: ioutil.NopCloser(bytes.NewBufferString(blogTitles)),
	}*/

	bytes.NewBufferString(blogTitles).WriteTo(w) //used this because not writing as bytes corrupts html!
}

// GetLatestBlogTitles gets the latest blog title headings from the url
// given and returns them as a list.
func GetLatestBlogTitles(url string) (string, error) {
	/*
	   ╔══════════════════════════════════════════════════════════════════════════════╗
	   ║ html static parts, tags, styles and scripts                                  ║
	   ╚══════════════════════════════════════════════════════════════════════════════╝
	*/
	htmlparthead := `<!DOCTYPE html><html><head>
    <title>Son Dakika Haberleri</title>
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8;">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
    .collapsible {
      background-color: #777;
      color: white;
      cursor: pointer;
      padding: 18px;
      width: 100%;
      border: none;
      text-align: left;
      outline: none;
      font-size: 17px;
    }
    
    .active, .collapsible:hover {
      background-color: #483855;
    }
    
    .content {
      padding: 0 18px;
      display: none;
      overflow: hidden;
      font-size: 17px;
      font-family: arial;
      background-color: #f1fff2;
    }
    </style>
    </head>
    <body style="background-color: #efefef;">    
    <h2><u>Son Dakika Haberleri</u></h2> `

	htmlpartbutton1 := `
<button type="button" class="collapsible">`

	htmlpartbutton2 := `
</button>
<div class="content">
<pre>`

	htmlpartbutton3 := `
</pre>
</div>`

	htmlpartend := `
<script>
var coll = document.getElementsByClassName("collapsible");
var i;

for (i = 0; i < coll.length; i++) {
	coll[i].addEventListener("click", function() {
	this.classList.toggle("active");
	var content = this.nextElementSibling;
	if (content.style.display === "block") {
		content.style.display = "none";
	} else {
		content.style.display = "block";
	}
	});
}
</script>

</body>
</html>`

	/*
	   ╔══════════════════════════════════════════════════════════════════════════════╗
	   ║ use .Header.Set to define content with proper character encoding             ║
	   ╚══════════════════════════════════════════════════════════════════════════════╝
	*/
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "text/html;charset=windows-1254")
	req.Header.Set("Content-Type", "text/html; charset=windows-1254")
	client := &http.Client{}
	resp, err := client.Do(req)

	/*
	   ╔══════════════════════════════════════════════════════════════════════════════╗
	   ║ parse dom document and iterate content to find news head-lines and detail-url║
	   ╚══════════════════════════════════════════════════════════════════════════════╝
	*/
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	resultstring := htmlparthead

	doc.Find(".anatxt a").Each(func(i int, s *goquery.Selection) {
		resultstring += htmlpartbutton1 + s.Text() + htmlpartbutton2
		linkid, _ := s.Attr("href")

		var linkurl string
		linkurl = strings.ReplaceAll(linkid, "javascript:openWindow('", " ~~~~~~~~~~ redacted-url-2 ~~~~~~~~~~ ")
		linkurl = strings.ReplaceAll(linkurl, "');", "")
		/*
		   ╔══════════════════════════════════════════════════════════════════════════════╗
		   ║ get news details using http.NewRequest and extact detail content using DOM   ║
		   ╚══════════════════════════════════════════════════════════════════════════════╝
		*/
		req, err := http.NewRequest("GET", linkurl, nil)
		if err != nil {
			log.Println(err)
		}

		req.Header.Set("Accept", "text/html;charset=windows-1254")
		req.Header.Set("Content-Type", "text/html; charset=windows-1254")

		resp, err := client.Do(req)

		/*
		   ╔══════════════════════════════════════════════════════════════════════════════╗
		   ║ parse dom document and iterate content to find news detail-text              ║
		   ╚══════════════════════════════════════════════════════════════════════════════╝
		*/
		doc, err := goquery.NewDocumentFromReader(resp.Body)

		doc.Find(".anatxt").Each(func(i int, s *goquery.Selection) {
			resultstring += s.Text() + htmlpartbutton3
		})
	})
	/*
	   ╔══════════════════════════════════════════════════════════════════════════════╗
	   ║ finalize page content add date and time stamp UTC bottom of the page         ║
	   ╚══════════════════════════════════════════════════════════════════════════════╝
	*/
	t := time.Now()
	resultstring += "<br><hr><br><pre>UTC now (RFC822Z) is: " + t.Format(time.RFC822Z) + "</pre><br><hr>"
	resultstring += htmlpartend
	return resultstring, nil
}

func main() {
	http.HandleFunc("/", index)
	http.ListenAndServe(":8080", nil)
}
