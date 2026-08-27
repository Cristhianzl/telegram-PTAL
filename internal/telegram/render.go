package telegram

import (
	"fmt"
	"html"
	"sort"
	"strings"
	"time"

	"github.com/Cristhianzl/telegram-PTAL/internal/engine"
	"github.com/Cristhianzl/telegram-PTAL/internal/githubapi"
)

// RenderEvents builds the text for a batch of events. Events of the same kind
// are grouped, so five approvals become one block instead of five notices.
func RenderEvents(events []engine.Event) string {
	if len(events) == 0 {
		return ""
	}
	if len(events) == 1 {
		return renderSingle(events[0])
	}

	groups := map[engine.Kind][]engine.Event{}
	var order []engine.Kind
	for _, e := range events {
		if _, seen := groups[e.Kind]; !seen {
			order = append(order, e.Kind)
		}
		groups[e.Kind] = append(groups[e.Kind], e)
	}

	var b strings.Builder
	for i, kind := range order {
		if i > 0 {
			b.WriteString("\n")
		}
		group := groups[kind]
		head := group[0]
		fmt.Fprintf(&b, "%s <b>%s</b>", head.Emoji(), esc(head.Title()))
		if len(group) > 1 {
			fmt.Fprintf(&b, " · %d", len(group))
		}
		b.WriteString("\n")
		for _, e := range group {
			fmt.Fprintf(&b, "  %s\n", prLine(e.PR))
			if e.Detail != "" {
				fmt.Fprintf(&b, "  <i>%s</i>\n", esc(e.Detail))
			}
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

// renderSingle gives the alert more room when it arrives alone, which is the
// case for urgent events.
func renderSingle(e engine.Event) string {
	pr := e.PR
	var b strings.Builder

	fmt.Fprintf(&b, "%s <b>%s</b>\n\n", e.Emoji(), esc(e.Title()))
	fmt.Fprintf(&b, "<a href=%q>%s</a>\n", pr.URL, esc(pr.Slug()))
	fmt.Fprintf(&b, "<b>%s</b>\n", esc(pr.Title))

	var meta []string
	if pr.Author != "" {
		meta = append(meta, "@"+esc(pr.Author))
	}
	if pr.Additions > 0 || pr.Deletions > 0 {
		meta = append(meta, fmt.Sprintf("+%d −%d", pr.Additions, pr.Deletions))
	}
	if s := pr.ChecksSymbol(); s != "" {
		meta = append(meta, s)
	}
	if pr.Mergeable == "CONFLICTING" {
		meta = append(meta, "⚠️ conflict")
	}
	if pr.IsDraft {
		meta = append(meta, "draft")
	}
	if len(meta) > 0 {
		fmt.Fprintf(&b, "%s\n", strings.Join(meta, " · "))
	}
	if e.Detail != "" {
		fmt.Fprintf(&b, "\n<i>%s</i>\n", esc(e.Detail))
	}
	if !pr.CreatedAt.IsZero() {
		fmt.Fprintf(&b, "\n<i>opened %s</i>", esc(humanAge(pr.CreatedAt)))
	}
	return strings.TrimRight(b.String(), "\n")
}

// prLine is the compact form of a pull request inside a list.
func prLine(pr *githubapi.PullRequest) string {
	title := truncateTitle(pr.Title, 58)
	line := fmt.Sprintf("<a href=%q>%s</a> %s", pr.URL, esc(pr.Slug()), esc(title))
	if s := pr.ChecksSymbol(); s != "" && (s == "❌ CI") {
		line += " " + s
	}
	return line
}

// RenderPanel builds the status panel that stays pinned in the conversation
// and is rewritten each cycle instead of becoming a new message.
func RenderPanel(snap *githubapi.Snapshot) string {
	var b strings.Builder
	b.WriteString("📋 <b>Your pull requests</b>\n")

	counts := snap.Counts()
	total := 0
	for _, n := range counts {
		total += n
	}
	if total == 0 {
		b.WriteString("\n<i>Nothing open in your name right now.</i>\n")
	}

	for _, bucket := range githubapi.AllBuckets {
		prs := snap.InBucket(bucket)
		if len(prs) == 0 {
			continue
		}
		fmt.Fprintf(&b, "\n<b>%s</b> · %d\n", esc(bucket.Label()), len(prs))
		limit := len(prs)
		if limit > 8 {
			limit = 8
		}
		for _, pr := range prs[:limit] {
			fmt.Fprintf(&b, "  %s\n", prLine(pr))
		}
		if len(prs) > limit {
			fmt.Fprintf(&b, "  <i>+%d…</i>\n", len(prs)-limit)
		}
	}

	fmt.Fprintf(&b, "\n<i>updated %s</i>", esc(snap.FetchedAt.Local().Format("15:04")))
	return b.String()
}

// RenderRepoList renders every open pull request in one repository.
//
// Unlike the panel, this is not about you: it shows authors, because the
// point of asking about a whole project is seeing who is waiting on what.
func RenderRepoList(repo, author string, prs []*githubapi.PullRequest, limit int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "📂 <b>%s</b>\n", esc(repo))

	if len(prs) == 0 {
		if author != "" {
			fmt.Fprintf(&b, "\n<i>No open pull requests by %s.</i>", esc(author))
		} else {
			b.WriteString("\n<i>No open pull requests.</i>")
		}
		return b.String()
	}
	if author != "" {
		fmt.Fprintf(&b, "<i>%d open by %s</i>\n\n", len(prs), esc(author))
	} else {
		fmt.Fprintf(&b, "<i>%d open</i>\n\n", len(prs))
	}

	shown := prs
	if limit > 0 && len(shown) > limit {
		shown = shown[:limit]
	}

	for _, pr := range shown {
		fmt.Fprintf(&b, "<a href=%q>#%d</a> %s\n", pr.URL, pr.Number, esc(truncateTitle(pr.Title, 54)))

		var meta []string
		if pr.Author != "" {
			meta = append(meta, "@"+esc(pr.Author))
		}
		if pr.IsDraft {
			meta = append(meta, "draft")
		}
		if s := pr.ChecksSymbol(); s != "" {
			meta = append(meta, s)
		}
		if pr.ReviewDecision == "APPROVED" {
			meta = append(meta, "✅ approved")
		}
		if !pr.UpdatedAt.IsZero() {
			meta = append(meta, humanAge(pr.UpdatedAt))
		}
		if len(meta) > 0 {
			fmt.Fprintf(&b, "  <i>%s</i>\n", strings.Join(meta, " · "))
		}
	}

	if len(prs) > len(shown) {
		fmt.Fprintf(&b, "\n<i>+%d more</i>", len(prs)-len(shown))
	}
	return b.String()
}

func truncateTitle(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return strings.TrimSpace(s[:n-1]) + "…"
}

// RenderSummary is the short text of the /status command.
func RenderSummary(snap *githubapi.Snapshot, lastSuccess time.Time) string {
	counts := snap.Counts()
	var parts []string
	for _, bucket := range githubapi.AllBuckets {
		if n := counts[bucket]; n > 0 {
			parts = append(parts, fmt.Sprintf("%s: %d", esc(bucket.Label()), n))
		}
	}
	sort.SliceStable(parts, func(i, j int) bool { return false })

	var b strings.Builder
	b.WriteString("🤖 <b>PTAL</b>\n\n")
	if snap.Viewer != "" {
		fmt.Fprintf(&b, "GitHub: <b>@%s</b>\n", esc(snap.Viewer))
	}
	if len(parts) == 0 {
		b.WriteString("No open pull requests in your name.\n")
	} else {
		b.WriteString(strings.Join(parts, "\n") + "\n")
	}
	if !lastSuccess.IsZero() {
		fmt.Fprintf(&b, "\n<i>last sync %s</i>", esc(humanAge(lastSuccess)))
	}
	if snap.RateRemaining > 0 {
		fmt.Fprintf(&b, "\n<i>rate limit: %d/5000</i>", snap.RateRemaining)
	}
	return b.String()
}

// humanAge writes how long ago something happened.
func humanAge(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%d min ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 48*time.Hour:
		return "yesterday"
	default:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	}
}

// EscapeHTML is esc, exported for callers outside this package.
func EscapeHTML(s string) string { return esc(s) }

// HumanAge is humanAge, exported for callers outside this package.
func HumanAge(t time.Time) string { return humanAge(t) }

// esc protects text coming from GitHub: pull request titles frequently
// contain <, > and &, and Telegram rejects the whole message if the HTML
// breaks.
func esc(s string) string { return html.EscapeString(s) }
