<script lang="ts">
	import { page } from '$app/stores';
</script>

<header class="bg-[#282828] border-b border-white/10 sticky top-0 z-50 backdrop-blur-sm bg-[#282828]/80">
	<div class="container mx-auto px-4">
		<nav class="flex items-center justify-between h-16">
			<div class="flex-1">
				<a href="/" class="text-xl font-heading font-bold text-[#FF6B6B] hover:text-[#FF5252] transition-colors">
					Barracuda SEO
				</a>
			</div>
			<div class="flex-none">
				<ul class="flex items-center gap-6">
					<li>
						<a href="/" class={`hover:text-[#FF6B6B] transition-colors ${$page.url.pathname === '/' ? 'text-[#FF6B6B]' : 'text-white'}`}>
							Home
						</a>
					</li>
					<li>
						<a href="/features" class={`hover:text-[#FF6B6B] transition-colors ${$page.url.pathname === '/features' ? 'text-[#FF6B6B]' : 'text-white'}`}>
							Features
						</a>
					</li>
					<li>
						<a href="/pricing" class={`hover:text-[#FF6B6B] transition-colors ${$page.url.pathname === '/pricing' ? 'text-[#FF6B6B]' : 'text-white'}`}>
							Pricing
						</a>
					</li>
					<li>
						<a href="/about" class={`hover:text-[#FF6B6B] transition-colors ${$page.url.pathname === '/about' ? 'text-[#FF6B6B]' : 'text-white'}`}>
							About
						</a>
					</li>
					<li>
						<a href="https://app.barracudaseo.com" class="bg-[#FF6B6B] hover:bg-[#FF5252] text-white px-4 py-2 rounded-lg font-medium transition-colors" target="_blank" rel="noopener noreferrer">
							Get Started
						</a>
					</li>
				</ul>
			</div>
		</nav>
	</div>
</header>

