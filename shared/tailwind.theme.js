/** @type {import('tailwindcss').Config} */
module.exports = {
	theme: {
		extend: {
			fontFamily: {
				heading: ['Sora', 'sans-serif'],
				body: ['DM Sans', 'sans-serif'],
				mono: ['JetBrains Mono', 'monospace']
			},
			colors: {
				primary: {
					DEFAULT: '#FF6B6B',
					focus: '#FF5252'
				},
				neutral: '#282828',
				'base-100': '#282828',
				'base-content': '#FFFFFF',
				info: '#3B82F6',
				success: '#15803D',
				warning: '#F59E0B',
				error: '#B91C1C'
			}
		}
	}
};

