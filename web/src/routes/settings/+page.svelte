<script lang="ts">
  import { onMount } from 'svelte';
  import {
    getProviders, createProvider, updateProvider, deleteProvider,
    getTools, type Provider, type Tool,
  } from '$lib/api';

  let providers = $state<Provider[]>([]);
  let tools = $state<Tool[]>([]);
  let showProviderForm = $state(false);
  let editingProvId = $state<number | null>(null);
  let provName = $state('');
  let provBaseURL = $state('');
  let provAPIKey = $state('');
  let saving = $state(false);

  onMount(() => { loadProviders(); loadTools(); });

  async function loadProviders() { try { providers = await getProviders(); } catch (e) { console.error(e); } }
  async function loadTools() { try { tools = await getTools(); } catch (e) { console.error(e); } }

  function resetForm() {
    showProviderForm = false; editingProvId = null;
    provName = ''; provBaseURL = ''; provAPIKey = '';
  }

  function handleAdd() { resetForm(); showProviderForm = true; }
  function handleEdit(prov: Provider) {
    showProviderForm = true; editingProvId = prov.id;
    provName = prov.name; provBaseURL = prov.base_url; provAPIKey = '';
  }

  async function handleSave() {
    if (!provName.trim() || !provBaseURL.trim() || saving) return;
    saving = true;
    try {
      if (editingProvId !== null) {
        await updateProvider(editingProvId, { name: provName.trim(), base_url: provBaseURL.trim(), api_key: provAPIKey });
      } else {
        await createProvider({ name: provName.trim(), base_url: provBaseURL.trim(), api_key: provAPIKey });
      }
      await loadProviders(); resetForm();
    } catch (e) { console.error(e); }
    finally { saving = false; }
  }

  async function handleDelete(prov: Provider) {
    if (!confirm(`Delete "${prov.name}"?`)) return;
    try { await deleteProvider(prov.id); providers = providers.filter(p => p.id !== prov.id); }
    catch (e) { console.error(e); }
  }
</script>

<div class="page">
  <section class="section">
    <div class="section-header">
      <h2>[+] Providers</h2>
      <button onclick={handleAdd} disabled={showProviderForm && editingProvId === null}>[+] Add</button>
    </div>

    {#if showProviderForm}
      <div class="card form-card">
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
          <button onclick={handleSave} disabled={saving}>{saving ? '[...]' : 'Save'}</button>
          <button class="btn-secondary" onclick={resetForm}>Cancel</button>
        </div>
      </div>
    {/if}

    <div class="provider-list">
      {#each providers as prov (prov.id)}
        <div class="card provider-card">
          <div>
            <div class="provider-name">{prov.name}</div>
            <div class="provider-url">{prov.base_url}</div>
          </div>
          <div class="provider-actions">
            <button class="btn-secondary" onclick={() => handleEdit(prov)}>Edit</button>
            <button class="btn-danger" onclick={() => handleDelete(prov)}>Delete</button>
          </div>
        </div>
      {:else}
        <p class="muted">[-] No providers configured.</p>
      {/each}
    </div>
  </section>

  <section class="section">
    <h2>[+] Available Tools</h2>
    <div class="tools-grid">
      {#each tools as tool (tool.id)}
        <div class="card tool-card">
          <div class="tool-header">
            <span class="tool-name">{tool.name}</span>
            <span class="tool-category">[{tool.category}]</span>
          </div>
          <p class="tool-description">{tool.description}</p>
        </div>
      {:else}
        <p class="muted">[-] No tools available.</p>
      {/each}
    </div>
  </section>
</div>

<style>
  .page { max-width: 800px; }
  .section { margin-bottom: 96px; }
  .section-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px; }
  h2 { margin: 0 0 12px 0; font-size: 16px; font-weight: 700; color: var(--ink); }
  .section-header h2 { margin: 0; }

  .card {
    background: var(--canvas);
    border: 1px solid var(--hairline);
    border-radius: 0;
    padding: 16px;
    margin-bottom: 8px;
  }
  .form-card { display: flex; flex-direction: column; gap: 12px; }
  .form-field { display: flex; flex-direction: column; gap: 4px; }
  .form-field label { font-size: 14px; color: var(--mute); line-height: 2; }
  .form-actions { display: flex; gap: 8px; }

  .btn-secondary {
    background: var(--canvas); color: var(--ink);
    border: 1px solid var(--hairline-strong); border-radius: var(--rounded-sm);
    padding: 4px 20px; height: 36px; font-family: inherit;
    font-size: 16px; font-weight: 500; line-height: 2; cursor: pointer;
  }
  .btn-secondary:active { background: var(--surface-card); }
  .btn-secondary:disabled { background: var(--surface-card); color: var(--ash); }
  .btn-danger { background: var(--danger); color: #fff; border-radius: var(--rounded-sm); }
  .btn-danger:active { background: var(--danger-active); }
  .btn-danger:disabled { background: var(--surface-card); color: var(--ash); }

  .provider-list { display: flex; flex-direction: column; }
  .provider-card { display: flex; align-items: center; justify-content: space-between; }
  .provider-name { font-size: 16px; font-weight: 500; color: var(--ink); }
  .provider-url { font-size: 14px; color: var(--mute); margin-top: 2px; }
  .provider-actions { display: flex; gap: 8px; }

  .tools-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); gap: 8px; }
  .tool-card { display: flex; flex-direction: column; gap: 8px; }
  .tool-header { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
  .tool-name { font-size: 16px; font-weight: 500; color: var(--ink); }
  .tool-category { font-size: 14px; color: var(--mute); }
  .tool-description { margin: 0; font-size: 16px; color: var(--body-c); line-height: 1.5; }
  .muted { color: var(--mute); font-size: 16px; }
</style>
