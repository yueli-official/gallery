// ==UserScript==
// @name         Gallery 图片发布
// @namespace    https://gallery.yuelili.com/
// @version      1.5.0
// @description  悬停网页图片，后台发布到 Gallery，支持作品角色、分组图片特征及完成通知。
// @match        https://*/*
// @match        http://*/*
// @grant        GM_getValue
// @grant        GM_setValue
// @grant        GM_deleteValue
// @grant        GM_xmlhttpRequest
// @grant        GM_notification
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
    fieldset{border:0;padding:0;margin:12px 0;min-width:0}legend{display:flex;align-items:center;width:100%}.add-value{margin-left:auto;border:0;border-radius:4px;background:none;color:var(--muted);font-size:11px;font-weight:400;padding:2px 5px}.add-value:hover{background:var(--field);color:var(--accent)}.choices{display:flex;flex-wrap:wrap;gap:6px}.choice{display:flex;align-items:center;gap:4px;margin:0;padding:4px 8px;border:1px solid var(--line);border-radius:6px;background:var(--field);font-size:12px}.choice:has(:checked),.feature-selected button{color:var(--accent);border-color:#e8a7c0;background:#fff0f5}.choice input{margin:0;accent-color:var(--accent)}.hint{color:var(--muted);font-weight:400;font-size:11px;margin-left:6px}.muted{color:var(--muted);font-size:11px}.inline{display:flex;align-items:center;gap:5px;font-size:12px}.inline input{margin:0;accent-color:var(--accent)}details{margin:14px 0 0;border-top:1px solid var(--line);padding-top:10px}summary{cursor:pointer;color:var(--muted);font-size:12px}#save-key{border:1px solid var(--line);border-radius:6px;background:var(--field);color:inherit;padding:5px 10px}#forget-key{border:0;background:none;color:var(--muted);padding:5px 8px}
    #identity-editor{margin:12px 0;padding:10px 12px;border:1px solid var(--line);border-radius:8px;background:var(--field)}#identity-heading{margin:0;font-size:12px;font-weight:600}.identity-actions{display:flex;justify-content:flex-end;gap:7px}.identity-actions button{border:1px solid var(--line);border-radius:5px;padding:5px 10px;background:var(--canvas);color:inherit;font-size:12px}#identity-save{border-color:var(--accent);background:var(--accent);color:white}#identity-error{color:#bd2b58;font-size:12px;white-space:pre-wrap;overflow-wrap:anywhere}#identity-error:empty{display:none}@media(prefers-color-scheme:dark){#identity-error{color:#ffafc8}}
    footer{justify-content:space-between;border-top:1px solid var(--line)}#progress{font-size:11px;color:var(--muted);overflow-wrap:anywhere;max-width:68%}#publish{flex:none;border:0;border-radius:7px;background:var(--accent);color:white;font-weight:600;padding:7px 18px}#error{margin:10px 0 0;color:#bd2b58;white-space:pre-wrap;font-size:12px;overflow-wrap:anywhere}#result{margin:10px 0 0;font-size:12px}#result:empty,#error:empty{display:none}#result a{margin-left:6px}
    .failure-details{margin-top:7px;padding-top:6px;color:var(--muted)}.failure-details pre{margin:6px 0 0;white-space:pre-wrap;overflow-wrap:anywhere;font:11px/1.5 ui-monospace,monospace}
    #tasks-toggle{position:fixed;right:16px;bottom:16px;pointer-events:auto;border:1px solid var(--line);border-radius:8px;background:var(--canvas);color:inherit;padding:9px 13px;box-shadow:0 3px 16px #0002;font-size:12px}#tasks-panel{position:fixed;right:12px;bottom:62px;width:min(380px,calc(100vw - 24px));pointer-events:auto;border:1px solid var(--line);border-radius:12px;background:var(--canvas);color:inherit;box-shadow:0 10px 40px #0003;overflow:hidden}#tasks-panel header{padding:10px 12px}#tasks-panel h2{font-size:13px}#tasks-close{border:0;background:none;color:var(--muted);font-size:21px;padding:0 4px}#task-list{max-height:min(60dvh,520px);overflow:auto;overscroll-behavior:contain;padding:0 12px}.task-item{padding:12px 0;border-bottom:1px solid var(--line)}.task-item:last-child{border-bottom:0}.task-title{margin:0 0 5px;font-size:12px;font-weight:600;overflow-wrap:anywhere}.task-progress{margin:0;font-size:12px;color:var(--muted);overflow-wrap:anywhere}.task-error{margin-top:7px;font-size:12px;color:#bd2b58;white-space:pre-wrap;overflow-wrap:anywhere}.task-error:empty{display:none}.task-actions{display:flex;flex-wrap:wrap;align-items:center;gap:10px;margin-top:8px;font-size:12px}.task-actions button{border:1px solid var(--line);border-radius:5px;padding:4px 8px;background:var(--field);color:inherit}.task-notice{font-size:11px;color:#bd2b58;margin:6px 0 0;overflow-wrap:anywhere}.task-notice:empty{display:none}@media(prefers-color-scheme:dark){.task-error,.task-notice{color:#ffafc8}}
    .work-tree{display:flex;flex-wrap:wrap;align-items:flex-start;gap:6px;width:100%}.work-family{display:grid;gap:5px}.work-family:has(.work-children){width:100%}.work-children{margin-left:9px;padding-left:9px;border-left:1px solid var(--line)}legend .add-value + .add-value{margin-left:3px}
    .feature-tree{display:grid;gap:7px;width:100%;min-width:0}.feature-tree details{margin:0;border:1px solid var(--line);border-radius:5px;padding:0;min-width:0}.feature-tree summary{padding:7px 8px;cursor:pointer;font-size:12px;overflow-wrap:anywhere}.feature-body{display:grid;gap:6px;padding:0 7px 7px}.feature-leaves,.feature-selected{display:flex;flex-wrap:wrap;gap:5px}.feature-tree .choice{white-space:normal;overflow-wrap:anywhere}.feature-selected:empty,.feature-limit:empty{display:none}.feature-selected button{font-size:12px;padding:4px 7px;text-align:left;overflow-wrap:anywhere}.feature-group-count{margin-left:6px;color:var(--muted);font-size:11px}.feature-limit{color:#a21d48;font-size:12px}.feature-tree button:focus-visible,.feature-tree summary:focus-visible{outline:2px solid #db2777;outline-offset:2px}
    @media(prefers-color-scheme:dark){:host{--canvas:#211e26;--field:#2b2731;--line:#423a49;--muted:#b2a7bb;color:#eae3ef;color-scheme:dark}.choice:has(:checked),.feature-selected button{background:#482338;border-color:#a24d71;color:#ffafcf}#error{color:#ffafc8}}
    @media(max-width:480px){main{padding:12px}header,footer{padding:11px 12px}#preview{height:120px}}
  </style>
  <button id="launch" type="button" hidden aria-label="发布这张图片到 Gallery"><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M12 16V4m-4 4 4-4 4 4M4 16v4h16v-4"/></svg>发布</button>
  <dialog aria-labelledby="heading"><header><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect x="3" y="3" width="18" height="18" rx="4"/><circle cx="8.5" cy="8.5" r="1.5"/><path d="m21 15-5-5L5 21"/></svg><h2 id="heading">发布到 Gallery</h2><button id="close" type="button" aria-label="关闭">×</button></header>
  <main><img id="preview" alt="待发布图片" referrerpolicy="no-referrer"><form id="form"><label><span class="caption">标题</span><input id="title" required maxlength="200" autocomplete="off"></label><div id="classification"></div><section id="identity-editor" hidden aria-labelledby="identity-heading"><h3 id="identity-heading"></h3><label><span class="caption">名称</span><input id="identity-name" maxlength="80" autocomplete="off"></label><label><span class="caption">链接标识</span><input id="identity-slug" maxlength="80" autocomplete="off" spellcheck="false"></label><label id="identity-parent-field"><span id="identity-parent-label" class="caption">所属作品</span><select id="identity-parent"></select></label><p class="muted">保存后会加入 Gallery 全站目录，需要分类读取和分类治理权限。<a href="https://gallery.yuelili.com/manage/classification" target="_blank" rel="noopener noreferrer">管理目录</a></p><div id="identity-error" role="alert"></div><div class="identity-actions"><button id="identity-cancel" type="button">取消</button><button id="identity-save" type="button">保存并选中</button></div></section><label><span class="caption">标签<span class="hint">逗号分隔</span></span><input id="tags" autocomplete="off"></label><label><span class="caption">说明</span><textarea id="description" maxlength="4000" rows="2"></textarea></label><label><span class="caption">来源页面</span><input id="source" type="url" autocomplete="off"></label><label class="inline"><input id="remember" type="checkbox" checked>记住本次分类</label><p class="muted">发布后在后台上传、审核并公开。刷新、离开或关闭本页会中断任务。需要令牌具有投稿、上传、读取自己的投稿和审核权限。</p></form>
  <details id="settings"><summary>开发者令牌</summary><label><input id="key" type="password" autocomplete="off" placeholder="粘贴 Gallery 的 pat_ 令牌" aria-label="Gallery 开发者令牌"></label><button id="save-key" type="button">保存令牌</button><button id="forget-key" type="button">清除</button><p class="muted"><a href="https://account.yuelili.com/developer-tokens" target="_blank" rel="noopener noreferrer">账户中心</a> → 月离图库</p></details><div id="error" role="alert"></div><div id="result" role="status"></div></main><footer><span id="progress" role="status">确认图片和分类后发布</span><button id="publish" type="button">发布</button></footer></dialog>
  <button id="tasks-toggle" type="button" hidden aria-label="查看发布任务" aria-controls="tasks-panel" aria-expanded="false">发布任务</button>
  <aside id="tasks-panel" hidden aria-label="发布任务"><header><h2>发布任务</h2><button id="tasks-close" type="button" aria-label="关闭发布任务">×</button></header><div id="task-list"></div></aside>`;
  const $ = id => root.getElementById(id);
  const dialog = root.querySelector('dialog');
  let target = null, chosen = null, options = null, busy = false, job = null, pointerFrame = 0, pointerTarget = null, identityDraft = null, editorRevision = 0;
  const tasks = [];
  const taskPrefix = 'gallery-publish-' + Date.now().toString(36) + '-' + Math.random().toString(36).slice(2, 8);
  let pumpingTasks = false;
  const error = failure => showFailure('error', failure);
  const progress = message => { $('progress').textContent = message; };
  const sleep = ms => new Promise(resolve => setTimeout(resolve, ms));

  function showFailure(id, failure) {
    renderFailure($(id), failure);
    $(id).scrollIntoView({ block: 'nearest' });
  }
  function renderFailure(container, failure) {
    container.replaceChildren();
    const message = typeof failure === 'string' ? failure : failure.message;
    if (!message) return;
    container.append(document.createTextNode(message));
    if (!failure.technicalDetails) return;
    const details = document.createElement('details'), summary = document.createElement('summary'), text = document.createElement('pre');
    details.className = 'failure-details'; summary.textContent = '技术详情'; text.textContent = failure.technicalDetails;
    details.append(summary, text); container.append(details);
  }

  function responseFailure(response, stage, context) {
    const statusMessages = { 400: '请求未通过检查，请展开技术详情查看原因。', 409: '当前对象已存在或状态已变化，请刷新后确认。', 413: '图片超过当前上传大小限制。', 415: '当前图片格式不受支持。', 429: '请求过于频繁，请稍后继续。' };
    const contextMessages = {
      source: { 401: '原图需要访问授权，请在来源网页确认。', 403: '原图服务器拒绝读取，请确认图片可在来源网页中打开。', 404: '原图不存在或链接已失效，请重新选择图片。', 429: '原图服务器限制了读取频率，请稍后重试。' },
      gallery: { 401: 'Gallery 令牌无效或已撤销，请重新配置。', 403: '当前令牌或账号缺少 Gallery 操作权限，请检查月离图库授权。' },
      asset: { 401: '媒体服务未接受当前令牌，请检查媒体上传授权。', 403: '当前令牌或账号缺少媒体上传权限，请检查月离图库的媒体授权。' },
      upload: { 401: '上传地址的授权无效或已失效，请关闭弹窗后重新选择图片上传。', 403: '上传地址拒绝写入，请关闭弹窗后重新选择图片上传。' },
    };
    // Standalone userscripts cannot import the site's runtime resolver. Map the
    // declared Asset Problem codes locally; never display server title/detail.
    const problemMessages = {
      'asset.upload.token_invalid': '上传凭据无效或已过期，请关闭弹窗后重新选择图片上传。',
      'asset.finalize.mismatch': '上传内容未通过完整性或格式检查，请关闭弹窗后重新上传。',
      'asset.image_process.invalid_request': '图片处理未通过检查，请换一张图片或联系管理员。',
      'asset.finalize.pixels_too_large': '图片像素超过当前限制，请缩小后上传。',
      'asset.derive.too_large': '处理后的图片超过当前限制，请缩小后上传。',
      'asset.upload.too_large': '图片超过当前上传大小限制，请压缩后上传。',
      'asset.upload.invalid_type': '图片格式不符合当前上传要求，请选择 JPEG、PNG 或 WebP 静态图片。',
      'asset.draft.unavailable': '本次上传草稿已失效，请关闭弹窗后重新选择图片。',
      'asset.reference.unavailable': '本次图片暂不可用，请到 Gallery 确认后再操作。',
    };
    let problem = {};
    if (context === 'gallery' || context === 'asset') {
      try { const parsed = JSON.parse(response.responseText || '{}'); if (parsed && parsed.status === response.status) problem = parsed; } catch {}
    }
    const code = typeof problem.code === 'string' && /^[a-z][a-z0-9_.]{0,119}$/.test(problem.code) ? problem.code : '';
    const traceId = typeof problem.traceId === 'string' && /^[a-zA-Z0-9_-]{1,128}$/.test(problem.traceId) ? problem.traceId : '';
    const message = (context === 'asset' && problemMessages[code])
      || contextMessages[context][response.status]
      || (context === 'source' ? '原图服务器返回 HTTP ' + response.status + '，请在来源网页确认图片可读取。' : statusMessages[response.status])
      || '服务返回 HTTP ' + response.status + '，请稍后检查。';
    const failure = new Error(stage + '失败：' + message);
    failure.httpStatus = response.status; failure.problemCode = code; failure.requestContext = context;
    const violations = Array.isArray(problem.violations) ? problem.violations.slice(0, 10).filter(item => item && typeof item.pointer === 'string' && /^\/[a-zA-Z0-9_./~-]{0,159}$/.test(item.pointer) && typeof item.code === 'string' && /^[a-z][a-z0-9_.]{0,119}$/.test(item.code)).map(item => item.pointer + '：' + item.code) : [];
    failure.technicalDetails = [stage, 'HTTP ' + response.status, code && '错误码：' + code, traceId && '追踪号：' + traceId, ...violations].filter(Boolean).join('\n');
    return failure;
  }

  function transport({ url, method = 'GET', headers = {}, data, responseType = 'text', redirect = 'error' }, stage, context) {
    return new Promise((resolve, reject) => {
      let settled = false, request;
      const finish = (failure, response) => { if (settled) return; settled = true; clearTimeout(timer); failure ? reject(failure) : resolve(response); };
      const interrupted = reason => new Error(stage + reason + (context === 'source'
        ? '请在来源网页确认图片可读取。'
        : ['GET', 'HEAD'].includes(method.toUpperCase()) ? '请稍后重试。' : '写入结果可能尚未确认，请先查看 Gallery。'));
      // anonymous/redirect use Tampermonkey's fetch mode, whose native timeout
      // can be ignored in Chromium. This timer owns the request deadline.
      const timer = setTimeout(() => { finish(interrupted('超时。')); request?.abort(); }, 45000);
      try { request = GM_xmlhttpRequest({
      url, method, headers, data, ...(responseType === 'blob' ? { responseType } : {}), redirect, anonymous: true, timeout: 45000,
      onload(response) {
        if (redirect === 'error' && response.finalUrl && response.finalUrl !== url) {
          finish(new Error(stage + '发生重定向，已停止。')); return;
        }
        if (response.status >= 200 && response.status < 300) { finish(null, response); return; }
        finish(responseFailure(response, stage, context));
      },
      onerror: () => finish(interrupted('网络失败。')),
      ontimeout: () => finish(interrupted('超时。')),
      onabort: () => finish(new Error(stage + '已中断。')),
    }); } catch { finish(new Error(stage + '未能发起，请检查油猴的站点访问权限。')); }
    });
  }
  async function api(origin, path, method = 'GET', body, stage = '请求', token = GM_getValue('gallery-key', '')) {
    if (![GALLERY, ASSET].includes(origin) || !path.startsWith('/api/v1/')) throw new Error('无效的 API 地址。');
    if (!token.startsWith('pat_')) throw new Error('请先保存 Gallery 开发者令牌。');
    const response = await transport({ url: origin + path, method, headers: { Accept: 'application/json', Authorization: 'Bearer ' + token, 'X-Yueli-Token-Site': SITE, ...(body !== undefined ? { 'Content-Type': 'application/json' } : {}) }, ...(body !== undefined ? { data: JSON.stringify(body) } : {}) }, stage, origin === GALLERY ? 'gallery' : 'asset');
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
      return input;
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
    function features(facet, selected) {
      const tree = group(facet.name, 'facet:' + facet.id, [], selected, false, '可选');
      addButton(tree.legend, '特征', { kind: 'facet_value', facetId: facet.id, requiresParent: true, featureValue: true });
      tree.choices.className = 'feature-tree';
      const chips = document.createElement('div'), picker = document.createElement('details'), summary = document.createElement('summary'), body = document.createElement('div'), search = document.createElement('input'), limit = document.createElement('p'), empty = document.createElement('p');
      chips.className = 'feature-selected'; chips.setAttribute('aria-label', '已选图片特征'); body.className = 'feature-body';
      search.type = 'search'; search.placeholder = '搜索图片特征'; search.setAttribute('aria-label', '搜索图片特征');
      search.addEventListener('keydown', event => { if (event.key === 'Enter') event.preventDefault(); });
      limit.className = 'feature-limit'; limit.setAttribute('role', 'status');
      empty.className = 'muted'; empty.textContent = '没有匹配的特征'; empty.hidden = true;
      picker.append(summary, body); body.append(search, empty); tree.choices.append(chips, picker, limit);
      const entries = [], branches = [];
      const children = parentId => facet.values.filter(node => (node.parentId || '') === parentId);
      function append(parentId, container, names = [], seen = new Set()) {
        const result = [];
        for (const node of children(parentId)) {
          if (seen.has(node.id)) continue;
          const path = [...names, node.name, node.slug], descendants = children(node.id);
          if (descendants.length) {
            const details = document.createElement('details'), heading = document.createElement('summary'), count = document.createElement('span'), content = document.createElement('div');
            heading.append(document.createTextNode(node.name)); count.className = 'feature-group-count'; heading.append(count); content.className = descendants.some(value => children(value.id).length) ? 'feature-body' : 'feature-leaves feature-body';
            details.append(heading, content); container.append(details);
            const branch = { details, count, entries: append(node.id, content, path, new Set([...seen, node.id])), wasOpen: false };
            branches.push(branch); result.push(...branch.entries);
          } else if (node.parentId) {
            const input = choice(container, node, 'facet:' + facet.id, selected, false), entry = { node, input, label: input.parentElement, path: path.join(' ').toLocaleLowerCase() };
            input.addEventListener('change', () => {
              if (input.checked && entries.filter(item => item.input.checked).length > facet.maxValues) { input.checked = false; limit.textContent = `最多选择 ${facet.maxValues} 项`; }
              else limit.textContent = '';
              updateSelection();
            });
            entries.push(entry); result.push(entry);
          }
        }
        return result;
      }
      append('', body);
      empty.hidden = entries.length > 0;
      function updateSelection() {
        const chosen = entries.filter(entry => entry.input.checked);
        summary.textContent = chosen.length ? `调整特征（${chosen.length}）` : '选择特征'; chips.replaceChildren();
        for (const entry of chosen) {
          const remove = document.createElement('button'); remove.type = 'button'; remove.className = 'choice'; remove.textContent = entry.node.name + ' ×'; remove.setAttribute('aria-label', '移除特征：' + entry.node.name);
          remove.addEventListener('click', () => { entry.input.checked = false; entry.input.dispatchEvent(new Event('change')); summary.focus(); }); chips.append(remove);
        }
        for (const branch of branches) { const n = branch.entries.filter(entry => entry.input.checked).length; branch.count.textContent = n ? n + ' 已选' : ''; }
      }
      let searching = false;
      search.addEventListener('input', () => {
        const term = search.value.trim().toLocaleLowerCase();
        if (term && !searching) for (const branch of branches) branch.wasOpen = branch.details.open;
        for (const entry of entries) entry.label.hidden = Boolean(term && !entry.path.includes(term));
        for (const branch of branches) { branch.details.hidden = !branch.entries.some(entry => !entry.label.hidden); branch.details.open = term ? true : branch.wasOpen; }
        empty.hidden = entries.some(entry => !entry.label.hidden);
        searching = Boolean(term);
      });
      updateSelection();
    }
    group('分类', 'category', options.categories, [defaults.categoryId], true, '', { kind: 'category', facetId: '' });
    for (const facet of options.facets.filter(item => item.maxValues > 0)) {
      const selected = defaults.facets?.find(item => item.facetId === facet.id)?.valueIds || [];
      if (facet.slug === 'image-features') { features(facet, selected); continue; }
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
        family.className = 'work-family'; family.setAttribute('role', 'group'); family.setAttribute('aria-label', work.name); heading.className = 'choices'; children.className = 'choices work-children';
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
    $('identity-parent-label').textContent = creation.featureValue ? '所属分组' : creation.requiresParent ? '所属作品' : '父级';
    const empty = document.createElement('option'); empty.value = ''; empty.textContent = creation.featureValue ? '请选择特征分组' : creation.requiresParent ? '请选择作品或角色分组' : '顶级'; parentSelect.append(empty);
    for (const node of nodes.filter(item => creation.featureValue ? nodes.some(child => child.parentId === item.id) : !creation.requiresParent || !item.parentId)) {
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
    if (draft.requiresParent && !parentId) { $('identity-error').textContent = draft.featureValue ? '请选择特征所属分组。' : '请选择所属作品；没有所属作品的选“独立角色”，出处不明确的选“待归属角色”。'; return; }
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
        if (!parent || parent.status !== 'active' || (draft.featureValue ? !existing.some(node => node.parentId === parentId) : draft.requiresParent && parent.parentId)) throw new Error('所属父级已变化，请关闭新增面板后重新选择。');
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
      const response = await transport({ url: GALLERY + '/api/v1/gallery/submission-options' }, '刷新分类', 'gallery');
      const refreshed = JSON.parse(response.responseText);
      if (!Array.isArray(refreshed.categories) || !Array.isArray(refreshed.facets)) throw new Error('刷新分类失败，请稍后继续。');
      const available = draft.kind === 'category' ? refreshed.categories : refreshed.facets.find(facet => facet.id === draft.facetId)?.values;
      if (!available?.some(item => item.id === draft.id && (item.parentId || '') === parentId)) throw new Error('项目已保存，但归属或可用状态已变化，请在管理目录检查。');
      if (draft.featureValue && available.some(item => item.parentId === draft.id)) throw new Error('这是特征分组，请选择或新增具体特征。');
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
      showFailure('identity-error', failure);
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
    const revision = ++editorRevision;
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
      const response = await transport({ url: GALLERY + '/api/v1/gallery/submission-options' }, '读取分类', 'gallery');
      if (revision !== editorRevision || !dialog.open) return;
      options = JSON.parse(response.responseText);
      if (options.facets?.some(facet => facet.slug === 'character')) throw new Error('Gallery 尚未更新作品与角色树，请先更新网站后再使用此版本。');
      if (!Array.isArray(options.categories) || !options.categories.length || !Array.isArray(options.facets)) throw new Error('Gallery 暂无可用分类。');
      renderClassification(GM_getValue('gallery-defaults', {})); progress('确认后将上传、审核并公开');
    } catch (failure) { if (revision !== editorRevision || !dialog.open) return; options = null; error(failure); $('publish').disabled = true; progress('分类读取失败，请重新打开'); }
  }
  async function canvasWebP(blob) {
    // Only call on this script's sRGB Canvas output. Stripping a source image's
    // ICC profile before color conversion would change its displayed colors.
    const bytes = new Uint8Array(await blob.arrayBuffer()), view = new DataView(bytes.buffer);
    const kind = at => String.fromCharCode(...bytes.subarray(at, at + 4));
    if (bytes.length < 12 || kind(0) !== 'RIFF' || kind(8) !== 'WEBP' || view.getUint32(4, true) + 8 !== bytes.length) throw new Error('浏览器生成的 WebP 不完整，请重新选择图片。');
    const chunks = [];
    for (let at = 12; at < bytes.length;) {
      if (at + 8 > bytes.length) throw new Error('浏览器生成的 WebP 不完整，请重新选择图片。');
      const type = kind(at), size = view.getUint32(at + 4, true), end = at + 8 + size + (size & 1);
      if (end > bytes.length) throw new Error('浏览器生成的 WebP 不完整，请重新选择图片。');
      if (['VP8X', 'VP8 ', 'VP8L', 'ALPH'].includes(type)) {
        const chunk = bytes.slice(at, end);
        if (type === 'VP8X') {
          if (size !== 10 || chunk[8] & 2) throw new Error('浏览器生成了不受支持的 WebP。');
          chunk[8] &= ~0x2c; // ICCP, EXIF and XMP flags; retain alpha and dimensions.
        }
        chunks.push(chunk);
      } else if (!['ICCP', 'EXIF', 'XMP '].includes(type)) throw new Error('浏览器生成了不受支持的 WebP。');
      at = end;
    }
    const header = bytes.slice(0, 12);
    new DataView(header.buffer).setUint32(4, 4 + chunks.reduce((sum, chunk) => sum + chunk.length, 0), true);
    return new Blob([header, ...chunks], { type: 'image/webp' });
  }
  function imageRequestHeaders(selection) {
    const source = new URL(selection.url), page = new URL(selection.page);
    const policy = (selection.referrerPolicy ?? (selection.element?.referrerPolicy || document.querySelector('meta[name="referrer"]')?.content || '')).trim().toLowerCase();
    // Use the actual selected Pixiv page's origin for its own image CDN. Never
    // borrow another site's origin, source cookies or Gallery authorization.
    if (source.protocol === 'https:' && source.hostname === 'i.pximg.net'
      && page.protocol === 'https:' && page.hostname === 'www.pixiv.net'
      && !['no-referrer', 'same-origin'].includes(policy)) {
      return { Referer: page.origin + '/' };
    }
    return {};
  }
  async function downloadImage(selection) {
    let blob;
    const url = selection.url, parsed = new URL(url);
    if (parsed.protocol === 'data:' || parsed.protocol === 'blob:') {
      const response = await fetch(url, { credentials: 'omit', referrerPolicy: 'no-referrer' });
      if (!response.ok) throw new Error('当前图片无法读取。'); blob = await response.blob();
    } else { const response = await transport({ url, headers: imageRequestHeaders(selection), responseType: 'blob', redirect: 'follow' }, '读取图片', 'source'); blob = response.response; }
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
      if (context.getContextAttributes().colorSpace !== 'srgb') throw new Error('浏览器无法将图片转换到 sRGB。');
      context.imageSmoothingEnabled = true; context.imageSmoothingQuality = 'high'; context.drawImage(bitmap, 0, 0, canvas.width, canvas.height);
      const normalized = await new Promise((resolve, reject) => canvas.toBlob(value => value?.type === 'image/webp' ? resolve(value) : reject(new Error('图片处理失败。')), 'image/webp', 1));
      if (normalized.size > MAX_BYTES) throw new Error('处理后的图片超过 20 MiB。');
      return canvasWebP(normalized);
    } finally { bitmap.close(); canvas.width = canvas.height = 1; }
  }
  function recordURL(submission) {
    return submission && ['published', 'duplicate'].includes(submission.outcome) && submission.imagePublicId
      ? GALLERY + '/images/' + encodeURIComponent(submission.imagePublicId) : GALLERY + '/submissions';
  }
  function showTasks(task) {
    $('tasks-panel').hidden = false; $('tasks-toggle').setAttribute('aria-expanded', 'true');
    if (!dialog.open) $('tasks-close').focus();
    if (task?.nodes) task.nodes.row.scrollIntoView({ block: 'nearest' });
  }
  function hideTasks() {
    $('tasks-panel').hidden = true; $('tasks-toggle').setAttribute('aria-expanded', 'false');
    $('tasks-toggle').focus();
  }
  function renderTasks() {
    const active = tasks.filter(task => ['queued', 'running'].includes(task.state)).length;
    const failed = tasks.filter(task => task.state === 'failed').length;
    $('tasks-toggle').hidden = !tasks.length;
    $('tasks-toggle').textContent = active ? `发布任务 · ${active} 进行中` : failed ? `发布任务 · ${failed} 未完成` : '发布任务 · 已完成';
    for (const task of tasks) {
      if (!task.nodes) {
        const row = document.createElement('section'), title = document.createElement('h3'), status = document.createElement('p'), failure = document.createElement('div'), actions = document.createElement('div'), retry = document.createElement('button'), settings = document.createElement('button'), link = document.createElement('a'), notice = document.createElement('p');
        row.className = 'task-item'; row.setAttribute('aria-label', task.payload.title);
        title.className = 'task-title'; title.textContent = task.payload.title;
        status.className = 'task-progress'; status.setAttribute('role', 'status');
        failure.className = 'task-error'; actions.className = 'task-actions'; notice.className = 'task-notice';
        retry.type = settings.type = 'button'; settings.textContent = '设置令牌';
        retry.addEventListener('click', () => retryTask(task)); settings.addEventListener('click', () => open());
        link.target = '_blank'; link.rel = 'noopener noreferrer';
        actions.append(retry, settings, link); row.append(title, status, failure, actions, notice); $('task-list').prepend(row);
        task.nodes = { row, status, failure, retry, settings, link, notice, renderedError: null };
      }
      const nodes = task.nodes;
      const message = task.state === 'queued' ? '等待上传' : task.progress;
      if (nodes.status.textContent !== message) nodes.status.textContent = message;
      if (nodes.renderedError !== task.error) { renderFailure(nodes.failure, task.error || ''); nodes.renderedError = task.error; }
      nodes.retry.hidden = task.state !== 'failed' || (task.uncertain && !task.submissionId);
      nodes.retry.textContent = task.submissionId ? '继续发布' : '重试';
      nodes.settings.hidden = task.state !== 'failed' || !['gallery', 'asset'].includes(task.error?.requestContext) || ![401, 403].includes(task.error?.httpStatus);
      nodes.link.hidden = !task.result && !task.submissionId && !task.uncertain;
      nodes.link.href = recordURL(task.result); nodes.link.textContent = task.done ? '查看图片' : '查看我的投稿';
      nodes.notice.textContent = task.notificationError || '';
    }
  }
  function notifyTask(task) {
    const success = task.state === 'done';
    const details = {
      title: success ? (task.result.outcome === 'duplicate' ? 'Gallery 已有这张图片' : 'Gallery 发布完成') : 'Gallery 发布未完成',
      text: task.payload.title + (success ? '' : '：' + task.error.message),
      tag: taskPrefix + '-' + task.id, url: recordURL(task.result), highlight: false,
      ...(success ? {} : { onclick: event => { event?.preventDefault(); window.focus(); showTasks(task); } }),
    };
    const failed = () => { task.notificationError = '系统通知未能发送。'; renderTasks(); };
    try { Promise.resolve(GM_notification(details)).catch(failed); } catch { failed(); }
  }
  function retryTask(task) {
    if (task.state !== 'failed' || (task.uncertain && !task.submissionId)) return;
    const token = GM_getValue('gallery-key', '');
    if (!token.startsWith('pat_')) {
      task.error = Object.assign(new Error('请先保存 Gallery 开发者令牌。'), { requestContext: 'gallery', httpStatus: 401 }); renderTasks(); return;
    }
    task.token = token; task.state = 'queued'; task.error = null; task.notificationError = ''; renderTasks(); void pumpTasks();
  }
  async function pumpTasks() {
    if (pumpingTasks) return;
    pumpingTasks = true;
    try {
      let task;
      while ((task = tasks.find(item => item.state === 'queued'))) await runTask(task);
    } finally { pumpingTasks = false; }
  }
  async function waitForSubmission(job, request, progress) {
    for (let attempt = 0; attempt < 60; attempt++) {
      const response = await request(GALLERY, '/api/v1/gallery/me/submissions?page=1&size=100', 'GET', undefined, '读取投稿状态');
      const item = response.items?.find(submission => submission.id === job.submissionId);
      if (!item) throw new Error('投稿记录暂不可见，请在 Gallery 确认后继续。');
      job.result = item; job.uncertain = false; renderTasks();
      if (['published', 'duplicate'].includes(item.outcome)) return item;
      if (['failed', 'rejected', 'withdrawn'].includes(item.outcome) || item.processingState === 'failed' || ['blocked', 'unavailable'].includes(item.safetyState)) throw new Error('当前图片未通过处理或安全检查，请到 Gallery 查看。');
      if (item.processingState === 'ready' && ['safe', 'uncertain'].includes(item.safetyState)) return item;
      progress('正在处理图片，等待审核…'); await sleep(2000);
    }
    throw new Error('图片仍在处理中。稍后点“继续发布”，会继续当前投稿。');
  }
  function publish() {
    if (busy || !chosen || !options || job?.state || identityDraft) return;
    error('');
    const token = GM_getValue('gallery-key', '');
    if (!token.startsWith('pat_')) { $('settings').open = true; error('请先保存 Gallery 开发者令牌。'); $('key').focus(); return; }
    let submissionPayload;
    try { submissionPayload = payload(); } catch (failure) { error(failure.message); return; }
    const existing = tasks.find(task => task.selection.url === chosen.url && (['queued', 'running'].includes(task.state) || task.uncertain));
    if (existing) { dialog.close(); showTasks(existing); return; }
    Object.assign(job, { id: tasks.length + 1, payload: submissionPayload, token, state: 'queued', progress: '等待上传', error: null, notificationError: '',
      selection: { url: chosen.url, page: chosen.page, title: chosen.title, referrerPolicy: chosen.element?.referrerPolicy || document.querySelector('meta[name="referrer"]')?.content || '' } });
    tasks.push(job);
    if ($('remember').checked) GM_setValue('gallery-defaults', { categoryId: submissionPayload.primaryCategoryId, facets: submissionPayload.facets });
    renderTasks(); dialog.close(); $('tasks-toggle').focus(); void pumpTasks();
  }
  async function runTask(job) {
    job.state = 'running'; job.error = null;
    const submissionPayload = job.payload;
    const token = job.token;
    const request = (origin, path, method, body, stage) => api(origin, path, method, body, stage, token);
    const progress = message => { job.progress = message; renderTasks(); };
    try {
      if (!job.submissionId) {
        if (!job.blob) { progress('读取并处理图片…'); job.blob = await downloadImage(job.selection); }
        if (!job.assetId) {
          progress('准备上传…');
          if (!job.draft) job.draft = await request(ASSET, '/api/v1/asset-references/drafts', 'POST', {}, '创建上传草稿');
          if (!job.draft.id) throw new Error('上传草稿响应异常。');
          if (!job.init) job.init = await request(ASSET, '/api/v1/assets/upload-init', 'POST', { filename: 'gallery-image.webp', mime: 'image/webp', size: job.blob.size, siteKey: 'gallery', profileKey: 'gallery-submission', category: 'gallery-submission', visibility: 'private', multipart: false, imageProcess: { execution: 'auto', target: 'stored', preprocessed: true } }, '初始化上传');
          let upload = new URL(job.init.uploadUrl, ASSET);
          // Asset's own blob path is delivered through its public HTTPS origin.
          // External object-storage signed URLs stay exactly as returned.
          if (upload.pathname.startsWith('/api/v1/assets/blob/')) upload = new URL(upload.pathname + upload.search, ASSET);
          if (upload.protocol !== 'https:' || !job.init.uploadToken) throw new Error('上传地址或凭据无效。');
          if (!job.uploaded) { progress('上传图片…'); job.uncertain = true; await transport({ url: upload.href, method: 'PUT', headers: { 'Content-Type': 'image/webp', ...job.init.uploadHeaders }, data: job.blob }, '上传图片', 'upload'); job.uploaded = true; job.uncertain = false; }
          progress('保存图片…'); job.uncertain = true;
          const finalized = await request(ASSET, '/api/v1/assets/finalize', 'POST', { uploadToken: job.init.uploadToken, draftSessionId: job.draft.id }, '保存图片');
          if (!finalized.asset?.id) throw new Error('图片保存结果未确认，请到 Gallery 查看。');
          job.assetId = finalized.asset.id; job.uncertain = false;
        }
        progress('创建投稿…'); job.uncertain = true;
        const created = await request(GALLERY, '/api/v1/gallery/submissions', 'POST', { ...submissionPayload, assetId: job.assetId }, '创建投稿');
        if (!created.submission?.id) throw new Error('投稿结果未确认，请查看我的投稿。');
        job.submissionId = created.submission.id; job.payload = submissionPayload; job.uncertain = false;
        job.result = created.submission; renderTasks();
      }
      const ready = await waitForSubmission(job, request, progress);
      let result = ready;
      if (!['published', 'duplicate'].includes(ready.outcome)) {
        progress('审核发布…');
        job.uncertain = true;
        const reviewed = await request(GALLERY, '/api/v1/gallery/admin/submissions/' + encodeURIComponent(job.submissionId) + '/review', 'POST', { decision: 'approve', note: '' }, '审核发布');
        if (!reviewed.submission || reviewed.submission.id !== job.submissionId || !['published', 'duplicate'].includes(reviewed.submission.outcome)) throw new Error('发布结果未确认，请到 Gallery 查看后继续。');
        result = reviewed.submission; job.uncertain = false;
      }
      job.done = true; job.state = 'done'; job.result = result; progress(result.outcome === 'duplicate' ? 'Gallery 已有这张图片' : '发布完成');
    } catch (failure) {
      if (failure.httpStatus >= 400 && failure.httpStatus < 500 && failure.httpStatus !== 408) job.uncertain = false;
      job.state = 'failed'; job.error = failure;
      progress(job.uncertain ? '结果尚未确认，请先查看 Gallery' : job.submissionId ? '投稿已保存，可继续发布' : '发布未完成');
    } finally {
      job.token = '';
      if (job.done) { job.blob = null; job.init = null; job.draft = null; }
      renderTasks(); notifyTask(job);
    }
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
  dialog.addEventListener('close', () => { editorRevision++; $('key').value = ''; $('preview').removeAttribute('src'); chosen = null; job = null; });
  $('save-key').addEventListener('click', () => { const token = $('key').value.trim(); if (!/^pat_[A-Za-z0-9]+$/.test(token)) { error('请填写有效的 pat_ 开发者令牌。'); return; } GM_setValue('gallery-key', token); error(''); progress('令牌已保存'); $('key').value = ''; $('settings').open = false; });
  $('forget-key').addEventListener('click', () => { GM_deleteValue('gallery-key'); $('key').value = ''; progress('令牌已清除'); });
  $('tasks-toggle').addEventListener('click', () => $('tasks-panel').hidden ? showTasks() : hideTasks());
  $('tasks-close').addEventListener('click', hideTasks);
  root.addEventListener('keydown', event => { if (event.key === 'Escape' && !dialog.open && !$('tasks-panel').hidden) { event.preventDefault(); hideTasks(); } });
  window.addEventListener('beforeunload', event => { if (tasks.some(task => ['queued', 'running'].includes(task.state))) { event.preventDefault(); event.returnValue = ''; } });
  GM_registerMenuCommand('Gallery 发布任务', () => { if (tasks.length) showTasks(); });
  GM_registerMenuCommand('Gallery 发布设置', () => open());
})();
