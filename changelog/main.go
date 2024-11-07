package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/provider"
)

type flags struct {
	save          bool
	diff          bool
	schemaFile    string
	changelogFile string
	reformat      bool
}

func main() {
	if err := exec(); err != nil {
		log.Fatal(err)
	}
}

func exec() error {
	if err := checkBetaMode(); err != nil {
		return err
	}

	flags, err := parseFlags()
	if err != nil {
		return err
	}

	p, err := loadProvider()
	if err != nil {
		return fmt.Errorf("failed to load provider: %w", err)
	}

	newMap, err := fromProvider(p)
	if err != nil {
		return fmt.Errorf("failed to process provider schema: %w", err)
	}

	if flags.save {
		return saveSchema(flags.schemaFile, newMap)
	}

	return processDiff(flags, newMap)
}

func checkBetaMode() error {
	if !util.IsBeta() {
		return fmt.Errorf("please enable beta mode, i.e. set %s=1", util.AivenEnableBeta)
	}
	return nil
}

func parseFlags() (*flags, error) {
	f := &flags{}
	flag.BoolVar(&f.save, "save", false, "output current schema")
	flag.BoolVar(&f.diff, "diff", false, "compare current schema with imported schema")
	flag.StringVar(&f.schemaFile, "schema", "", "schema file path (for save/diff)")
	flag.StringVar(&f.changelogFile, "changelog", "", "changelog output file path")
	flag.BoolVar(&f.reformat, "reformat", false, "reformat the whole changelog file")
	flag.Parse()

	if f.save == f.diff {
		return nil, fmt.Errorf("either --save or --diff must be set")
	}

	if f.diff && f.schemaFile == "" {
		return nil, fmt.Errorf("schema file path is required when using --diff")
	}

	return f, nil
}

func loadProvider() (*schema.Provider, error) {
	p, err := provider.Provider("dev")
	if err != nil {
		return nil, err
	}
	return p, nil
}

func saveSchema(filePath string, schema ItemMap) error {
	if filePath == "" {
		return json.NewEncoder(os.Stdout).Encode(&schema)
	}

	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create schema file: %w", err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ") // Pretty print the JSON
	return encoder.Encode(&schema)
}

func processDiff(flags *flags, newMap ItemMap) error {
	oldMap, err := loadSchemaFile(flags.schemaFile)
	if err != nil {
		return err
	}

	entries, err := diffItemMaps(oldMap, newMap)
	if err != nil {
		return fmt.Errorf("failed to generate diff: %w", err)
	}

	if flags.changelogFile == "" {
		return printChangelog(entries)
	}

	return writeChangelog(flags.changelogFile, flags.reformat, entries)
}

func loadSchemaFile(path string) (ItemMap, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open schema file: %w", err)
	}
	defer f.Close()

	var oldMap ItemMap
	if err := json.NewDecoder(f).Decode(&oldMap); err != nil {
		return nil, fmt.Errorf("failed to parse schema file: %w", err)
	}

	return oldMap, nil
}

func printChangelog(entries []string) error {
	for _, l := range entries {
		fmt.Printf("- %s\n", l)
	}
	return nil
}

func writeChangelog(filePath string, reformat bool, entries []string) error {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read changelog file: %w", err)
	}

	s, err := updateChangelog(string(b), defaultLineMaxLength, reformat, entries...)
	if err != nil {
		return fmt.Errorf("failed to format changelog: %w", err)
	}

	return os.WriteFile(filePath, []byte(s), 0644)
}

func fromProvider(p *schema.Provider) (ItemMap, error) {
	// Item names might clash between resources and data sources
	// Splits into separate maps
	sourceMaps := map[RootType]map[string]*schema.Resource{
		ResourceRootType:   p.ResourcesMap,
		DataSourceRootType: p.DataSourcesMap,
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
		item.ElementType = elem.Type
	case *schema.Resource:
		for k, child := range elem.Schema {
			items = append(items, walkSchema(k, child, item)...)
		}
	}

	return items
}
