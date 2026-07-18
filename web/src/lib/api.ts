const BASE = '/api';

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  });
  if (res.status === 204) return undefined as T;
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Request failed');
  return data;
}

// Types
export interface Provider {
  id: number;
  name: string;
  base_url: string;
  created_at: string;
}

export interface Tool {
  id: number;
  name: string;
  description: string;
  category: string;
  parameters?: unknown;
}

export interface Agent {
  id: number;
  name: string;
  system_prompt: string;
  provider_id: number;
  model: string;
  temperature: number;
  max_tokens: number;
  tools: Tool[];
  created_at: string;
  updated_at: string;
}

export interface Conversation {
  id: number;
  agent_id: number;
  title: string;
  created_at: string;
  updated_at: string;
}

export interface Message {
  id: number;
  conversation_id: number;
  role: string;
  content: string;
  tool_call_id?: string;
  tool_name?: string;
  created_at: string;
}

export interface Setting {
  key: string;
  value: string;
}

// Providers
export const getProviders = () => request<Provider[]>('/providers');
export const createProvider = (data: { name: string; base_url: string; api_key: string }) =>
  request<Provider>('/providers', { method: 'POST', body: JSON.stringify(data) });
export const updateProvider = (id: number, data: { name: string; base_url: string; api_key: string }) =>
  request('/providers/' + id, { method: 'PUT', body: JSON.stringify(data) });
export const deleteProvider = (id: number) =>
  request('/providers/' + id, { method: 'DELETE' });

// Agents
export const getAgents = () => request<Agent[]>('/agents');
export const getAgent = (id: number) => request<Agent>('/agents/' + id);
export const createAgent = (data: Partial<Agent> & { name: string; system_prompt: string; model: string }) =>
  request<Agent>('/agents', { method: 'POST', body: JSON.stringify(data) });
export const updateAgent = (id: number, data: Partial<Agent>) =>
  request('/agents/' + id, { method: 'PUT', body: JSON.stringify(data) });
export const deleteAgent = (id: number) =>
  request('/agents/' + id, { method: 'DELETE' });

// Tools
export const getTools = () => request<Tool[]>('/tools');

// Conversations
export const getConversations = (agentId: number) =>
  request<Conversation[]>(`/conversations/agents/${agentId}`);
export const createConversation = (agentId: number, title?: string) =>
  request<Conversation>(`/conversations/agents/${agentId}`, { method: 'POST', body: JSON.stringify({ title }) });
export const getConversation = (id: number) =>
  request<Conversation & { messages: Message[] }>('/conversations/' + id);
export const deleteConversation = (id: number) =>
  request('/conversations/' + id, { method: 'DELETE' });

// Messages
export const getMessages = (conversationId: number) =>
  request<Message[]>('/conversations/' + conversationId + '/messages');
export const sendMessage = (conversationId: number, content: string) =>
  request<{ role: string; content: string }>(`/conversations/${conversationId}/messages`, {
    method: 'POST',
    body: JSON.stringify({ role: 'user', content }),
  });

// Settings
export const getSettings = () => request<Setting[]>('/settings');
export const updateSettings = (data: Record<string, string>) =>
  request('/settings', { method: 'PUT', body: JSON.stringify(data) });
