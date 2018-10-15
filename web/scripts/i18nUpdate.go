package scripts

import (
	"path/filepath"

	"regexp"

	"io/ioutil"

	"fmt"

	"strings"

	"bytes"
	"encoding/xml"

	"github.com/hexya-erp/hexya/cmd"
	"github.com/hexya-erp/hexya/hexya/i18n"
)

func updateFuncJS(messages cmd.MessageMap, lang, path, moduleName string) cmd.MessageMap {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return messages
	}
	rx := regexp.MustCompile(`_t\("(.*?)"\)`)
	for i, line := range strings.Split(string(content), "\n") {
		matches := rx.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			translated := i18n.TranslateCustom(lang, match[1], moduleName)
			if translated == match[1] {
				translated = ""
			}
			msgRef := cmd.MessageRef{MsgId: match[1]}
			msg := cmd.GetOrCreateMessage(messages, msgRef, translated)
			msg.ExtractedComment += fmt.Sprintf("js:%s,%d\n", filepath.Base(path), i)
			messages[msgRef] = msg
		}
	}
	return messages
}

type Node struct {
	XMLName xml.Name
	Content []byte `xml:",innerxml"`
	Nodes   []Node `xml:",any"`
}

func walk(nodes []Node, f func(Node, string) (bool, string), str string) {
	for _, n := range nodes {
		if ok, strNew := f(n, str); ok {
			walk(n.Nodes, f, strNew)
		}
	}
}

func updateFuncXML(messages cmd.MessageMap, lang, path, moduleName string) cmd.MessageMap {
	data, err := ioutil.ReadFile(path)
	buf := bytes.NewBuffer(data)
	dec := xml.NewDecoder(buf)
	var n Node
	err = dec.Decode(&n)
	if err != nil {
		panic(err)
	}
	walk([]Node{n}, func(n Node, path string) (bool, string) {
		content := strings.TrimSpace(string(n.Content))
		if len(content) > 0 && !strings.HasPrefix(content, "<") {
			translated := i18n.TranslateCustom(lang, content, moduleName)
			if translated == content {
				translated = ""
			}
			msgRef := cmd.MessageRef{MsgId: content}
			msg := cmd.GetOrCreateMessage(messages, msgRef, translated)
			msg.ExtractedComment += fmt.Sprintf("xml:%s\n", path+"/"+n.XMLName.Local)
			messages[msgRef] = msg
		}
		return true, path + "/" + n.XMLName.Local
	}, ".")
	return messages
}

func UpdateFunc(messages cmd.MessageMap, lang, path, moduleName string) cmd.MessageMap {
	if filepath.Ext(path) == ".js" {
		return updateFuncJS(messages, lang, path, moduleName)
	}
	return updateFuncXML(messages, lang, path, moduleName)
}
