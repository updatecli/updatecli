package yaml

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/log"

	"github.com/updatecli/updatecli/pkg/core/result"

	goyaml "github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

func (y Yaml) goYamlTarget(valueToWrite string, resultTarget *result.Target, dryRun bool) (notChanged int, ignoredFiles int, err error) {
	nodeToWrite, err := goyaml.ValueToNode(valueToWrite)
	if err != nil {
		return 0, ignoredFiles, fmt.Errorf("parsing value to write: %w", err)
	}

	if y.spec.Comment != "" {
		if err := setNodeComment(nodeToWrite, y.spec.Comment); err != nil {
			log.Errorf("error setting comment: %s", err)
		}
	}

	keys := y.spec.getKeys()

	resultTargetFilesMap := map[string]bool{}

	for filePath := range y.files {
		originFilePath := y.files[filePath].originalFilePath
		fileNotChanged := 0
		fileKeysProcessed := 0

		yamlFile, err := parser.ParseBytes([]byte(y.files[filePath].content), parser.ParseComments)
		if err != nil {
			return 0, ignoredFiles, fmt.Errorf("parsing yaml file: %w", err)
		}

		// A documentindex addressing no document of the file leaves the keys loop
		// below evaluating nothing at all, reporting neither a match nor a miss, so
		// the file would be silently ignored and the target would succeed without
		// having updated anything. The yamlpath engine and the source both report
		// this as a missing key.
		if y.spec.DocumentIndex != nil && (*y.spec.DocumentIndex < 0 || *y.spec.DocumentIndex >= len(yamlFile.Docs)) {
			if y.spec.SearchPattern {
				log.Debugf("ignoring file %q as it holds %d document(s) and documentindex is %d", originFilePath, len(yamlFile.Docs), *y.spec.DocumentIndex)
				ignoredFiles++
				continue
			}

			return 0, ignoredFiles, fmt.Errorf("documentindex %d addresses no document of file %q, which holds %d document(s)", *y.spec.DocumentIndex, originFilePath, len(yamlFile.Docs))
		}

		// Process each key for this file
		for _, key := range keys {
			urlPath, err := goyaml.PathString(key)
			if err != nil {
				return 0, ignoredFiles, fmt.Errorf("crafting yamlpath query for key %q: %w", key, err)
			}

			keyNotFound := []string{}
			errMsg := []string{}
			// keyProcessed reports whether the key was resolved, or created, in at
			// least one of the evaluated documents.
			keyProcessed := false
			keyChanged := false

			for index, doc := range yamlFile.Docs {
				if y.spec.DocumentIndex != nil {
					if index != *y.spec.DocumentIndex {
						continue
					}
				}

				// goccy reports a missing key as a nil node rather than an error,
				// but reports an intermediate null value ("key:") as an invalid
				// query. Both mean the key is absent, so when we are allowed to
				// create it we let our own walker decide rather than FilterNode.
				// Aliases and merge keys are invisible to that query, so a document
				// that uses them gets a second walk before the key is declared missing.
				docIndex := indexDocument(doc)
				node, matched, missing, filterErr, throughAlias, locateErr := locateDocumentTargets(doc, urlPath, key, docIndex)
				if locateErr != nil {
					return 0, ignoredFiles, fmt.Errorf("searching for key %q in document index %d of file %q: %w", key, index, originFilePath, locateErr)
				}

				switch {
				// goccy's selector replacement only rewrites the positions that
				// already hold the key, and createmissingkey cannot create the
				// others either, so a partial match would quietly write less than
				// the manifest asks for. spec.searchpattern asks for that
				// tolerance explicitly, so it updates what it found instead.
				case filterErr == nil && len(matched) > 0 && missing > 0 && !y.spec.SearchPattern:
					errMsg = append(errMsg, fmt.Sprintf("key %q from file %q is missing from %d of the %d nodes it selects in document index %d", key, originFilePath, missing, missing+len(matched), index))

				// A null value is an existing node that holds nothing, so when we
				// may create the key we overwrite it instead of updating it.
				case filterErr == nil && len(matched) > 0 && (!y.spec.CreateMissingKey || !isNullNode(node)):
					if missing > 0 {
						log.Debugf("key %q from file %q is missing from %d of the %d nodes it selects in document %d, updating the %d nodes holding it", key, originFilePath, missing, missing+len(matched), index, len(matched))
					}

					changed, err := y.updateNode(yamlFile, index, doc, urlPath, matched, key, valueToWrite, nodeToWrite, resultTarget, originFilePath, dryRun, docIndex, throughAlias)
					if err != nil {
						return 0, ignoredFiles, err
					}
					keyProcessed = true
					keyChanged = keyChanged || changed

				case y.spec.CreateMissingKey:
					changed, err := y.createKey(doc, key, valueToWrite, resultTarget, originFilePath, dryRun)
					if err != nil {
						return 0, ignoredFiles, err
					}
					keyProcessed = true
					keyChanged = keyChanged || changed

				case filterErr != nil:
					errMsg = append(errMsg, fmt.Sprintf("searching for key %q in document index %d: %s", key, index, filterErr.Error()))

				case y.spec.SearchPattern:
					// If search pattern is true then we don't want to return an error
					// as we are probably trying to identify a file matching the key
					log.Debugf("ignoring key %q from file %q in document %d as it does not exist", key, originFilePath, index)

				default:
					keyNotFound = append(keyNotFound, fmt.Sprintf("couldn't find key %q from file %q in document %d", key, originFilePath, index))
				}
			}

			if len(errMsg) > 0 {
				return 0, ignoredFiles, fmt.Errorf("errors occurred:\n%s", strings.Join(errMsg, "\n"))
			}

			if len(keyNotFound) > 0 {
				for _, msg := range keyNotFound {
					log.Errorln(msg)
				}
				return 0, ignoredFiles, fmt.Errorf("key not found from file %q", originFilePath)
			}

			if !keyProcessed {
				continue
			}

			fileKeysProcessed++

			if !keyChanged {
				fileNotChanged++
				continue
			}

			if _, ok := resultTargetFilesMap[filePath]; !ok {
				resultTarget.Files = append(resultTarget.Files, y.files[filePath].filePath)
				resultTargetFilesMap[filePath] = true
			}

			resultTarget.Changed = true
			resultTarget.Result = result.ATTENTION
		}

		// If no keys were processed for this file (all were ignored), count as ignored
		if fileKeysProcessed == 0 {
			ignoredFiles++
			continue
		}

		// If all processed keys in this file were unchanged, count as not changed
		if fileNotChanged == fileKeysProcessed {
			notChanged++
		}

		// Update file content
		f := y.files[filePath]
		f.content = yamlFile.String()
		y.files[filePath] = f

		if !dryRun {
			newFile, err := os.Create(y.files[filePath].filePath)
			if err != nil {
				return 0, ignoredFiles, fmt.Errorf("creating file %q: %w", originFilePath, err)
			}
			defer newFile.Close()

			err = y.contentRetriever.WriteToFile(
				y.files[filePath].content,
				y.files[filePath].filePath)
			if err != nil {
				return 0, ignoredFiles, fmt.Errorf("saving file %q: %w", originFilePath, err)
			}
		}
	}

	return notChanged, ignoredFiles, nil
}

// createKey inserts a key that does not exist yet in doc. When the target also
// appends to an array, the key is created holding the value as its sole entry.
func (y Yaml) createKey(doc *ast.DocumentNode, key, valueToWrite string, resultTarget *result.Target, originFilePath string, dryRun bool) (bool, error) {
	elements, err := splitYamlPathKey(key)
	if err != nil {
		return false, fmt.Errorf("cannot create key %q: %w", key, err)
	}

	var leafValue interface{} = valueToWrite
	if y.spec.AppendToArray {
		leafValue = []interface{}{valueToWrite}
	}

	leaf, err := createMissingKey(doc, key, elements, leafValue)
	if err != nil {
		return false, err
	}

	if y.spec.Comment != "" {
		if y.spec.AppendToArray {
			// The comment belongs to the sole entry of the created sequence.
			if sequence, ok := leaf.(*ast.SequenceNode); ok && len(sequence.Values) == 1 {
				leaf = sequence.Values[0]
			}
		}
		if err := setNodeComment(leaf, y.spec.Comment); err != nil {
			log.Errorf("error setting comment: %s", err)
		}
	}

	resultTarget.Description = fmt.Sprintf("%s\nkey %q%screated in file %q with value %q",
		resultTarget.Description,
		key,
		shouldMessage(dryRun),
		originFilePath,
		valueToWrite)

	return true, nil
}

// updateNode either appends to the sequences addressed by key, or replaces the value
// of the nodes it addresses. matched holds every node the key selected, which is more
// than one for a wildcard or a recursive selector. It reports whether the document
// was modified.
//
// Aliases are resolved before the value is compared or written. An alias update
// rewrites the anchored value so every reference keeps its *alias, and a direct
// update of an anchored node keeps the &anchor. ReplaceWithNode stays in place for
// matches that never touch an anchor or an alias.
func (y Yaml) updateNode(yamlFile *ast.File, index int, doc *ast.DocumentNode, urlPath *goyaml.Path, matched []ast.Node, key, valueToWrite string, nodeToWrite ast.Node, resultTarget *result.Target, originFilePath string, dryRun bool, docIndex *anchorIndex, throughAlias bool) (bool, error) {
	resolved, err := docIndex.resolveMatches(matched)
	if err != nil {
		return false, fmt.Errorf("resolving key %q in file %q: %w", key, originFilePath, err)
	}
	values := resolvedValues(resolved)

	if y.spec.AppendToArray {
		return y.appendResolvedSequences(values, key, valueToWrite, resultTarget, originFilePath, dryRun, docIndex)
	}

	oldVersion := matchedValues(values)
	resultTarget.Information = oldVersion

	// Every selected node must already hold the value for the target to be a no-op,
	// as ReplaceWithNode rewrites all of them at once.
	if alreadySet(values, valueToWrite) {
		resultTarget.Description = fmt.Sprintf("%s\nkey %q already set to %q, from file %q",
			resultTarget.Description,
			key,
			valueToWrite,
			originFilePath)
		return false, nil
	}

	if err := docIndex.checkReplacementAnchors(values); err != nil {
		return false, fmt.Errorf("replacing yaml key %q: %w", key, err)
	}

	if matchesUseAnchor(resolved, throughAlias) {
		if err := y.updateInPlace(resolved, docIndex, valueToWrite); err != nil {
			return false, fmt.Errorf("replacing yaml key %q: %w", key, err)
		}
	} else if err := replaceMatchedNode(yamlFile, index, doc, urlPath, nodeToWrite, key); err != nil {
		return false, err
	}

	resultTarget.Description = fmt.Sprintf("%s\nkey %q%supdated from %q to %q, in file %q",
		resultTarget.Description,
		key,
		shouldMessage(dryRun),
		oldVersion,
		valueToWrite,
		originFilePath)

	return true, nil
}

// appendResolvedSequences appends valueToWrite to each sequence selected by key.
// Sequences reached through several aliases of one anchor are appended to once.
func (y Yaml) appendResolvedSequences(values []ast.Node, key, valueToWrite string, resultTarget *result.Target, originFilePath string, dryRun bool, docIndex *anchorIndex) (bool, error) {
	sequences, err := appendTargetSequences(values, key, originFilePath)
	if err != nil {
		return false, err
	}

	appended := false
	seen := map[*ast.SequenceNode]struct{}{}
	for _, sequence := range sequences {
		if _, ok := seen[sequence]; ok {
			continue
		}
		seen[sequence] = struct{}{}

		sequenceAppended, err := appendToSequence(sequence, valueToWrite, y.spec.Comment, docIndex)
		if err != nil {
			return false, fmt.Errorf("appending to key %q from file %q: %w", key, originFilePath, err)
		}
		appended = appended || sequenceAppended
	}

	if !appended {
		resultTarget.Description = fmt.Sprintf("%s\nkey %q already contains %q, from file %q",
			resultTarget.Description,
			key,
			valueToWrite,
			originFilePath)
		return false, nil
	}

	resultTarget.Description = fmt.Sprintf("%s\nvalue %q%sappended to key %q, in file %q",
		resultTarget.Description,
		valueToWrite,
		shouldMessage(dryRun),
		key,
		originFilePath)

	return true, nil
}

// replaceMatchedNode rewrites every node selected by urlPath. It is the update used
// when none of those nodes is an alias or an anchor.
func replaceMatchedNode(yamlFile *ast.File, index int, doc *ast.DocumentNode, urlPath *goyaml.Path, nodeToWrite ast.Node, key string) error {
	tmpYAMLFile := ast.File{
		Name: yamlFile.Name,
	}
	tmpYAMLFile.Docs = append(tmpYAMLFile.Docs, doc)
	if err := urlPath.ReplaceWithNode(&tmpYAMLFile, nodeToWrite); err != nil {
		return fmt.Errorf("replacing yaml key %q: %w", key, err)
	}
	yamlFile.Docs[index].Body = tmpYAMLFile.Docs[0].Body

	return nil
}

// shouldMessage crafts the message fragment depending on whether Updatecli runs in
// dry run mode or not.
func shouldMessage(dryRun bool) string {
	if dryRun {
		return " should be "
	}
	return " "
}

// matchedValues renders the value(s) held by the matched nodes, for the target
// description and information.
func matchedValues(matched []ast.Node) string {
	if len(matched) == 1 {
		return matched[0].String()
	}

	values := make([]string, 0, len(matched))
	for _, node := range matched {
		values = append(values, strings.TrimSpace(node.String()))
	}

	return strings.Join(values, ", ")
}

// alreadySet reports whether every matched node already holds valueToWrite. The
// decoded value is compared so that folded and literal scalars (">-", "|") are not
// flagged as changed by their formatting markers. See issue #8295.
func alreadySet(matched []ast.Node, valueToWrite string) bool {
	for _, node := range matched {
		var decoded string
		if err := goyaml.NodeToValue(node, &decoded); err != nil || decoded != valueToWrite {
			return false
		}
	}

	return len(matched) > 0
}

// anchorIndex resolves aliases inside one YAML document. YAML 1.2 gives each
// document its own anchor scope, and an alias binds to the most recent preceding
// anchor of that name. parents records the mapping or sequence slot that holds a
// node so an inline sibling can be replaced without disturbing an alias next to it.
type anchorIndex struct {
	aliasToAnchor map[*ast.AliasNode]*ast.AnchorNode
	parents       map[ast.Node]nodeParent
	hasAlias      bool
	hasMergeKey   bool
}

// nodeParent is the slot that owns a node in the document tree.
type nodeParent struct {
	mapping  *ast.MappingValueNode
	sequence *ast.SequenceNode
	index    int
	anchor   *ast.AnchorNode
	document *ast.DocumentNode
}

// resolvedMatch is a selected node after aliases and anchors have been followed.
// anchor is the node whose value must change when the match is an alias or an
// anchor; it is nil for an ordinary inline value.
type resolvedMatch struct {
	value  ast.Node
	anchor *ast.AnchorNode
}

// anchorWalker indexes anchors and aliases in document order.
type anchorWalker struct {
	anchors map[string]*ast.AnchorNode
	index   *anchorIndex
}

// Visit records parent slots and binds each alias to the most recent preceding anchor.
func (w *anchorWalker) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *ast.DocumentNode:
		w.index.setParent(n.Body, nodeParent{document: n})
	case *ast.MappingValueNode:
		if n.Key != nil && n.Key.IsMergeKey() {
			w.index.hasMergeKey = true
		}
		w.index.setParent(n.Value, nodeParent{mapping: n})
	case *ast.SequenceNode:
		for i, value := range n.Values {
			w.index.setParent(value, nodeParent{sequence: n, index: i})
		}
	case *ast.AnchorNode:
		w.index.setParent(n.Value, nodeParent{anchor: n})
		if name := anchorName(n); name != "" {
			w.anchors[name] = n
		}
	case *ast.AliasNode:
		w.index.hasAlias = true
		if anchor, ok := w.anchors[aliasName(n)]; ok {
			w.index.aliasToAnchor[n] = anchor
		}
	}

	return w
}

func (idx *anchorIndex) setParent(child ast.Node, parent nodeParent) {
	if child == nil {
		return
	}
	if _, ok := idx.parents[child]; ok {
		return
	}
	idx.parents[child] = parent
}

// indexDocument walks doc in document order and indexes its anchors and aliases.
func indexDocument(doc *ast.DocumentNode) *anchorIndex {
	idx := &anchorIndex{
		aliasToAnchor: map[*ast.AliasNode]*ast.AnchorNode{},
		parents:       map[ast.Node]nodeParent{},
	}
	if doc == nil || doc.Body == nil {
		return idx
	}

	ast.Walk(&anchorWalker{
		anchors: map[string]*ast.AnchorNode{},
		index:   idx,
	}, doc)

	return idx
}

// locateDocumentTargets returns the nodes selected by key. When goccy's path query
// cannot see through an alias or a merge key, the document's own anchor index is
// used instead. throughAlias reports that the match was found by that fallback, so
// the update must edit the anchored node rather than call ReplaceWithNode.
func locateDocumentTargets(doc *ast.DocumentNode, urlPath *goyaml.Path, key string, idx *anchorIndex) (ast.Node, []ast.Node, int, error, bool, error) {
	node, filterErr := urlPath.FilterNode(doc.Body)

	var matched []ast.Node
	var missing int
	if filterErr == nil {
		var err error
		matched, missing, err = matchedNodes(node, key)
		if err != nil {
			return nil, nil, 0, nil, false, err
		}
	}

	recursive := hasRecursiveSelector(key)
	if (filterErr != nil || len(matched) == 0 || missing > 0 || recursive) && idx != nil && (idx.hasAlias || idx.hasMergeKey) {
		found, foundMissing, err := idx.matchPath(doc.Body, key)
		if err != nil {
			return nil, nil, 0, nil, false, err
		}
		// goccy stops evaluating a path at a recursive selector and ignores any
		// trailing segments, so it can return an ancestor node even when the
		// full path has no match. Prefer matchPath's full-path result for
		// recursive paths even when found is empty.
		if recursive || (len(found) > 0 && (filterErr != nil || len(matched) == 0 || hasAdditionalNodes(found, matched) || foundMissing < missing)) {
			return node, found, foundMissing, nil, true, nil
		}
	}

	return node, matched, missing, filterErr, false, nil
}

func hasRecursiveSelector(key string) bool {
	elements, err := splitYamlPathKey(key)
	if err != nil {
		return false
	}
	for _, element := range elements {
		if strings.HasPrefix(element.raw, "..") {
			return true
		}
	}

	return false
}

func hasAdditionalNodes(found, matched []ast.Node) bool {
	if len(found) > len(matched) {
		return true
	}

	seen := make(map[ast.Node]struct{}, len(matched))
	for _, node := range matched {
		seen[node] = struct{}{}
	}
	for _, node := range found {
		if _, ok := seen[node]; !ok {
			return true
		}
	}

	return false
}

// resolveMatches follows every alias in nodes, unwrapping anchors along the way.
// An unresolved alias or an alias cycle is an error: guessing would write a scalar
// over the *alias and leave the anchored value stale.
func (idx *anchorIndex) resolveMatches(nodes []ast.Node) ([]resolvedMatch, error) {
	matches := make([]resolvedMatch, 0, len(nodes))
	for _, node := range nodes {
		value, anchor, err := idx.resolve(node)
		if err != nil {
			return nil, err
		}
		matches = append(matches, resolvedMatch{value: value, anchor: anchor})
	}

	return matches, nil
}

// resolve follows chained aliases until it reaches the value they share. The
// returned anchor is the one whose Value holds that result, so updating it keeps
// every alias in the chain intact.
func (idx *anchorIndex) resolve(node ast.Node) (ast.Node, *ast.AnchorNode, error) {
	seenAliases := map[*ast.AliasNode]struct{}{}
	seenAnchors := map[*ast.AnchorNode]struct{}{}
	current := node
	var anchor *ast.AnchorNode

	for {
		switch typed := current.(type) {
		case *ast.AliasNode:
			if _, ok := seenAliases[typed]; ok {
				return nil, nil, fmt.Errorf("cycle in yaml alias %q", aliasName(typed))
			}
			seenAliases[typed] = struct{}{}

			next, ok := idx.aliasToAnchor[typed]
			if !ok || next == nil {
				return nil, nil, fmt.Errorf("unresolved yaml alias %q", aliasName(typed))
			}
			anchor = next
			current = next.Value
		case *ast.AnchorNode:
			if _, ok := seenAnchors[typed]; ok {
				return nil, nil, fmt.Errorf("cycle in yaml anchor %q", anchorName(typed))
			}
			seenAnchors[typed] = struct{}{}
			anchor = typed
			current = typed.Value
		default:
			return current, anchor, nil
		}
	}
}

func resolvedValues(matches []resolvedMatch) []ast.Node {
	values := make([]ast.Node, len(matches))
	for i, match := range matches {
		values[i] = match.value
	}

	return values
}

// matchesUseAnchor reports whether the update has to edit anchors in place.
// throughAlias means the path itself crossed an alias or a merge key, so goccy's
// ReplaceWithNode cannot see the node it has to change.
func matchesUseAnchor(matches []resolvedMatch, throughAlias bool) bool {
	if throughAlias {
		return true
	}
	for _, match := range matches {
		if match.anchor != nil {
			return true
		}
	}

	return false
}

// subtreeAnchorWalker collects anchors defined and aliases used within a
// replaced subtree so replacements that would orphan or rebind external
// aliases can be rejected before mutating the document.
type subtreeAnchorWalker struct {
	anchors     map[*ast.AnchorNode]struct{}
	anchorOrder []*ast.AnchorNode
	aliases     map[*ast.AliasNode]struct{}
}

func (w *subtreeAnchorWalker) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *ast.AnchorNode:
		if _, ok := w.anchors[n]; !ok {
			w.anchors[n] = struct{}{}
			w.anchorOrder = append(w.anchorOrder, n)
		}
	case *ast.AliasNode:
		w.aliases[n] = struct{}{}
	}

	return w
}

// checkReplacementAnchors verifies that replacing subtrees will not remove or
// rebind any anchor definition nested inside them that is referenced by an
// alias outside those subtrees.
func (idx *anchorIndex) checkReplacementAnchors(subtrees []ast.Node) error {
	if idx == nil || len(idx.aliasToAnchor) == 0 {
		return nil
	}

	walker := &subtreeAnchorWalker{
		anchors: map[*ast.AnchorNode]struct{}{},
		aliases: map[*ast.AliasNode]struct{}{},
	}
	seenSubtrees := map[ast.Node]struct{}{}
	for _, subtree := range subtrees {
		if subtree == nil {
			continue
		}
		if _, ok := seenSubtrees[subtree]; ok {
			continue
		}
		seenSubtrees[subtree] = struct{}{}
		ast.Walk(walker, subtree)
	}

	if len(walker.anchors) == 0 {
		return nil
	}

	danglingAnchors := map[*ast.AnchorNode]struct{}{}
	for alias, anchor := range idx.aliasToAnchor {
		if _, removed := walker.anchors[anchor]; !removed {
			continue
		}
		if _, aliasRemoved := walker.aliases[alias]; aliasRemoved {
			continue
		}
		danglingAnchors[anchor] = struct{}{}
	}

	if len(danglingAnchors) == 0 {
		return nil
	}

	for _, anchor := range walker.anchorOrder {
		if _, ok := danglingAnchors[anchor]; ok {
			return fmt.Errorf("cannot remove anchor %q because it is referenced by an alias outside the replaced subtree", anchorName(anchor))
		}
	}

	return nil
}

// updateInPlace writes valueToWrite into each anchored value once, and replaces
// inline siblings through the slot that holds them.
func (y Yaml) updateInPlace(matches []resolvedMatch, idx *anchorIndex, valueToWrite string) error {
	updatedAnchors := map[*ast.AnchorNode]struct{}{}
	updatedValues := map[ast.Node]struct{}{}

	for _, match := range matches {
		if match.anchor != nil {
			if _, ok := updatedAnchors[match.anchor]; ok {
				continue
			}
			updatedAnchors[match.anchor] = struct{}{}

			fresh, err := y.freshValueNode(valueToWrite)
			if err != nil {
				return err
			}
			replaceAnchorValue(match.anchor, fresh)
			continue
		}

		if match.value == nil {
			return fmt.Errorf("cannot update an empty yaml node")
		}
		if _, ok := updatedValues[match.value]; ok {
			continue
		}
		updatedValues[match.value] = struct{}{}

		fresh, err := y.freshValueNode(valueToWrite)
		if err != nil {
			return err
		}
		parent, ok := idx.parents[match.value]
		if !ok {
			return fmt.Errorf("cannot locate parent of yaml node %q", match.value.String())
		}
		if err := parent.replace(fresh); err != nil {
			return err
		}
	}

	return nil
}

// freshValueNode builds the scalar written into an anchor or an inline sibling.
// Each call returns a new node: column adjustment mutates the node it is given,
// and the comment belongs on that updated value.
func (y Yaml) freshValueNode(valueToWrite string) (ast.Node, error) {
	node, err := goyaml.ValueToNode(valueToWrite)
	if err != nil {
		return nil, fmt.Errorf("parsing value to write: %w", err)
	}
	if y.spec.Comment != "" {
		if err := setNodeComment(node, y.spec.Comment); err != nil {
			log.Errorf("error setting comment: %s", err)
		}
	}

	return node, nil
}

func replaceAnchorValue(anchor *ast.AnchorNode, value ast.Node) {
	if anchor.Value != nil && anchor.Value.GetToken() != nil && value != nil && value.GetToken() != nil {
		column := anchor.Value.GetToken().Position.Column - value.GetToken().Position.Column
		value.AddColumn(column)
	}
	anchor.Value = value
}

func (p nodeParent) replace(value ast.Node) error {
	switch {
	case p.mapping != nil:
		return p.mapping.Replace(value)
	case p.sequence != nil:
		return p.sequence.Replace(p.index, value)
	case p.anchor != nil:
		replaceAnchorValue(p.anchor, value)
		return nil
	case p.document != nil:
		p.document.Body = value
		return nil
	default:
		return fmt.Errorf("yaml node has no parent slot")
	}
}

const (
	maxMatchDepth   = 1024
	maxMatchSteps   = 100000
	maxMatchResults = 10000
)

type matchState struct {
	node ast.Node
	step int
}

type matchResult struct {
	nodes   []ast.Node
	missing int
	empty   bool
}

type pathMatcher struct {
	idx        *anchorIndex
	elements   []pathElement
	memo       map[matchState]matchResult
	inProgress map[matchState]struct{}
	steps      int
}

// matchPath walks key across aliases and << merge keys. It returns the nodes that
// hold the addressed value (still wrapped in an alias or an anchor when that is
// how they appear in the document) together with how many wildcard positions
// lacked a match.
func (idx *anchorIndex) matchPath(root ast.Node, key string) ([]ast.Node, int, error) {
	if root == nil {
		return nil, 0, nil
	}

	elements, err := splitYamlPathKey(key)
	if err != nil {
		return nil, 0, err
	}

	matcher := &pathMatcher{
		idx:        idx,
		elements:   elements,
		memo:       map[matchState]matchResult{},
		inProgress: map[matchState]struct{}{},
	}

	res, _, err := matcher.matchFrom(root, 0, 0)
	if err != nil {
		return nil, 0, err
	}

	return dedupeASTNodes(res.nodes), res.missing, nil
}

func (m *pathMatcher) consumeStep() error {
	m.steps++
	if m.steps > maxMatchSteps {
		return fmt.Errorf("yaml path matching exceeded maximum traversal limit of %d steps", maxMatchSteps)
	}

	return nil
}

func addMissing(a, b int) int {
	if b <= 0 {
		return a
	}
	const maxMissing = 1<<30 - 1
	if a >= maxMissing || b >= maxMissing-a {
		return maxMissing
	}

	return a + b
}

func (m *pathMatcher) appendUniqueNodes(dst []ast.Node, seen map[ast.Node]struct{}, src []ast.Node) ([]ast.Node, error) {
	for _, n := range src {
		if err := m.consumeStep(); err != nil {
			return nil, err
		}
		if n == nil {
			continue
		}
		if _, ok := seen[n]; ok {
			continue
		}
		if len(dst) >= maxMatchResults {
			return nil, fmt.Errorf("yaml path matching exceeded maximum result limit of %d nodes", maxMatchResults)
		}
		seen[n] = struct{}{}
		dst = append(dst, n)
	}

	return dst, nil
}

func (m *pathMatcher) matchFrom(node ast.Node, step int, depth int) (matchResult, bool, error) {
	if step == len(m.elements) {
		if node == nil {
			return matchResult{}, false, nil
		}
		return matchResult{nodes: []ast.Node{node}}, false, nil
	}

	if depth > maxMatchDepth {
		return matchResult{}, false, fmt.Errorf("yaml path matching exceeded maximum depth of %d", maxMatchDepth)
	}
	if err := m.consumeStep(); err != nil {
		return matchResult{}, false, err
	}

	unwrapped, err := m.idx.unwrapContainer(node)
	if err != nil {
		return matchResult{}, false, err
	}
	if unwrapped == nil {
		return matchResult{}, false, nil
	}

	state := matchState{node: unwrapped, step: step}
	if cached, ok := m.memo[state]; ok {
		return cached, false, nil
	}
	if _, ok := m.inProgress[state]; ok {
		return matchResult{}, true, nil
	}

	m.inProgress[state] = struct{}{}
	defer delete(m.inProgress, state)

	element := m.elements[step]
	var (
		res      matchResult
		hitCycle bool
	)
	switch {
	case element.isKey:
		res, hitCycle, err = m.matchKey(unwrapped, element.name, step+1, depth+1)
	case element.raw == "[*]":
		res, hitCycle, err = m.matchAll(unwrapped, step+1, depth+1)
	case strings.HasPrefix(element.raw, ".."):
		res, hitCycle, err = m.matchRecursive(unwrapped, strings.TrimPrefix(element.raw, ".."), step, depth+1)
	case strings.HasPrefix(element.raw, "[") && strings.HasSuffix(element.raw, "]"):
		res, hitCycle, err = m.matchIndex(unwrapped, element.raw, step+1, depth+1)
	default:
		return matchResult{}, false, nil
	}
	if err != nil {
		return matchResult{}, false, err
	}

	if !hitCycle {
		m.memo[state] = res
	}

	return res, hitCycle, nil
}

func (m *pathMatcher) matchKey(node ast.Node, name string, nextStep int, depth int) (matchResult, bool, error) {
	mapping := asMapping(node)
	if mapping == nil {
		return matchResult{}, false, nil
	}

	value, found, err := m.lookupMappingKey(mapping, name, depth, map[*ast.MappingNode]struct{}{})
	if err != nil || !found {
		return matchResult{}, false, err
	}

	return m.matchFrom(value, nextStep, depth)
}

// lookupMappingKey resolves key name on mapping according to YAML shallow-merge
// precedence: local keys override << merges, and an earlier << merge target that
// defines name shadows later merge targets even if remaining path elements do not
// match inside its value.
func (m *pathMatcher) lookupMappingKey(mapping *ast.MappingNode, name string, depth int, seen map[*ast.MappingNode]struct{}) (ast.Node, bool, error) {
	if mapping == nil {
		return nil, false, nil
	}
	if depth > maxMatchDepth {
		return nil, false, fmt.Errorf("yaml path matching exceeded maximum depth of %d", maxMatchDepth)
	}
	if _, ok := seen[mapping]; ok {
		return nil, false, nil
	}
	seen[mapping] = struct{}{}

	if err := m.consumeStep(); err != nil {
		return nil, false, err
	}

	for _, value := range mapping.Values {
		if value == nil || value.Key == nil || value.Key.IsMergeKey() {
			continue
		}
		if mappingKeyName(value.Key) == name {
			return value.Value, true, nil
		}
	}

	for _, value := range mapping.Values {
		if value == nil || value.Key == nil || !value.Key.IsMergeKey() {
			continue
		}

		targets, err := m.idx.mergeTargets(value.Value)
		if err != nil {
			return nil, false, err
		}
		for _, target := range targets {
			if err := m.consumeStep(); err != nil {
				return nil, false, err
			}
			targetMapping := asMapping(target)
			if targetMapping == nil {
				continue
			}
			val, found, err := m.lookupMappingKey(targetMapping, name, depth+1, seen)
			if err != nil {
				return nil, false, err
			}
			if found {
				return val, true, nil
			}
		}
	}

	return nil, false, nil
}

func (m *pathMatcher) matchAll(node ast.Node, nextStep int, depth int) (matchResult, bool, error) {
	sequence, ok := node.(*ast.SequenceNode)
	if !ok || sequence == nil {
		return matchResult{}, false, nil
	}

	if len(sequence.Values) == 0 {
		return matchResult{empty: true}, false, nil
	}

	var (
		matched  []ast.Node
		missing  int
		hitCycle bool
		sawEmpty bool
	)
	seenNodes := map[ast.Node]struct{}{}
	for _, value := range sequence.Values {
		if err := m.consumeStep(); err != nil {
			return matchResult{}, false, err
		}
		if value == nil {
			if nextStep < len(m.elements) {
				missing = addMissing(missing, 1)
			}
			continue
		}

		sub, subCycle, err := m.matchFrom(value, nextStep, depth)
		if err != nil {
			return matchResult{}, false, err
		}
		hitCycle = hitCycle || subCycle
		sawEmpty = sawEmpty || sub.empty

		if nextStep < len(m.elements) {
			missing = addMissing(missing, sub.missing)
			if len(sub.nodes) == 0 && sub.missing == 0 && !sub.empty {
				missing = addMissing(missing, 1)
			}
		}

		matched, err = m.appendUniqueNodes(matched, seenNodes, sub.nodes)
		if err != nil {
			return matchResult{}, false, err
		}
	}

	return matchResult{
		nodes:   matched,
		missing: missing,
		empty:   len(matched) == 0 && missing == 0 && sawEmpty,
	}, hitCycle, nil
}

func (m *pathMatcher) matchIndex(node ast.Node, raw string, nextStep int, depth int) (matchResult, bool, error) {
	sequence, ok := node.(*ast.SequenceNode)
	if !ok || sequence == nil || len(raw) < 3 {
		return matchResult{}, false, nil
	}

	index, convErr := strconv.Atoi(raw[1 : len(raw)-1])
	if convErr != nil || index < 0 || index >= len(sequence.Values) || sequence.Values[index] == nil {
		return matchResult{}, false, nil
	}

	return m.matchFrom(sequence.Values[index], nextStep, depth)
}

func (m *pathMatcher) effectiveMappingValues(mapping *ast.MappingNode, depth int) ([]*ast.MappingValueNode, error) {
	hasMerge := false
	for _, value := range mapping.Values {
		if value != nil && value.Key != nil && value.Key.IsMergeKey() {
			hasMerge = true
			break
		}
	}
	if !hasMerge {
		return mapping.Values, nil
	}

	var entries []*ast.MappingValueNode
	defined := map[string]struct{}{}
	seen := map[*ast.MappingNode]struct{}{}
	if err := m.collectMappingValues(mapping, depth, defined, seen, &entries); err != nil {
		return nil, err
	}

	return entries, nil
}

func (m *pathMatcher) collectMappingValues(mapping *ast.MappingNode, depth int, defined map[string]struct{}, seen map[*ast.MappingNode]struct{}, out *[]*ast.MappingValueNode) error {
	if mapping == nil {
		return nil
	}
	if depth > maxMatchDepth {
		return fmt.Errorf("yaml path matching exceeded maximum depth of %d", maxMatchDepth)
	}
	if _, ok := seen[mapping]; ok {
		return nil
	}
	seen[mapping] = struct{}{}

	if err := m.consumeStep(); err != nil {
		return err
	}

	for _, value := range mapping.Values {
		if value == nil || value.Key == nil || value.Key.IsMergeKey() {
			continue
		}
		keyName := mappingKeyName(value.Key)
		if _, exists := defined[keyName]; exists {
			continue
		}
		defined[keyName] = struct{}{}
		*out = append(*out, value)
	}

	for _, value := range mapping.Values {
		if value == nil || value.Key == nil || !value.Key.IsMergeKey() {
			continue
		}

		targets, err := m.idx.mergeTargets(value.Value)
		if err != nil {
			return err
		}
		for _, target := range targets {
			if err := m.consumeStep(); err != nil {
				return err
			}
			targetMapping := asMapping(target)
			if targetMapping == nil {
				continue
			}
			if err := m.collectMappingValues(targetMapping, depth+1, defined, seen, out); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *pathMatcher) matchRecursive(node ast.Node, name string, step int, depth int) (matchResult, bool, error) {
	var (
		matched  []ast.Node
		missing  int
		hitCycle bool
		sawEmpty bool
	)
	seenNodes := map[ast.Node]struct{}{}

	if mapping := asMapping(node); mapping != nil {
		values, err := m.effectiveMappingValues(mapping, depth)
		if err != nil {
			return matchResult{}, false, err
		}
		for _, value := range values {
			if value == nil || value.Key == nil {
				continue
			}
			if err := m.consumeStep(); err != nil {
				return matchResult{}, false, err
			}
			if mappingKeyName(value.Key) == name {
				sub, subCycle, err := m.matchFrom(value.Value, step+1, depth)
				if err != nil {
					return matchResult{}, false, err
				}
				hitCycle = hitCycle || subCycle
				sawEmpty = sawEmpty || sub.empty
				missing = addMissing(missing, sub.missing)
				matched, err = m.appendUniqueNodes(matched, seenNodes, sub.nodes)
				if err != nil {
					return matchResult{}, false, err
				}
			}
			sub, subCycle, err := m.matchFrom(value.Value, step, depth)
			if err != nil {
				return matchResult{}, false, err
			}
			hitCycle = hitCycle || subCycle
			sawEmpty = sawEmpty || sub.empty
			missing = addMissing(missing, sub.missing)
			matched, err = m.appendUniqueNodes(matched, seenNodes, sub.nodes)
			if err != nil {
				return matchResult{}, false, err
			}
		}

		return matchResult{
			nodes:   matched,
			missing: missing,
			empty:   len(matched) == 0 && missing == 0 && sawEmpty,
		}, hitCycle, nil
	}

	if sequence, ok := node.(*ast.SequenceNode); ok && sequence != nil {
		for _, value := range sequence.Values {
			if value == nil {
				continue
			}
			if err := m.consumeStep(); err != nil {
				return matchResult{}, false, err
			}
			sub, subCycle, err := m.matchFrom(value, step, depth)
			if err != nil {
				return matchResult{}, false, err
			}
			hitCycle = hitCycle || subCycle
			sawEmpty = sawEmpty || sub.empty
			missing = addMissing(missing, sub.missing)
			matched, err = m.appendUniqueNodes(matched, seenNodes, sub.nodes)
			if err != nil {
				return matchResult{}, false, err
			}
		}
	}

	return matchResult{
		nodes:   matched,
		missing: missing,
		empty:   len(matched) == 0 && missing == 0 && sawEmpty,
	}, hitCycle, nil
}

// mergeTargets returns the mappings introduced by a << value. A sequence of
// aliases (<<: [*defaults, *extra]) contributes one mapping per entry, earlier
// entries first so they keep precedence.
func (idx *anchorIndex) mergeTargets(node ast.Node) ([]ast.Node, error) {
	node, err := idx.unwrapContainer(node)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, nil
	}

	sequence, ok := node.(*ast.SequenceNode)
	if !ok {
		return []ast.Node{node}, nil
	}

	targets := make([]ast.Node, 0, len(sequence.Values))
	for _, value := range sequence.Values {
		resolved, err := idx.unwrapContainer(value)
		if err != nil {
			return nil, err
		}
		if resolved != nil {
			targets = append(targets, resolved)
		}
	}

	return targets, nil
}

// unwrapContainer follows aliases and anchors until the node can be indexed or
// searched as a mapping, a sequence, or a scalar.
func (idx *anchorIndex) unwrapContainer(node ast.Node) (ast.Node, error) {
	seenAliases := map[*ast.AliasNode]struct{}{}
	seenAnchors := map[*ast.AnchorNode]struct{}{}
	for node != nil {
		switch typed := node.(type) {
		case *ast.AliasNode:
			if _, ok := seenAliases[typed]; ok {
				return nil, fmt.Errorf("cycle in yaml alias %q", aliasName(typed))
			}
			seenAliases[typed] = struct{}{}

			anchor, ok := idx.aliasToAnchor[typed]
			if !ok || anchor == nil {
				return nil, fmt.Errorf("unresolved yaml alias %q", aliasName(typed))
			}
			node = anchor.Value
		case *ast.AnchorNode:
			if _, ok := seenAnchors[typed]; ok {
				return nil, fmt.Errorf("cycle in yaml anchor %q", anchorName(typed))
			}
			seenAnchors[typed] = struct{}{}
			node = typed.Value
		default:
			return node, nil
		}
	}

	return node, nil
}

func asMapping(node ast.Node) *ast.MappingNode {
	switch typed := node.(type) {
	case *ast.MappingNode:
		return typed
	case *ast.MappingValueNode:
		return &ast.MappingNode{Values: []*ast.MappingValueNode{typed}}
	default:
		return nil
	}
}

func mappingKeyName(key ast.MapKeyNode) string {
	if key == nil {
		return ""
	}
	if value, ok := key.(*ast.StringNode); ok {
		return value.Value
	}

	token := key.GetToken()
	if token == nil {
		return strings.TrimSpace(key.String())
	}

	name := token.Value
	if len(name) > 1 && name[0] == '"' {
		if unquoted, err := strconv.Unquote(name); err == nil {
			return unquoted
		}
	}
	if len(name) > 1 && name[0] == '\'' && name[len(name)-1] == '\'' {
		return name[1 : len(name)-1]
	}

	return name
}

func dedupeASTNodes(nodes []ast.Node) []ast.Node {
	if len(nodes) < 2 {
		return nodes
	}

	seen := map[ast.Node]struct{}{}
	deduped := make([]ast.Node, 0, len(nodes))
	for _, node := range nodes {
		if node == nil {
			continue
		}
		if _, ok := seen[node]; ok {
			continue
		}
		seen[node] = struct{}{}
		deduped = append(deduped, node)
	}

	return deduped
}

func anchorName(node *ast.AnchorNode) string {
	if node == nil || node.Name == nil {
		return ""
	}
	if value, ok := node.Name.(*ast.StringNode); ok {
		return value.Value
	}

	return node.Name.String()
}

func aliasName(node *ast.AliasNode) string {
	if node == nil || node.Value == nil {
		return ""
	}
	if value, ok := node.Value.(*ast.StringNode); ok {
		return value.Value
	}

	return node.Value.String()
}
