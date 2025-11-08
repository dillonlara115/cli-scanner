import { supabase } from './supabase.js';

export const getApiUrl = () => import.meta.env.VITE_CLOUD_RUN_API_URL || 'http://localhost:8080';

async function authorizedRequest(path, { method = 'GET', body, headers = {} } = {}) {
  const { data: { session } } = await supabase.auth.getSession();
  if (!session) {
    throw new Error('Not authenticated');
  }

  const requestHeaders = new Headers(headers);
  requestHeaders.set('Authorization', `Bearer ${session.access_token}`);

  let requestBody = body;
  if (body && !(body instanceof FormData) && typeof body === 'object' && !(body instanceof Blob)) {
    requestHeaders.set('Content-Type', 'application/json');
    requestBody = JSON.stringify(body);
  }

  const response = await fetch(`${getApiUrl()}${path}`, {
    method,
    headers: requestHeaders,
    body: requestBody,
  });

  return response;
}

async function authorizedJSON(path, options = {}) {
  try {
    const response = await authorizedRequest(path, options);

    if (!response.ok) {
      let message = `Request failed with status ${response.status}`;
      try {
        const errorPayload = await response.json();
        message = errorPayload.error || errorPayload.message || message;
      } catch (_) {
        // Ignore JSON parse errors
      }
      throw new Error(message);
    }

    if (response.status === 204) {
      return { data: null, error: null };
    }

    const data = await response.json();
    return { data, error: null };
  } catch (error) {
    console.error('API request failed:', error);
    return { data: null, error };
  }
}

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

// Update project settings
export async function updateProjectSettings(projectId, settings) {
  try {
    const { data: { user } } = await supabase.auth.getUser();
    if (!user) throw new Error('Not authenticated');

    // Get current project to merge settings
    const { data: project, error: fetchError } = await supabase
      .from('projects')
      .select('settings')
      .eq('id', projectId)
      .single();

    if (fetchError) throw fetchError;

    // Merge with existing settings
    const updatedSettings = {
      ...(project.settings || {}),
      ...settings
    };

    const { data, error } = await supabase
      .from('projects')
      .update({ settings: updatedSettings })
      .eq('id', projectId)
      .select()
      .single();

    if (error) throw error;
    return { data, error: null };
  } catch (error) {
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

export async function fetchProjectGSCStatus(projectId) {
  if (!projectId) return { data: null, error: new Error('projectId is required') };
  return authorizedJSON(`/api/v1/projects/${projectId}/gsc/status`);
}

export async function fetchProjectGSCProperties(projectId) {
  if (!projectId) return { data: null, error: new Error('projectId is required') };
  return authorizedJSON(`/api/v1/projects/${projectId}/gsc/properties`);
}

export async function updateProjectGSCProperty(projectId, propertyUrl, propertyType = null) {
  if (!projectId) return { data: null, error: new Error('projectId is required') };
  if (!propertyUrl) return { data: null, error: new Error('propertyUrl is required') };
  return authorizedJSON(`/api/v1/projects/${projectId}/gsc/property`, {
    method: 'POST',
    body: {
      property_url: propertyUrl,
      property_type: propertyType,
    },
  });
}

export async function triggerProjectGSCSync(projectId, options = {}) {
  if (!projectId) return { data: null, error: new Error('projectId is required') };
  return authorizedJSON(`/api/v1/projects/${projectId}/gsc/trigger-sync`, {
    method: 'POST',
    body: options,
  });
}

export async function fetchProjectGSCDimensions(projectId, type, params = {}) {
  if (!projectId) return { data: null, error: new Error('projectId is required') };
  if (!type) return { data: null, error: new Error('type is required') };

  const searchParams = new URLSearchParams({ type });
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') {
      searchParams.set(key, value.toString());
    }
  });

  return authorizedJSON(`/api/v1/projects/${projectId}/gsc/dimensions?${searchParams.toString()}`);
}

// Fetch link graph for a crawl
export async function fetchCrawlGraph(crawlId) {
  if (!crawlId) return { data: null, error: new Error('crawlId is required') };
  return authorizedJSON(`/api/v1/crawls/${crawlId}/graph`);
}
