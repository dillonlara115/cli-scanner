# Setup Notes

## Tailwind v4 and DaisyUI Compatibility

This project is currently using **Tailwind CSS v4** (via `@tailwindcss/vite`), which uses CSS-based configuration. 

**Important:** DaisyUI may not fully support Tailwind v4 yet. If you encounter issues with DaisyUI classes not working:

### Option 1: Use Tailwind v3 (Recommended for DaisyUI)
If you need full DaisyUI support, downgrade to Tailwind v3:

```bash
npm install -D tailwindcss@^3 postcss autoprefixer
npm uninstall @tailwindcss/vite
```

Then update `vite.config.ts` to remove the `tailwindcss()` plugin and use PostCSS instead.

### Option 2: Use Standard Tailwind Classes
The components are written to work with standard Tailwind utility classes. DaisyUI classes (like `btn`, `card`, etc.) can be replaced with standard Tailwind classes if needed.

### Current Configuration
- Tailwind v4 via `@tailwindcss/vite` plugin
- Theme configured in `src/app.css` using `@theme` directive
- DaisyUI installed but may need Tailwind v3 for full functionality

## Testing

Run the dev server to test:
```bash
npm run dev
```

If DaisyUI classes don't work, you may need to switch to Tailwind v3 or update components to use standard Tailwind classes.

