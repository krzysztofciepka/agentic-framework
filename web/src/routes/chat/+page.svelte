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

  let convListEl: HTMLDivElement;

  onMount(() => {
    loadAgents();
  });

  async function loadAgents() {
    try {
      agents = await getAgents();
    } catch (e) {
      console.error(e);
    }
  }

  async function handleAgentChange(agentId: number) {
    selectedAgent = agents.find(a => a.id === agentId) ?? null;
    selectedConv = null;
    messages = [];
    if (selectedAgent) {
      try {
        conversations = await getConversations(selectedAgent.id);
      } catch (e) {
        console.error(e);
        conversations = [];
      }
    } else {
      conversations = [];
    }
  }

  async function selectConversation(conv: Conversation) {
    try {
      const full = await getConversation(conv.id);
      selectedConv = full;
      messages = full.messages ?? [];
    } catch (e) {
      console.error(e);
    }
  }

  async function handleNewChat() {
    if (!selectedAgent) return;
    try {
      const conv = await createConversation(selectedAgent.id);
      conversations = [...conversations, conv];
      selectedConv = conv;
      messages = [];
    } catch (e) {
      console.error(e);
    }
  }

  async function handleSend() {
    const content = input.trim();
    if (!content || !selectedConv || sending) return;

    const userMsg: Message = {
      id: Date.now(),
      conversation_id: selectedConv.id,
      role: 'user',
      content,
      created_at: new Date().toISOString(),
    };

    messages = [...messages, userMsg];
    input = '';
    sending = true;

    try {
      const response = await sendMessage(selectedConv.id, content);
      const assistantMsg: Message = {
        id: Date.now() + 1,
        conversation_id: selectedConv.id,
        role: response.role,
        content: response.content,
        created_at: new Date().toISOString(),
      };
      messages = [...messages, assistantMsg];

      try {
        conversations = await getConversations(selectedAgent!.id);
      } catch (_) {}
    } catch (e: any) {
      const errorMsg: Message = {
        id: Date.now() + 2,
        conversation_id: selectedConv.id,
        role: 'error',
        content: e?.message ?? 'Failed to send message',
        created_at: new Date().toISOString(),
      };
      messages = [...messages, errorMsg];
    } finally {
      sending = false;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  }
</script>

<div class="chat-page">
  <aside class="chat-sidebar">
    <div class="agent-select">
      <label for="agent-select">Agent</label>
      <select id="agent-select" onchange={(e) => handleAgentChange(parseInt(e.target.value))}>
        <option value="">Select an agent</option>
        {#each agents as agent}
          <option value={agent.id}>{agent.name}</option>
        {/each}
      </select>
    </div>

    <button class="new-chat-btn" onclick={handleNewChat} disabled={!selectedAgent}>
      + New Chat
    </button>

    <div class="conv-list" bind:this={convListEl}>
      {#each conversations as conv (conv.id)}
        <button
          class="conv-item"
          class:active={selectedConv?.id === conv.id}
          onclick={() => selectConversation(conv)}
        >
          <span class="conv-title">{conv.title || 'Untitled'}</span>
          <span class="conv-date">{new Date(conv.created_at).toLocaleDateString()}</span>
        </button>
      {:else}
        {#if selectedAgent}
          <p class="muted">No conversations yet.</p>
        {:else}
          <p class="muted">Select an agent to view conversations.</p>
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
        {:else}
          <div class="empty-state">
            <p class="muted">Send a message to start.</p>
          </div>
        {/each}
      </div>

      <div class="input-area">
        <textarea
          placeholder="Type a message... (Shift+Enter for newline)"
          value={input}
          oninput={(e) => input = e.target.value}
          onkeydown={handleKeydown}
          disabled={sending}
          rows={3}
        ></textarea>
        <button class="send-btn" onclick={handleSend} disabled={sending || !input.trim()}>
          {sending ? 'Sending...' : 'Send'}
        </button>
      </div>
    {:else}
      <div class="empty-main">
        <p class="muted">Select a conversation to begin chatting.</p>
      </div>
    {/if}
  </section>
</div>

<style>
  .chat-page {
    display: flex;
    height: 100%;
    margin: -2rem;
  }

  .chat-sidebar {
    width: 260px;
    background: var(--bg-surface);
    border-right: 1px solid var(--border);
    padding: 1.25rem 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    flex-shrink: 0;
    overflow: hidden;
  }

  .agent-select {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .agent-select label {
    font-size: 0.8rem;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .agent-select select {
    width: 100%;
  }

  .new-chat-btn {
    width: 100%;
    padding: 0.5rem;
    font-size: 0.85rem;
  }

  .conv-list {
    flex: 1;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    margin-top: 0.25rem;
  }

  .conv-item {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
    background: transparent;
    color: var(--text);
    text-align: left;
    padding: 0.5rem 0.75rem;
    border-radius: 6px;
    font-size: 0.85rem;
    width: 100%;
  }

  .conv-item:hover {
    background: var(--bg-elevated);
  }

  .conv-item.active {
    background: var(--accent);
    color: #000;
  }

  .conv-title {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .conv-date {
    font-size: 0.75rem;
    opacity: 0.7;
  }

  .chat-main {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
    height: 100%;
  }

  .messages-panel {
    flex: 1;
    overflow-y: auto;
    padding: 1.5rem 2rem;
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .message {
    max-width: 70%;
    padding: 0.75rem 1rem;
    border-radius: 10px;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .msg-user {
    align-self: flex-end;
    background: var(--accent);
    color: #000;
  }

  .msg-user .msg-role {
    display: none;
  }

  .msg-assistant {
    align-self: flex-start;
    background: var(--bg-elevated);
    color: var(--text);
  }

  .msg-assistant .msg-role {
    display: none;
  }

  .msg-tool {
    align-self: flex-start;
    background: var(--bg-surface);
    color: var(--text-muted);
    font-style: italic;
    font-size: 0.8rem;
  }

  .msg-tool .msg-role {
    font-weight: 600;
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-style: normal;
  }

  .msg-error {
    align-self: flex-start;
    background: var(--danger);
    color: #fff;
    font-size: 0.85rem;
  }

  .msg-error .msg-role {
    display: none;
  }

  .msg-content {
    white-space: pre-wrap;
    word-break: break-word;
  }

  .empty-state {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .empty-main {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .input-area {
    display: flex;
    gap: 0.75rem;
    padding: 1rem 1.5rem;
    border-top: 1px solid var(--border);
    background: var(--bg-surface);
  }

  .input-area textarea {
    flex: 1;
    min-height: 48px;
    resize: none;
  }

  .send-btn {
    flex-shrink: 0;
    align-self: flex-end;
  }

  .muted {
    color: var(--text-muted);
    font-size: 0.9rem;
  }
</style>
