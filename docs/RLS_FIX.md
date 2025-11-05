# Fixing Infinite Recursion in RLS Policies

## Problem

The `project_members` table had RLS policies that caused infinite recursion:
- Policies were checking the `project_members` table to verify membership
- This triggered the same policy check, causing infinite recursion

## Solution

Run the migration `20240321_fix_project_members_rls.sql` which:

1. **Drops the problematic policies**
2. **Creates new policies that avoid recursion** by:
   - Checking `projects.owner_id` directly instead of querying `project_members`
   - Allowing users to view their own membership record directly
   - Using the `projects` table as the source of truth for ownership

## How to Apply

### Option 1: Via Supabase Dashboard (Recommended)

1. Go to your Supabase project dashboard
2. Navigate to **SQL Editor**
3. Copy the contents of `supabase/migrations/20240321_fix_project_members_rls.sql`
4. Paste and run it

### Option 2: Via Supabase CLI

```bash
# If you have Supabase CLI set up locally
supabase db push
```

Or manually:

```bash
supabase migration new fix_project_members_rls
# Copy the SQL into the new migration file
supabase db push
```

## Verification

After applying the migration:

1. Try signing up for a new account
2. Create a project
3. The infinite recursion error should be gone
4. You should be able to view projects and project members

## What Changed

### Before (Problematic)
```sql
-- This caused recursion because it queries project_members
create policy "Project members can view project members"
  on public.project_members
  for select
  using (
    exists (
      select 1
      from public.project_members pm  -- ❌ Circular reference!
      where pm.project_id = project_members.project_id
        and pm.user_id = auth.uid()
    )
  );
```

### After (Fixed)
```sql
-- This avoids recursion by checking projects table directly
create policy "Project members can view project members"
  on public.project_members
  for select
  using (
    -- Check ownership via projects table (no recursion)
    exists (
      select 1
      from public.projects p  -- ✅ Direct check, no recursion
      where p.id = project_members.project_id
        and p.owner_id = auth.uid()
    )
    -- OR user viewing their own record
    or project_members.user_id = auth.uid()
  );
```

