// InDel Worker App — minimal JS (Jinja2 SSR first, AJAX only where needed)

// ── Auto-dismiss alerts after 5 s ─────────────────────────────────────────────
document.querySelectorAll('.alert').forEach(el => {
  setTimeout(() => {
    el.style.transition = 'opacity .4s';
    el.style.opacity = '0';
    setTimeout(() => el.remove(), 400);
  }, 5000);
});

// ── Active nav link highlighting (bottom nav + any fallback nav) ─────────────
const currentPath = window.location.pathname;
document.querySelectorAll('[data-nav-item], .sidebar-nav a').forEach(link => {
  if (link.getAttribute('href') === currentPath ||
      (link.getAttribute('href') !== '/' && currentPath.startsWith(link.getAttribute('href')))) {
    link.classList.add('active');
  }
});

// ── Order accept confirmation ──────────────────────────────────────────────────
document.querySelectorAll('[data-confirm]').forEach(btn => {
  btn.addEventListener('click', e => {
    if (!confirm(btn.dataset.confirm)) e.preventDefault();
  });
});

// ── Earnings chart (simple bar chart from data attrs) ─────────────────────────
const chartEl = document.getElementById('earnings-chart');
if (chartEl) {
  const items = JSON.parse(chartEl.dataset.history || '[]');
  if (items.length > 0) {
    const maxVal = Math.max(...items.flatMap(i => [i.actual || 0, i.baseline || 0]), 1);
    const fragment = document.createDocumentFragment();
    items.slice(-8).forEach(item => {
      const group = document.createElement('div');
      group.className = 'bar-group';
      ['actual', 'baseline'].forEach(key => {
        const bar = document.createElement('div');
        bar.className = `bar bar-${key}`;
        bar.style.height = `${Math.round(((item[key] || 0) / maxVal) * 90)}px`;
        bar.title = `${key}: ₹${item[key] || 0}`;
        group.appendChild(bar);
      });
      fragment.appendChild(group);
    });
    chartEl.appendChild(fragment);
  }
}

// ── Policy SHAP bar widths ─────────────────────────────────────────────────────
document.querySelectorAll('[data-shap-bar]').forEach(bar => {
  const val = parseFloat(bar.dataset.shapBar) || 0;
  bar.style.width = `${Math.min(Math.abs(val) * 100, 100)}%`;
});

// ── Dev tools — optional polling of order count (demo-only) ───────────────────
const devPoll = document.getElementById('dev-order-count');
if (devPoll) {
  setInterval(() => {
    fetch('/orders')
      .then(r => r.text())
      .then(html => {
        const parser = new DOMParser();
        const doc = parser.parseFromString(html, 'text/html');
        const count = doc.querySelectorAll('.order-card').length;
        devPoll.textContent = count + ' orders';
      }).catch(() => {});
  }, 15000);
}
