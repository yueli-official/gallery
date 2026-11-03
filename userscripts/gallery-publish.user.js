// ==UserScript==
// @name         Gallery 图片发布
// @namespace    https://gallery.yuelili.com/
// @version      1.2.0
// @description  悬停网页图片，上传并发布到 Gallery，支持新增分类、作品和角色。
// @match        https://*/*
// @match        http://*/*
// @grant        GM_getValue
// @grant        GM_setValue
// @grant        GM_deleteValue
// @grant        GM_xmlhttpRequest
// @grant        GM_registerMenuCommand
// @connect      *
// @run-at       document-idle
// @noframes
// ==/UserScript==

(() => {
  'use strict';
  const GALLERY = 'https://gallery.yuelili.com';
  const ASSET = 'https://asset.yuelili.com';
  const SITE = 'gallery-main-web';
  const MAX_BYTES = 20 * 1024 * 1024;
  const host = document.createElement('div');
  host.style.cssText = 'position:fixed;inset:0;z-index:2147483647;pointer-events:none;';
  document.documentElement.append(host);
  const root = host.attachShadow({ mode: 'closed' });
  root.innerHTML = `<style>
    :host{all:initial;font:13px/1.5 system-ui,-apple-system,"PingFang SC","Microsoft YaHei",sans-serif;color:#28252c;--canvas:#fff;--field:#f7f7f9;--line:#e8e5eb;--muted:#85808b;--accent:#db3d78;color-scheme:light}
    *{box-sizing:border-box}button,input,textarea,select{font:inherit}button{cursor:pointer}button:disabled{opacity:.5;cursor:default}a{color:var(--accent);text-decoration:none}a:hover{text-decoration:underline}[hidden]{display:none!important}
    #launch{position:fixed;pointer-events:auto;display:flex;align-items:center;gap:5px;padding:6px 11px;border:1px solid #ffffff66;border-radius:7px;background:#25202cdb;backdrop-filter:blur(10px);color:white;font-size:12px;line-height:18px;box-shadow:0 2px 9px #0002}#launch:hover{background:#cf356e}
    svg{flex:none}dialog{pointer-events:auto;width:min(440px,calc(100vw - 24px));max-height:86dvh;margin:auto;padding:0;overflow:hidden;border:1px solid var(--line);border-radius:14px;background:var(--canvas);color:inherit;box-shadow:0 18px 70px #0003}dialog::backdrop{background:#17121c6b;backdrop-filter:blur(3px)}
    header,footer{display:flex;align-items:center;gap:10px;padding:12px 16px}header{border-bottom:1px solid var(--line)}header svg{color:var(--accent)}h2{font-size:14px;font-weight:600;margin:0;flex:1}#close{border:0;background:none;color:var(--muted);font-size:21px;line-height:1;padding:2px 4px}
    main{padding:14px 16px;overflow:auto;max-height:calc(86dvh - 108px);scrollbar-width:thin}#preview{display:block;max-width:100%;height:150px;object-fit:contain;margin:0 auto 12px;border-radius:6px;background:var(--field)}label,.field{display:block;margin:10px 0}.caption,legend{display:block;font-size:12px;font-weight:600;margin:0 0 5px}input:not([type=checkbox]):not([type=radio]),textarea,select{width:100%;border:1px solid var(--line);border-radius:6px;background:var(--field);color:inherit;padding:7px 9px;outline-offset:2px}textarea{resize:vertical;min-height:54px}input:focus-visible,textarea:focus-visible,select:focus-visible,button:focus-visible,summary:focus-visible{outline:2px solid var(--accent);outline-offset:2px}
    fieldset{border:0;padding:0;margin:12px 0;min-width:0}legend{display:flex;align-items:center;width:100%}.add-value{margin-left:auto;border:0;border-radius:4px;background:none;color:var(--muted);font-size:11px;font-weight:400;padding:2px 5px}.add-value:hover{background:var(--field);color:var(--accent)}.choices{display:flex;flex-wrap:wrap;gap:6px}.choice{display:flex;align-items:center;gap:4px;margin:0;padding:4px 8px;border:1px solid var(--line);border-radius:6px;background:var(--field);font-size:12px}.choice:has(:checked){color:var(--accent);border-color:#e8a7c0;background:#fff0f5}.choice input{margin:0;accent-color:var(--accent)}.hint{color:var(--muted);font-weight:400;font-size:11px;margin-left:6px}.muted{color:var(--muted);font-size:11px}.inline{display:flex;align-items:center;gap:5px;font-size:12px}.inline input{margin:0;accent-color:var(--accent)}details{margin:14px 0 0;border-top:1px solid var(--line);padding-top:10px}summary{cursor:pointer;color:var(--muted);font-size:12px}#save-key{border:1px solid var(--line);border-radius:6px;background:var(--field);color:inherit;padding:5px 10px}#forget-key{border:0;background:none;color:var(--muted);padding:5px 8px}
    #identity-editor{margin:12px 0;padding:10px 12px;border:1px solid var(--line);border-radius:8px;background:var(--field)}#identity-heading{margin:0;font-size:12px;font-weight:600}.identity-actions{display:flex;justify-content:flex-end;gap:7px}.identity-actions button{border:1px solid var(--line);border-radius:5px;padding:5px 10px;background:var(--canvas);color:inherit;font-size:12px}#identity-save{border-color:var(--accent);background:var(--accent);color:white}#identity-error{color:#bd2b58;font-size:12px;white-space:pre-wrap;overflow-wrap:anywhere}#identity-error:empty{display:none}@media(prefers-color-scheme:dark){#identity-error{color:#ffafc8}}
    footer{justify-content:space-between;border-top:1px solid var(--line)}#progress{font-size:11px;color:var(--muted);overflow-wrap:anywhere;max-width:68%}#publish{flex:none;border:0;border-radius:7px;background:var(--accent);color:white;font-weight:600;padding:7px 18px}#error{margin:10px 0 0;color:#bd2b58;white-space:pre-wrap;font-size:12px;overflow-wrap:anywhere}#result{margin:10px 0 0;font-size:12px}#result:empty,#error:empty{display:none}#result a{margin-left:6px}
    .work-tree{display:grid;gap:9px;width:100%}.work-family{display:grid;gap:5px}.work-children{margin-left:9px;padding-left:9px;border-left:1px solid var(--line)}legend .add-value + .add-value{margin-left:3px}
    @media(prefers-color-scheme:dark){:host{--canvas:#211e26;--field:#2b2731;--line:#423a49;--muted:#b2a7bb;color:#eae3ef;color-scheme:dark}.choice:has(:checked){background:#482338;border-color:#a24d71;color:#ffafcf}#error{color:#ffafc8}}
    @media(max-width:480px){main{padding:12px}header,footer{padding:11px 12px}#preview{height:120px}}
  </style>
  <button id="launch" type="button" hidden aria-label="发布这张图片到 Gallery"><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M12 16V4m-4 4 4-4 4 4M4 16v4h16v-4"/></svg>发布</button>
  <dialog aria-labelledby="heading"><header><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect x="3" y="3" width="18" height="18" rx="4"/><circle cx="8.5" cy="8.5" r="1.5"/><path d="m21 15-5-5L5 21"/></svg><h2 id="heading">发布到 Gallery</h2><button id="close" type="button" aria-label="关闭">×</button></header>
  <main><img id="preview" alt="待发布图片" referrerpolicy="no-referrer"><form id="form"><label><span class="caption">标题</span><input id="title" required maxlength="200" autocomplete="off"></label><div id="classification"></div><section id="identity-editor" hidden aria-labelledby="identity-heading"><h3 id="identity-heading"></h3><label><span class="caption">名称</span><input id="identity-name" maxlength="80" autocomplete="off"></label><label><span class="caption">链接标识</span><input id="identity-slug" maxlength="80" autocomplete="off" spellcheck="false"></label><label id="identity-parent-field"><span id="identity-parent-label" class="caption">所属作品</span><select id="identity-parent"></select></label><p class="muted">保存后会加入 Gallery 全站目录，需要分类读取和分类治理权限。<a href="https://gallery.yuelili.com/manage/classification" target="_blank" rel="noopener noreferrer">管理目录</a></p><p id="identity-error" role="alert"></p><div class="identity-actions"><button id="identity-cancel" type="button">取消</button><button id="identity-save" type="button">保存并选中</button></div></section><label><span class="caption">标签<span class="hint">逗号分隔</span></span><input id="tags" autocomplete="off"></label><label><span class="caption">说明</span><textarea id="description" maxlength="4000" rows="2"></textarea></label><label><span class="caption">来源页面</span><input id="source" type="url" autocomplete="off"></label><label class="inline"><input id="remember" type="checkbox" checked>记住本次分类</label><p class="muted">图片处理完成后自动审核并公开，需要令牌具有投稿、上传、读取自己的投稿和审核权限。</p></form>
  <details id="settings"><summary>开发者令牌</summary><label><input id="key" type="password" autocomplete="off" placeholder="粘贴 Gallery 的 pat_ 令牌" aria-label="Gallery 开发者令牌"></label><button id="save-key" type="button">保存令牌</button><button id="forget-key" type="button">清除</button><p class="muted"><a href="https://account.yuelili.com/developer-tokens" target="_blank" rel="noopener noreferrer">账户中心</a> → 月离图库</p></details><p id="error" role="alert"></p><div id="result" role="status"></div></main><footer><span id="progress" role="status">确认图片和分类后发布</span><button id="publish" type="button">发布</button></footer></dialog>`;
  const $ = id => root.getElementById(id);
  const dialog = root.querySelector('dialog');
  let target = null, chosen = null, options = null, busy = false, job = null, pointerFrame = 0, pointerTarget = null, identityDraft = null;
  const error = message => { $('error').textContent = message; };
  const progress = message => { $('progress').textContent = message; };
  const sleep = ms => new Promise(resolve => setTimeout(resolve, ms));

  function transport({ url, method = 'GET', headers = {}, data, responseType = 'text', redirect = 'error' }, stage) {
    return new Promise((resolve, reject) => {
      let settled = false, request;
      const finish = (failure, response) => { if (settled) return; settled = true; clearTimeout(timer); failure ? reject(failure) : resolve(response); };
      // anonymous/redirect use Tampermonkey's fetch mode, whose native timeout
      // can be ignored in Chromium. This timer owns the request deadline.
      const timer = setTimeout(() => { finish(new Error(stage + '超时。写入结果可能尚未确认，请先查看 Gallery。')); request?.abort(); }, 45000);
      try { request = GM_xmlhttpRequest({
      url, method, headers, data, ...(responseType === 'blob' ? { responseType } : {}), redirect, anonymous: true, timeout: 45000,
      onload(response) {
        if (redirect === 'error' && response.finalUrl && response.finalUrl !== url) {
          finish(new Error(stage + '发生重定向，已停止。')); return;
        }
        if (response.status >= 200 && response.status < 300) { finish(null, response); return; }
        const messages = { 401: '令牌无效或已撤销，请重新配置。', 403: '当前令牌或账号权限不足，请检查月离图库权限。', 409: '当前对象已存在或状态已变化，请刷新后确认。', 413: '图片超过当前上传大小限制。', 415: '当前图片格式不受支持。', 429: '请求过于频繁，请稍后继续。' };
        let detail = '';
        try { const problem = JSON.parse(response.responseText || '{}'); if (problem.traceId) detail = '\n追踪号：' + String(problem.traceId); } catch {}
        const failure = new Error(stage + '失败：' + (messages[response.status] || '服务返回 HTTP ' + response.status + '，请稍后检查。') + detail);
        failure.httpStatus = response.status; finish(failure);
      },
      onerror: () => finish(new Error(stage + '网络失败。写入结果可能尚未确认，请先查看 Gallery。')),
      ontimeout: () => finish(new Error(stage + '超时。写入结果可能尚未确认，请先查看 Gallery。')),
      onabort: () => finish(new Error(stage + '已中断。')),
    }); } catch { finish(new Error(stage + '未能发起，请检查油猴的站点访问权限。')); }
    });
  }
  async function api(origin, path, method = 'GET', body, stage = '请求') {
    if (![GALLERY, ASSET].includes(origin) || !path.startsWith('/api/v1/')) throw new Error('无效的 API 地址。');
    const token = GM_getValue('gallery-key', '');
    if (!token.startsWith('pat_')) throw new Error('请先保存 Gallery 开发者令牌。');
    const response = await transport({ url: origin + path, method, headers: { Accept: 'application/json', Authorization: 'Bearer ' + token, 'X-Yueli-Token-Site': SITE, ...(body !== undefined ? { 'Content-Type': 'application/json' } : {}) }, ...(body !== undefined ? { data: JSON.stringify(body) } : {}) }, stage);
    if (response.status === 204) return null;
    try { return JSON.parse(response.responseText); } catch { throw new Error(stage + '返回了无效响应，请查看 Gallery 后再操作。'); }
  }
  function imageAt(element) {
    for (let node = element, depth = 0; node && node !== document.body && node !== document.documentElement && depth < 4; node = node.parentElement, depth++) {
      if (node === host) return null;
      const rect = node.getBoundingClientRect();
      if (rect.width < 80 || rect.height < 80) continue;
      let url = '';
      if (node.tagName === 'IMG') url = node.currentSrc || node.src;
      else { const background = getComputedStyle(node).backgroundImage; const match = background.match(/^url\(["']?(.+?)["']?\)$/); if (match) url = match[1]; }
      if (url && URL.canParse(url, location.href)) {
        const parsed = new URL(url, location.href);
        if (['http:', 'https:', 'blob:', 'data:'].includes(parsed.protocol)) return { element: node, url: parsed.href, title: (node.getAttribute('alt') || document.title || '网页图片').trim().slice(0, 200), page: location.href };
      }
    }
    return null;
  }
  function placeButton() {
    if (!target || !target.element.isConnected || dialog.open) { $('launch').hidden = true; return; }
    const rect = target.element.getBoundingClientRect();
    if (rect.bottom <= 0 || rect.right <= 0 || rect.top >= innerHeight || rect.left >= innerWidth) { $('launch').hidden = true; return; }
    $('launch').style.left = Math.max(8, Math.min(innerWidth - 76, rect.right - 72)) + 'px';
    $('launch').style.top = Math.max(8, Math.min(innerHeight - 40, rect.top + 8)) + 'px';
    $('launch').hidden = false;
  }
  document.addEventListener('pointermove', event => {
    if (event.composedPath().includes(host) || dialog.open) return;
    pointerTarget = event.target;
    if (pointerFrame) return;
    pointerFrame = requestAnimationFrame(() => { pointerFrame = 0; target = imageAt(pointerTarget); placeButton(); });
  }, { passive: true });
  document.addEventListener('pointerleave', () => { $('launch').hidden = true; });
  window.addEventListener('scroll', () => { $('launch').hidden = true; target = null; }, { passive: true, capture: true });
  window.addEventListener('resize', placeButton, { passive: true });

  function renderClassification(defaults) {
    const container = $('classification'); container.replaceChildren();
    function addButton(legend, label, creation) {
      const add = document.createElement('button'); add.type = 'button'; add.className = 'add-value'; add.textContent = '＋ ' + label; add.setAttribute('aria-label', '新增' + label); add.addEventListener('click', () => openIdentity({ ...creation, label })); legend.append(add);
    }
    function choice(choices, value, name, selected, single) {
      const item = document.createElement('label'), input = document.createElement('input');
      item.className = 'choice'; input.type = single ? 'radio' : 'checkbox'; input.name = name; input.value = value.id; input.checked = selected.includes(value.id);
      if (single) input.required = true;
      item.append(input, document.createTextNode(value.name)); choices.append(item);
    }
    function group(label, name, values, selected, single, hint = '', creation) {
      const field = document.createElement('fieldset'), legend = document.createElement('legend'), choices = document.createElement('div');
      legend.textContent = label;
      if (hint) { const note = document.createElement('span'); note.className = 'hint'; note.textContent = hint; legend.append(note); }
      if (creation) addButton(legend, label, creation);
      choices.className = 'choices'; field.append(legend, choices);
      for (const value of values) choice(choices, value, name, selected, single);
      container.append(field);
      return { field, legend, choices };
    }
    group('分类', 'category', options.categories, [defaults.categoryId], true, '', { kind: 'category', facetId: '' });
    for (const facet of options.facets.filter(item => item.maxValues > 0)) {
      const selected = defaults.facets?.find(item => item.facetId === facet.id)?.valueIds || [];
      if (facet.slug !== 'work') {
        group(facet.name, 'facet:' + facet.id, facet.values, selected, facet.maxValues === 1 && facet.minValues > 0, facet.minValues ? `至少 ${facet.minValues} 项` : '可选', { kind: 'facet_value', facetId: facet.id });
        continue;
      }
      const tree = group(facet.name, 'facet:' + facet.id, [], selected, false, '可选');
      addButton(tree.legend, '作品', { kind: 'facet_value', facetId: facet.id, workRoot: true });
      addButton(tree.legend, '角色', { kind: 'facet_value', facetId: facet.id, requiresParent: true });
      tree.choices.className = 'work-tree';
      for (const work of facet.values.filter(item => !item.parentId)) {
        const family = document.createElement('div'), heading = document.createElement('div'), children = document.createElement('div');
        family.className = 'work-family'; heading.className = 'choices'; children.className = 'choices work-children';
        choice(heading, work, 'facet:' + facet.id, selected, false); family.append(heading);
        function appendChildren(parent) {
          for (const value of facet.values.filter(item => item.parentId === parent.id)) { choice(children, value, 'facet:' + facet.id, selected, false); appendChildren(value); }
        }
        appendChildren(work);
        if (children.childNodes.length) family.append(children);
        tree.choices.append(family);
      }
    }
  }
  function values(name) { return [...$('form').querySelectorAll('input:checked')].filter(input => input.name === name).map(input => input.value); }
  function currentChoices() { return { categoryId: values('category')[0] || '', facets: options.facets.map(facet => ({ facetId: facet.id, valueIds: values('facet:' + facet.id) })) }; }
  // The management catalog uses database UUIDs; submission options and create
  // results use public IDs. Compare their identities in the public ID space.
  function publicId(value) {
    if (!/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(value || '')) return value || '';
    return btoa(String.fromCharCode(...value.replace(/-/g, '').match(/../g).map(byte => parseInt(byte, 16)))).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
  }
  function catalogPublicIds(catalog) {
    const node = item => ({ ...item, id: publicId(item.id), parentId: publicId(item.parentId) });
    return { ...catalog, categories: catalog.categories.map(node), facets: catalog.facets.map(facet => ({ ...facet, id: publicId(facet.id), values: facet.values.map(node) })) };
  }
  function openIdentity(creation) {
    if (busy || job?.submissionId || job?.uncertain) return;
    if (identityDraft?.id || identityDraft?.uncertain) { $('identity-editor').hidden = false; $('identity-editor').scrollIntoView({ block: 'nearest' }); return; }
    identityDraft = { ...creation, id: '', uncertain: false, slugEdited: false };
    const parentField = $('identity-parent-field'), parentSelect = $('identity-parent');
    parentSelect.replaceChildren();
    const nodes = creation.kind === 'category' ? options.categories : options.facets.find(facet => facet.id === creation.facetId)?.values || [];
    parentField.hidden = Boolean(creation.workRoot);
    $('identity-parent-label').textContent = creation.requiresParent ? '所属作品' : '父级';
    const empty = document.createElement('option'); empty.value = ''; empty.textContent = creation.requiresParent ? '请选择作品或角色分组' : '顶级'; parentSelect.append(empty);
    for (const node of nodes.filter(item => !creation.requiresParent || !item.parentId)) {
      const item = document.createElement('option'); item.value = node.id; item.textContent = node.name; parentSelect.append(item);
    }
    if (creation.requiresParent) {
      const selected = values('facet:' + creation.facetId), roots = [...new Set(nodes.filter(node => selected.includes(node.id)).map(node => node.parentId || node.id))];
      if (roots.length === 1) parentSelect.value = roots[0];
    }
    $('identity-heading').textContent = '新增' + creation.label; $('identity-name').value = ''; $('identity-slug').value = ''; $('identity-error').textContent = ''; $('identity-save').textContent = '保存并选中';
    $('identity-editor').hidden = false; lock(false); $('identity-name').focus();
  }
  async function saveIdentity() {
    if (busy || !identityDraft || identityDraft.uncertain || job?.submissionId || job?.uncertain) return;
    const draft = identityDraft;
    $('identity-error').textContent = '';
    const name = $('identity-name').value.trim(), slug = $('identity-slug').value.trim().toLowerCase();
    const parentId = draft.workRoot ? '' : $('identity-parent').value;
    if (draft.requiresParent && !parentId) { $('identity-error').textContent = '请选择所属作品；没有所属作品的选“独立角色”，出处不明确的选“待归属角色”。'; return; }
    if (!name || [...name].length > 80 || !slug || [...slug].length > 80 || /[\s/?#]/.test(slug)) { $('identity-error').textContent = '请填写名称和有效的链接标识，均不超过 80 字；标识不能含空格、/、? 或 #。'; return; }
    if (!GM_getValue('gallery-key', '').startsWith('pat_')) { $('settings').open = true; $('identity-error').textContent = '请先保存有分类治理权限的 Gallery 令牌。'; return; }
    const selection = currentChoices();
    lock(true); $('identity-save').textContent = '保存中…';
    try {
      const result = await api(GALLERY, '/api/v1/gallery/admin/classification', 'GET', undefined, '读取分类目录');
      if (!result.catalog || !Array.isArray(result.catalog.categories) || !Array.isArray(result.catalog.facets)) throw new Error('分类目录返回异常。');
      const catalog = catalogPublicIds(result.catalog);
      const existing = draft.kind === 'category' ? catalog.categories : catalog.facets.find(facet => facet.id === draft.facetId)?.values;
      if (!Array.isArray(existing)) throw new Error('当前维度已变化，请关闭面板后重新读取。');
      if (parentId) {
        const parent = existing.find(item => item.id === parentId);
        if (!parent || parent.status !== 'active' || (draft.requiresParent && parent.parentId)) throw new Error('所属父级已变化，请关闭新增面板后重新选择。');
      }
      let identity = draft.id ? existing.find(item => item.id === draft.id) : existing.find(item => (item.name.trim() === name && (item.parentId || '') === parentId) || item.slug === slug);
      if (draft.id && identity && (identity.parentId || '') !== parentId) throw new Error('该项目的归属已被修改，请到管理目录确认后重新选择。');
      if (!draft.id && identity) {
        if (identity.name.trim() !== name) throw new Error('链接标识已被“' + identity.name + '”使用，请修改。');
        if ((identity.parentId || '') !== parentId) throw new Error('同标识项目属于其他父级，请修改链接标识或到管理目录调整归属。');
        if (identity.status !== 'active') throw new Error('同名项目已存在但未启用，请先到管理目录处理。');
        draft.id = identity.id;
      }
      if (!draft.id) {
        draft.uncertain = true;
        const created = await api(GALLERY, '/api/v1/gallery/admin/classification/identities', 'POST', { kind: draft.kind, name, slug, parentId, facetId: draft.facetId }, '新增' + draft.label);
        if (!created.identity?.id) throw new Error('新增结果未确认，请先到管理目录查看，不要重复新增。');
        draft.id = created.identity.id; draft.uncertain = false; identity = { id: draft.id, status: 'draft', parentId };
      }
      if (!identity) throw new Error('已新增项目暂未出现在目录，请稍后继续。');
      if (identity.status !== 'active') {
        if (identity.status !== 'draft') throw new Error('项目状态已被修改，请到管理目录确认。');
        const command = { operation: 'set_status', kind: draft.kind, id: draft.id, status: 'active', childPlan: [], deleteAllRelated: false };
        const prepared = await api(GALLERY, '/api/v1/gallery/admin/classification/governance/preview', 'POST', { command }, '预览启用');
        const preview = prepared.preview;
        if (preview?.outcome !== 'planned' || !preview.plan?.expectedCatalogRevision || !preview.plan.expectedRequestToken) throw new Error('暂时无法启用该项目，请到管理目录检查。');
        const plan = preview.plan;
        const activated = await api(GALLERY, '/api/v1/gallery/admin/classification/governance/execute', 'POST', { command, expectedCatalogRevision: plan.expectedCatalogRevision, expectedRequestToken: plan.expectedRequestToken, expectedImpactToken: plan.expectedImpactToken }, '启用' + draft.label);
        if (!activated.execution?.applied) throw new Error('启用结果尚未确认，请稍后继续。');
      }
      const response = await transport({ url: GALLERY + '/api/v1/gallery/submission-options' }, '刷新分类');
      const refreshed = JSON.parse(response.responseText);
      if (!Array.isArray(refreshed.categories) || !Array.isArray(refreshed.facets)) throw new Error('刷新分类失败，请稍后继续。');
      const available = draft.kind === 'category' ? refreshed.categories : refreshed.facets.find(facet => facet.id === draft.facetId)?.values;
      if (!available?.some(item => item.id === draft.id && (item.parentId || '') === parentId)) throw new Error('项目已保存，但归属或可用状态已变化，请在管理目录检查。');
      if (refreshed.facets.some(facet => facet.slug === 'character')) throw new Error('Gallery 尚未更新作品与角色树，请先更新网站后再使用此版本。');
      options = refreshed;
      if (draft.kind === 'category') selection.categoryId = draft.id;
      else {
        const facet = options.facets.find(item => item.id === draft.facetId), selected = selection.facets.find(item => item.facetId === draft.facetId);
        if (selected) selected.valueIds = facet.maxValues === 1 ? [draft.id] : [...new Set([...selected.valueIds, draft.id])];
        else selection.facets.push({ facetId: draft.facetId, valueIds: [draft.id] });
      }
      renderClassification(selection); $('identity-editor').hidden = true; identityDraft = null; progress('已选中“' + name + '”');
    } catch (failure) {
      if (failure.httpStatus >= 400 && failure.httpStatus < 500 && failure.httpStatus !== 408) draft.uncertain = false;
      $('identity-error').textContent = failure.message;
      $('identity-save').textContent = draft.id ? '继续完成' : '保存并选中';
    } finally { lock(false); }
  }
  function payload() {
    if (!$('form').reportValidity()) throw new Error('请补全标题、分类和来源链接。');
    const categoryId = values('category')[0];
    if (!options.categories.some(category => category.id === categoryId)) throw new Error('请选择有效分类。');
    const facets = options.facets.map(facet => {
      const valueIds = values('facet:' + facet.id);
      if (valueIds.length < facet.minValues || (facet.maxValues > 0 && valueIds.length > facet.maxValues)) throw new Error(`“${facet.name}”需选择 ${facet.minValues}–${facet.maxValues || '不限'} 项。`);
      return { facetId: facet.id, valueIds };
    });
    return { title: $('title').value.trim(), description: $('description').value.trim(), sourceUrl: $('source').value.trim(), altText: chosen.title, categoryIds: [categoryId], primaryCategoryId: categoryId, facets, tags: [...new Set($('tags').value.split(/[,，\n]/).map(tag => tag.trim()).filter(Boolean))] };
  }
  function lock(value) {
    busy = value;
    $('publish').disabled = value || Boolean(job?.uncertain || job?.done || identityDraft);
    $('close').disabled = value;
    for (const input of $('form').querySelectorAll('input,textarea,select,button')) input.disabled = value || Boolean(job?.submissionId || job?.uncertain);
    for (const id of ['key', 'save-key', 'forget-key']) $(id).disabled = value;
    $('identity-name').disabled = $('identity-slug').disabled = $('identity-parent').disabled = value || Boolean(identityDraft?.id || identityDraft?.uncertain);
    $('identity-save').disabled = value || Boolean(identityDraft?.uncertain);
  }
  async function open(selection = null) {
    if (busy) return;
    if (dialog.open) return;
    chosen = selection; job = selection ? { blob: null, assetId: '', submissionId: '', uncertain: false, done: false } : null; identityDraft = null; $('identity-editor').hidden = true;
    error(''); $('result').replaceChildren(); $('launch').hidden = true; $('preview').hidden = !selection;
    $('form').hidden = !selection; $('publish').hidden = !selection; $('publish').textContent = '发布';
    $('title').value = selection?.title || ''; $('source').value = selection?.page || ''; $('description').value = ''; $('tags').value = '';
    $('key').value = GM_getValue('gallery-key', ''); $('settings').open = !selection || !$('key').value;
    if (selection) $('preview').src = selection.url; else $('preview').removeAttribute('src');
    dialog.showModal(); lock(false);
    progress(selection ? '读取分类…' : '令牌只发送给 Gallery 和 Asset');
    if (!selection) return;
    try {
      const response = await transport({ url: GALLERY + '/api/v1/gallery/submission-options' }, '读取分类');
      options = JSON.parse(response.responseText);
      if (options.facets?.some(facet => facet.slug === 'character')) throw new Error('Gallery 尚未更新作品与角色树，请先更新网站后再使用此版本。');
      if (!Array.isArray(options.categories) || !options.categories.length || !Array.isArray(options.facets)) throw new Error('Gallery 暂无可用分类。');
      renderClassification(GM_getValue('gallery-defaults', {})); progress('确认后将上传、审核并公开');
    } catch (failure) { options = null; error(failure.message); $('publish').disabled = true; progress('分类读取失败，请重新打开'); }
  }
  async function downloadImage(url) {
    let blob;
    const parsed = new URL(url);
    if (parsed.protocol === 'data:' || parsed.protocol === 'blob:') {
      const response = await fetch(url, { credentials: 'omit', referrerPolicy: 'no-referrer' });
      if (!response.ok) throw new Error('当前图片无法读取。'); blob = await response.blob();
    } else { const response = await transport({ url, responseType: 'blob', redirect: 'follow' }, '读取图片'); blob = response.response; }
    if (!(blob instanceof Blob) || blob.size <= 0 || blob.size > MAX_BYTES) throw new Error('图片为空或超过 20 MiB。');
    const bytes = new Uint8Array(await blob.arrayBuffer());
    const ascii = (offset, length) => String.fromCharCode(...bytes.subarray(offset, offset + length));
    let mime = '';
    if (bytes[0] === 255 && bytes[1] === 216 && bytes[2] === 255) mime = 'image/jpeg';
    if (bytes[0] === 137 && ascii(1, 3) === 'PNG') {
      mime = 'image/png'; const view = new DataView(bytes.buffer);
      for (let at = 8; at + 12 <= bytes.length;) { if (ascii(at + 4, 4) === 'acTL') throw new Error('暂不支持动画图片。'); const size = view.getUint32(at); if (size > bytes.length - at - 12) break; at += size + 12; }
    }
    if (ascii(0, 4) === 'RIFF' && ascii(8, 4) === 'WEBP') {
      mime = 'image/webp'; const view = new DataView(bytes.buffer);
      for (let at = 12; at + 8 <= bytes.length;) { const type = ascii(at, 4); if (type === 'ANIM' || type === 'ANMF' || (type === 'VP8X' && (bytes[at + 8] & 2))) throw new Error('暂不支持动画图片。'); const size = view.getUint32(at + 4, true); if (size > bytes.length - at - 8) break; at += size + 8 + (size & 1); }
    }
    if (!mime) throw new Error('仅支持 JPEG、PNG 和 WebP 静态图片。');
    const bitmap = await createImageBitmap(new Blob([blob], { type: mime }), { imageOrientation: 'from-image' });
    const canvas = document.createElement('canvas');
    try {
      const scale = Math.min(1, 4096 / Math.max(bitmap.width, bitmap.height)); canvas.width = Math.max(1, Math.round(bitmap.width * scale)); canvas.height = Math.max(1, Math.round(bitmap.height * scale));
      const context = canvas.getContext('2d', { colorSpace: 'srgb' }); if (!context) throw new Error('浏览器无法处理图片。');
      context.imageSmoothingEnabled = true; context.imageSmoothingQuality = 'high'; context.drawImage(bitmap, 0, 0, canvas.width, canvas.height);
      const normalized = await new Promise((resolve, reject) => canvas.toBlob(value => value?.type === 'image/webp' ? resolve(value) : reject(new Error('图片处理失败。')), 'image/webp', 1));
      if (normalized.size > MAX_BYTES) throw new Error('处理后的图片超过 20 MiB。');
      return normalized;
    } finally { bitmap.close(); canvas.width = canvas.height = 1; }
  }
  function showRecord(submission) {
    const container = $('result'); container.replaceChildren();
    const published = submission.outcome === 'published';
    const duplicate = submission.outcome === 'duplicate';
    const link = document.createElement('a');
    link.href = (published || duplicate) && submission.imageId ? GALLERY + '/images/' + encodeURIComponent(submission.imageId) : GALLERY + '/submissions';
    link.target = '_blank'; link.rel = 'noopener noreferrer'; link.textContent = published ? '查看图片' : '查看投稿';
    container.append(document.createTextNode(published ? '已发布到 Gallery' : duplicate ? 'Gallery 已有这张图片' : '投稿已保存'), link);
  }
  async function waitForSubmission() {
    for (let attempt = 0; attempt < 60; attempt++) {
      const response = await api(GALLERY, '/api/v1/gallery/me/submissions?page=1&size=100', 'GET', undefined, '读取投稿状态');
      const item = response.items?.find(submission => submission.id === job.submissionId);
      if (!item) throw new Error('投稿记录暂不可见，请在 Gallery 确认后继续。');
      showRecord(item);
      if (['published', 'duplicate'].includes(item.outcome)) return item;
      if (['failed', 'rejected', 'withdrawn'].includes(item.outcome) || item.processingState === 'failed' || ['blocked', 'unavailable'].includes(item.safetyState)) throw new Error('当前图片未通过处理或安全检查，请到 Gallery 查看。');
      if (item.processingState === 'ready' && ['safe', 'uncertain'].includes(item.safetyState)) return item;
      progress('正在处理图片，等待审核…'); await sleep(2000);
    }
    throw new Error('图片仍在处理中。稍后点“继续发布”，会继续当前投稿。');
  }
  async function publish() {
    if (busy || !chosen || !options || job?.done || job?.uncertain || identityDraft) return;
    error('');
    if (!GM_getValue('gallery-key', '').startsWith('pat_')) { $('settings').open = true; error('请先保存 Gallery 开发者令牌。'); $('key').focus(); return; }
    let submissionPayload;
    try { submissionPayload = job.payload || payload(); } catch (failure) { error(failure.message); return; }
    lock(true);
    try {
      if (!job.submissionId) {
        if (!job.blob) { progress('读取并处理图片…'); job.blob = await downloadImage(chosen.url); }
        if (!job.assetId) {
          progress('准备上传…');
          if (!job.draft) job.draft = await api(ASSET, '/api/v1/asset-references/drafts', 'POST', {}, '创建上传草稿');
          if (!job.draft.id) throw new Error('上传草稿响应异常。');
          if (!job.init) job.init = await api(ASSET, '/api/v1/assets/upload-init', 'POST', { filename: 'gallery-image.webp', mime: 'image/webp', size: job.blob.size, siteKey: 'gallery', profileKey: 'gallery-submission', category: 'gallery-submission', visibility: 'private', multipart: false, imageProcess: { execution: 'auto', target: 'stored', preprocessed: true } }, '初始化上传');
          let upload = new URL(job.init.uploadUrl, ASSET);
          // Asset's own blob path is delivered through its public HTTPS origin.
          // External object-storage signed URLs stay exactly as returned.
          if (upload.pathname.startsWith('/api/v1/assets/blob/')) upload = new URL(upload.pathname + upload.search, ASSET);
          if (upload.protocol !== 'https:' || !job.init.uploadToken) throw new Error('上传地址或凭据无效。');
          if (!job.uploaded) { progress('上传图片…'); job.uncertain = true; await transport({ url: upload.href, method: 'PUT', headers: { 'Content-Type': 'image/webp', ...job.init.uploadHeaders }, data: job.blob }, '上传图片'); job.uploaded = true; job.uncertain = false; }
          progress('保存图片…'); job.uncertain = true;
          const finalized = await api(ASSET, '/api/v1/assets/finalize', 'POST', { uploadToken: job.init.uploadToken, draftSessionId: job.draft.id }, '保存图片');
          if (!finalized.asset?.id) throw new Error('图片保存结果未确认，请到 Gallery 查看。');
          job.assetId = finalized.asset.id; job.uncertain = false;
        }
        progress('创建投稿…'); job.uncertain = true;
        const created = await api(GALLERY, '/api/v1/gallery/submissions', 'POST', { ...submissionPayload, assetId: job.assetId }, '创建投稿');
        if (!created.submission?.id) throw new Error('投稿结果未确认，请查看我的投稿。');
        job.submissionId = created.submission.id; job.payload = submissionPayload; job.uncertain = false;
        if ($('remember').checked) GM_setValue('gallery-defaults', { categoryId: submissionPayload.primaryCategoryId, facets: submissionPayload.facets });
        showRecord(created.submission);
      }
      const ready = await waitForSubmission();
      let result = ready;
      if (!['published', 'duplicate'].includes(ready.outcome)) {
        progress('审核发布…');
        const reviewed = await api(GALLERY, '/api/v1/gallery/admin/submissions/' + encodeURIComponent(job.submissionId) + '/review', 'POST', { decision: 'approve', note: '' }, '审核发布');
        if (!reviewed.submission || reviewed.submission.id !== job.submissionId || !['published', 'duplicate'].includes(reviewed.submission.outcome)) throw new Error('发布结果未确认，请到 Gallery 查看后继续。');
        result = reviewed.submission;
      }
      job.done = true; showRecord(result); progress('发布完成'); $('publish').textContent = '完成';
    } catch (failure) {
      if (failure.httpStatus >= 400 && failure.httpStatus < 500 && failure.httpStatus !== 408) job.uncertain = false;
      error(failure.message);
      progress(job.uncertain ? '结果尚未确认，请先查看 Gallery' : job.submissionId ? '投稿已保存，可继续发布' : '发布未完成');
      if (job.uncertain && !job.submissionId) { const link = document.createElement('a'); link.href = GALLERY + '/submissions'; link.target = '_blank'; link.rel = 'noopener noreferrer'; link.textContent = '查看我的投稿'; $('result').replaceChildren(link); }
      $('publish').textContent = job.submissionId ? '继续发布' : '发布';
    } finally { lock(false); }
  }
  $('launch').addEventListener('click', event => { event.preventDefault(); event.stopPropagation(); if (target) open({ ...target }); });
  $('publish').addEventListener('click', publish);
  $('identity-name').addEventListener('input', () => { if (identityDraft && !identityDraft.slugEdited && !identityDraft.id) $('identity-slug').value = $('identity-name').value.normalize('NFKC').trim().toLowerCase().replace(/[^\p{L}\p{N}_-]+/gu, '-').replace(/^-+|-+$/g, '').slice(0, 80); });
  $('identity-slug').addEventListener('input', () => { if (identityDraft) identityDraft.slugEdited = true; });
  $('identity-save').addEventListener('click', saveIdentity);
  $('identity-cancel').addEventListener('click', () => { if (!busy) { identityDraft = null; $('identity-editor').hidden = true; lock(false); } });
  $('form').addEventListener('submit', event => { event.preventDefault(); identityDraft ? saveIdentity() : publish(); });
  $('close').addEventListener('click', () => { if (!busy) dialog.close(); });
  dialog.addEventListener('cancel', event => { if (busy) event.preventDefault(); });
  dialog.addEventListener('close', () => { $('key').value = ''; $('preview').removeAttribute('src'); chosen = null; job = null; });
  $('save-key').addEventListener('click', () => { const token = $('key').value.trim(); if (!/^pat_[A-Za-z0-9]+$/.test(token)) { error('请填写有效的 pat_ 开发者令牌。'); return; } GM_setValue('gallery-key', token); error(''); progress('令牌已保存'); $('key').value = ''; $('settings').open = false; });
  $('forget-key').addEventListener('click', () => { GM_deleteValue('gallery-key'); $('key').value = ''; progress('令牌已清除'); });
  GM_registerMenuCommand('Gallery 发布设置', () => open());
})();
