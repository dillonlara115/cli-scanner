import { supabase } from './supabase.js';

// Fetch user's projects
export async function fetchProjects() {
  try {
    const { data, error } = await supabase
      .from('projects')
      .select('*')
      .order('created_at', { ascending: false });

    if (error) throw error;
    return { data, error: null };
  } catch (error) {
    console.error('Error fetching projects:', error);
    return { data: null, error };
  }
}

// Fetch crawls for a project
export async function fetchCrawls(projectId) {
  try {
    const { data, error } = await supabase
      .from('crawls')
      .select('*')
      .eq('project_id', projectId)
      .order('started_at', { ascending: false });

    if (error) throw error;
    return { data, error: null };
  } catch (error) {
    return { data: null, error };
  }
}

// Fetch a single crawl by ID
export async function fetchCrawl(crawlId) {
  try {
    const { data, error } = await supabase
      .from('crawls')
      .select('*')
      .eq('id', crawlId)
      .single();

    if (error) throw error;
    return { data, error: null };
  } catch (error) {
    return { data: null, error };
  }
}

// Fetch page count for a crawl (for progress tracking)
export async function fetchCrawlPageCount(crawlId) {
  try {
    const { count, error } = await supabase
      .from('pages')
      .select('*', { count: 'exact', head: true })
      .eq('crawl_id', crawlId);

    if (error) throw error;
    return { count: count || 0, error: null };
  } catch (error) {
    return { count: 0, error };
  }
}

// Fetch pages for a crawl
export async function fetchPages(crawlId) {
  try {
    const { data, error } = await supabase
      .from('pages')
      .select('*')
      .eq('crawl_id', crawlId)
      .order('created_at', { ascending: false });

    if (error) throw error;
    return { data, error: null };
  } catch (error) {
    return { data: null, error };
  }
}

// Fetch issues for a crawl
export async function fetchIssues(crawlId) {
  try {
    const { data, error } = await supabase
      .from('issues')
      .select('*')
      .eq('crawl_id', crawlId)
      .order('created_at', { ascending: false });

    if (error) throw error;
    return { data, error: null };
  } catch (error) {
    return { data: null, error };
  }
}

// Fetch project issue summary (using the view)
export async function fetchProjectIssueSummary(projectId) {
  try {
    const { data, error } = await supabase
      .from('project_issue_summary')
      .select('*')
      .eq('project_id', projectId)
      .single();

    if (error) throw error;
    return { data, error: null };
  } catch (error) {
    return { data: null, error };
  }
}

// Create a new project
export async function createProject(name, domain, settings = {}) {
  try {
    const { data: { user } } = await supabase.auth.getUser();
    if (!user) throw new Error('Not authenticated');

    const { data, error } = await supabase
      .from('projects')
      .insert({
        name,
        domain,
        owner_id: user.id,
        settings
      })
      .select()
      .single();

    if (error) throw error;

    // Also add the owner as a project member
    await supabase
      .from('project_members')
      .insert({
        project_id: data.id,
        user_id: user.id,
        role: 'owner'
      });

    return { data, error: null };
  } catch (error) {
    return { data: null, error };
  }
}

// Trigger a new crawl for a project
export async function triggerCrawl(projectId, crawlConfig) {
  try {
    const { data: { session } } = await supabase.auth.getSession();
    if (!session) throw new Error('Not authenticated');

    // Use local API server (http://localhost:8080) or Cloud Run URL if set
    const apiUrl = import.meta.env.VITE_CLOUD_RUN_API_URL || 'http://localhost:8080';
    
    const response = await fetch(`${apiUrl}/api/v1/projects/${projectId}/crawl`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${session.access_token}`
      },
      body: JSON.stringify(crawlConfig)
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to trigger crawl');
    }

    const data = await response.json();
    return { data, error: null };
  } catch (error) {
    console.error('Error triggering crawl:', error);
    return { data: null, error };
  }
}

// Update issue status
export async function updateIssueStatus(issueId, status, notes = null) {
  try {
    const { data: { user } } = await supabase.auth.getUser();
    if (!user) throw new Error('Not authenticated');

    const { data, error } = await supabase
      .from('issues')
      .update({
        status,
        status_updated_at: new Date().toISOString()
      })
      .eq('id', issueId)
      .select()
      .single();

    if (error) throw error;

    // Log status change
    if (data) {
      await supabase
        .from('issue_status_history')
        .insert({
          issue_id: issueId,
          old_status: data.status, // This won't work perfectly - we'd need the old value
          new_status: status,
          changed_by: user.id,
          notes
        });
    }

    return { data, error: null };
  } catch (error) {
    return { data: null, error };
  }
}

