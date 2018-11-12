package scripts

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hexya-erp/hexya/hexya/i18n"
	"github.com/hexya-erp/hexya/hexya/i18n/translations"
	"github.com/hexya-erp/hexya/hexya/tools/strutils"
)

func addToTranslationMap(messages translations.MessageMap, lang, moduleName, value, extractedCmt string) translations.MessageMap {
	translated := i18n.TranslateCustom(lang, value, moduleName)
	if translated == value {
		translated = ""
	}
	msgRef := translations.MessageRef{MsgId: value}
	msg := translations.GetOrCreateMessage(messages, msgRef, translated)
	msg.ExtractedComment += extractedCmt
	messages[msgRef] = msg
	return messages
}

func updateFuncJS(messages translations.MessageMap, lang, path, moduleName string) translations.MessageMap {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return messages
	}
	var tVar = `_t`
	var coreVar = `core`
	rxT := regexp.MustCompile(tVar + `\("(.*?)"\)`)
	rxTVar := regexp.MustCompile(`var (.*?) = ` + coreVar + `\._t`)
	rxCoreVar := regexp.MustCompile(`var (.*?) = require\(web\.core\)`)
	for i, line := range strings.Split(string(content), "\n") {
		switch {
		case rxT.MatchString(line):
			matches := rxT.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				addToTranslationMap(messages, lang, moduleName, match[1], fmt.Sprintf("js:%s,%d\n", filepath.Base(path), i))
			}
		case rxTVar.MatchString(line):
			matches := rxTVar.FindStringSubmatch(line)
			tVar = matches[1]
			rxT = regexp.MustCompile(tVar + `\("(.*?)"\)`)
		case rxCoreVar.MatchString(line):
			matches := rxCoreVar.FindStringSubmatch(line)
			coreVar = matches[1]
			rxTVar = regexp.MustCompile(`var (.*?) = ` + coreVar + `\._t`)
		}
	}
	return messages
}

// A Node is an XML Node used for walking down the tree.
type Node struct {
	XMLName xml.Name
	Content []byte     `xml:",innerxml"`
	Nodes   []Node     `xml:",any"`
	Attrs   []xml.Attr `xml:",attr"`
}

func walk(nodes []Node, f func(Node, string) (bool, string), str string) {
	for _, n := range nodes {
		if ok, strNew := f(n, str); ok {
			walk(n.Nodes, f, strNew)
		}
	}
}

func updateFuncXML(messages translations.MessageMap, lang, xmlPath, moduleName string) translations.MessageMap {
	data, err := ioutil.ReadFile(xmlPath)
	buf := bytes.NewBuffer(data)
	dec := xml.NewDecoder(buf)
	var n Node
	err = dec.Decode(&n)
	if err != nil {
		panic(fmt.Errorf("unable to read file %s: %s", xmlPath, err))
	}
	walk([]Node{n}, func(n Node, xmlPath string) (bool, string) {
		content := strings.TrimSpace(string(n.Content))
		for _, attr := range n.Attrs {
			if strutils.IsInStringSlice(attr.Name.Local, []string{`title`, `alt`, `label`, `placeholder`}) && len(attr.Value) > 0 {
				addToTranslationMap(messages, lang, moduleName, attr.Value, fmt.Sprintf("xml:%s\n", path.Join(xmlPath, n.XMLName.Local)))
			}
		}
		if len(content) > 0 && !strings.HasPrefix(content, "<") {
			addToTranslationMap(messages, lang, moduleName, content, fmt.Sprintf("xml:%s\n", path.Join(xmlPath, n.XMLName.Local)))
		}
		return true, path.Join(xmlPath, n.XMLName.Local)
	}, ".")
	return messages
}

// UpdateFunc is the function that extracts strings to translate from XML and JS files.
func UpdateFunc(messages translations.MessageMap, lang, path, moduleName string) translations.MessageMap {
	if filepath.Ext(path) == ".js" {
		return updateFuncJS(messages, lang, path, moduleName)
	}
	return updateFuncXML(messages, lang, path, moduleName)
}
