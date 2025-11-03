<script>
  export let issues = [];
  export let navigateToTab = null;
  export let enrichedIssues = {}; // Map of enriched issue data

  // Recommendation data for each issue type
  const recommendations = {
    missing_h1: {
      title: "Add H1 Heading",
      impact: "High",
      description: "Each page should have exactly one H1 heading that clearly describes the page content.",
      codeSnippet: `<h1>Your Page Title Here</h1>`,
      explanation: "Place the H1 near the top of your main content, ideally within the first 100 words.",
      resources: [
        { name: "H1 Tag Best Practices", url: "https://moz.com/learn/seo/h1-tag" },
        { name: "Google H1 Guidelines", url: "https://developers.google.com/search/docs/appearance/heading-titles" }
      ]
    },
    missing_title: {
      title: "Add Page Title",
      impact: "Critical",
      description: "Every page must have a unique, descriptive title tag for SEO and user experience.",
      codeSnippet: `<title>Your Page Title - Your Brand Name</title>`,
      explanation: "Place in the <head> section. Keep it under 60 characters for optimal display in search results.",
      resources: [
        { name: "Title Tag Best Practices", url: "https://moz.com/learn/seo/title-tag" },
        { name: "Google Title Guidelines", url: "https://developers.google.com/search/docs/appearance/title-link" }
      ]
    },
    missing_meta_description: {
      title: "Add Meta Description",
      impact: "Medium",
      description: "Meta descriptions improve click-through rates from search results.",
      codeSnippet: `<meta name="description" content="A compelling description of your page content (150-160 characters).">`,
      explanation: "Place in the <head> section. Write compelling copy that encourages clicks while accurately describing the content.",
      resources: [
        { name: "Meta Description Guide", url: "https://moz.com/learn/seo/meta-description" },
        { name: "Writing Effective Descriptions", url: "https://developers.google.com/search/docs/appearance/snippet" }
      ]
    },
    long_title: {
      title: "Shorten Page Title",
      impact: "Medium",
      description: "Titles over 60 characters may be truncated in search results.",
      codeSnippet: `<title>Shorter, More Focused Title</title>`,
      explanation: "Keep the most important keywords at the beginning. Aim for 50-60 characters.",
      resources: [
        { name: "Title Length Guidelines", url: "https://moz.com/learn/seo/title-tag" }
      ]
    },
    short_title: {
      title: "Expand Page Title",
      impact: "Low",
      description: "Very short titles may not provide enough context for search engines and users.",
      codeSnippet: `<title>Descriptive Title That Provides Context</title>`,
      explanation: "Include primary keywords and brand name. Aim for at least 30 characters.",
      resources: [
        { name: "Title Tag Best Practices", url: "https://moz.com/learn/seo/title-tag" }
      ]
    },
    long_meta_description: {
      title: "Shorten Meta Description",
      impact: "Low",
      description: "Meta descriptions over 160 characters may be truncated in search results.",
      codeSnippet: `<meta name="description" content="Concise description within 150-160 characters.">`,
      explanation: "Keep the most compelling information at the beginning. Aim for 150-160 characters.",
      resources: [
        { name: "Meta Description Best Practices", url: "https://moz.com/learn/seo/meta-description" }
      ]
    },
    short_meta_description: {
      title: "Expand Meta Description",
      impact: "Low",
      description: "Very short meta descriptions may not be compelling enough to encourage clicks.",
      codeSnippet: `<meta name="description" content="A more detailed description that provides value and encourages clicks.">`,
      explanation: "Include key benefits and a call-to-action. Aim for at least 120 characters.",
      resources: [
        { name: "Meta Description Guide", url: "https://moz.com/learn/seo/meta-description" }
      ]
    },
    multiple_h1: {
      title: "Use Single H1 Per Page",
      impact: "Medium",
      description: "Having multiple H1 tags can confuse search engines about your page's main topic.",
      codeSnippet: `<h1>Main Page Heading</h1>
<h2>Subsection Heading</h2>
<h2>Another Subsection</h2>`,
      explanation: "Use only one H1 for the main page heading. Use H2-H6 for subsections.",
      resources: [
        { name: "H1 Tag Best Practices", url: "https://moz.com/learn/seo/h1-tag" }
      ]
    },
    empty_h1: {
      title: "Add Content to H1 Tag",
      impact: "High",
      description: "Empty H1 tags provide no value to users or search engines.",
      codeSnippet: `<h1>Meaningful Heading Text</h1>`,
      explanation: "Ensure your H1 contains descriptive text that summarizes the page content.",
      resources: [
        { name: "H1 Tag Guidelines", url: "https://moz.com/learn/seo/h1-tag" }
      ]
    },
    missing_image_alt: {
      title: "Add Alt Text to Images",
      impact: "Medium",
      description: "Alt text improves accessibility and helps search engines understand images.",
      codeSnippet: `<img src="image.jpg" alt="Descriptive text explaining what the image shows">`,
      explanation: "Write concise, descriptive alt text that explains the image's content and purpose.",
      resources: [
        { name: "Alt Text Best Practices", url: "https://moz.com/learn/seo/alt-text" },
        { name: "Accessibility Guidelines", url: "https://www.w3.org/WAI/tutorials/images/" }
      ]
    },
    large_image: {
      title: "Optimize Image Size",
      impact: "Medium",
      description: "Large images slow down page load times, impacting user experience and SEO.",
      codeSnippet: `<!-- Optimize images before uploading -->
<img src="optimized-image.jpg" 
     srcset="image-400w.jpg 400w, image-800w.jpg 800w"
     sizes="(max-width: 400px) 400px, 800px"
     alt="Image description">`,
      explanation: "Compress images, use modern formats (WebP), and implement responsive images with srcset.",
      resources: [
        { name: "Image Optimization Guide", url: "https://web.dev/fast/#optimize-your-images" },
        { name: "PageSpeed Insights", url: "https://pagespeed.web.dev/" }
      ]
    },
    slow_response: {
      title: "Improve Page Speed",
      impact: "High",
      description: "Slow-loading pages hurt user experience and search rankings.",
      codeSnippet: `<!-- Enable compression -->
<!-- Use CDN for static assets -->
<!-- Minimize HTTP requests -->
<!-- Optimize server response time -->`,
      explanation: "Optimize images, enable caching, use a CDN, minimize JavaScript/CSS, and improve server response time.",
      resources: [
        { name: "PageSpeed Optimization", url: "https://web.dev/fast/" },
        { name: "Core Web Vitals", url: "https://web.dev/vitals/" },
        { name: "Google PageSpeed Insights", url: "https://pagespeed.web.dev/" }
      ]
    },
    redirect_chain: {
      title: "Simplify Redirect Chains",
      impact: "Medium",
      description: "Redirect chains waste crawl budget and slow down page loads.",
      codeSnippet: `<!-- Update links to point directly to final destination -->
<a href="https://example.com/final-page">Link Text</a>

<!-- Or configure server to redirect directly -->
<!-- Apache: Redirect 301 /old-page /final-page -->
<!-- Nginx: return 301 /final-page; -->`,
      explanation: "Update internal links to point directly to the final destination. Use 301 redirects for permanent moves.",
      resources: [
        { name: "Redirect Best Practices", url: "https://moz.com/learn/seo/redirection" },
        { name: "HTTP Status Codes", url: "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status" }
      ]
    },
    no_canonical: {
      title: "Add Canonical Tag",
      impact: "Low",
      description: "Canonical tags help prevent duplicate content issues.",
      codeSnippet: `<link rel="canonical" href="https://example.com/canonical-page-url">`,
      explanation: "Place in the <head> section. Point to the preferred version of the page (usually self-referential).",
      resources: [
        { name: "Canonical Tag Guide", url: "https://moz.com/learn/seo/canonicalization" },
        { name: "Google Canonical Guidelines", url: "https://developers.google.com/search/docs/crawling-indexing/consolidate-duplicate-urls" }
      ]
    },
    broken_link: {
      title: "Fix Broken Links",
      impact: "Medium",
      description: "Broken links hurt user experience and may indicate crawlability issues.",
      codeSnippet: `<!-- Update link to correct URL -->
<a href="https://example.com/correct-page">Link Text</a>

<!-- Or remove if page no longer exists -->
<!-- Or redirect to relevant alternative page -->`,
      explanation: "Update links to point to valid pages, remove dead links, or redirect to relevant alternatives.",
      resources: [
        { name: "Link Building Best Practices", url: "https://moz.com/learn/seo/internal-link" },
        { name: "HTTP Status Codes", url: "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status" }
      ]
    }
  };

  // Group issues by type and get unique recommendations
  $: issueTypes = [...new Set(issues.map(i => i.type))];
  
  $: recommendationsToShow = issueTypes.map(type => {
    const rec = recommendations[type];
    if (!rec) return null;
    
    // Count affected pages for this issue type
    const affectedPages = issues.filter(i => i.type === type).length;
    
    return {
      issueType: type,
      ...rec,
      affectedPages,
      severity: issues.find(i => i.type === type)?.severity || 'info'
    };
  }).filter(r => r !== null);

  // Sort by impact (High > Medium > Low) and affected pages
  // If enriched data available, use enriched priority for sorting
  $: sortedRecommendations = [...recommendationsToShow].sort((a, b) => {
    // First, check if we have enriched priority for these issue types
    const enrichedA = issues.find(i => i.type === a.issueType && enrichedIssues[`${i.url}|${i.type}`]);
    const enrichedB = issues.find(i => i.type === b.issueType && enrichedIssues[`${i.url}|${i.type}`]);
    
    if (enrichedA && enrichedB) {
      const priorityA = enrichedIssues[`${enrichedA.url}|${enrichedA.type}`]?.enriched_priority || 0;
      const priorityB = enrichedIssues[`${enrichedB.url}|${enrichedB.type}`]?.enriched_priority || 0;
      if (priorityA !== priorityB) {
        return priorityB - priorityA; // Descending
      }
    }
    
    // Fallback to impact and affected pages
    const impactOrder = { 'Critical': 4, 'High': 3, 'Medium': 2, 'Low': 1 };
    const impactDiff = impactOrder[b.impact] - impactOrder[a.impact];
    if (impactDiff !== 0) return impactDiff;
    return b.affectedPages - a.affectedPages;
  });

  const copyToClipboard = async (text) => {
    try {
      await navigator.clipboard.writeText(text);
      // Show temporary feedback
      return true;
    } catch (err) {
      console.error('Failed to copy:', err);
      return false;
    }
  };

  const getImpactColor = (impact) => {
    switch (impact) {
      case 'Critical': return 'badge-error';
      case 'High': return 'badge-warning';
      case 'Medium': return 'badge-info';
      case 'Low': return 'badge-ghost';
      default: return 'badge-ghost';
    }
  };
</script>

<div class="card bg-base-100 shadow">
  <div class="card-body">
    <h2 class="card-title mb-4">Actionable Recommendations</h2>
    
    {#if sortedRecommendations.length === 0}
      <div class="alert alert-success">
        <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span>No issues found! Your site is looking great.</span>
      </div>
    {:else}
      <div class="space-y-6">
        {#each sortedRecommendations as rec}
          <div class="card bg-base-200 shadow-md">
            <div class="card-body">
              <div class="flex items-start justify-between mb-2">
                <div class="flex-1">
                  <h3 class="card-title text-lg">{rec.title}</h3>
                  <div class="flex gap-2 mt-1">
                    <span class="badge {getImpactColor(rec.impact)}">{rec.impact} Impact</span>
                    <span class="badge badge-ghost">{rec.affectedPages} page{rec.affectedPages !== 1 ? 's' : ''} affected</span>
                  </div>
                </div>
              </div>
              
              <p class="text-base-content/80 mb-4">{rec.description}</p>
              
              <!-- Code Snippet -->
              <div class="mb-4">
                <div class="flex items-center justify-between mb-2">
                  <span class="text-sm font-semibold">Code Example:</span>
                  <button 
                    class="btn btn-xs btn-ghost"
                    on:click={() => copyToClipboard(rec.codeSnippet)}
                    title="Copy to clipboard"
                  >
                    📋 Copy
                  </button>
                </div>
                <pre class="bg-base-300 p-4 rounded-lg overflow-x-auto text-sm"><code>{rec.codeSnippet}</code></pre>
              </div>
              
              <p class="text-sm text-base-content/70 mb-4">{rec.explanation}</p>
              
              <!-- Resources -->
              {#if rec.resources && rec.resources.length > 0}
                <div class="divider"></div>
                <div>
                  <span class="text-sm font-semibold mb-2 block">Learn More:</span>
                  <div class="flex flex-wrap gap-2">
                    {#each rec.resources as resource}
                      <a 
                        href={resource.url} 
                        target="_blank" 
                        rel="noopener noreferrer"
                        class="btn btn-xs btn-outline"
                      >
                        {resource.name} ↗
                      </a>
                    {/each}
                  </div>
                </div>
              {/if}
              
              <!-- Link to view issues -->
              <div class="card-actions justify-end mt-4">
                {#if navigateToTab}
                  <button 
                    class="btn btn-sm btn-primary"
                    on:click={() => navigateToTab('issues', { type: rec.issueType })}
                  >
                    View {rec.affectedPages} Issue{rec.affectedPages !== 1 ? 's' : ''}
                  </button>
                {:else}
                  <a 
                    href="#issues" 
                    class="btn btn-sm btn-primary"
                  >
                    View {rec.affectedPages} Issue{rec.affectedPages !== 1 ? 's' : ''}
                  </a>
                {/if}
              </div>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<style>
  pre code {
    white-space: pre;
    font-family: 'Courier New', monospace;
  }
</style>

