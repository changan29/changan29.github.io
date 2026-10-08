(() => {
  const root = document.documentElement;
  try { const saved = localStorage.getItem('blog-theme'); if (['light','dark'].includes(saved)) root.dataset.theme = saved; } catch {}
  document.querySelector('#theme-toggle')?.addEventListener('click', () => {
    const dark = root.dataset.theme ? root.dataset.theme === 'dark' : matchMedia('(prefers-color-scheme: dark)').matches;
    root.dataset.theme = dark ? 'light' : 'dark';
    try { localStorage.setItem('blog-theme', root.dataset.theme); } catch {}
  });
  const search = document.querySelector('#search');
  if (!search) return;
  const list = document.querySelector('#post-list'), results = document.querySelector('#search-results');
  const pagination = document.querySelector('#pagination'), status = document.querySelector('#search-status');
  let indexPromise, revision = 0;
  search.addEventListener('input', async () => {
    const current = ++revision, query = search.value.trim().toLocaleLowerCase();
    results.replaceChildren();
    if (!query) { list.hidden = false; pagination.hidden = false; results.hidden = true; status.textContent = ''; return; }
    list.hidden = true; pagination.hidden = true; results.hidden = false;
    status.textContent = '正在搜索…';
    try {
      indexPromise ??= fetch(search.dataset.index).then(response => { if (!response.ok) throw Error('Search unavailable'); return response.json(); });
      const index = await indexPromise;
      if (revision !== current) return;
      const matches = index.filter(post => [post.title, ...(post.tags || []), post.text].join(' ').toLocaleLowerCase().includes(query));
      status.textContent = `找到 ${matches.length} 篇文章`;
      for (const post of matches) {
        const article = document.createElement('article'); article.className = 'post-card';
        const time = document.createElement('div'); time.className = 'post-meta'; time.textContent = post.date;
        const heading = document.createElement('h2'), link = document.createElement('a'); link.textContent = post.title; link.href = post.url; heading.append(link);
        const excerpt = document.createElement('p'); excerpt.textContent = post.text.slice(0, 150) + '…';
        article.append(time, heading, excerpt); results.append(article);
      }
    } catch { if (revision === current) { indexPromise = null; status.textContent = '搜索暂时不可用，请通过归档浏览文章。'; } }
  });
})();
