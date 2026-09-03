package yaml

import (
	"fmt"
	"strings"

	goyaml "github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

// defaultYamlIndent is the indentation width used when a document gives no hint.
const defaultYamlIndent = 2

// pathElement is a single step of a yaml key path, as produced by splitYamlPathKey.
type pathElement struct {
	// raw is the element as it appears in the key, including its leading
	// separator (".name", ".'a.b'", "[0]", "[*]", "..name"). Concatenating the
	// raw form of the leading elements after "$" rebuilds a valid path prefix.
	raw string
	// name is the map key this element selects. Only meaningful when isKey is true.
	name string
	// isKey reports whether the element selects a map key, as opposed to a
	// sequence index ("[0]", "[*]") or a recursive descent ("..name").
	isKey bool
}

// splitYamlPathKey splits a sanitized yaml key (see sanitizeYamlPathKey) into its
// successive elements. It mirrors the grammar implemented by goccy's PathString so
// that a prefix rebuilt from these elements is always a path goccy can parse.
func splitYamlPathKey(key string) ([]pathElement, error) {
	buf := []rune(key)
	if len(buf) == 0 || buf[0] != '$' {
		return nil, fmt.Errorf("path %q must start with %q", key, "$")
	}

	var elements []pathElement
	cursor := 1

	for cursor < len(buf) {
		switch buf[cursor] {
		case '.':
			// "..name" is a recursive descent, which we cannot create through.
			if cursor+1 < len(buf) && buf[cursor+1] == '.' {
				start := cursor
				cursor += 2
				for cursor < len(buf) && buf[cursor] != '.' && buf[cursor] != '[' {
					cursor++
				}
				elements = append(elements, pathElement{raw: string(buf[start:cursor])})
				continue
			}

			start := cursor
			cursor++ // skip '.'

			// A quoted key runs until the closing quote, with '\' escaping.
			if cursor < len(buf) && buf[cursor] == '\'' {
				cursor++ // skip opening quote
				var name strings.Builder
				closed := false
				for cursor < len(buf) {
					if buf[cursor] == '\\' && cursor+1 < len(buf) {
						cursor++
						name.WriteRune(buf[cursor])
						cursor++
						continue
					}
					if buf[cursor] == '\'' {
						cursor++ // skip closing quote
						closed = true
						break
					}
					name.WriteRune(buf[cursor])
					cursor++
				}
				if !closed {
					return nil, fmt.Errorf("could not find end delimiter for key in path %q", key)
				}
				elements = append(elements, pathElement{
					raw:   string(buf[start:cursor]),
					name:  name.String(),
					isKey: true,
				})
				continue
			}

			nameStart := cursor
			for cursor < len(buf) && buf[cursor] != '.' && buf[cursor] != '[' {
				cursor++
			}
			if nameStart == cursor {
				return nil, fmt.Errorf("could not find key by empty name in path %q", key)
			}
			elements = append(elements, pathElement{
				raw:   string(buf[start:cursor]),
				name:  string(buf[nameStart:cursor]),
				isKey: true,
			})

		case '[':
			start := cursor
			for cursor < len(buf) && buf[cursor] != ']' {
				cursor++
			}
			if cursor >= len(buf) {
				return nil, fmt.Errorf("could not find closing %q in path %q", "]", key)
			}
			cursor++ // skip ']'
			elements = append(elements, pathElement{raw: string(buf[start:cursor])})

		default:
			return nil, fmt.Errorf("invalid character %q at %d in path %q", string(buf[cursor]), cursor, key)
		}
	}

	return elements, nil
}

// pathPrefix rebuilds the path addressing the first n elements.
func pathPrefix(elements []pathElement, n int) string {
	var sb strings.Builder
	sb.WriteString("$")
	for i := 0; i < n; i++ {
		sb.WriteString(elements[i].raw)
	}
	return sb.String()
}

// deepestExistingNode walks key element by element and returns the node addressed by
// the longest prefix that resolves, together with the number of elements it consumed.
// A nil node with depth 0 means the document itself is empty.
func deepestExistingNode(body ast.Node, elements []pathElement) (ast.Node, int, error) {
	// An empty document has nothing to walk.
	if body == nil {
		return nil, 0, nil
	}

	node := body
	depth := 0

	for i := 1; i <= len(elements); i++ {
		prefix := pathPrefix(elements, i)

		path, err := goyaml.PathString(prefix)
		if err != nil {
			return nil, 0, fmt.Errorf("crafting yamlpath query for key %q: %w", prefix, err)
		}

		found, err := path.FilterNode(body)
		if err != nil {
			// goccy reports descending into a scalar as an invalid query.
			return nil, 0, fmt.Errorf("searching for key %q: %w", prefix, err)
		}

		if found == nil {
			break
		}

		// An explicit null ("key:", "key: ~") carries no mapping to merge onto.
		// Stopping here leaves the null as part of the missing path so that its
		// parent mapping adopts the generated subtree in its place.
		if _, isNull := found.(*ast.NullNode); isNull {
			break
		}

		node = found
		depth = i
	}

	return node, depth, nil
}

// buildMissingSubtree renders the missing elements as a nested mapping ending on
// leafValue. Letting goccy encode the Go value keeps the token positions coherent,
// which is what ast.Merge relies on to re-indent the subtree onto its new parent.
func buildMissingSubtree(missing []pathElement, leafValue interface{}, indent int) (ast.Node, error) {
	value := leafValue
	for i := len(missing) - 1; i >= 0; i-- {
		value = map[string]interface{}{missing[i].name: value}
	}

	node, err := goyaml.ValueToNode(value, goyaml.Indent(indent), goyaml.IndentSequence(true))
	if err != nil {
		return nil, fmt.Errorf("crafting yaml node: %w", err)
	}

	return node, nil
}

// detectIndent reports the indentation width used by the block mappings of a
// document, so that a generated subtree matches the surrounding style. It falls
// back to defaultYamlIndent when the document holds no nested block mapping.
func detectIndent(node ast.Node) int {
	mapping, ok := node.(*ast.MappingNode)
	if !ok || mapping.IsFlowStyle {
		return defaultYamlIndent
	}

	for _, value := range mapping.Values {
		child, ok := value.Value.(*ast.MappingNode)
		if !ok || child.IsFlowStyle || len(child.Values) == 0 {
			continue
		}

		indent := child.Values[0].Key.GetToken().Position.Column - value.Key.GetToken().Position.Column
		if indent > 0 {
			return indent
		}

		// Keep looking deeper when this level is inconclusive.
		if indent := detectIndent(child); indent != defaultYamlIndent {
			return indent
		}
	}

	return defaultYamlIndent
}

// subtreeLeaf descends depth levels into a subtree built by buildMissingSubtree and
// returns the leaf node holding the value.
func subtreeLeaf(node ast.Node, depth int) ast.Node {
	current := node
	for i := 0; i < depth; i++ {
		mapping, ok := current.(*ast.MappingNode)
		if !ok || len(mapping.Values) != 1 {
			return nil
		}
		current = mapping.Values[0].Value
	}
	return current
}

// isNullNode reports whether node is an explicit yaml null ("key:", "key: ~").
// Such a node exists but holds nothing to descend into, so key creation treats it
// as an absent key and overwrites it.
func isNullNode(node ast.Node) bool {
	_, ok := node.(*ast.NullNode)
	return ok
}

// mergeable normalizes a node so that ast.Merge accepts it as a destination. goccy
// currently always exposes maps as *ast.MappingNode, but a single mapping pair can
// surface as *ast.MappingValueNode, which ast.Merge rejects.
func mergeable(node ast.Node) (ast.Node, bool) {
	if mappingValue, ok := node.(*ast.MappingValueNode); ok {
		return ast.Mapping(mappingValue.GetToken(), false, mappingValue), true
	}
	return node, false
}

// createMissingKey creates the key addressed by elements inside doc and sets it to
// leafValue. It returns the newly inserted leaf node so that a comment can be
// attached to it. The caller is responsible for checking that the key is absent.
func createMissingKey(doc *ast.DocumentNode, key string, elements []pathElement, leafValue interface{}) (ast.Node, error) {
	parent, depth, err := deepestExistingNode(doc.Body, elements)
	if err != nil {
		return nil, err
	}

	missing := elements[depth:]
	if len(missing) == 0 {
		return nil, fmt.Errorf("key %q already exists", key)
	}

	// Creating through a sequence index is ambiguous: there is no sensible way to
	// materialize element N of a sequence that does not have one.
	for _, element := range missing {
		if !element.isKey {
			return nil, fmt.Errorf("cannot create key %q: unsupported element %q in a missing path", key, element.raw)
		}
	}

	subtree, err := buildMissingSubtree(missing, leafValue, detectIndent(doc.Body))
	if err != nil {
		return nil, err
	}

	leaf := subtreeLeaf(subtree, len(missing))
	if leaf == nil {
		return nil, fmt.Errorf("cannot create key %q: unexpected generated yaml node", key)
	}

	// An empty document has no node to merge onto, so the subtree becomes the body.
	if parent == nil {
		doc.Body = subtree
		return leaf, nil
	}

	target, wrapped := mergeable(parent)
	if err := ast.Merge(target, subtree); err != nil {
		return nil, fmt.Errorf("creating key %q: %w", key, err)
	}

	if wrapped {
		if err := replaceNode(doc, elements, depth, target); err != nil {
			return nil, fmt.Errorf("creating key %q: %w", key, err)
		}
	}

	return leaf, nil
}

// replaceNode re-attaches node at the position addressed by the first depth elements.
func replaceNode(doc *ast.DocumentNode, elements []pathElement, depth int, node ast.Node) error {
	if depth == 0 {
		doc.Body = node
		return nil
	}

	path, err := goyaml.PathString(pathPrefix(elements, depth))
	if err != nil {
		return err
	}

	file := &ast.File{Docs: []*ast.DocumentNode{doc}}

	return path.ReplaceWithNode(file, node)
}

// appendToSequence appends valueNode to seq unless the sequence already holds value.
// It reports whether the sequence was modified.
func appendToSequence(seq *ast.SequenceNode, value string, comment string) (bool, error) {
	for _, element := range seq.Values {
		// Compare the decoded value so that quoted and folded scalars are not
		// mistaken for a different entry.
		var decoded string
		if err := goyaml.NodeToValue(element, &decoded); err != nil {
			continue
		}
		if decoded == value {
			return false, nil
		}
	}

	appended, err := goyaml.ValueToNode([]interface{}{value}, goyaml.IndentSequence(true))
	if err != nil {
		return false, fmt.Errorf("crafting yaml node: %w", err)
	}

	appendedSeq, ok := appended.(*ast.SequenceNode)
	if !ok {
		return false, fmt.Errorf("unexpected generated yaml node of type %s", appended.Type())
	}

	if comment != "" {
		if err := setNodeComment(appendedSeq.Values[0], comment); err != nil {
			return false, err
		}
	}

	// Merging keeps SequenceNode.ValueHeadComments in sync with Values, which a
	// direct append to Values would not.
	if err := ast.Merge(seq, appendedSeq); err != nil {
		return false, fmt.Errorf("appending to sequence: %w", err)
	}

	return true, nil
}
