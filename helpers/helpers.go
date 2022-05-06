package helpers

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"gitlab.com/toby3d/telegraph"
)

func WriteFileJSON(filename string, obj interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Error opening file")
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(obj); err != nil {
		return fmt.Errorf("Error writing to file")
	}

	return nil
}

func JoinStringMaps(maps ...map[string]string) map[string]string {
	ret := make(map[string]string)

	for _, m := range maps {
		for key, value := range m {
			ret[key] = value
		}
	}

	return ret
}

func SortedValuesByKey(m map[string]string) []string {
	var (
		keys   = make([]string, 0, len(m))
		values = make([]string, 0, len(m))
	)

	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		values = append(values, m[key])
	}

	return values
}

func CreateDomFromImages(images []string) []telegraph.Node {
	result := make([]telegraph.Node, 0, len(images))
	for _, img := range images {
		result = append(result, telegraph.NodeElement{
			Tag:      "img",
			Attrs:    map[string]string{"src": img},
			Children: nil,
		})
	}

	return result
}
