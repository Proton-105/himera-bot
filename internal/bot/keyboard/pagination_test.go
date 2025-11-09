package keyboard_test

import (
	"testing"

	"github.com/Proton-105/himera-bot/internal/bot/keyboard"
	"github.com/Proton-105/himera-bot/internal/testutil"
)

type mockTranslator struct {
	translations map[string]string
	lang         string
}

func (m *mockTranslator) T(key string) string {
	if val, ok := m.translations[key]; ok {
		return val
	}
	return key
}

func (m *mockTranslator) Lang() string {
	if m.lang == "" {
		return "en"
	}
	return m.lang
}

func TestPaginationButtons(t *testing.T) {
	translator := &mockTranslator{
		translations: map[string]string{
			"pagination.pagination_prev": "◀️ Prev",
			"pagination.pagination_next": "Next ▶️",
			"pagination.pagination_page": "Page {{.Page}}/{{.Total}}",
		},
	}

	testCases := []struct {
		name      string
		page      int
		total     int
		wantTexts []string
		wantData  []string
	}{
		{
			name:      "first page",
			page:      1,
			total:     5,
			wantTexts: []string{"Page 1/5", "Next ▶️"},
			wantData:  []string{"1", "2"},
		},
		{
			name:      "middle page",
			page:      3,
			total:     5,
			wantTexts: []string{"◀️ Prev", "Page 3/5", "Next ▶️"},
			wantData:  []string{"2", "3", "4"},
		},
		{
			name:      "last page",
			page:      5,
			total:     5,
			wantTexts: []string{"◀️ Prev", "Page 5/5"},
			wantData:  []string{"4", "5"},
		},
		{
			name:      "single page",
			page:      1,
			total:     1,
			wantTexts: []string{"Page 1/1"},
			wantData:  []string{"1"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buttons := keyboard.PaginationButtons(translator, "history", tc.page, tc.total)
			testutil.AssertEqual(t, len(tc.wantTexts), len(buttons))

			for i := range tc.wantTexts {
				testutil.AssertEqual(t, tc.wantTexts[i], buttons[i].Text)
				testutil.AssertEqual(t, "history", buttons[i].Unique)
				testutil.AssertEqual(t, tc.wantData[i], buttons[i].Data)
			}
		})
	}
}
