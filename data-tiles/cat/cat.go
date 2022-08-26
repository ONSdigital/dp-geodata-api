package cat

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/ONSdigital/dp-geodata-api/data-tiles/types"
)

// LoadCategories loads the categories.txt file into a []types.Category
func LoadCategories(fname string) ([]types.Category, error) {
	log.Printf("Loading %s", fname)
	buf, err := os.ReadFile(fname)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(buf), "\n")

	var cats []types.Category
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		cats = append(cats, types.Category(line))
	}
	return cats, nil
}

// LoadMetrics loads the categories named in cats, plus corresponding totals categories
func LoadMetrics(cats []types.Category, dir string) (map[types.Category]map[types.Geocode]types.Value, error) {
	metrics := map[types.Category]map[types.Geocode]types.Value{}

	allcats, err := addTotals(cats)
	if err != nil {
		return nil, err
	}

	for _, cat := range allcats {
		log.Printf("Loading %s", cat)
		csv, err := loadCatfile(filepath.Join(dir, string(cat)+".CSV"))
		if err != nil {
			return nil, err
		}

		values := map[types.Geocode]types.Value{}
		for i, row := range csv {
			if i == 0 {
				continue // skip header line
			}
			if len(row) != 2 {
				return nil, fmt.Errorf("%s: %v: not enough fields", cat, row)
			}
			v, err := strconv.ParseFloat(row[1], 64)
			if err != nil {
				return nil, fmt.Errorf("%s: %v: %w", cat, row, err)
			}
			values[types.Geocode(row[0])] = types.Value(v)
		}
		metrics[cat] = values

	}
	return metrics, nil
}

// addTotals adds totals categories to list of categories we want
func addTotals(cats []types.Category) ([]types.Category, error) {
	totcats := map[types.Category]bool{}

	for _, cat := range cats {
		totcat, err := GuessTotalsCat(cat)
		if err != nil {
			return nil, err
		}
		totcats[totcat] = true
	}

	newcats := make([]types.Category, len(cats))
	copy(newcats, cats)
	for cat := range totcats {
		newcats = append(newcats, cat)
	}
	return newcats, nil
}

// loadCatfile loads a single category CSV
func loadCatfile(fname string) ([][]string, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return csv.NewReader(f).ReadAll()
}

var catRegex = regexp.MustCompile(`^([A-Z]+)([0-9]+)([A-Z]+)([0-9]+)$`)

// GuessTotalsCat figures out the category holding totals.
// So far this means just changing the numeric part to 1.
// "QS402EW0012" --> "QS402EW0001"
func GuessTotalsCat(cat types.Category) (types.Category, error) {
	matches := catRegex.FindStringSubmatch(string(cat))
	if len(matches) != 5 {
		return "", errors.New("can't parse category code")
	}

	n, err := strconv.Atoi(matches[4])
	if err != nil {
		return "", err
	}

	if n == 1 {
		return "", errors.New("category is already the totals category")
	}

	digits := len(matches[4])
	s := fmt.Sprintf("%s%s%s%0*d", matches[1], matches[2], matches[3], digits, 1)
	return types.Category(s), nil
}
