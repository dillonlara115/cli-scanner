<script>
  import { signIn, signUp, signOut } from '../lib/auth.js';
  import { user } from '../lib/auth.js';
  import { link, push } from 'svelte-spa-router';
  import { Plug } from 'lucide-svelte';
  import Logo from './Logo.svelte';

  let isSignUp = false;
  let email = '';
  let password = '';
  let firstName = '';
  let lastName = '';
  let showPassword = false;
  let loading = false;
  let error = null;
  let success = null;

  $: isAuthenticated = $user !== null;

  async function handleSubmit() {
    loading = true;
    error = null;
    success = null;

    try {
      const displayName = isSignUp ? `${firstName} ${lastName}`.trim() : '';
      if (isSignUp) {
        const { data, error: signUpError } = await signUp(email, password, displayName);
        if (signUpError) throw signUpError;
        success = 'Account created successfully!';
      } else {
        const { data, error: signInError } = await signIn(email, password);
        if (signInError) throw signInError;
        success = 'Signed in successfully!';
      }
    } catch (err) {
      error = err.message || 'An error occurred';
    } finally {
      loading = false;
    }
  }

  async function handleSignOut() {
    loading = true;
    error = null;
    try {
      const { error: signOutError } = await signOut();
      if (signOutError) throw signOutError;
    } catch (err) {
      error = err.message || 'An error occurred';
    } finally {
      loading = false;
    }
  }
</script>

{#if isAuthenticated}
  <div class="dropdown dropdown-end">
    <button type="button" tabindex="0" class="btn btn-ghost">
      <div class="avatar placeholder">
        <div class="bg-neutral text-neutral-content rounded-full w-8">
          <span class="text-xs">{$user?.email?.charAt(0).toUpperCase() || 'U'}</span>
        </div>
      </div>
      <span class="ml-2">{$user?.email || 'User'}</span>
    </button>
    <ul tabindex="0" role="menu" class="dropdown-content menu bg-base-100 rounded-box z-[1] w-52 p-2 shadow">
      <li>
        <a href="/integrations" use:link role="menuitem">
          <Plug class="w-5 h-5" />
          Integrations
        </a>
      </li>
      <li>
        <button type="button" role="menuitem" on:click={handleSignOut} class="text-error w-full text-left">
          Sign Out
        </button>
      </li>
    </ul>
  </div>
{:else}
  <div class="min-h-screen bg-[#282828] flex">
    <!-- Left Side - Form -->
    <div class="flex-1 flex flex-col justify-center px-8 lg:px-16 xl:px-24">
      <!-- Logo -->
      <div class="absolute top-6 left-6 lg:left-8">
        <Logo size="md" />
      </div>

      <!-- Form Container -->
      <div class="w-full max-w-md mx-auto">
        <h1 class="text-4xl font-bold text-white mb-2">
          {isSignUp ? 'Create Account' : 'Log in to Barracuda'}
        </h1>
        
        {#if !isSignUp}
          <p class="text-gray-400 mb-8">
            Don't have an account yet? 
            <button 
              class="text-[#FF6B6B] hover:text-[#FF5252] font-medium"
              on:click={() => isSignUp = true}
            >
              Sign up for free
            </button>
          </p>
        {:else}
          <p class="text-gray-400 mb-8">
            Already have an account? 
            <button 
              class="text-[#FF6B6B] hover:text-[#FF5252] font-medium"
              on:click={() => isSignUp = false}
            >
              Log in
            </button>
          </p>
        {/if}

        {#if error}
          <div class="bg-red-900/30 border border-red-700 text-red-200 px-4 py-3 rounded-lg mb-4">
            {error}
          </div>
        {/if}

        {#if success}
          <div class="bg-green-900/30 border border-green-700 text-green-200 px-4 py-3 rounded-lg mb-4">
            {success}
          </div>
        {/if}

        <form on:submit|preventDefault={handleSubmit} class="space-y-4">
          {#if isSignUp}
            <!-- First Name and Last Name side by side -->
            <div class="grid grid-cols-2 gap-4">
              <div>
                <input
                  type="text"
                  placeholder="First name"
                  class="w-full bg-[#282828] border border-gray-600 rounded-lg px-4 py-3 text-white placeholder-gray-400 focus:outline-none focus:border-[#FF6B6B] transition-colors"
                  bind:value={firstName}
                  required
                />
              </div>
              <div>
                <input
                  type="text"
                  placeholder="Last name"
                  class="w-full bg-[#282828] border border-gray-600 rounded-lg px-4 py-3 text-white placeholder-gray-400 focus:outline-none focus:border-[#FF6B6B] transition-colors"
                  bind:value={lastName}
                  required
                />
              </div>
            </div>
          {/if}

          <!-- Email -->
          <div>
            <input
              type="email"
              placeholder={isSignUp ? "Email address" : "name@work-email.com"}
              class="w-full bg-[#282828] border border-gray-600 rounded-lg px-4 py-3 text-white placeholder-gray-400 focus:outline-none focus:border-[#FF6B6B] transition-colors"
              bind:value={email}
              required
            />
          </div>

          <!-- Password -->
          <div class="relative">
            {#if showPassword}
              <input
                type="text"
                placeholder="Password"
                class="w-full bg-[#282828] border border-gray-600 rounded-lg px-4 py-3 pr-12 text-white placeholder-gray-400 focus:outline-none focus:border-[#FF6B6B] transition-colors"
                bind:value={password}
                required
                minlength="6"
              />
            {:else}
              <input
                type="password"
                placeholder="Password"
                class="w-full bg-[#282828] border border-gray-600 rounded-lg px-4 py-3 pr-12 text-white placeholder-gray-400 focus:outline-none focus:border-[#FF6B6B] transition-colors"
                bind:value={password}
                required
                minlength="6"
              />
            {/if}
            <button
              type="button"
              class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-white"
              on:click={() => showPassword = !showPassword}
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                {#if showPassword}
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.29 3.29m0 0A9.97 9.97 0 015 12c0 1.65.404 3.203 1.117 4.562M12 19c-1.65 0-3.203-.404-4.562-1.117M9.878 9.878L12 12m-2.122-2.122L9.88 9.88m2.242 2.242L14.12 14.12" />
                {:else}
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                {/if}
              </svg>
            </button>
          </div>

          {#if !isSignUp}
            <div class="flex items-center justify-between text-sm">
              <button type="button" class="text-[#FF6B6B] hover:text-[#FF5252]">
                Forgot password?
              </button>
            </div>
          {/if}

          <!-- Submit Button -->
          <button
            type="submit"
            class="w-full bg-[#FF6B6B] hover:bg-[#FF5252] text-white font-medium py-3 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            disabled={loading}
          >
            {#if loading}
              <span class="loading loading-spinner loading-sm"></span>
            {:else}
              {isSignUp ? 'Create Account' : 'Log in'}
            {/if}
          </button>
        </form>

        {#if isSignUp}
          <p class="text-sm text-gray-500 mt-6 text-center">
            By creating an account, I agree with Barracuda's 
            <button type="button" class="text-[#FF6B6B] hover:underline bg-transparent border-none p-0 cursor-pointer">Privacy Policy</button> 
            and 
            <button type="button" class="text-[#FF6B6B] hover:underline bg-transparent border-none p-0 cursor-pointer">Terms of Service</button>.
          </p>
        {/if}
      </div>
    </div>

    <!-- Right Side - Decorative Graphic -->
    <div class="hidden lg:flex flex-1 items-center justify-center bg-[#282828] relative overflow-hidden">
      <div class="w-full h-full flex items-center justify-center p-12">
        <!-- Decorative Illustration -->
        <svg 
          width="600" 
          height="600" 
          viewBox="0 0 600 600" 
          class="w-full h-full max-w-2xl"
          fill="none" 
          xmlns="http://www.w3.org/2000/svg"
        >
          <!-- Retro Computer Monitor -->
          <g opacity="0.8">
            <!-- Monitor -->
            <rect x="200" y="150" width="200" height="140" rx="4" stroke="white" stroke-width="2" fill="none"/>
            <rect x="210" y="160" width="180" height="120" fill="#282828"/>
            <!-- Screen Content -->
            <text x="220" y="200" font-family="monospace" font-size="12" fill="#00FF00">
              <tspan x="220" dy="15">const crawl = () =></tspan>
              <tspan x="220" dy="15">  return pages;</tspan>
            </text>
            <text x="220" y="240" font-family="monospace" font-size="12" fill="#FF0000">
              <tspan x="220" dy="15">// SEO Analysis</tspan>
            </text>
            <!-- Monitor Stand -->
            <rect x="280" y="290" width="40" height="20" rx="2" stroke="white" stroke-width="2" fill="none"/>
            <!-- Computer Tower -->
            <rect x="270" y="310" width="60" height="80" rx="2" stroke="white" stroke-width="2" fill="none"/>
            <circle cx="285" cy="330" r="3" fill="white"/>
            <circle cx="315" cy="330" r="3" fill="white"/>
          </g>

          <!-- Painter's Palette -->
          <g opacity="0.8">
            <path 
              d="M 200 400 Q 250 450 300 400 Q 350 450 400 400 Q 350 380 300 400 Q 250 380 200 400 Z" 
              stroke="white" 
              stroke-width="2" 
              fill="none"
            />
            <!-- Paint Blobs -->
            <circle cx="240" cy="410" r="12" fill="#FFB6C1"/>
            <circle cx="280" cy="405" r="12" fill="#FFD700"/>
            <circle cx="320" cy="410" r="12" fill="#87CEEB"/>
            <circle cx="260" cy="425" r="12" fill="#FF69B4"/>
            <circle cx="300" cy="430" r="12" fill="#FFA500"/>
          </g>

          <!-- Flying Birdhouses -->
          <g opacity="0.8">
            <!-- Birdhouse 1 -->
            <g transform="translate(120, 100)">
              <rect x="0" y="0" width="40" height="50" rx="2" stroke="#FF6B6B" stroke-width="2" fill="none"/>
              <circle cx="20" cy="15" r="8" stroke="white" stroke-width="2" fill="none"/>
              <rect x="18" y="20" width="4" height="8" fill="white"/>
              <!-- Wings -->
              <path d="M -10 25 Q -5 20 0 25" stroke="white" stroke-width="2" fill="none"/>
              <path d="M 50 25 Q 45 20 40 25" stroke="white" stroke-width="2" fill="none"/>
              <!-- Glow -->
              <circle cx="20" cy="60" r="8" fill="#FFD700" opacity="0.6"/>
            </g>
            
            <!-- Birdhouse 2 -->
            <g transform="translate(450, 120)">
              <rect x="0" y="0" width="40" height="50" rx="2" stroke="#FF6B6B" stroke-width="2" fill="none"/>
              <circle cx="20" cy="15" r="8" stroke="white" stroke-width="2" fill="none"/>
              <rect x="18" y="20" width="4" height="8" fill="white"/>
              <!-- Wings -->
              <path d="M -10 25 Q -5 20 0 25" stroke="white" stroke-width="2" fill="none"/>
              <path d="M 50 25 Q 45 20 40 25" stroke="white" stroke-width="2" fill="none"/>
              <!-- Glow -->
              <circle cx="20" cy="60" r="8" fill="#FFD700" opacity="0.6"/>
            </g>
          </g>
        </svg>
      </div>
    </div>
  </div>
{/if}
