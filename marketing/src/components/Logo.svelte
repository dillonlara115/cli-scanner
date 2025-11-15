<script lang="ts">
	export let size: 'sm' | 'md' | 'lg' = 'md';
	
	$: iconSize = size === 'sm' ? 'w-6 h-6' : size === 'md' ? 'w-8 h-8' : 'w-10 h-10';
	$: textSize = size === 'sm' ? 'text-sm' : size === 'md' ? 'text-xl' : 'text-2xl';
	$: iconTextSize = size === 'sm' ? 'text-xs' : size === 'md' ? 'text-xs' : 'text-sm';
</script>

<div class="flex items-center gap-2">
	<!-- Logo icon -->
	<div class="{iconSize} bg-gradient-to-br from-[#8ec07c] via-[#458588] to-[#d79921] rounded-sm flex items-center justify-center">
		<span class="text-white font-bold {iconTextSize}">B</span>
	</div>
	<span class="text-white {textSize} font-semibold">barracuda</span>
</div>

