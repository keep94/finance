package categories

import (
	"testing"

	"github.com/keep94/finance/fin"
)

type catPopularities map[fin.Cat]int

func (c catPopularities) Popularity(cat fin.Cat) int {
	return c[cat]
}

func TestMostPopularFirst(t *testing.T) {
	popularities := catPopularities{
		toCat("0:1"): 5,
		toCat("0:2"): 4,
		toCat("0:3"): 3,
		toCat("0:4"): 2,
		toCat("0:5"): 1,
	}
	catDetails := []CatDetail{
		createCatDetail(toCat("0:3"), "", true),
		createCatDetail(toCat("0:6"), "", true),
		createCatDetail(toCat("0:1"), "", true),
		createCatDetail(toCat("0:7"), "", true),
		createCatDetail(toCat("0:4"), "", true),
		createCatDetail(toCat("0:2"), "", true),
		createCatDetail(toCat("0:5"), "", true),
	}
	result := MostPopularFirst(catDetails, popularities, 0.8, -1)
	verifyCatDetailList(
		t,
		result, "0:1", "0:2", "0:3", "0:4", "0:5",
		"0:3", "0:6", "0:1", "0:7", "0:4", "0:2", "0:5")
	result = MostPopularFirst(catDetails, nil, 0.8, -1)
	verifyCatDetailList(
		t,
		result,
		"0:3", "0:6", "0:1", "0:7", "0:4", "0:2", "0:5")
	result = MostPopularFirst(catDetails, popularities, 1.0, -1)
	verifyCatDetailList(
		t,
		result, "0:1", "0:2", "0:3", "0:4", "0:5",
		"0:3", "0:6", "0:1", "0:7", "0:4", "0:2", "0:5")
	result = MostPopularFirst(catDetails, popularities, 0.8, 3)
	verifyCatDetailList(
		t,
		result, "0:1", "0:2", "0:3",
		"0:3", "0:6", "0:1", "0:7", "0:4", "0:2", "0:5")
	result = MostPopularFirst(catDetails, popularities, 0.8, 2)
	verifyCatDetailList(
		t,
		result, "0:1", "0:2",
		"0:3", "0:6", "0:1", "0:7", "0:4", "0:2", "0:5")
	result = MostPopularFirst(catDetails, popularities, 0.8, 1)
	verifyCatDetailList(
		t,
		result, "0:1",
		"0:3", "0:6", "0:1", "0:7", "0:4", "0:2", "0:5")
	result = MostPopularFirst(catDetails, popularities, 0.8, 0)
	verifyCatDetailList(
		t,
		result,
		"0:3", "0:6", "0:1", "0:7", "0:4", "0:2", "0:5")
	result = MostPopularFirst(catDetails, popularities, 0.0, -1)
	verifyCatDetailList(
		t,
		result,
		"0:3", "0:6", "0:1", "0:7", "0:4", "0:2", "0:5")
}

func verifyCatDetailList(t *testing.T, result []CatDetail, catIds ...string) {
	t.Helper()
	if len(catIds) != len(result) {
		t.Errorf("Expected %d CatDetails, got %d", len(catIds), len(result))
		return
	}
	for i := range catIds {
		if catIds[i] != result[i].Id().String() {
			t.Errorf(
				"Expected %s at posit %d, got %s",
				catIds[i], i, result[i].Id().String())
			return
		}
	}
}
