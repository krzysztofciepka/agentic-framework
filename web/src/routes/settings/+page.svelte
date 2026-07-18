<script lang="ts">
  import { onMount } from 'svelte';
  import {
    getProviders, createProvider, updateProvider, deleteProvider,
    getTools, getSettings,
    type Provider, type Tool, type Setting,
  } from '$lib/api';

  let providers = $state<Provider[]>([]);
  let tools = $state<Tool[]>([]);
  let settings = $state<Setting[]>([]);

  let showProviderForm = $state(false);
  let editingProvId = $state<number | null>(null);
  let provName = $state('');
  let provBaseURL = $state('');
  let provAPIKey = $state('');

  let saving = $state(false);

  onMount(() => {
    loadProviders();
    loadTools();
    loadSettings();
  });

  async function loadProviders() {
    try { providers = await getProviders(); } catch (e) { console.error(e); }
  }

  async function loadTools() {
    try { tools = await getTools(); } catch (e) { console.error(e); }
  }

  async function loadSettings() {
    try { settings = await getSettings(); } catch (e) { console.error(e); }
  }

  function resetForm() {
    showProviderForm = false;
    editingProvId = null;
    provName = '';
    provBaseURL = '';
    provAPIKey = '';
  }

  function handleAdd() {
    resetForm();
    showProviderForm = true;
  }

  function handleEdit(prov: Provider) {
    showProviderForm = true;
    editingProvId = prov.id;
    provName = prov.name;
    provBaseURL = prov.base_url;
    provAPIKey = '';
  }

  function handleCancel() {
    resetForm();
  }

  async function handleSave() {
    if (!provName.trim() || !provBaseURL.trim() || saving) return;
    saving = true;
    try {
      if (editingProvId !== null) {
        await updateProvider(editingProvId, {
          name: provName.trim(),
          base_url: provBaseURL.trim(),
          api_key: provAPIKey,
        });
      } else {
        await createProvider({
          name: provName.trim(),
          base_url: provBaseURL.trim(),
          api_key: provAPIKey,
        });
      }
      await loadProviders();
      resetForm();
    } catch (e) {
      console.error(e);
    } finally {
      saving = false;
    }
  }

  async function handleDelete(prov: Provider) {
    if (!confirm(`Delete provider "${prov.name}"?`)) return;
    try {
      await deleteProvider(prov.id);
      providers = providers.filter(p => p.id !== prov.id);
    } catch (e) {
      console.error(e);
    }
  }
</script>

<section class="settings-section">
  <div class="section-header">
    <h2>Providers</h2>
    <button onclick={handleAdd} disabled={showProviderForm && editingProvId === null}>+ Add</button>
  </div>

  {#if showProviderForm}
    <div class="form-card">
      <div class="form-field">
        <label for="prov-name">Name</label>
        <input id="prov-name" value={provName} oninput={(e) => provName = e.target.value} placeholder="Provider name" />
      </div>
      <div class="form-field">
        <label for="prov-url">Base URL</label>
        <input id="prov-url" value={provBaseURL} oninput={(e) => provBaseURL = e.target.value} placeholder="https://api.example.com/v1" />
      </div>
      <div class="form-field">
        <label for="prov-key">API Key</label>
        <input id="prov-key" type="password" value={provAPIKey} oninput={(e) => provAPIKey = e.target.value} placeholder={editingProvId !== null ? '(unchanged if empty)' : 'API key'} />
      </div>
      <div class="form-actions">
        <button onclick={handleSave} disabled={saving}>
          {saving ? 'Saving...' : 'Save'}
        </button>
        <button class="btn-secondary" onclick={handleCancel}>Cancel</button>
      </div>
    </div>
  {/if}

  <div class="provider-list">
    {#each providers as prov (prov.id)}
      <div class="provider-card">
        <div class="provider-info">
          <span class="provider-name">{prov.name}</span>
          <span class="provider-url">{prov.base_url}</span>
        </div>
        <div class="provider-actions">
          <button class="btn-small" onclick={() => handleEdit(prov)}>Edit</button>
          <button class="btn-small btn-danger" onclick={() => handleDelete(prov)}>Delete</button>
        </div>
      </div>
    {:else}
      <p class="muted">No providers configured.</p>
    {/each}
  </div>
</section>

<section class="settings-section">
  <h2>Available Tools</h2>

  <div class="tools-grid">
    {#each tools as tool (tool.id)}
      <div class="tool-card">
        <div class="tool-header">
          <span class="tool-name">{tool.name}</span>
          <span class="tool-category">{tool.category}</span>
        </div>
        <p class="tool-description">{tool.description}</p>
      </div>
    {:else}
      <p class="muted">No tools available.</p>
    {/each}
  </div>
</section>

<style>
  .settings-section {
    margin-bottom: 2.5rem;
  }

  .section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 1rem;
  }

  h2 {
    margin: 0 0 1rem 0;
    font-size: 1.25rem;
    font-weight: 600;
  }

  .section-header h2 {
    margin: 0;
  }

  .form-card {
    background: var(--bg-surface);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 1.25rem;
    margin-bottom: 1.25rem;
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .form-field {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
  }

  .form-field label {
    font-size: 0.85rem;
    color: var(--text-muted);
  }

  .form-field input {
    width: 100%;
  }

  .form-actions {
    display: flex;
    gap: 0.5rem;
  }

  .btn-secondary {
    background: var(--bg-elevated);
    color: var(--text);
    border: 1px solid var(--border);
  }

  .btn-small {
    padding: 0.3rem 0.6rem;
    font-size: 0.8rem;
  }

  .btn-danger {
    background: var(--danger);
    color: #fff;
  }

  .provider-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .provider-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: var(--bg-surface);
    border: 1px solid var(--border);
    border-radius: 6px;
    padding: 0.75rem 1rem;
  }

  .provider-info {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
  }

  .provider-name {
    font-weight: 600;
    font-size: 0.95rem;
  }

  .provider-url {
    font-size: 0.8rem;
    color: var(--text-muted);
  }

  .provider-actions {
    display: flex;
    gap: 0.4rem;
  }

  .tools-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
    gap: 0.75rem;
  }

  .tool-card {
    background: var(--bg-surface);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .tool-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
  }

  .tool-name {
    font-weight: 600;
    font-size: 0.95rem;
  }

  .tool-category {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    background: var(--bg-elevated);
    color: var(--accent);
    padding: 0.15rem 0.5rem;
    border-radius: 12px;
    white-space: nowrap;
  }

  .tool-description {
    margin: 0;
    font-size: 0.85rem;
    color: var(--text-muted);
    line-height: 1.4;
  }

  .muted {
    color: var(--text-muted);
    font-size: 0.9rem;
  }
</style>
