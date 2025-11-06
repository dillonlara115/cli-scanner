<script lang="ts">
	import Pricing from '../../components/sections/Pricing.svelte';
</script>

<svelte:head>
	<title>Pricing - Barracuda SEO</title>
	<meta name="description" content="Start with a free 100-page audit. Upgrade to Pro for 10,000+ pages, recommendations, integrations, and team collaboration." />
</svelte:head>

<div class="py-20 bg-[#282828]">
	<div class="container mx-auto px-4">
		<div class="text-center mb-16">
			<h1 class="text-5xl md:text-6xl font-heading font-bold mb-4 text-white">Pricing</h1>
			<p class="text-xl text-white/70 max-w-2xl mx-auto">
				Simple, transparent pricing for teams of all sizes
			</p>
		</div>
	</div>
</div>

<Pricing />

<section class="py-20 bg-[#282828]">
	<div class="container mx-auto px-4">
		<div class="max-w-3xl mx-auto text-center">
			<h2 class="text-3xl font-heading font-bold mb-4 text-white">Frequently Asked Questions</h2>
			<div class="space-y-6 mt-8 text-left">
				<div class="bg-[#1a1a1a] rounded-lg p-6 border border-white/10">
					<h3 class="text-xl font-heading font-bold mb-3 text-[#FF6B6B]">Can I try before I buy?</h3>
					<p class="text-white/70">
						Yes! The Free plan includes all core crawling features. Upgrade to Pro or Team when you need cloud workspace and team collaboration.
					</p>
				</div>
				<div class="bg-[#1a1a1a] rounded-lg p-6 border border-white/10">
					<h3 class="text-xl font-heading font-bold mb-3 text-[#FF6B6B]">What's the difference between Pro and Team?</h3>
					<p class="text-white/70">
						Pro is perfect for individual users and small teams. Team includes unlimited projects, advanced analytics, API access, and dedicated support for larger organizations.
					</p>
				</div>
				<div class="bg-[#1a1a1a] rounded-lg p-6 border border-white/10">
					<h3 class="text-xl font-heading font-bold mb-3 text-[#FF6B6B]">Do you offer refunds?</h3>
					<p class="text-white/70">
						Yes, we offer a 30-day money-back guarantee. If you're not satisfied, contact us for a full refund.
					</p>
				</div>
			</div>
		</div>
	</div>
</section>

