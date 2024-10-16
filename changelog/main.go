package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/provider"
)

func main() {
	err := exec()
	if err != nil {
		log.Fatal(err)
	}
}

func exec() error {
	if !util.IsBeta() {
		return fmt.Errorf("please enable beta mode, i.e. set %s=1", util.AivenEnableBeta)
	}

	save := flag.Bool("save", false, "output current schema")
	diff := flag.Bool("diff", false, "compare current schema with imported schema")
	changelog := flag.String("changelog", "", "write changes to file")
	flag.Parse()

	if *save == *diff {
		return fmt.Errorf("either --save or --diff must be set")
	}

	// Loads the current Provider schema
	p, err := provider.Provider("dev")
	if err != nil {
		return err
	}

	newMap, err := fromProvider(p)
	if err != nil {
		return err
	}

	if *save {
		return json.NewEncoder(os.Stdout).Encode(&newMap)
	}

	// Outputs diff with the current Provider schema
	if *diff {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}

		var oldMap ItemMap
		err = json.Unmarshal(b, &oldMap)
		if err != nil {
			return err
		}

		entries, err := diffItemMaps(oldMap, newMap)
		if err != nil {
			return err
		}

		// todo: write to file
		if *changelog == "" {
			for _, l := range entries {
				fmt.Printf("-  %s\n", l)
			}
		}
	}

	return nil
}

func fromProvider(p *schema.Provider) (ItemMap, error) {
	// Item names might clash between resources and data sources
	// Splits into separate maps
	sourceMaps := map[ResourceType]map[string]*schema.Resource{
		ResourceKind:   p.ResourcesMap,
		DataSourceKind: p.DataSourcesMap,
	}

	items := make(ItemMap)
	for kind, m := range sourceMaps {
		items[kind] = make(map[string]*Item)
		for name, r := range m {
			res := &Item{
				Name:        name,
				Path:        name,
				Description: r.Description,
				Type:        schema.TypeList,
			}
			for k, v := range r.Schema {
				walked := walkSchema(k, v, res)
				for i := range walked {
					item := walked[i]
					items[kind][item.Path] = item
				}
			}
		}
	}
	return items, nil
}

type ResourceType string

const (
	ResourceKind   ResourceType = "resource"
	DataSourceKind ResourceType = "datasource"
)

type ItemMap map[ResourceType]map[string]*Item

type Item struct {
	Name string `json:"name"` // TF field name
	Path string `json:"path"` // full path to the item, e.g. `aiven_project.project`

	// TF schema fields
	Description string           `json:"description"`
	ForceNew    bool             `json:"forceNew"`
	Optional    bool             `json:"optional"`
	Sensitive   bool             `json:"sensitive"`
	MaxItems    int              `json:"maxItems"`
	Deprecated  string           `json:"deprecated"`
	Type        schema.ValueType `json:"type"`
	ElemType    schema.ValueType `json:"elemType"`
}

func walkSchema(name string, this *schema.Schema, parent *Item) []*Item {
	item := &Item{
		Name:        name,
		Path:        fmt.Sprintf("%s.%s", parent.Path, name),
		ForceNew:    this.ForceNew,
		Optional:    this.Optional,
		Sensitive:   this.Sensitive,
		MaxItems:    this.MaxItems,
		Description: this.Description,
		Deprecated:  this.Deprecated,
		Type:        this.Type,
	}

	items := []*Item{item}

	// Properties
	switch elem := this.Elem.(type) {
	case *schema.Schema:
		item.ElemType = elem.Type
	case *schema.Resource:
		for k, child := range elem.Schema {
			items = append(items, walkSchema(k, child, item)...)
		}
	}

	return items
}
