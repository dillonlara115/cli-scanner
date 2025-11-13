-- Add INSERT policy for profiles table
-- Allows users to create their own profile when they first sign up

create policy "Users can create their own profile"
  on public.profiles
  for insert
  with check (auth.uid() = id);


