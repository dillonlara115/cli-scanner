<script lang="ts">
	import Features from '../../components/sections/Features.svelte';
</script>

<svelte:head>
	<title>Features - Barracuda SEO</title>
	<meta name="description" content="Discover, analyze, prioritize, and integrate with Barracuda SEO. Crawl thousands of URLs, visualize site structure, and get actionable insights." />
</svelte:head>

<div class="py-20 bg-[#282828]">
	<div class="container mx-auto px-4">
		<div class="text-center mb-16">
			<h1 class="text-5xl md:text-6xl font-heading font-bold mb-4 text-white">Features</h1>
			<p class="text-xl text-white/70 max-w-2xl mx-auto">
				Everything you need to crawl, analyze, and optimize your website's SEO
			</p>
		</div>
	</div>
</div>

<Features />

<section class="py-20 bg-[#1a1a1a]">
	<div class="container mx-auto px-4">
		<div class="max-w-4xl mx-auto">
			<h2 class="text-3xl font-heading font-bold mb-8 text-center text-white">Technical Specifications</h2>
			<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
				<div class="bg-[#282828] rounded-lg p-6 border border-white/10">
					<h3 class="text-xl font-heading font-bold mb-3 text-[#FF6B6B]">Performance</h3>
					<ul class="list-disc list-inside space-y-2 text-white/70">
						<li>100-500 pages/minute crawl speed</li>
						<li>Configurable worker pool (default: 10)</li>
						<li>~50-100 MB memory for 1000 pages</li>
						<li>Concurrent request handling</li>
					</ul>
				</div>
				<div class="bg-[#282828] rounded-lg p-6 border border-white/10">
					<h3 class="text-xl font-heading font-bold mb-3 text-[#FF6B6B]">SEO Analysis</h3>
					<ul class="list-disc list-inside space-y-2 text-white/70">
						<li>Title & meta description detection</li>
						<li>Heading structure analysis (H1-H6)</li>
						<li>Image alt text validation</li>
						<li>Canonical URL tracking</li>
						<li>Response time monitoring</li>
					</ul>
				</div>
				<div class="bg-[#282828] rounded-lg p-6 border border-white/10">
					<h3 class="text-xl font-heading font-bold mb-3 text-[#FF6B6B]">Issue Detection</h3>
					<ul class="list-disc list-inside space-y-2 text-white/70">
						<li>Missing or duplicate H1 tags</li>
						<li>Missing meta descriptions</li>
						<li>Broken links (404 errors)</li>
						<li>Large images (>100KB)</li>
						<li>Redirect chains</li>
					</ul>
				</div>
				<div class="bg-[#282828] rounded-lg p-6 border border-white/10">
					<h3 class="text-xl font-heading font-bold mb-3 text-[#FF6B6B]">Export Formats</h3>
					<ul class="list-disc list-inside space-y-2 text-white/70">
						<li>CSV export with all fields</li>
						<li>JSON export for programmatic use</li>
						<li>Link graph visualization</li>
						<li>Filtered issue exports</li>
					</ul>
				</div>
			</div>
		</div>
	</div>
</section>

