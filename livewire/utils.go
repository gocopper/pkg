package livewire

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"html/template"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocopper/copper/cerrors"
)

func updateHTML(html template.HTML, attrs map[string]string, footer string) (template.HTML, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(html)))
	if err != nil {
		return "", cerrors.New(err, "failed to parse html", map[string]interface{}{
			"html": html,
		})
	}
	root := doc.Find("body>*:first-child")
	// todo: validate root node

	for k, v := range attrs {
		root.SetAttr(k, v)
	}

	out, err := goquery.OuterHtml(root)
	if err != nil {
		return "", cerrors.New(err, "failed to render html", nil)
	}
	out += fmt.Sprintf("\n%s\n", footer)

	return template.HTML(out), nil
}

func htmlHash(h template.HTML) string {
	return fmt.Sprintf("%08x", crc32.ChecksumIEEE([]byte(h)))
}

func mergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	for i := range maps {
		for k, v := range maps[i] {
			merged[k] = v
		}
	}

	return merged
}

func structToMap(i interface{}) (ret map[string]interface{}, err error) {
	j, err := json.Marshal(i)
	if err != nil {
		return nil, cerrors.New(err, "failed to marshal struct to json", nil)
	}

	err = json.Unmarshal(j, &ret)
	if err != nil {
		return nil, cerrors.New(err, "failed to unmarshal json to map[string]interface{}", nil)
	}

	return ret, nil
}
