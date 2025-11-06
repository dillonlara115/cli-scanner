const shared = require('../shared/tailwind.theme.js');
const daisyui = require('daisyui');

/** @type {import('tailwindcss').Config} */
module.exports = {
	...shared,
	content: ['./src/**/*.{svelte,js,ts}'],
	plugins: [
		daisyui
	],
	daisyui: {
		themes: [
			{
				barracuda: {
					primary: '#FF6B6B',
					'primary-focus': '#FF5252',
					neutral: '#282828',
					'base-100': '#282828',
					'base-content': '#FFFFFF',
					info: '#3B82F6',
					success: '#15803D',
					warning: '#F59E0B',
					error: '#B91C1C'
				}
			}
		],
		darkTheme: 'barracuda',
		base: true,
		styled: true,
		utils: true,
		prefix: '',
		logs: true,
		themeRoot: ':root'
	}
};

