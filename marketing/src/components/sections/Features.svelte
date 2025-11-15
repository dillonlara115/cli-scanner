<script lang="ts">
	import { Search, BarChart3, Target, Plug, Users } from 'lucide-svelte';
</script>

<section class="py-20 bg-[#3c3836]">
	<div class="container mx-auto px-4">
		<div class="text-center mb-16">
			<h2 class="text-4xl md:text-5xl font-heading font-bold mb-4 text-white">Powerful Features</h2>
			<p class="text-xl text-white/70 max-w-2xl mx-auto">
				Everything you need to crawl, analyze, and optimize your website's SEO
			</p>
		</div>
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-8">
			<div class="bg-[#2d2826] rounded-lg shadow-xl p-6 border border-white/10">
				<div class="flex items-center gap-3 mb-3">
					<div class="p-2 bg-[#8ec07c]/10 rounded-lg">
						<Search class="w-5 h-5 text-[#8ec07c]" />
					</div>
					<h3 class="text-xl font-heading font-bold text-[#8ec07c]">Discover</h3>
				</div>
				<p class="text-white/70">
					Crawl up to tens of thousands of URLs and identify broken links, duplicate tags, missing meta data, slow pages, and more — fast.
				</p>
			</div>
			<div class="bg-[#2d2826] rounded-lg shadow-xl p-6 border border-white/10">
				<div class="flex items-center gap-3 mb-3">
					<div class="p-2 bg-[#8ec07c]/10 rounded-lg">
						<BarChart3 class="w-5 h-5 text-[#8ec07c]" />
					</div>
					<h3 class="text-xl font-heading font-bold text-[#8ec07c]">Analyze</h3>
				</div>
				<p class="text-white/70">
					Visualize your site architecture, link structure, and SEO health inside a clean, intuitive web dashboard.
				</p>
			</div>
			<div class="bg-[#2d2826] rounded-lg shadow-xl p-6 border border-white/10">
				<div class="flex items-center gap-3 mb-3">
					<div class="p-2 bg-[#8ec07c]/10 rounded-lg">
						<Target class="w-5 h-5 text-[#8ec07c]" />
					</div>
					<h3 class="text-xl font-heading font-bold text-[#8ec07c]">Prioritize</h3>
				</div>
				<p class="text-white/70">
					Get clear issue summaries with built-in severity and impact scoring to focus on what truly matters.
				</p>
			</div>
			<div class="bg-[#2d2826] rounded-lg shadow-xl p-6 border border-[#8ec07c]/30 relative">
				<div class="absolute -top-2 -right-2 bg-[#8ec07c] text-[#3c3836] px-2 py-1 rounded text-xs font-bold">Pro</div>
				<div class="flex items-center gap-3 mb-3">
					<div class="p-2 bg-[#8ec07c]/10 rounded-lg">
						<Users class="w-5 h-5 text-[#8ec07c]" />
					</div>
					<h3 class="text-xl font-heading font-bold text-[#8ec07c]">Collaborate</h3>
				</div>
				<p class="text-white/70">
					Invite teammates with role-based permissions. Perfect for agencies and teams who need shared project access and seamless collaboration.
				</p>
			</div>
			<div class="bg-[#2d2826] rounded-lg shadow-xl p-6 border border-[#8ec07c]/30 relative">
				<div class="absolute -top-2 -right-2 bg-[#8ec07c] text-[#3c3836] px-2 py-1 rounded text-xs font-bold">Pro</div>
				<div class="flex items-center gap-3 mb-3">
					<div class="p-2 bg-[#8ec07c]/10 rounded-lg">
						<Plug class="w-5 h-5 text-[#8ec07c]" />
					</div>
					<h3 class="text-xl font-heading font-bold text-[#8ec07c]">Integrate</h3>
				</div>
				<p class="text-white/70">
					Connect Google Search Console, Analytics, and Microsoft Clarity to enrich your audit data and track performance over time.
				</p>
			</div>
		</div>
	</div>
</section>

