# Dashboard Improvements Roadmap

This document tracks planned improvements for the Barracuda web dashboard to better serve SEO strategists in identifying, organizing, and solving technical SEO issues.

## Status Legend
- [ ] Not Started
- [🚧] In Progress
- [✅] Completed
- [⏸️] Paused/Blocked

---

## Phase 1: High Impact, Quick Wins

### [✅] Export Issues Functionality
**Priority:** High  
**Impact:** High  
**Effort:** Low-Medium

- [✅] Add "Export Issues" button to IssuesPanel
- [✅] Export filtered issues to CSV
- [✅] Export filtered issues to JSON
- [✅] Include all issue fields: URL, type, severity, message, recommendation
- [✅] Preserve current filters when exporting

**Files to modify:**
- `web/src/components/IssuesPanel.svelte`

**Notes:**
- Use browser download API for file export
- Format CSV headers appropriately
- Include timestamp in filename

---

### [✅] Dashboard Navigation Links
**Priority:** High  
**Impact:** High  
**Effort:** Low

- [✅] Make "Total Issues" card clickable → navigate to Issues tab
- [✅] Add "View All Issues" button in Issue Types card
- [✅] Add quick action buttons in dashboard:
  - [✅] "Fix Critical Issues" → filters to errors
  - [✅] "View Slow Pages" → navigate to Results with performance filter

**Files to modify:**
- `web/src/components/SummaryCard.svelte`
- `web/src/components/Dashboard.svelte`

**Notes:**
- Pass navigation function as prop or use event system
- Highlight target tab when navigating

---

### [✅] Enhanced Issue Filtering & Search
**Priority:** High  
**Impact:** High  
**Effort:** Low-Medium

- [✅] Add search box to filter issues by URL or message text
- [✅] Add "Group by URL" toggle
- [✅] Add "Group by Type" toggle
- [✅] Add "Affected Pages" count for each issue type
- [✅] Filter by affected pages count

**Files to modify:**
- `web/src/components/IssuesPanel.svelte`

**Notes:**
- Search should be real-time (reactive)
- Grouping should be toggleable
- Show count badges for grouped items

---

### [ ] Issue Priority Scoring
**Priority:** Medium  
**Impact:** High  
**Effort:** Medium

- [ ] Calculate priority score: `severity_weight * pages_affected`
- [ ] Display priority score in issue cards
- [ ] Add "Sort by Priority" option
- [ ] Highlight top 10 priority issues

**Files to modify:**
- `web/src/components/IssuesPanel.svelte`
- May need backend calculation: `internal/analyzer/analyzer.go`

**Notes:**
- Severity weights: error=10, warning=5, info=1
- Pages affected = count of unique URLs with this issue type
- Display as badge or numeric score

---

### [ ] Page-Level Issue View
**Priority:** High  
**Impact:** High  
**Effort:** Medium

- [ ] Make ResultsTable rows clickable
- [ ] Show modal/sidebar with page details + all issues for that page
- [ ] Add "View Issues" button in table row
- [ ] Filter ResultsTable: "Show pages with issues only" checkbox
- [ ] Add issue count badge to each row

**Files to modify:**
- `web/src/components/ResultsTable.svelte`
- New component: `web/src/components/PageDetailModal.svelte` (optional)

**Notes:**
- Could use modal or expandable row
- Link to Issues tab filtered by URL
- Show page metadata: title, meta desc, headings, etc.

---

## Phase 2: Medium Priority

### [ ] Issue Status Tracking
**Priority:** Medium  
**Impact:** High  
**Effort:** Medium

- [ ] Add status field to issues: "New", "In Progress", "Fixed", "Ignored"
- [ ] Status change buttons in issue cards
- [ ] Filter by status
- [ ] Bulk status actions (select multiple → mark as fixed)
- [ ] Persist status in browser localStorage or backend

**Files to modify:**
- `web/src/components/IssuesPanel.svelte`
- May need backend support: `cmd/serve.go` (localStorage initially)

**Notes:**
- Start with localStorage for MVP
- Consider backend storage for multi-user scenarios
- Add visual indicators (colors/icons) for status

---

### [ ] Actionable Recommendations Panel
**Priority:** Medium  
**Impact:** Medium  
**Effort:** Medium

- [ ] Create dedicated Recommendations component
- [ ] Show quick fixes with copy-paste solutions
- [ ] Code snippets for common fixes (HTML/JS examples)
- [ ] Estimated impact indicator (potential SEO improvement)
- [ ] Link to documentation/articles

**Files to modify:**
- New component: `web/src/components/RecommendationsPanel.svelte`
- Update `web/src/components/Dashboard.svelte` to include it

**Notes:**
- Could be a new tab or sidebar
- Recommendations should be contextual based on issue type
- Include external links to SEO best practices

---

### [ ] Export Full Report
**Priority:** Medium  
**Impact:** Medium  
**Effort:** Medium-High

- [ ] Generate comprehensive PDF report
- [ ] Generate HTML report (printable)
- [ ] Include: summary, all issues, recommendations, page list
- [ ] Customizable report templates
- [ ] Executive summary section

**Files to modify:**
- New component: `web/src/components/ReportGenerator.svelte`
- May need library: jsPDF or similar

**Notes:**
- PDF generation requires client-side library
- HTML export is simpler (print CSS)
- Include branding/logo

---

### [ ] Enhanced Search & Filter
**Priority:** Medium  
**Impact:** Medium  
**Effort:** Low-Medium

- [ ] Advanced search in IssuesPanel (regex support?)
- [ ] URL pattern matching
- [ ] Filter by multiple severities/types at once
- [ ] Save filter presets
- [ ] Filter by date range (if crawl timestamps available)

**Files to modify:**
- `web/src/components/IssuesPanel.svelte`

**Notes:**
- Keep search simple for MVP
- Advanced features can be added later

---

## Phase 3: Advanced Features

### [ ] Progress Tracking Over Time
**Priority:** Low  
**Impact:** Medium  
**Effort:** High

- [ ] Compare multiple crawl results
- [ ] Issue trend charts (line graph)
- [ ] Before/after comparison view
- [ ] Issue resolution rate calculation
- [ ] Visual progress indicators

**Files to modify:**
- Backend: `cmd/serve.go` (store historical data)
- New component: `web/src/components/ProgressTracker.svelte`

**Notes:**
- Requires backend storage for historical data
- Could use localStorage initially for single-user
- CSV/JSON import for comparison

---

### [ ] Priority Matrix Visualization
**Priority:** Low  
**Impact:** Medium  
**Effort:** Medium

- [ ] 2x2 matrix: Severity vs Impact
- [ ] Color-code critical issues
- [ ] Interactive chart (click to filter)
- [ ] Sortable priority list view

**Files to modify:**
- New component: `web/src/components/PriorityMatrix.svelte`
- May need chart library: Chart.js or similar

**Notes:**
- Can use SVG or canvas for visualization
- DaisyUI might have chart components

---

### [ ] Collaboration Features
**Priority:** Low  
**Impact:** Low-Medium  
**Effort:** High

- [ ] Assign issues to team members
- [ ] Add notes/comments to issues
- [ ] Share specific issues via link
- [ ] Issue activity log

**Files to modify:**
- Backend: `cmd/serve.go` (multi-user support)
- New components for collaboration features

**Notes:**
- Requires backend authentication
- Significant scope increase
- Consider for v2.0

---

### [ ] Integration Suggestions
**Priority:** Low  
**Impact:** Low  
**Effort:** Low

- [ ] Links to relevant tools (PageSpeed Insights, Search Console)
- [ ] API integration hints
- [ ] Export to project management tools (Jira, Trello)

**Files to modify:**
- `web/src/components/IssuesPanel.svelte`
- New component: `web/src/components/Integrations.svelte`

**Notes:**
- Mostly external links
- Could add export formats for PM tools

---

### [ ] Performance Insights
**Priority:** Low  
**Impact:** Medium  
**Effort:** Medium

- [ ] Core Web Vitals summary card
- [ ] Response time distribution chart
- [ ] Identify slow pages with issues
- [ ] Performance recommendations

**Files to modify:**
- `web/src/components/SummaryCard.svelte`
- New component: `web/src/components/PerformanceChart.svelte`

**Notes:**
- Requires additional data collection
- May need to integrate with PageSpeed Insights API

---

### [ ] Link Analysis Improvements
**Priority:** Low  
**Impact:** Medium  
**Effort:** Medium

- [ ] Broken link detection (highlight in graph)
- [ ] Orphan pages visualization (no inbound links)
- [ ] Most linked pages list
- [ ] Link equity visualization

**Files to modify:**
- `web/src/components/LinkGraph.svelte`
- `web/src/components/ResultsTable.svelte`

**Notes:**
- Enhance existing link graph component
- Add new visualizations

---

## Additional Enhancements

### [ ] Contextual Help & Tooltips
**Priority:** Low  
**Impact:** Low-Medium  
**Effort:** Low

- [ ] Info icons with explanations
- [ ] "Why this matters" tooltips
- [ ] SEO best practices links
- [ ] Industry benchmarks comparison

**Files to modify:**
- All components (add tooltips)
- Create `web/src/components/Tooltip.svelte`

**Notes:**
- DaisyUI has tooltip components
- Add help text for each issue type

---

### [ ] Issue Templates/Checklists
**Priority:** Low  
**Impact:** Low  
**Effort:** Medium

- [ ] Pre-built SEO audit checklists
- [ ] Custom issue templates
- [ ] Template library for common SEO issues

**Files to modify:**
- New component: `web/src/components/Checklist.svelte`

**Notes:**
- Could be a separate feature
- Useful for structured audits

---

## Implementation Notes

### Technical Considerations
- **State Management:** Consider using Svelte stores for shared state
- **Data Persistence:** localStorage for MVP, backend for multi-user
- **Export Libraries:** 
  - CSV: Native browser API or `papaparse`
  - JSON: Native `JSON.stringify`
  - PDF: `jsPDF` or `pdfmake`
- **Chart Libraries:** Chart.js, Plotly.js, or D3.js

### UI/UX Patterns
- Use DaisyUI components where possible
- Maintain consistent color scheme (error=red, warning=orange, info=blue)
- Add loading states for async operations
- Implement error boundaries

### Testing Strategy
- Test export functionality with various issue counts
- Test filtering with large datasets
- Test navigation between tabs
- Test responsive design on mobile

---

## Quick Reference: File Structure

```
web/src/components/
├── Dashboard.svelte          # Main dashboard container
├── SummaryCard.svelte        # Dashboard overview stats
├── ResultsTable.svelte       # Page results table
├── IssuesPanel.svelte        # Issues list and filters
├── LinkGraph.svelte          # Link visualization
└── [New Components]
    ├── PageDetailModal.svelte
    ├── RecommendationsPanel.svelte
    ├── ReportGenerator.svelte
    ├── PriorityMatrix.svelte
    └── ProgressTracker.svelte
```

---

## Next Steps

1. **Review and prioritize** - Confirm which Phase 1 items to tackle first
2. **Create GitHub issues** - Break down into actionable tasks
3. **Design mockups** - Sketch UI for new features
4. **Start with Export Issues** - Quick win, high impact
5. **Add navigation links** - Easy improvement to dashboard

---

**Last Updated:** {{ date }}
**Status:** Planning Phase

