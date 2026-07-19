<script lang="ts">
  import { onMount } from 'svelte';
  import {
    getAgents, createAgent, updateAgent, deleteAgent,
    getProviders, getTools, getSettings,
    type Agent, type Provider, type Tool,
  } from '$lib/api';

  let agents = $state<Agent[]>([]);
  let providers = $state<Provider[]>([]);
  let tools = $state<Tool[]>([]);
  let modelSuggestions = $state<string[]>([]);
  let loading = $state(true);
  let submitting = $state(false);
  let showForm = $state(false);
  let editingId = $state<number | null>(null);

  let formName = $state('');
  let formSystemPrompt = $state('');
  let formProviderId = $state(0);
  let formModel = $state('deepseek-v4-pro');
  let formTemperature = $state(0.7);
  let formMaxTokens = $state(4096);
  let formToolIds = $state<number[]>([]);

  async function loadModels() {
    try {
      const settings = await getSettings();
      const raw = settings.find(s => s.key === 'openode_go_models')?.value || '';
      modelSuggestions = raw.split(',').map(s => s.trim()).filter(Boolean);
    } catch (_) {}
  }

  async function loadData() {
    loading = true;
    try {
      const [a, p, t] = await Promise.all([getAgents(), getProviders(), getTools()]);
      agents = a; providers = p; tools = t;
      loadModels();
    } catch (e) { console.error(e); }
    finally { loading = false; }
  }

  function resetForm() {
    formName = ''; formSystemPrompt = ''; formProviderId = 0;
    formModel = 'deepseek-v4-pro'; formTemperature = 0.7; formMaxTokens = 4096;
    formToolIds = []; editingId = null; showForm = false;
  }

  function startCreate() { resetForm(); showForm = true; }

  function startEdit(agent: Agent) {
    formName = agent.name; formSystemPrompt = agent.system_prompt;
    formProviderId = agent.provider_id; formModel = agent.model;
    formTemperature = agent.temperature; formMaxTokens = agent.max_tokens;
    formToolIds = agent.tools.map(t => t.id);
    editingId = agent.id; showForm = true;
  }

  async function handleSubmit() {
    submitting = true;
    try {
      const data = {
        name: formName, system_prompt: formSystemPrompt,
        provider_id: formProviderId, model: formModel,
        temperature: formTemperature, max_tokens: formMaxTokens,
        tool_ids: formToolIds,
      };
      if (editingId != null) {
        await updateAgent(editingId, data as Partial<Agent>);
      } else {
        await createAgent(data as Partial<Agent> & { name: string; system_prompt: string; model: string });
      }
      resetForm(); await loadData();
    } catch (e) { console.error(e); }
    finally { submitting = false; }
  }

  async function handleDelete(agent: Agent) {
    if (confirm(`Delete agent "${agent.name}"?`)) {
      try { await deleteAgent(agent.id); await loadData(); }
      catch (e) { console.error(e); }
    }
  }

  function toggleTool(id: number) {
    if (formToolIds.includes(id)) formToolIds = formToolIds.filter(tid => tid !== id);
    else formToolIds = [...formToolIds, id];
  }

  onMount(() => { loadData(); });
</script>

<div class="page">
  <div class="page-header">
    <h1>Agents</h1>
    <button onclick={startCreate} disabled={showForm}>[+] New Agent</button>
  </div>

  {#if showForm}
    <div class="card form-card">
      <h2>{editingId != null ? '[x] Edit Agent' : '[+] Create Agent'}</h2>
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
            <option value={0} disabled>[-] Select a provider</option>
            {#each providers as p}
              <option value={p.id}>{p.name}</option>
            {/each}
          </select>
        </div>
        <div class="field">
          <label for="model">Model</label>
          <input id="model" type="text" required placeholder="deepseek-v4-pro" value={formModel} oninput={(e) => formModel = e.target.value} list="model-list" />
          <datalist id="model-list">
            {#each modelSuggestions as m}
              <option value={m} />
            {/each}
          </datalist>
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
          </div>
        </div>
        <div class="form-actions">
          <button type="submit" disabled={submitting || !formName || !formSystemPrompt || !formModel}>
            {editingId != null ? 'Update' : 'Create'}
          </button>
          <button type="button" class="btn-secondary" onclick={resetForm}>Cancel</button>
        </div>
      </form>
    </div>
  {/if}

  {#if loading}
    <p class="muted">[-] Loading...</p>
  {:else if agents.length === 0}
    <p class="muted">[-] No agents configured.</p>
  {:else}
    <div class="agent-list">
      {#each agents as agent (agent.id)}
        <div class="card agent-card">
          <div>
            <div class="agent-name">{agent.name}</div>
            <div class="agent-meta">{agent.model}<span class="sep">|</span>{agent.tools.length} tool(s)</div>
          </div>
          <div class="agent-actions">
            <button class="btn-secondary" onclick={() => startEdit(agent)}>Edit</button>
            <button class="btn-danger" onclick={() => handleDelete(agent)}>Delete</button>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .page { max-width: 800px; }
  .page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; }
  .page-header h1 { margin: 0; font-size: 16px; font-weight: 700; color: var(--ink); }

  .card {
    background: var(--canvas);
    border: 1px solid var(--hairline);
    border-radius: 0;
    padding: 16px;
    margin-bottom: 8px;
  }
  .form-card h2 { margin: 0 0 16px 0; font-size: 16px; font-weight: 700; color: var(--ink); }
  .field { margin-bottom: 12px; display: flex; flex-direction: column; gap: 4px; }
  .field label, .field-label { font-size: 14px; font-weight: 400; color: var(--mute); line-height: 2; }
  select { width: 100%; }
  .tool-checkboxes {
    display: flex; flex-direction: column; gap: 4px;
    max-height: 200px; overflow-y: auto;
    padding: 8px;
    background: var(--surface-soft);
    border: 1px solid var(--hairline);
    border-radius: var(--rounded-sm);
  }
  .tool-checkbox {
    display: flex; align-items: flex-start; gap: 8px;
    cursor: pointer; font-size: 16px; padding: 4px 0;
    color: var(--ink);
  }
  .tool-label { display: flex; flex-direction: column; gap: 2px; }
  .tool-label strong { font-weight: 500; }
  .tool-label small { color: var(--mute); font-size: 14px; font-weight: 400; }
  .form-actions { display: flex; gap: 8px; margin-top: 16px; }

  .btn-secondary {
    background: var(--canvas);
    color: var(--ink);
    border: 1px solid var(--hairline-strong);
    border-radius: var(--rounded-sm);
    padding: 4px 20px;
    height: 36px;
    font-family: inherit;
    font-size: 16px;
    font-weight: 500;
    line-height: 2;
    cursor: pointer;
  }
  .btn-secondary:active { background: var(--surface-card); }
  .btn-danger { background: var(--danger); color: #fff; border-radius: var(--rounded-sm); }
  .btn-danger:active { background: var(--danger-active); }
  .btn-danger:disabled { background: var(--surface-card); color: var(--ash); }

  .agent-list { display: flex; flex-direction: column; }
  .agent-card { display: flex; align-items: center; justify-content: space-between; gap: 16px; }
  .agent-name { font-size: 16px; font-weight: 500; color: var(--ink); }
  .agent-meta { font-size: 14px; color: var(--mute); margin-top: 2px; }
  .sep { margin: 0 8px; color: var(--hairline-strong); }
  .agent-actions { display: flex; gap: 8px; flex-shrink: 0; }
  .muted { color: var(--mute); font-size: 16px; }
</style>
