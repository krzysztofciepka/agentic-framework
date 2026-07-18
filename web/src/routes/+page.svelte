<script lang="ts">
  import { onMount } from 'svelte';
  import {
    getAgents, createAgent, updateAgent, deleteAgent,
    getProviders, getTools,
    type Agent, type Provider, type Tool,
  } from '$lib/api';

  let agents = $state<Agent[]>([]);
  let providers = $state<Provider[]>([]);
  let tools = $state<Tool[]>([]);
  let loading = $state(true);
  let submitting = $state(false);
  let showForm = $state(false);
  let editingId = $state<number | null>(null);

  let formName = $state('');
  let formSystemPrompt = $state('');
  let formProviderId = $state(0);
  let formModel = $state('gpt-4o');
  let formTemperature = $state(0.7);
  let formMaxTokens = $state(4096);
  let formToolIds = $state<number[]>([]);

  async function loadData() {
    loading = true;
    try {
      const [a, p, t] = await Promise.all([getAgents(), getProviders(), getTools()]);
      agents = a;
      providers = p;
      tools = t;
    } catch (e) {
      console.error(e);
    } finally {
      loading = false;
    }
  }

  function resetForm() {
    formName = '';
    formSystemPrompt = '';
    formProviderId = 0;
    formModel = 'gpt-4o';
    formTemperature = 0.7;
    formMaxTokens = 4096;
    formToolIds = [];
    editingId = null;
    showForm = false;
  }

  function startCreate() {
    resetForm();
    showForm = true;
  }

  function startEdit(agent: Agent) {
    formName = agent.name;
    formSystemPrompt = agent.system_prompt;
    formProviderId = agent.provider_id;
    formModel = agent.model;
    formTemperature = agent.temperature;
    formMaxTokens = agent.max_tokens;
    formToolIds = agent.tools.map(t => t.id);
    editingId = agent.id;
    showForm = true;
  }

  async function handleSubmit() {
    submitting = true;
    try {
      const data = {
        name: formName,
        system_prompt: formSystemPrompt,
        provider_id: formProviderId,
        model: formModel,
        temperature: formTemperature,
        max_tokens: formMaxTokens,
        tool_ids: formToolIds,
      };
      if (editingId != null) {
        await updateAgent(editingId, data as Partial<Agent>);
      } else {
        await createAgent(data as Partial<Agent> & { name: string; system_prompt: string; model: string });
      }
      resetForm();
      await loadData();
    } catch (e) {
      console.error(e);
    } finally {
      submitting = false;
    }
  }

  async function handleDelete(agent: Agent) {
    if (confirm(`Delete agent "${agent.name}"?`)) {
      try {
        await deleteAgent(agent.id);
        await loadData();
      } catch (e) {
        console.error(e);
      }
    }
  }

  function toggleTool(id: number) {
    if (formToolIds.includes(id)) {
      formToolIds = formToolIds.filter(tid => tid !== id);
    } else {
      formToolIds = [...formToolIds, id];
    }
  }

  onMount(() => {
    loadData();
  });
</script>

<div class="agents-page">
  <header class="page-header">
    <h1>Agents</h1>
    <button onclick={startCreate} disabled={showForm}>+ New Agent</button>
  </header>

  {#if showForm}
    <div class="card form-card">
      <h2>{editingId != null ? 'Edit Agent' : 'Create Agent'}</h2>
      <form onsubmit={(e: SubmitEvent) => { e.preventDefault(); handleSubmit(); }}>
        <div class="field">
          <label for="name">Name</label>
          <input id="name" type="text" required value={formName} oninput={(e) => formName = e.target.value} />
        </div>

        <div class="field">
          <label for="system_prompt">System Prompt</label>
          <textarea id="system_prompt" required value={formSystemPrompt} oninput={(e) => formSystemPrompt = e.target.value}></textarea>
        </div>

        <div class="field">
          <label for="provider">Provider</label>
          <select id="provider" required value={formProviderId} onchange={(e) => formProviderId = parseInt(e.target.value)}>
            <option value={0} disabled>Select a provider</option>
            {#each providers as p}
              <option value={p.id}>{p.name}</option>
            {/each}
          </select>
        </div>

        <div class="field">
          <label for="model">Model</label>
          <input id="model" type="text" required placeholder="gpt-4o" value={formModel} oninput={(e) => formModel = e.target.value} />
        </div>

        <div class="field">
          <label for="temperature">Temperature: {formTemperature.toFixed(1)}</label>
          <input id="temperature" type="range" min="0" max="2" step="0.1" value={formTemperature} oninput={(e) => formTemperature = parseFloat(e.target.value)} />
        </div>

        <div class="field">
          <label for="max_tokens">Max Tokens</label>
          <input id="max_tokens" type="number" min="100" max="128000" value={formMaxTokens} oninput={(e) => formMaxTokens = parseInt(e.target.value)} />
        </div>

        <div class="field">
          <span class="field-label">Tools</span>
          <div class="tool-checkboxes">
            {#each tools as tool}
              <label class="tool-checkbox">
                <input type="checkbox" checked={formToolIds.includes(tool.id)} onchange={() => toggleTool(tool.id)} />
                <span class="tool-label">
                  <strong>{tool.name}</strong>
                  <small>{tool.description}</small>
                </span>
              </label>
            {/each}
            {#if tools.length === 0}
              <p class="muted">No tools available.</p>
            {/if}
          </div>
        </div>

        <div class="form-actions">
          <button type="submit" disabled={submitting || !formName || !formSystemPrompt || !formModel}>
            {editingId != null ? 'Update' : 'Create'}
          </button>
          <button type="button" class="cancel" onclick={resetForm}>Cancel</button>
        </div>
      </form>
    </div>
  {/if}

  <div class="agent-list">
    {#if loading}
      <p class="muted">Loading agents...</p>
    {:else if agents.length === 0}
      <p class="muted">No agents configured.</p>
    {:else}
      {#each agents as agent (agent.id)}
        <div class="card agent-card">
          <div class="agent-info">
            <h3>{agent.name}</h3>
            <div class="agent-meta">
              <span>Model: {agent.model}</span>
              <span>Tools: {agent.tools.length}</span>
            </div>
          </div>
          <div class="agent-actions">
            <button class="secondary" onclick={() => startEdit(agent)}>Edit</button>
            <button class="danger" onclick={() => handleDelete(agent)}>Delete</button>
          </div>
        </div>
      {/each}
    {/if}
  </div>
</div>

<style>
  .agents-page {
    max-width: 800px;
  }

  .page-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 1.5rem;
  }

  .page-header h1 {
    margin: 0;
    font-size: 1.5rem;
  }

  .card {
    background: var(--bg-surface);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 1.25rem 1.5rem;
    margin-bottom: 1rem;
  }

  .form-card h2 {
    margin: 0 0 1.25rem 0;
    font-size: 1.1rem;
  }

  .field {
    margin-bottom: 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
  }

  .field label, .field .field-label {
    font-size: 0.85rem;
    color: var(--text-muted);
  }

  select {
    width: 100%;
  }

  .tool-checkboxes {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    max-height: 200px;
    overflow-y: auto;
    padding: 0.5rem;
    background: var(--bg);
    border: 1px solid var(--border);
    border-radius: 6px;
  }

  .tool-checkbox {
    display: flex;
    align-items: flex-start;
    gap: 0.5rem;
    cursor: pointer;
    font-size: 0.85rem;
    padding: 0.25rem 0;
  }

  .tool-checkbox input[type="checkbox"] {
    margin-top: 0.15rem;
    accent-color: var(--accent);
  }

  .tool-label {
    display: flex;
    flex-direction: column;
    gap: 0.1rem;
  }

  .tool-label small {
    color: var(--text-muted);
    font-size: 0.8rem;
  }

  .form-actions {
    display: flex;
    gap: 0.75rem;
    margin-top: 1.5rem;
  }

  .agent-list {
    display: flex;
    flex-direction: column;
  }

  .agent-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
  }

  .agent-info h3 {
    margin: 0 0 0.35rem 0;
    font-size: 1rem;
  }

  .agent-meta {
    display: flex;
    gap: 1.25rem;
    font-size: 0.85rem;
    color: var(--text-muted);
  }

  .agent-actions {
    display: flex;
    gap: 0.5rem;
    flex-shrink: 0;
  }

  button.secondary {
    background: var(--bg-elevated);
    color: var(--text);
  }

  button.danger {
    background: var(--danger);
    color: #fff;
  }

  button.cancel {
    background: var(--bg-elevated);
    color: var(--text);
  }

  .muted {
    color: var(--text-muted);
    font-size: 0.9rem;
  }
</style>
