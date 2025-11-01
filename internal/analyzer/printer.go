package analyzer

import (
	"fmt"
	"os"
	"text/tabwriter"
)

// PrintSummary prints a formatted summary to stdout
func PrintSummary(summary *Summary) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintf(os.Stdout, "\n")
	fmt.Fprintf(os.Stdout, "═══════════════════════════════════════════════════════════\n")
	fmt.Fprintf(os.Stdout, "                    SEO Analysis Summary                    \n")
	fmt.Fprintf(os.Stdout, "═══════════════════════════════════════════════════════════\n")
	fmt.Fprintf(os.Stdout, "\n")

	// Overall stats
	fmt.Fprintf(w, "Total Pages Crawled:\t%d\n", summary.TotalPages)
	fmt.Fprintf(w, "Total Issues Found:\t%d\n", summary.TotalIssues)
	fmt.Fprintf(w, "Average Response Time:\t%d ms\n", summary.AverageResponseTime)
	fmt.Fprintf(w, "Pages with Errors:\t%d\n", summary.PagesWithErrors)
	fmt.Fprintf(w, "Pages with Redirects:\t%d\n", summary.PagesWithRedirects)
	fmt.Fprintf(w, "Total Internal Links:\t%d\n", summary.TotalInternalLinks)
	fmt.Fprintf(w, "Total External Links:\t%d\n", summary.TotalExternalLinks)
	fmt.Fprintf(w, "\n")

	// Issues by severity
	severityCounts := summary.GetIssueCountBySeverity()
	if len(severityCounts) > 0 {
		fmt.Fprintf(os.Stdout, "Issues by Severity:\n")
		fmt.Fprintf(w, "  Errors:\t%d\n", severityCounts["error"])
		fmt.Fprintf(w, "  Warnings:\t%d\n", severityCounts["warning"])
		fmt.Fprintf(w, "  Info:\t%d\n", severityCounts["info"])
		fmt.Fprintf(w, "\n")
	}

	// Top issues by type
	if len(summary.IssuesByType) > 0 {
		fmt.Fprintf(os.Stdout, "Issues by Type:\n")
		topIssues := summary.GetTopIssues(10)
		for _, issueType := range topIssues {
			count := summary.IssuesByType[issueType]
			icon := getIssueIcon(issueType)
			fmt.Fprintf(w, "  %s %s:\t%d\n", icon, formatIssueType(issueType), count)
		}
		fmt.Fprintf(w, "\n")
	}

	// Slowest pages
	if len(summary.SlowestPages) > 0 {
		fmt.Fprintf(os.Stdout, "Slowest Pages (>2s):\n")
		for i, page := range summary.SlowestPages {
			if i >= 5 {
				break
			}
			fmt.Fprintf(w, "  %s\t%d ms\n", page.URL, page.ResponseTime)
		}
		fmt.Fprintf(w, "\n")
	}

	// Top issues detail
	if len(summary.Issues) > 0 {
		fmt.Fprintf(os.Stdout, "Top Issues:\n")
		// Group by type and show first few examples
		issueGroups := make(map[IssueType][]Issue)
		for _, issue := range summary.Issues {
			issueGroups[issue.Type] = append(issueGroups[issue.Type], issue)
		}

		topTypes := summary.GetTopIssues(5)
		for _, issueType := range topTypes {
			issues := issueGroups[issueType]
			if len(issues) == 0 {
				continue
			}
			icon := getIssueIcon(issueType)
			fmt.Fprintf(os.Stdout, "\n  %s %s:\n", icon, formatIssueType(issueType))
			
			// Show first 3 examples
			for i := 0; i < 3 && i < len(issues); i++ {
				issue := issues[i]
				fmt.Fprintf(w, "    • %s\n", issue.URL)
				fmt.Fprintf(w, "      %s\n", issue.Message)
				if issue.Recommendation != "" {
					fmt.Fprintf(w, "      Recommendation: %s\n", issue.Recommendation)
				}
			}
			if len(issues) > 3 {
				fmt.Fprintf(w, "    ... and %d more\n", len(issues)-3)
			}
		}
	}

	fmt.Fprintf(os.Stdout, "\n")
	fmt.Fprintf(os.Stdout, "═══════════════════════════════════════════════════════════\n")
}

func getIssueIcon(issueType IssueType) string {
	switch issueType {
	case IssueMissingH1, IssueMissingTitle, IssueMissingMetaDesc, IssueBrokenLink, IssueEmptyH1:
		return "🔴"
	case IssueLongTitle, IssueLongMetaDesc, IssueShortTitle, IssueShortMetaDesc, IssueMultipleH1, IssueRedirectChain, IssueLargeImage, IssueMissingImageAlt:
		return "⚠️"
	case IssueNoCanonical, IssueSlowResponse:
		return "ℹ️"
	default:
		return "•"
	}
}

func formatIssueType(issueType IssueType) string {
	switch issueType {
	case IssueMissingH1:
		return "Missing H1"
	case IssueMissingMetaDesc:
		return "Missing Meta Description"
	case IssueMissingTitle:
		return "Missing Title"
	case IssueLongTitle:
		return "Long Title"
	case IssueLongMetaDesc:
		return "Long Meta Description"
	case IssueShortTitle:
		return "Short Title"
	case IssueShortMetaDesc:
		return "Short Meta Description"
	case IssueLargeImage:
		return "Large Images (>100KB)"
	case IssueMissingImageAlt:
		return "Missing Image Alt Text"
	case IssueSlowResponse:
		return "Slow Response"
	case IssueRedirectChain:
		return "Redirect Chain"
	case IssueNoCanonical:
		return "No Canonical"
	case IssueBrokenLink:
		return "Broken Links"
	case IssueMultipleH1:
		return "Multiple H1 Tags"
	case IssueEmptyH1:
		return "Empty H1 Tag"
	default:
		return string(issueType)
	}
}

