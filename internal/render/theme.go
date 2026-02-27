package render

import "github.com/charmbracelet/lipgloss"

// LensBadge returns a single coloured header line identifying the active lens.
// It sits above the Glamour-rendered content, not wrapped around it.
func LensBadge(lens string) string {
	style := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1)

	switch lens {
	case "why":
		return style.Foreground(lipgloss.Color("10")).Render("◆ WHY") + "\n"
	case "importance":
		return style.Foreground(lipgloss.Color("11")).Render("◆ IMPORTANCE") + "\n"
	case "cli-tools":
		return style.Foreground(lipgloss.Color("14")).Render("◆ CLI TOOLS") + "\n"
	case "arch":
		return style.Foreground(lipgloss.Color("12")).Render("◆ ARCH") + "\n"
	case "used":
		return style.Foreground(lipgloss.Color("13")).Render("◆ USED") + "\n"
	case "gotchas":
		return style.Foreground(lipgloss.Color("9")).Render("◆ GOTCHAS") + "\n"
	case "refs":
		return style.Foreground(lipgloss.Color("8")).Render("◆ REFS") + "\n"
	default:
		return ""
	}
}
