package gconfig

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type trieNode struct {
	nodes map[rune]*trieNode
	match bool
}

const (
	maxASCII           = '\u007f'
	deltaChar          = 'a' - 'A'
	chStar             = '*'
	filterTextFileName = "filtertext.txt"
	filterNames        = " ~@#$%^&*()_+-={[}]:;\"'|\\<,>.?/`\t\r\n"
)

func toUpper(r rune) rune {
	if r >= 'a' && r <= 'z' {
		r -= deltaChar
	}
	return r
}

var (
	filterTextRoot = &trieNode{nodes: make(map[rune]*trieNode)}
	filterNameRoot = &trieNode{nodes: make(map[rune]*trieNode)}
)

func init() {
	for _, text := range []rune(filterNames) {
		addNode(filterNameRoot, string(text))
	}
}

func addNode(root *trieNode, text string) {
	chars := []rune(strings.ToUpper(text))
	l := len(chars)
	if l == 0 {
		return
	}
	node := root
	for i := 0; i < l; i++ {
		ch := chars[i]
		if _, ok := node.nodes[ch]; !ok {
			node.nodes[ch] = &trieNode{nodes: make(map[rune]*trieNode)}
		}
		node = node.nodes[ch]
	}
	node.match = true
}

func QueryName(text string) bool {
	return queryText(filterNameRoot, text) || QueryText(text)
}

func QueryText(text string) bool {
	return queryText(filterTextRoot, text)
}

func queryText(root *trieNode, text string) bool {
	chars := []rune(text)
	l := len(chars)
	if l == 0 {
		return false
	}

	nodes := root.nodes
	for i := 0; i < l; i++ {
		ch := toUpper(chars[i])
		node, ok := nodes[ch]
		if !ok {
			continue
		}
		if node.match {
			return true
		}
		nodes = node.nodes
		for j := i + 1; j < l; j++ {
			ch = toUpper(chars[j])
			node, ok := nodes[ch]
			if !ok {
				break
			}
			if node.match {
				return true
			}
			nodes = node.nodes
		}
		nodes = root.nodes
	}
	return false
}

func FilterText(text string) string {
	chars := []rune(text)
	l := len(chars)
	if l == 0 {
		return text
	}

	nodes := filterTextRoot.nodes
	for i := 0; i < l; i++ {
		ch := toUpper(chars[i])
		node, ok := nodes[ch]
		if !ok {
			continue
		}
		if node.match {
			chars[i] = chStar
			continue
		}
		nodes = node.nodes
		pos := 0
		for j := i + 1; j < l; j++ {
			ch = toUpper(chars[j])
			node, ok := nodes[ch]
			if !ok {
				if pos > 0 {
					for idx := i; idx <= pos; idx++ {
						chars[idx] = chStar
					}
					i = pos
				}
				break
			}
			if node.match {
				pos = j
				if j+1 == l {
					for idx := i; idx <= pos; idx++ {
						chars[idx] = chStar
					}
					i = j
					break
				}
			}
			nodes = node.nodes
		}
		nodes = filterTextRoot.nodes
	}
	return string(chars)
}

func LoadFilterTexts(path string) {
	f, err := os.Open(fmt.Sprintf("%s/%s", path, filterTextFileName))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	filter := &trieNode{nodes: make(map[rune]*trieNode)}
	for {
		line, _, _ := reader.ReadLine()
		if len(line) == 0 {
			break
		}
		addNode(filter, strings.Trim(string(line), " "))
	}
	filterTextRoot = filter
}
