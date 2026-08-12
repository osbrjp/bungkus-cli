package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
)

// orderedObj preserves top-level key order and the exact raw bytes of every
// value, so re-emitting an untouched package.json section changes nothing the
// user wrote. Interior bytes of untouched values are preserved verbatim; only
// each emitted level's key/indent framing is normalized to two-space style.
type orderedObj struct {
	keys []string
	vals map[string]json.RawMessage
}

func parseOrderedObj(data []byte) (*orderedObj, error) {
	o := &orderedObj{vals: map[string]json.RawMessage{}}
	dec := json.NewDecoder(bytes.NewReader(data))
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return nil, fmt.Errorf("expected a JSON object")
	}
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, ok := keyTok.(string)
		if !ok {
			return nil, fmt.Errorf("expected an object key")
		}
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return nil, err
		}
		// Duplicate keys: last value wins, the key is emitted once.
		if _, dup := o.vals[key]; !dup {
			o.keys = append(o.keys, key)
		}
		o.vals[key] = raw
	}
	if _, err := dec.Token(); err != nil { // consume the closing brace
		return nil, err
	}
	return o, nil
}

// set replaces an existing key's value or appends the key at the end.
func (o *orderedObj) set(key string, v json.RawMessage) {
	if _, ok := o.vals[key]; !ok {
		o.keys = append(o.keys, key)
	}
	o.vals[key] = v
}

// marshal emits the object with two-space indentation at this level; value
// bytes are written verbatim, preserving their interior formatting.
func (o *orderedObj) marshal(indent string) []byte {
	if len(o.keys) == 0 {
		return []byte("{}")
	}
	var buf bytes.Buffer
	buf.WriteString("{\n")
	for i, k := range o.keys {
		kb, _ := json.Marshal(k)
		buf.WriteString(indent + "  ")
		buf.Write(kb)
		buf.WriteString(": ")
		buf.Write(bytes.TrimSpace(o.vals[k]))
		if i < len(o.keys)-1 {
			buf.WriteByte(',')
		}
		buf.WriteByte('\n')
	}
	buf.WriteString(indent + "}")
	return buf.Bytes()
}

// depSections are every package.json section that can already carry a
// dependency; a package present in ANY of them is never re-added.
var depSections = []string{"dependencies", "devDependencies", "peerDependencies", "optionalDependencies"}

// MergeAddPackages additively merges add into raw package.json bytes. It never
// changes an existing script or dependency version; a dependency already
// present in any dependency section is skipped entirely (never duplicated or
// moved between sections). New keys are appended in sorted order after the
// existing ones; missing sections are created at the end of the file only when
// they gain a key. When nothing is added the original bytes are returned
// unchanged and rep.PkgJSONChanged stays false.
func MergeAddPackages(orig []byte, add Packages, rep *AddReport) ([]byte, error) {
	top, err := parseOrderedObj(orig)
	if err != nil {
		return nil, fmt.Errorf("could not parse package.json: %w", err)
	}

	sections := map[string]*orderedObj{}
	getSection := func(name string) (*orderedObj, error) {
		if o, ok := sections[name]; ok {
			return o, nil
		}
		o := &orderedObj{vals: map[string]json.RawMessage{}}
		if raw, ok := top.vals[name]; ok {
			var err error
			if o, err = parseOrderedObj(raw); err != nil {
				return nil, fmt.Errorf("package.json %q is not an object: %w", name, err)
			}
		}
		sections[name] = o
		return o, nil
	}

	present := map[string]bool{}
	for _, name := range depSections {
		o, err := getSection(name)
		if err != nil {
			return nil, err
		}
		for _, k := range o.keys {
			present[k] = true
		}
	}

	changed := map[string]bool{}

	scripts, err := getSection("scripts")
	if err != nil {
		return nil, err
	}
	for _, k := range slices.Sorted(maps.Keys(add.Scripts)) {
		if _, ok := scripts.vals[k]; ok {
			rep.ScriptsSkipped = append(rep.ScriptsSkipped, k)
			continue
		}
		v, _ := json.Marshal(add.Scripts[k])
		scripts.set(k, v)
		rep.ScriptsAdded = append(rep.ScriptsAdded, k+" = "+add.Scripts[k])
		changed["scripts"] = true
	}

	addDeps := func(section string, entries map[string]string) error {
		o, err := getSection(section)
		if err != nil {
			return err
		}
		for _, k := range slices.Sorted(maps.Keys(entries)) {
			if present[k] {
				rep.DepsSkipped = append(rep.DepsSkipped, k)
				continue
			}
			v, _ := json.Marshal(entries[k])
			o.set(k, v)
			present[k] = true
			rep.DepsAdded = append(rep.DepsAdded, k+" "+entries[k])
			changed[section] = true
		}
		return nil
	}
	if err := addDeps("dependencies", add.Dependencies); err != nil {
		return nil, err
	}
	if err := addDeps("devDependencies", add.DevDependencies); err != nil {
		return nil, err
	}

	if len(changed) == 0 {
		return orig, nil
	}
	rep.PkgJSONChanged = true
	for name := range changed {
		top.set(name, sections[name].marshal("  "))
	}
	return append(top.marshal(""), '\n'), nil
}
