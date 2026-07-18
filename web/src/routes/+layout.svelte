<script>
  import '../app.css';
  let { children } = $props();
</script>

<div class="app-shell">
  <nav class="sidebar">
    <div class="logo">[agentic]</div>
    <a href="/" class="nav-item">[+] Agents</a>
    <a href="/chat" class="nav-item">[+] Chat</a>
    <a href="/settings" class="nav-item">[+] Settings</a>
  </nav>
  <main class="content">
    {@render children?.()}
  </main>
</div>

<style>
  .app-shell { display: flex; height: 100vh; }
  .sidebar {
    width: 220px;
    background: var(--canvas);
    border-right: 1px solid var(--hairline);
    padding: 16px 12px;
    display: flex;
    flex-direction: column;
    gap: 4px;
    flex-shrink: 0;
  }
  .logo {
    font-size: 16px;
    font-weight: 700;
    color: var(--ink);
    margin-bottom: 16px;
    padding: 0 8px;
  }
  .nav-item {
    padding: 8px 8px;
    border-radius: var(--rounded-sm);
    color: var(--mute);
    font-size: 16px;
    font-weight: 500;
    text-decoration: none;
  }
  .nav-item:hover {
    background: var(--surface-soft);
    color: var(--ink);
    text-decoration: none;
  }
  .content {
    flex: 1;
    overflow-y: auto;
    padding: 24px 32px;
    background: var(--canvas);
  }
</style>
