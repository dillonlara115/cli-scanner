-- Fix infinite recursion in project_members RLS policies
-- The issue: policies were checking project_members table, which triggers the same policy check, causing recursion

-- Create a SECURITY DEFINER function to check membership without RLS recursion
-- This function bypasses RLS to check if a user is a project member
create or replace function public.is_project_member(project_uuid uuid, user_uuid uuid)
returns boolean
language plpgsql
security definer
set search_path = public
as $$
begin
  return exists (
    select 1
    from public.project_members pm
    where pm.project_id = project_uuid
      and pm.user_id = user_uuid
  )
  or exists (
    select 1
    from public.projects p
    where p.id = project_uuid
      and p.owner_id = user_uuid
  );
end;
$$;

-- Drop existing policies
drop policy if exists "Project members can view project members" on public.project_members;
drop policy if exists "Project owners can manage members" on public.project_members;

-- New SELECT policy: Use SECURITY DEFINER function to avoid recursion
-- This allows members to see all members without causing recursion
create policy "Project members can view project members"
  on public.project_members
  for select
  using (
    public.is_project_member(project_members.project_id, auth.uid())
  );

-- New INSERT policy: Only project owners can add members
-- Check ownership via projects table directly (no recursion)
create policy "Project owners can add members"
  on public.project_members
  for insert
  with check (
    exists (
      select 1
      from public.projects p
      where p.id = project_members.project_id
        and p.owner_id = auth.uid()
    )
  );

-- New UPDATE policy: Only project owners can update members
-- Check ownership via projects table directly (no recursion)
create policy "Project owners can update members"
  on public.project_members
  for update
  using (
    exists (
      select 1
      from public.projects p
      where p.id = project_members.project_id
        and p.owner_id = auth.uid()
    )
  );

-- New DELETE policy: Project owners can remove members, or users can remove themselves
create policy "Project owners can remove members"
  on public.project_members
  for delete
  using (
    -- User is the project owner
    exists (
      select 1
      from public.projects p
      where p.id = project_members.project_id
        and p.owner_id = auth.uid()
    )
    -- OR user is removing themselves
    or project_members.user_id = auth.uid()
  );

