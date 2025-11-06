export interface MetaTags {
	title?: string;
	description?: string;
	ogImage?: string;
	ogType?: string;
}

export function getMetaTags(meta: MetaTags = {}) {
	const title = meta.title 
		? `${meta.title} - Barracuda SEO`
		: 'Barracuda SEO - Fast Website Crawler & SEO Analysis Tool';
	
	const description = meta.description || 
		'A fast, lightweight SEO website crawler CLI tool. Crawl, analyze, and optimize your site\'s SEO in minutes.';

	return {
		title,
		description,
		ogTitle: title,
		ogDescription: description,
		ogImage: meta.ogImage || '/og-image.png',
		ogType: meta.ogType || 'website'
	};
}

