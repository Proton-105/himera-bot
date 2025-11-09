package keyboard

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Proton-105/himera-bot/internal/i18n"
)

// PaginationButtons returns up to three inline buttons (prev, current page, next)
// allowing the caller to paginate lists using a shared action prefix.
func PaginationButtons(t i18n.Translator, action string, page, totalPages int) []InlineButton {
	if totalPages < 1 {
		totalPages = 1
	}
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	buttons := make([]InlineButton, 0, 3)

	if page > 1 {
		buttons = append(buttons, InlineButton{
			Text:   translated(t, "pagination.pagination_prev", "◀️ Prev"),
			Unique: action,
			Data:   strconv.Itoa(page - 1),
		})
	}

	buttons = append(buttons, InlineButton{
		Text:   paginationLabel(t, page, totalPages),
		Unique: action,
		Data:   strconv.Itoa(page),
	})

	if page < totalPages {
		buttons = append(buttons, InlineButton{
			Text:   translated(t, "pagination.pagination_next", "Next ▶️"),
			Unique: action,
			Data:   strconv.Itoa(page + 1),
		})
	}

	return buttons
}

func translated(t i18n.Translator, key, fallback string) string {
	if t == nil {
		return fallback
	}

	text := strings.TrimSpace(t.T(key))
	if text == "" || text == key {
		return fallback
	}

	return text
}

func paginationLabel(t i18n.Translator, page, total int) string {
	label := translated(t, "pagination.pagination_page", "")
	if label == "" {
		label = "Page {{.Page}}/{{.Total}}"
	}

	label = strings.ReplaceAll(label, "{{.Page}}", strconv.Itoa(page))
	label = strings.ReplaceAll(label, "{{.Total}}", strconv.Itoa(total))

	if strings.Contains(label, "{{") {
		return fmt.Sprintf("Page %d/%d", page, total)
	}

	return label
}
