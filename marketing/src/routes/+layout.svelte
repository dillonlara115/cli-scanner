<script lang="ts">
	import '../app.css';
	import Header from '../components/layout/Header.svelte';
	import Footer from '../components/layout/Footer.svelte';

	let { children } = $props();
</script>

<svelte:head>
	<title>Barracuda SEO - Web-Based SEO Crawler & Auditing Tool</title>
	<meta name="description" content="Barracuda is a web-based SEO crawler and auditing tool that helps you discover technical issues, prioritize fixes, and improve your site's performance." />
	<meta name="viewport" content="width=device-width, initial-scale=1" />
</svelte:head>

<div class="min-h-screen flex flex-col" data-theme="barracuda">
	<Header />
	<main class="flex-grow">
		{@render children()}
	</main>
	<Footer />
</div>
