// Load and save content.json
package content

import (
	"encoding/json"
	"io"
	"os"
)

type Content struct {
	Meta        Meta         `json:"meta"`
	TopicGroups []TopicGroup `json:"content"`
}

type Meta struct {
	FirstCreatedAt                      string `json:"first_created_at"`
	CantabularMetadataSource            string `json:"cantabular_metadata_source"`
	FilteredToAtlasContentAt            string `json:"filtered_to_atlas_content_at"`
	RichContentSpecFileUsedToFilter     string `json:"rich_content_spec_file_used_to_filter"`
	LegendStrsUpdatedAt                 string `json:"legend_strs_updated_at"`
	LegendStrsFileUsed                  string `json:"legend_strs_file_used"`
	PlaceholderVariablesDescsInsertedAt string `json:"placeholder_variable_descs_inserted_at"`
}

type TopicGroup struct {
	Name   string  `json:"name"`
	Slug   string  `json:"slug"`
	Desc   string  `json:"desc"`
	Topics []Topic `json:"topics"`
}

type Topic struct {
	Name      string     `json:"name"`
	Slug      string     `json:"slug"`
	Desc      string     `json:"desc"`
	Variables []Variable `json:"variables"`
}

type Variable struct {
	Name            string           `json:"name"`
	Code            string           `json:"code"`
	Slug            string           `json:"slug"`
	Desc            string           `json:"desc"`
	Units           string           `json:"units"`
	Classifications []Classification `json:"classifications"`
}

type Classification struct {
	Code                        string     `json:"code"`
	Slug                        string     `json:"slug"`
	Desc                        string     `json:"desc"`
	ChoroplethDefault           bool       `json:"choropleth_default,omitempty"`
	DotDensityDefault           bool       `json:"dot_density_default,omitempty"`
	Comparison2011DataAvailable bool       `json:"comparison_2011_data_available"`
	Categories                  []Category `json:"categories"`
}

type Category struct {
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	Code       string `json:"code"`
	LegendStr1 string `json:"legend_str_1"`
	LegendStr2 string `json:"legend_str_2"`
	LegendStr3 string `json:"legend_str_3"`
}

func Load(r io.Reader) (*Content, error) {
	dec := json.NewDecoder(r)
	var c Content
	if err := dec.Decode(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

func LoadName(name string) (*Content, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return Load(f)
}

func (c *Content) Save(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	return enc.Encode(c)
}
