<script lang="ts">
	import { page } from '$app/stores';
	import Logo from '../Logo.svelte';
</script>

<header class="bg-[#3c3836] border-b border-white/10 sticky top-0 z-50 backdrop-blur-sm bg-[#3c3836]/80">
	<div class="container mx-auto px-4">
		<nav class="flex items-center justify-between h-16">
			<div class="flex-1">
				<a href="/" class="hover:opacity-80 transition-opacity">
					<Logo size="md" />
				</a>
			</div>
			<div class="flex-none">
				<ul class="flex items-center gap-6">
					<li>
						<a href="/" class={`hover:text-[#8ec07c] transition-colors ${$page.url.pathname === '/' ? 'text-[#8ec07c]' : 'text-white'}`}>
							Home
						</a>
					</li>
					<li>
						<a href="/features" class={`hover:text-[#8ec07c] transition-colors ${$page.url.pathname === '/features' ? 'text-[#8ec07c]' : 'text-white'}`}>
							Features
						</a>
					</li>
					<li>
						<a href="/pricing" class={`hover:text-[#8ec07c] transition-colors ${$page.url.pathname === '/pricing' ? 'text-[#8ec07c]' : 'text-white'}`}>
							Pricing
						</a>
					</li>
					<li>
						<a href="/about" class={`hover:text-[#8ec07c] transition-colors ${$page.url.pathname === '/about' ? 'text-[#8ec07c]' : 'text-white'}`}>
							About
						</a>
					</li>
					<li>
						<a href="https://app.barracudaseo.com" class="bg-[#8ec07c] hover:bg-[#a0d28c] text-[#3c3836] px-4 py-2 rounded-lg font-medium transition-colors" target="_blank" rel="noopener noreferrer">
							Get Started
						</a>
					</li>
				</ul>
			</div>
		</nav>
	</div>
</header>

