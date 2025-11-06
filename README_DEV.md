# 🧭 Barracuda Marketing Site — Developer Guide

**Stack:** SvelteKit + Tailwind + DaisyUI + Vercel
**Editor:** Cursor IDE (optimized for AI pair-programming)

---

## ⚙️ Quick Start

```bash
# clone & open project
git clone https://github.com/<your-org>/barracuda.git
cd marketing

# install dependencies
npm install

# start dev server
npm run dev

# build for production
npm run build && npm run preview
```

Site runs at → [http://localhost:5173](http://localhost:5173)

---

## 🧩 Project Structure

```
marketing/
├── src/
│   ├── routes/
│   │   ├── +layout.svelte       # global layout + meta
│   │   ├── +page.svelte         # homepage
│   │   ├── features/+page.svelte
│   │   ├── pricing/+page.svelte
│   │   ├── blog/[slug]/+page.svelte
│   │   └── about/+page.svelte
│   ├── components/
│   │   ├── layout/              # header, footer, nav
│   │   ├── ui/                  # buttons, cards, modals
│   │   └── sections/            # hero, features, pricing, etc.
│   ├── lib/                     # meta tags, helpers, constants
│   ├── styles/                  # theme + typography
│   └── app.html
├── tailwind.config.cjs
├── postcss.config.cjs
├── svelte.config.js
├── vite.config.js
└── README_DEV.md                # this file
```

---

## 🎨 Design System

### Colors (DaisyUI `barracuda` theme)

| Token           | Hex       | Purpose          |
| --------------- | --------- | ---------------- |
| `primary`       | `#FF6B6B` | CTA / highlights |
| `primary-focus` | `#FF5252` | hover states     |
| `neutral`       | `#282828` | base background  |
| `base-100`      | `#282828` | main BG          |
| `base-content`  | `#FFFFFF` | default text     |
| `info`          | `#3B82F6` | links / info     |
| `success`       | `#15803D` | success states   |
| `warning`       | `#F59E0B` | warnings         |
| `error`         | `#B91C1C` | errors           |

Typography
Import from google fonts
* **Heading:** `Sora 700`
* **Body:** `DM Sans 400–500`
* **Code:** `JetBrains Mono 400–500`

---

## 🧱 Tailwind Setup

Shared theme lives at `/shared/tailwind.theme.js`

```js
const shared = require('../shared/tailwind.theme.js');
module.exports = { ...shared, content: ['./src/**/*.{svelte,js,ts}'] };
```

Use classes:

```html
<h1 class="font-heading text-3xl text-primary">Heading</h1>
<p class="font-body text-base text-base-content">Paragraph text</p>
<code class="font-mono text-info">barracuda crawl</code>
```

---

## 🧩 Component Conventions

* **Layout components:** `layout/Header.svelte`, `layout/Footer.svelte`
* **Sections:** standalone marketing blocks (Hero, Features, Pricing)
* **UI components:** atomic buttons/cards using DaisyUI classes
* **Animation:** `@motionone/svelte` or native Tailwind transitions
* **File naming:** `PascalCase.svelte` for components, lowercase for routes

---

## 🧠 Cursor IDE Prompts

Paste into Cursor’s chat to bootstrap new sections quickly:

> *System prompt:*
> “We are building a dark SaaS marketing site for Barracuda SEO using SvelteKit + DaisyUI + Tailwind.
> Use the brand theme (primary #FF6B6B, base #282828) and fonts Sora + DM Sans + JetBrains Mono.
> Generate responsive, accessible Svelte components consistent with the app UI.”

### Example Tasks

* “Create a responsive Features grid with 3 cards using DaisyUI cards.”
* “Build a Pricing section with three tiers: Free, Pro ($49), Team ($149).”
* “Add smooth scroll + intersection fade-in animation to Hero section.”

Cursor will auto-generate components under `src/components/sections/`.

---

## 🌐 Deployment

**Platform:** Vercel
**Settings**

```
Framework: SvelteKit
Build command: npm run build
Output directory: .svelte-kit/output
```

**Domains**

| Site      | URL                  | Source     |
| --------- | -------------------- | ---------- |
| Marketing | barracudaseo.com     | marketing/ |
| App       | app.barracudaseo.com | app/       |

Environment vars (optional):

```
PUBLIC_SUPABASE_URL
PUBLIC_SUPABASE_ANON_KEY
```

---

## 🧠 Best Practices

* Maintain consistency with the app theme (use shared theme + typography).
* Keep components small and composable.
* Use semantic HTML for SEO.
* Optimize images (WebP + responsive sizes).
* Add `<svelte:head>` metadata in each route for title/description.
* Run Lighthouse audits locally (`npm run preview` + Chrome DevTools).

---

## 🧩 Future Enhancements

| Feature          | Status | Notes                                              |
| ---------------- | ------ | -------------------------------------------------- |
| Blog w/ Markdown | ⏳      | `/blog/[slug].md` using `@sveltejs/adapter-static` |
| CMS Integration  | ⏳      | Sanity / Supabase                                  |
| SEO Schema       | ⏳      | JSON-LD injection in `<svelte:head>`               |
| Analytics        | ⏳      | Plausible or GA4 script                            |
| Docs             | ⏳      | `/docs` static site (future)                       |

---

## 🧰 Commands Summary

| Task                 | Command                  |
| -------------------- | ------------------------ |
| Run dev server       | `npm run dev`            |
| Build for production | `npm run build`          |
| Preview local build  | `npm run preview`        |
| Format code          | `npx prettier --write .` |

---

### 💡 Tip for AI Workflow

When generating new components in Cursor:

1. Highlight the section placeholder (e.g., `<!-- Features Section -->`)
2. Type `/ai` → prompt Cursor:
   *“Create a responsive features section using DaisyUI cards in dark theme.”*
3. Review and tweak directly — Cursor will auto-import theme and Tailwind utilities.

---

**Last Updated:** {{ date }}
Maintainer: @dillonlara

---
