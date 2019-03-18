package proxy

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"log"
)

type WidgetSuggestItem struct {
	Slug     string `json:"slug"`
	Subtitle string `json:"subtitle"`
	Title    string `json:"title"`
}

func createItemsFromRawData(rd io.Reader) (items []WidgetSuggestItem, err error) {
	buf, err := ioutil.ReadAll(rd)
	if err != nil {
		return
	}

	var data []map[string]interface{}
	err = json.Unmarshal(buf, &data)
	if err != nil {
		return
	}

	for i, dataItem := range data {
		widgetItem, err := createWidgetSuggestItem(dataItem)
		if err != nil {
			// just log - no error propagation
			log.Println(errors.Wrapf(err, "item %d", i))
			continue
		}

		items = append(items, widgetItem)
	}

	return
}

func createWidgetSuggestItem(data map[string]interface{}) (w WidgetSuggestItem, err error) {
	var t string
	var ok bool
	if t, ok = data["type"].(string); !ok {
		err = errors.New("wrong format (no type)")
		return
	}

	var subtitleField string
	switch t {
	case "city":
		subtitleField = "country_name"
	case "airport":
		subtitleField = "city_name"
	default:
		err = fmt.Errorf("unsupported type %v", t)
		return
	}

	var fieldsToPtrs = map[string]*string{
		subtitleField: &w.Subtitle,
		"code":        &w.Slug,
		"name":        &w.Title,
	}

	for field, ptr := range fieldsToPtrs {
		if v, ok := data[field].(string); ok {
			*ptr = v
		} else {
			err = fmt.Errorf("field %s has wrong format or not exists", field)
			return
		}
	}

	return
}
