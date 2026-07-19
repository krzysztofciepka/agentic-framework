<script lang="ts">
  import { onMount } from 'svelte';
  import {
    getAgents, getConversations, getConversation,
    createConversation, sendMessage,
    type Agent, type Conversation, type Message,
  } from '$lib/api';

  let agents = $state<Agent[]>([]);
  let selectedAgent = $state<Agent | null>(null);
  let conversations = $state<Conversation[]>([]);
  let selectedConv = $state<(Conversation & { messages?: Message[] }) | null>(null);
  let messages = $state<Message[]>([]);
  let input = $state('');
  let sending = $state(false);

  onMount(async () => {
    await loadAgents();
    const params = new URLSearchParams(window.location.search);
    const convId = params.get('conv');
    if (convId) {
      try {
        const full = await getConversation(+convId);
        if (full && full.messages) {
          selectedAgent = agents.find(a => a.id === full.agent_id) ?? null;
          if (selectedAgent) {
            try { conversations = await getConversations(selectedAgent.id); } catch (_) {}
          }
          selectedConv = full;
          messages = full.messages;
        }
      } catch (_) {}
    }
  });

  async function loadAgents() {
    try { agents = await getAgents(); } catch (e) { console.error(e); }
  }

  async function handleAgentChange(agentId: number) {
    selectedAgent = agents.find(a => a.id === agentId) ?? null;
    selectedConv = null; messages = [];
    if (selectedAgent) {
      try { conversations = await getConversations(selectedAgent.id); }
      catch (e) { console.error(e); conversations = []; }
    } else { conversations = []; }
  }

  async function selectConversation(conv: Conversation) {
    try {
      const full = await getConversation(conv.id);
      selectedConv = full; messages = full.messages ?? [];
      const url = new URL(window.location.href);
      url.searchParams.set('conv', String(conv.id));
      if (selectedAgent) url.searchParams.set('agent', String(selectedAgent.id));
      window.history.pushState({}, '', url.toString());
    } catch (e) { console.error(e); }
  }

  async function handleNewChat() {
    if (!selectedAgent) return;
    try {
      const conv = await createConversation(selectedAgent.id);
      conversations = [conv, ...conversations];
      selectedConv = conv; messages = [];
    } catch (e) { console.error(e); }
  }

  async function handleSend() {
    const content = input.trim();
    if (!content || !selectedConv || sending) return;
    const userMsg: Message = { id: Date.now(), conversation_id: selectedConv.id, role: 'user', content, created_at: new Date().toISOString() };
    messages = [...messages, userMsg]; input = ''; sending = true;
    try {
      const response = await sendMessage(selectedConv.id, content);
      const assistantMsg: Message = { id: Date.now() + 1, conversation_id: selectedConv.id, role: response.role, content: response.content, created_at: new Date().toISOString() };
      messages = [...messages, assistantMsg];
      try { conversations = await getConversations(selectedAgent!.id); } catch (_) {}
    } catch (e: any) {
      const errorMsg: Message = { id: Date.now() + 2, conversation_id: selectedConv.id, role: 'error', content: e?.message ?? 'Failed', created_at: new Date().toISOString() };
      messages = [...messages, errorMsg];
    } finally { sending = false; }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); handleSend(); }
  }
</script>

<div class="chat-page">
  <aside class="chat-sidebar">
    <div class="agent-select">
      <label for="agent-select">Agent</label>
      <select id="agent-select" onchange={(e) => handleAgentChange(parseInt(e.target.value))}>
        <option value="">[-] Select</option>
        {#each agents as agent}
          <option value={agent.id}>{agent.name}</option>
        {/each}
      </select>
    </div>
    <button class="new-chat-btn btn-secondary" onclick={handleNewChat} disabled={!selectedAgent}>
      [+] New Chat
    </button>
    <div class="conv-list">
      {#each conversations as conv (conv.id)}
        <button class="conv-item" class:active={selectedConv?.id === conv.id} onclick={() => selectConversation(conv)}>
          {conv.title || 'Untitled'}
        </button>
      {:else}
        {#if selectedAgent}
          <p class="muted">[-] No conversations</p>
        {:else}
          <p class="muted">[-] Select an agent</p>
        {/if}
      {/each}
    </div>
  </aside>

  <section class="chat-main">
    {#if selectedConv}
      <div class="messages-panel">
        {#each messages as msg (msg.id)}
          <div class="message" class:msg-user={msg.role === 'user'} class:msg-assistant={msg.role === 'assistant'} class:msg-tool={msg.role === 'tool'} class:msg-error={msg.role === 'error'}>
            <div class="msg-role">{msg.role}</div>
            <div class="msg-content">{msg.content}</div>
          </div>
        {/each}
      </div>
      <div class="input-area">
        <textarea placeholder="Type a message... (Shift+Enter for newline)" value={input} oninput={(e) => input = e.target.value} onkeydown={handleKeydown} disabled={sending} rows={3}></textarea>
        <button class="send-btn" onclick={handleSend} disabled={sending || !input.trim()}>
          {sending ? '...' : 'Send'}
        </button>
      </div>
    {:else}
      <div class="empty-main">
        <p class="muted">[-] Select a conversation to begin.</p>
      </div>
    {/if}
  </section>
</div>

<style>
  .chat-page { display: flex; height: 100%; margin: -24px -32px; }
  .chat-sidebar {
    width: 260px; background: var(--canvas);
    border-right: 1px solid var(--hairline);
    padding: 16px; display: flex; flex-direction: column; gap: 8px;
    flex-shrink: 0; overflow: hidden;
  }
  .agent-select { display: flex; flex-direction: column; gap: 4px; }
  .agent-select label { font-size: 14px; color: var(--mute); line-height: 2; }
  .agent-select select { width: 100%; }
  .new-chat-btn { width: 100%; }
  .conv-list { flex: 1; overflow-y: auto; display: flex; flex-direction: column; gap: 2px; }
  .conv-item {
    display: block; background: transparent; color: var(--mute);
    text-align: left; padding: 8px 8px; border-radius: var(--rounded-sm);
    font-size: 16px; font-weight: 400; width: 100%; height: auto;
    white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
  }
  .conv-item:hover { background: var(--surface-soft); color: var(--ink); }
  .conv-item.active { background: var(--surface-dark); color: var(--on-dark); }
  .chat-main { flex: 1; display: flex; flex-direction: column; min-width: 0; height: 100%; }
  .messages-panel { flex: 1; overflow-y: auto; padding: 24px 32px; display: flex; flex-direction: column; gap: 12px; }
  .message { max-width: 70%; padding: 8px 12px; border-radius: var(--rounded-sm); display: flex; flex-direction: column; gap: 2px; }
  .msg-user { align-self: flex-end; background: var(--surface-dark); color: var(--on-dark); }
  .msg-user .msg-role { display: none; }
  .msg-assistant { align-self: flex-start; background: var(--surface-soft); color: var(--ink); }
  .msg-assistant .msg-role { display: none; }
  .msg-tool { align-self: flex-start; background: var(--canvas); color: var(--mute); border: 1px solid var(--hairline); font-size: 14px; }
  .msg-tool .msg-role { font-weight: 500; font-size: 14px; color: var(--mute); line-height: 2; }
  .msg-error { align-self: flex-start; background: var(--danger); color: #fff; }
  .msg-error .msg-role { display: none; }
  .msg-content { white-space: pre-wrap; word-break: break-word; font-size: 16px; line-height: 1.5; }
  .input-area {
    display: flex; gap: 8px; padding: 12px 16px;
    border-top: 1px solid var(--hairline); background: var(--canvas);
  }
  .input-area textarea { flex: 1; min-height: 48px; resize: none; }
  .send-btn { flex-shrink: 0; align-self: flex-end; }
  .empty-main { flex: 1; display: flex; align-items: center; justify-content: center; }
  .btn-secondary {
    background: var(--canvas); color: var(--ink);
    border: 1px solid var(--hairline-strong); border-radius: var(--rounded-sm);
    padding: 4px 20px; height: 36px; font-family: inherit;
    font-size: 16px; font-weight: 500; line-height: 2; cursor: pointer;
  }
  .btn-secondary:active { background: var(--surface-card); }
  .btn-secondary:disabled { background: var(--surface-card); color: var(--ash); cursor: not-allowed; opacity: 1; }
  .muted { color: var(--mute); font-size: 16px; }
</style>
