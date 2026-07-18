<script>
  import '../app.css';
  let { children } = $props();
</script>

<div class="app-shell">
  <nav class="sidebar">
    <div class="logo">Agentic Framework</div>
    <a href="/" class="nav-item">Agents</a>
    <a href="/chat" class="nav-item">Chat</a>
    <a href="/settings" class="nav-item">Settings</a>
  </nav>
  <main class="content">
    {@render children?.()}
  </main>
</div>

<style>
  .app-shell { display: flex; height: 100vh; }
  .sidebar {
    width: 220px; background: var(--bg-surface);
    border-right: 1px solid var(--border);
    padding: 1.5rem 1rem; display: flex; flex-direction: column; gap: 0.25rem; flex-shrink: 0;
  }
  .logo { font-size: 1.1rem; font-weight: 700; color: var(--accent); margin-bottom: 1.5rem; }
  .nav-item {
    padding: 0.5rem 0.75rem; border-radius: 6px; color: var(--text-muted); font-size: 0.9rem; text-decoration: none;
  }
  .nav-item:hover { background: var(--bg-elevated); color: var(--text); }
  .content { flex: 1; overflow-y: auto; padding: 2rem; }
</style>
