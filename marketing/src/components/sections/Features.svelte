<script lang="ts">
	import { Search, BarChart3, Target, Plug } from 'lucide-svelte';
</script>

<section class="py-20 bg-[#282828]">
	<div class="container mx-auto px-4">
		<div class="text-center mb-16">
			<h2 class="text-4xl md:text-5xl font-heading font-bold mb-4 text-white">Powerful Features</h2>
			<p class="text-xl text-white/70 max-w-2xl mx-auto">
				Everything you need to crawl, analyze, and optimize your website's SEO
			</p>
		</div>
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8">
			<div class="bg-[#1a1a1a] rounded-lg shadow-xl p-6 border border-white/10">
				<div class="flex items-center gap-3 mb-3">
					<div class="p-2 bg-[#FF6B6B]/10 rounded-lg">
						<Search class="w-5 h-5 text-[#FF6B6B]" />
					</div>
					<h3 class="text-xl font-heading font-bold text-[#FF6B6B]">Discover</h3>
				</div>
				<p class="text-white/70">
					Crawl up to tens of thousands of URLs and identify broken links, duplicate tags, missing meta data, slow pages, and more — fast.
				</p>
			</div>
			<div class="bg-[#1a1a1a] rounded-lg shadow-xl p-6 border border-white/10">
				<div class="flex items-center gap-3 mb-3">
					<div class="p-2 bg-[#FF6B6B]/10 rounded-lg">
						<BarChart3 class="w-5 h-5 text-[#FF6B6B]" />
					</div>
					<h3 class="text-xl font-heading font-bold text-[#FF6B6B]">Analyze</h3>
				</div>
				<p class="text-white/70">
					Visualize your site architecture, link structure, and SEO health inside a clean, intuitive web dashboard.
				</p>
			</div>
			<div class="bg-[#1a1a1a] rounded-lg shadow-xl p-6 border border-white/10">
				<div class="flex items-center gap-3 mb-3">
					<div class="p-2 bg-[#FF6B6B]/10 rounded-lg">
						<Target class="w-5 h-5 text-[#FF6B6B]" />
					</div>
					<h3 class="text-xl font-heading font-bold text-[#FF6B6B]">Prioritize</h3>
				</div>
				<p class="text-white/70">
					Get clear issue summaries with built-in severity and impact scoring to focus on what truly matters.
				</p>
			</div>
			<div class="bg-[#1a1a1a] rounded-lg shadow-xl p-6 border border-white/10">
				<div class="flex items-center gap-3 mb-3">
					<div class="p-2 bg-[#FF6B6B]/10 rounded-lg">
						<Plug class="w-5 h-5 text-[#FF6B6B]" />
					</div>
					<h3 class="text-xl font-heading font-bold text-[#FF6B6B]">Integrate</h3>
				</div>
				<p class="text-white/70">
					Connect Google Search Console, Analytics, and Clarity to enrich your audit data and track performance over time. <span class="text-white/50 text-sm">(Pro)</span>
				</p>
			</div>
		</div>
	</div>
</section>

