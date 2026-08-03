package generate

import (
	"encoding/xml"
	"io"
	"regexp"
	"strings"
)

// xnode is an XML element with the text layout the converter needs: Text is
// what sits before the first child, Tail is what sits after the element's own
// end tag. Go's struct unmarshalling folds all character data of an element
// together, which loses the distinction, and the distinction is what tells a
// suggestion's words apart from the words around it.
type xnode struct {
	Name     string
	Attr     map[string]string
	Text     string
	Tail     string
	Children []*xnode
}

// entityDecl matches an internal DTD entity declaration. The Ukrainian rule
// files define punctuation classes that way and then reference them inside
// tokens, so a parser that cannot resolve them reads no rules at all.
var entityDecl = regexp.MustCompile(`<!ENTITY\s+([A-Za-z_][\w.-]*)\s+"([^"]*)"\s*>`)

var predefined = strings.NewReplacer(
	"&lt;", "<",
	"&gt;", ">",
	"&quot;", `"`,
	"&apos;", "'",
	"&amp;", "&",
)

// parseXML reads a document into a tree. Entities declared in the internal
// subset are resolved first, the way the reference implementation's parser
// resolved them.
func parseXML(data []byte) (*xnode, error) {
	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	decoder.Entity = map[string]string{}
	for _, m := range entityDecl.FindAllSubmatch(data, -1) {
		// An entity body carries the predefined entities unexpanded, and they
		// are expanded once when the entity is declared, not again where it is
		// used.
		decoder.Entity[string(m[1])] = predefined.Replace(string(m[2]))
	}

	var root *xnode
	var stack []*xnode
	var tail *xnode

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := token.(type) {
		case xml.StartElement:
			node := &xnode{Name: t.Name.Local, Attr: map[string]string{}}
			for _, a := range t.Attr {
				node.Attr[a.Name.Local] = a.Value
			}
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, node)
			} else if root == nil {
				root = node
			}
			stack = append(stack, node)
			tail = nil
		case xml.CharData:
			if len(stack) == 0 {
				continue
			}
			if tail != nil {
				tail.Tail += string(t)
			} else {
				stack[len(stack)-1].Text += string(t)
			}
		case xml.EndElement:
			if len(stack) == 0 {
				continue
			}
			tail = stack[len(stack)-1]
			stack = stack[:len(stack)-1]
		}
	}
	return root, nil
}

// find returns the first direct child with this name.
func (n *xnode) find(name string) *xnode {
	for _, c := range n.Children {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// findAll returns every direct child with this name. Tokens nested inside an
// <and> or an <or> are deliberately out of reach: they express a condition this
// converter cannot carry across.
func (n *xnode) findAll(name string) []*xnode {
	var out []*xnode
	for _, c := range n.Children {
		if c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

// descendants walks the whole subtree, the node itself included.
func (n *xnode) descendants(name string, fn func(*xnode)) {
	if n.Name == name {
		fn(n)
	}
	for _, c := range n.Children {
		c.descendants(name, fn)
	}
}

// innerText is every piece of character data in the subtree, in document order.
func (n *xnode) innerText() string {
	var b strings.Builder
	b.WriteString(n.Text)
	for _, c := range n.Children {
		b.WriteString(c.innerText())
		b.WriteString(c.Tail)
	}
	return b.String()
}
