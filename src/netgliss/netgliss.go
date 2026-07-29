package netgliss

import(
	"path/filepath"
	"fmt"
	"encoding/json"
	"os"
	"path"
	"strconv"
	"strings"
)

type List struct {
	Servers []Server
}

// Returns a new *List
func New() *List {
	return &List {
		Servers: make([]Server, 0),
	}
}

// Loads servers from a file using a Glob pattern
func (l *List) LoadServers(globPattern string) (error) {
paths, err := filepath.Glob(globPattern)
	if err != nil {
		return fmt.Errorf("invalid glob pattern %q: %s", globPattern, err.Error())
	}
	if len(paths) == 0 {
		return fmt.Errorf("no files match '%q' ", globPattern)
	}

	var servers []Server
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error while opening file: %s: %s", path, err.Error())
		}
		err = json.Unmarshal(data, &servers)
		if err != nil {
			return fmt.Errorf("error while reading JSON: %s, %s", path, err.Error())
		}
	}
	l.Servers = servers
	return nil
}


func (l *List) SearchServers(globPattern, field string) ([]Server, error) {
	// Pattern validity depends only on the pattern, so check it once up front.
	if _, err := path.Match(globPattern, ""); err != nil {
		return nil, fmt.Errorf("invalid pattern %q: %w", globPattern, err)
	}

	var results []Server

	for _, server := range l.Servers {
		var value string

		switch field {
		case "name":
			value = server.Name
		case "version":
			value = server.Version
		case "description":
			value = server.Description
		case "ip":
			value = server.IP
		case "tags":
			value = strings.Join(server.Tags, ",") // convert []string to string
		case "languages":
			value = server.Languages
		case "requiresMods":
			value = strconv.FormatBool(server.RequiresMods)
		default:
			return nil, fmt.Errorf("unknown field %q", field)
		}

		if ok, _ := path.Match(globPattern, value); ok {
			results = append(results, server)
		}
	}

	return results, nil
}

