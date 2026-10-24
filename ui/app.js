(function () {
  'use strict';

  var BASE = document.body.getAttribute('data-base') || '/goralens';
  var TOKEN_KEY = document.body.getAttribute('data-token-key') || '';

  function storedToken() {
    if (!TOKEN_KEY) return '';
    try { return localStorage.getItem(TOKEN_KEY) || ''; } catch (e) { return ''; }
  }
  var API = BASE + '/api';
  var PAGE = 50;

  var ICONS = {
    requests: ['M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20z', 'M2 12h20', 'M12 2a15 15 0 0 1 4 10 15 15 0 0 1-4 10 15 15 0 0 1-4-10 15 15 0 0 1 4-10z'],
    commands: ['M4 17l6-6-6-6', 'M12 19h8'],
    schedule: ['M4 5h16a1 1 0 0 1 1 1v14a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V6a1 1 0 0 1 1-1z', 'M3 10h18', 'M8 3v4', 'M16 3v4', 'M12 14v3h2'],
    jobs: ['M12 2l10 5-10 5L2 7z', 'M2 12l10 5 10-5', 'M2 17l10 5 10-5'],
    exceptions: ['M10.3 3.9L1.8 18a2 2 0 0 0 1.7 3h17a2 2 0 0 0 1.7-3L13.7 3.9a2 2 0 0 0-3.4 0z', 'M12 9v4', 'M12 17h.01'],
    logs: ['M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z', 'M14 2v6h6', 'M16 13H8', 'M16 17H8', 'M10 9H8'],
    dumps: ['M16 18l6-6-6-6', 'M8 6l-6 6 6 6'],
    queries: ['M12 2c4.4 0 8 1.3 8 3s-3.6 3-8 3-8-1.3-8-3 3.6-3 8-3z', 'M4 5v14c0 1.7 3.6 3 8 3s8-1.3 8-3V5', 'M4 12c0 1.7 3.6 3 8 3s8-1.3 8-3'],
    models: ['M21 8l-9-5-9 5 9 5 9-5z', 'M3 8v8l9 5 9-5V8', 'M12 13v8'],
    events: ['M13 2L3 14h9l-1 8 10-12h-9l1-8z'],
    mail: ['M4 4h16a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2z', 'M22 6l-10 7L2 6'],
    cache: ['M22 12H2', 'M5.5 5.1L2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.5-6.9A2 2 0 0 0 16.8 4H7.2a2 2 0 0 0-1.7 1.1z', 'M6 16h.01', 'M10 16h.01'],
    redis: ['M3 4h18v6H3z', 'M3 14h18v6H3z', 'M7 7h.01', 'M7 17h.01'],
    gates: ['M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z', 'M9 12l2 2 4-4'],
    'http-client': ['M7 17L17 7', 'M8 7h9v9'],
    monitoring: ['M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z', 'M12 9a3 3 0 1 0 0 6 3 3 0 0 0 0-6z'],
    search: ['M11 4a7 7 0 1 0 0 14 7 7 0 0 0 0-14z', 'M21 21l-4.3-4.3'],
    refresh: ['M21 12a9 9 0 1 1-2.6-6.4', 'M21 3v6h-6'],
    trash: ['M3 6h18', 'M8 6V4h8v2', 'M19 6l-1 14H6L5 6', 'M10 11v6', 'M14 11v6'],
    sun: ['M12 7a5 5 0 1 0 0 10 5 5 0 0 0 0-10z', 'M12 1v2', 'M12 21v2', 'M4.2 4.2l1.4 1.4', 'M18.4 18.4l1.4 1.4', 'M1 12h2', 'M21 12h2', 'M4.2 19.8l1.4-1.4', 'M18.4 5.6l1.4-1.4'],
    moon: ['M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8z'],
    system: ['M3 4h18v12H3z', 'M8 20h8', 'M12 16v4'],
    menu: ['M3 6h18', 'M3 12h18', 'M3 18h18'],
    x: ['M18 6L6 18', 'M6 6l12 12'],
    copy: ['M9 9h11a1 1 0 0 1 1 1v11a1 1 0 0 1-1 1H9a1 1 0 0 1-1-1V10a1 1 0 0 1 1-1z', 'M5 15H4a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1h10a1 1 0 0 1 1 1v1'],
    check: ['M20 6L9 17l-5-5'],
    back: ['M19 12H5', 'M12 19l-7-7 7-7'],
    play: ['M6 4l14 8-14 8z'],
    pause: ['M7 4h3v16H7z', 'M14 4h3v16h-3z'],
    clock: ['M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20z', 'M12 6v6l4 2'],
    tag: ['M20.6 13.4l-7.2 7.2a2 2 0 0 1-2.8 0L2 12V2h10l8.6 8.6a2 2 0 0 1 0 2.8z', 'M7 7h.01'],
    layers: ['M12 2l10 5-10 5L2 7z', 'M2 12l10 5 10-5'],
    alert: ['M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20z', 'M12 8v4', 'M12 16h.01'],
    route: ['M6 19a2 2 0 1 0 0-4 2 2 0 0 0 0 4z', 'M18 9a2 2 0 1 0 0-4 2 2 0 0 0 0 4z', 'M6 15V9a4 4 0 0 1 4-4h6', 'M18 9v6a4 4 0 0 1-4 4H8'],
    user: ['M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2', 'M12 3a4 4 0 1 0 0 8 4 4 0 0 0 0-8z'],
    server: ['M3 4h18v7H3z', 'M3 13h18v7H3z'],
    hash: ['M4 9h16', 'M4 15h16', 'M10 3L8 21', 'M16 3l-2 18'],
    file: ['M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z', 'M14 2v6h6']
  };

  function icon(name) {
    var svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
    svg.setAttribute('viewBox', '0 0 24 24');
    svg.setAttribute('class', 'icon');
    svg.setAttribute('aria-hidden', 'true');
    (ICONS[name] || ICONS.file).forEach(function (d) {
      var p = document.createElementNS('http://www.w3.org/2000/svg', 'path');
      p.setAttribute('d', d);
      svg.appendChild(p);
    });
    return svg;
  }

  var SCREENS = [
    { type: 'request', slug: 'requests', title: 'Requests', group: 'Application', hint: 'HTTP requests handled by the application appear here.' },
    { type: 'client_request', slug: 'http-client', title: 'HTTP Client', group: 'Application', hint: 'Outgoing requests made through the Goravel HTTP client or goralens.Transport().' },
    { type: 'command', slug: 'commands', title: 'Commands', group: 'Application', hint: 'Artisan commands run through ./artisan appear here.' },
    { type: 'schedule', slug: 'schedule', title: 'Schedule', group: 'Application', hint: 'Scheduled tasks appear when the scheduler runs them.' },
    { type: 'job', slug: 'jobs', title: 'Jobs', group: 'Application', hint: 'Queued and synchronous jobs, with their status.' },
    { type: 'event', slug: 'events', title: 'Events', group: 'Application', hint: 'Dispatched events and the listeners that handled them.' },
    { type: 'gate', slug: 'gates', title: 'Gates', group: 'Application', hint: 'Authorization checks and whether they were allowed.' },
    { type: 'mail', slug: 'mail', title: 'Mail', group: 'Application', hint: 'Mail sent or queued by the application, with a preview.' },
    { type: 'exception', slug: 'exceptions', title: 'Exceptions', group: 'Diagnostics', hint: 'Panics and error level logs, grouped by type and location.' },
    { type: 'log', slug: 'logs', title: 'Logs', group: 'Diagnostics', hint: 'Log entries written through the goralens log channel.' },
    { type: 'dump', slug: 'dumps', title: 'Dumps', group: 'Diagnostics', hint: 'Values passed to goralens.Dump(v).' },
    { type: 'query', slug: 'queries', title: 'Queries', group: 'Data', hint: 'Database queries run through the ORM, with their bindings and timing.' },
    { type: 'model', slug: 'models', title: 'Models', group: 'Data', hint: 'Model creates, updates and deletes with the changed fields.' },
    { type: 'cache', slug: 'cache', title: 'Cache', group: 'Data', hint: 'Cache hits, misses, writes and deletes.' },
    { type: 'redis', slug: 'redis', title: 'Redis', group: 'Data', hint: 'Redis commands sent through the redishook package.' }
  ];
  var bySlug = {}, byType = {};
  SCREENS.forEach(function (s) { bySlug[s.slug] = s; byType[s.type] = s; });

  var state = { status: null, entries: [], screen: null, autoRefresh: false, timer: null, tag: '', complete: false };
  var view = document.getElementById('view');

  function h(tag, attrs) {
    var el = document.createElement(tag);
    if (attrs) {
      Object.keys(attrs).forEach(function (k) {
        var v = attrs[k];
        if (v === null || v === undefined || v === false) return;
        if (k === 'class') el.className = v;
        else if (k.indexOf('on') === 0) el.addEventListener(k.slice(2), v);
        else el.setAttribute(k, v === true ? '' : v);
      });
    }
    for (var i = 2; i < arguments.length; i++) append(el, arguments[i]);
    return el;
  }

  function append(el, child) {
    if (child === null || child === undefined || child === false) return;
    if (Array.isArray(child)) { child.forEach(function (c) { append(el, c); }); return; }
    el.appendChild(child instanceof Node ? child : document.createTextNode(String(child)));
  }

  function clear(el) { while (el.firstChild) el.removeChild(el.firstChild); }

  function escapeHtml(s) {
    return String(s).replace(/[&<>"']/g, function (c) {
      return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c];
    });
  }

  function request(method, path, body) {
    var opts = { method: method, headers: { 'Accept': 'application/json' }, credentials: 'same-origin' };
    if (method !== 'GET') {
      opts.headers['X-Goralens'] = '1';
      opts.headers['Content-Type'] = 'application/json';
      opts.body = JSON.stringify(body || {});
    }
    var token = storedToken();
    if (token) opts.headers['Authorization'] = 'Bearer ' + token;
    return fetch(API + path, opts).then(function (res) {
      return res.json().catch(function () { return {}; }).then(function (data) {
        if (!res.ok) {
          var err = new Error(data.message || ('Request failed (' + res.status + ')'));
          err.status = res.status;
          throw err;
        }
        return data;
      });
    });
  }

  function timeAgo(iso) {
    var seconds = Math.max(0, Math.round((Date.now() - new Date(iso).getTime()) / 1000));
    if (seconds < 5) return 'just now';
    if (seconds < 60) return seconds + 's ago';
    if (seconds < 3600) return Math.floor(seconds / 60) + 'm ago';
    if (seconds < 86400) return Math.floor(seconds / 3600) + 'h ago';
    return Math.floor(seconds / 86400) + 'd ago';
  }

  function fullTime(iso) { return new Date(iso).toLocaleString(); }

  function toast(text) {
    var t = h('div', { class: 'toast', role: 'status' }, text);
    document.body.appendChild(t);
    setTimeout(function () { t.remove(); }, 1800);
  }

  function badge(text, kind, extra) { return h('span', { class: 'badge' + (kind ? ' ' + kind : '') + (extra ? ' ' + extra : '') }, text); }

  function methodBadge(method) {
    var m = String(method || '').toUpperCase();
    var kind = { GET: 'info', HEAD: 'info', POST: 'success', PUT: 'warning', PATCH: 'warning', DELETE: 'danger', OPTIONS: '' }[m];
    return badge(m || '—', kind, 'method');
  }

  function statusKind(code) {
    code = Number(code) || 0;
    return code >= 500 ? 'danger' : code >= 400 ? 'warning' : code >= 300 ? 'info' : code ? 'success' : '';
  }

  function statusBadge(code, large) { return badge(Number(code) || '—', statusKind(code), large ? 'lg' : null); }

  function levelKind(level) {
    return { debug: '', info: 'info', notice: 'info', warning: 'warning', error: 'danger', fatal: 'danger', panic: 'danger', critical: 'danger' }[level] || '';
  }

  function levelBadge(level) { return badge(level || 'log', levelKind(level)); }

  function resultKind(text) {
    return { processed: 'success', allowed: 'success', hit: 'success', sent: 'success', queued: 'info', set: 'info', created: 'success',
      updated: 'info', failed: 'danger', denied: 'danger', deleted: 'danger', missed: 'warning', forget: 'warning', flush: 'warning' }[text] || '';
  }

  function resultBadge(text) { return badge(text || '—', resultKind(text)); }

  function ms(v) {
    if (v === undefined || v === null) return '—';
    if (v >= 1000) return (Math.round(v / 10) / 100) + 's';
    if (v < 1) return (Math.round(v * 100) / 100) + 'ms';
    return (Math.round(v * 10) / 10) + 'ms';
  }

  function duration(v, warn, danger) {
    var cls = 'duration';
    if (v >= (danger || 1000)) cls += ' very-slow';
    else if (v >= (warn || 300)) cls += ' slow';
    return h('span', { class: cls }, ms(v));
  }

  function short(v, n) {
    var s = typeof v === 'string' ? v : JSON.stringify(v);
    s = (s || '').replace(/\s+/g, ' ').trim();
    return s.length > n ? s.slice(0, n) + '…' : s;
  }

  function basename(path) { return String(path || '').split('/').pop(); }

  function fileLine(file, line) { return file ? basename(file) + ':' + line : ''; }

  function copy(text, button) {
    var done = function () {
      if (!button) return toast('Copied to clipboard');
      clear(button);
      append(button, [icon('check'), 'Copied']);
      setTimeout(function () { clear(button); append(button, [icon('copy'), 'Copy']); }, 1500);
    };
    if (navigator.clipboard && window.isSecureContext) navigator.clipboard.writeText(text).then(done, function () { fallbackCopy(text); done(); });
    else { fallbackCopy(text); done(); }
  }

  function fallbackCopy(text) {
    var ta = h('textarea', { style: 'position:fixed;opacity:0' });
    ta.value = text;
    document.body.appendChild(ta);
    ta.select();
    try { document.execCommand('copy'); } catch (e) {}
    ta.remove();
  }

  function highlightJson(text) {

    return escapeHtml(text).replace(
      /(&quot;(?:\\.|[^&]|&(?!quot;))*?&quot;)(\s*:)?|\b(true|false)\b|\bnull\b|-?\b\d+(?:\.\d+)?(?:[eE][+-]?\d+)?\b/g,
      function (m, str, colon, bool) {
        if (str) return colon ? '<span class="json-key">' + str + '</span>' + colon : '<span class="json-string">' + str + '</span>';
        if (bool) return '<span class="json-boolean">' + m + '</span>';
        if (m === 'null') return '<span class="json-null">null</span>';
        return '<span class="json-number">' + m + '</span>';
      });
  }

  var SQL_KEYWORDS = 'SELECT|FROM|WHERE|AND|OR|INSERT|INTO|VALUES|UPDATE|SET|DELETE|JOIN|LEFT|RIGHT|INNER|OUTER|FULL|CROSS|ON|GROUP BY|ORDER BY|LIMIT|OFFSET|RETURNING|AS|IN|IS|NOT|NULL|DESC|ASC|COUNT|DISTINCT|HAVING|UNION|ALL|CREATE|TABLE|ALTER|DROP|LIKE|ILIKE|BETWEEN|EXISTS|CASE|WHEN|THEN|ELSE|END|WITH|ON CONFLICT|DO NOTHING|DO UPDATE|BEGIN|COMMIT|ROLLBACK|TRUE|FALSE';

  function formatSql(sql) {
    return String(sql || '').replace(/\s+/g, ' ').trim()
      .replace(/\s+(FROM|WHERE|GROUP BY|ORDER BY|HAVING|LIMIT|OFFSET|RETURNING|VALUES|SET|ON CONFLICT|UNION(?: ALL)?|(?:LEFT |RIGHT |INNER |OUTER |FULL |CROSS )?JOIN)\s+/gi, '\n$1 ')
      .replace(/\s+(AND|OR)\s+/gi, '\n  $1 ');
  }

  function highlightSql(sql) {
    return escapeHtml(sql).replace(new RegExp("(&#39;(?:[^&]|&(?!#39;))*?&#39;)|(\\$\\d+|\\?)|\\b(" + SQL_KEYWORDS + ")\\b|\\b(\\d+(?:\\.\\d+)?)\\b", 'gi'),
      function (m, str, ph, kw, num) {
        if (str) return '<span class="sql-string">' + str + '</span>';
        if (ph) return '<span class="sql-placeholder">' + ph + '</span>';
        if (kw) return '<span class="sql-keyword">' + kw.toUpperCase() + '</span>';
        return '<span class="sql-number">' + num + '</span>';
      });
  }

  function codeBlock(value, kind) {
    var text;
    if (kind === 'sql') text = formatSql(value);
    else if (kind === 'text') text = value === undefined || value === null ? '' : String(value);
    else text = typeof value === 'string' ? value : JSON.stringify(value === undefined ? null : value, null, 2);
    var pre = h('pre', { tabindex: '0' });
    if (kind === 'sql') pre.innerHTML = highlightSql(text);
    else if (kind === 'text') pre.textContent = text;
    else pre.innerHTML = highlightJson(text);
    var copyBtn = h('button', { class: 'button small', type: 'button', 'aria-label': 'Copy to clipboard', onclick: function () { copy(text, copyBtn); } }, icon('copy'), 'Copy');
    var box = h('div', { class: 'code' }, pre, h('div', { class: 'code-actions' }, copyBtn));
    if (text.split('\n').length > 18 || text.length > 2400) {
      box.classList.add('collapsed');
      var toggle = h('button', { class: 'button small ghost', type: 'button', 'aria-expanded': 'false', onclick: function () {
        var collapsed = box.classList.toggle('collapsed');
        toggle.setAttribute('aria-expanded', String(!collapsed));
        toggle.textContent = collapsed ? 'Show everything' : 'Show less';
      } }, 'Show everything');
      return h('div', null, box, h('div', { class: 'code-toggle' }, toggle));
    }
    return box;
  }

  function card(title, body, opts) {
    opts = opts || {};
    return h('section', { class: 'card' + (opts.class ? ' ' + opts.class : '') },
      title ? h('div', { class: 'card-head' }, h('h2', null, opts.icon ? icon(opts.icon) : null, title), opts.meta ? h('span', { class: 'meta' }, opts.meta) : null) : null,
      body);
  }

  function props(rows) {
    var dl = h('dl', { class: 'props' });
    rows.forEach(function (r) {
      if (r[1] === undefined || r[1] === null || r[1] === '') return;
      dl.appendChild(h('div', { class: 'prop' }, h('dt', null, r[0]), h('dd', null, r[1])));
    });
    return dl.children.length ? card(null, dl) : null;
  }

  var tabsSeq = 0;
  function tabs(items, title) {
    items = items.filter(function (i) { return i[1] !== undefined && i[1] !== null && !(typeof i[1] === 'object' && !(i[1] instanceof Node) && Object.keys(i[1]).length === 0 && i[2] !== true); });
    if (!items.length) return null;
    var id = 'tabs' + (++tabsSeq);
    var body = h('div', { role: 'tabpanel', id: id + '-panel' });
    var bar = h('div', { class: 'tabs', role: 'tablist', 'aria-label': title || 'Sections' });
    var buttons = [];
    function show(i, focus) {
      buttons.forEach(function (b, j) {
        b.setAttribute('aria-selected', String(i === j));
        b.tabIndex = i === j ? 0 : -1;
      });
      if (focus) buttons[i].focus();
      clear(body);
      var content = items[i][1];
      body.appendChild(content instanceof Node ? content : codeBlock(content));
    }
    items.forEach(function (item, i) {
      var b = h('button', { type: 'button', role: 'tab', 'aria-controls': id + '-panel', onclick: function () { show(i); },
        onkeydown: function (ev) {
          if (ev.key === 'ArrowRight') show((i + 1) % items.length, true);
          if (ev.key === 'ArrowLeft') show((i - 1 + items.length) % items.length, true);
        } }, item[0]);
      buttons.push(b);
      bar.appendChild(b);
    });
    show(0);
    return h('section', { class: 'card' }, bar, body);
  }

  function entryLink(e) { return '#/' + (byType[e.type] ? byType[e.type].slug : 'requests') + '/' + e.uuid; }

  function tagChips(tags, limit) {
    if (!tags || !tags.length) return null;
    var shown = tags.slice(0, limit || tags.length);
    var chips = shown.map(function (t) {
      return h('button', { class: 'chip', type: 'button', title: 'Filter by ' + t, onclick: function (ev) { ev.stopPropagation(); filterByTag(t); } }, icon('tag'), t);
    });
    if (tags.length > shown.length) chips.push(h('span', { class: 'chip', title: tags.slice(shown.length).join(', ') }, '+' + (tags.length - shown.length)));
    return chips;
  }

  var COLUMNS = {
    request: [['Verb', function (c) { return methodBadge(c.method); }, 'nowrap'], ['Path', function (c) { return h('span', { class: 'mono' }, c.uri); }, 'truncate', true],
      ['Status', function (c) { return statusBadge(c.response_status); }, 'nowrap'], ['Duration', function (c) { return duration(c.duration); }, 'num']],
    client_request: [['Verb', function (c) { return methodBadge(c.method); }, 'nowrap'], ['URI', function (c) { return h('span', { class: 'mono' }, c.uri); }, 'truncate', true],
      ['Status', function (c) { return c.error ? badge('error', 'danger') : statusBadge(c.response_status); }, 'nowrap'], ['Duration', function (c) { return duration(c.duration, 500, 2000); }, 'num']],
    command: [['Command', function (c) { return h('span', { class: 'mono' }, c.command); }, 'truncate', true],
      ['Exit code', function (c) { return badge(String(c.exit_code), c.exit_code ? 'danger' : 'success'); }, 'nowrap'], ['Duration', function (c) { return duration(c.duration, 5000, 30000); }, 'num']],
    schedule: [['Task', function (c) { return h('span', { class: 'mono' }, c.name || c.command); }, 'truncate', true], ['Expression', function (c) { return h('span', { class: 'mono muted' }, c.expression); }, 'nowrap hide-sm'],
      ['Status', function (c) { return resultBadge(c.status); }, 'nowrap']],
    job: [['Job', function (c) { return h('span', { class: 'mono' }, c.name); }, 'truncate', true],
      ['Queue', function (c) { return h('span', { class: 'muted' }, [c.connection, c.queue].filter(Boolean).join(' / ') || (c.sync ? 'sync' : '—')); }, 'nowrap hide-sm'],
      ['Status', function (c) { return resultBadge(c.status); }, 'nowrap']],
    exception: [['Exception', function (c) { return h('div', null, h('div', { class: 'mono', style: 'font-weight:600' }, c.class), h('div', { class: 'muted', style: 'margin-top:2px' }, short(c.message, 150))); }, 'truncate', true],
      ['Location', function (c) { return h('span', { class: 'mono muted' }, fileLine(c.file, c.line)); }, 'nowrap hide-sm'],
      ['Count', function (c) { return badge(String(c.occurrences || 1), (c.occurrences || 1) > 1 ? 'danger' : ''); }, 'num']],
    log: [['Level', function (c) { return levelBadge(c.level); }, 'nowrap'], ['Message', function (c) { return short(c.message, 200); }, 'truncate', true]],
    dump: [['Dump', function (c) { return h('span', { class: 'mono' }, short(c.dump, 160)); }, 'truncate', true], ['Location', function (c) { return h('span', { class: 'mono muted' }, fileLine(c.file, c.line)); }, 'nowrap hide-sm']],
    query: [['Query', function (c) { return h('span', { class: 'mono' }, short(c.sql, 200)); }, 'truncate', true],
      ['Location', function (c) { return h('span', { class: 'mono muted' }, fileLine(c.file, c.line)); }, 'nowrap hide-sm'],
      ['Duration', function (c) { return h('span', null, c.slow ? badge('slow', 'warning') : null, ' ', duration(c.time, 100, 500)); }, 'num']],
    model: [['Model', function (c) { return h('span', { class: 'mono' }, c.model + (c.key !== null && c.key !== undefined ? ':' + c.key : '')); }, 'truncate', true],
      ['Action', function (c) { return resultBadge(c.action); }, 'nowrap']],
    event: [['Event', function (c) { return h('span', { class: 'mono' }, c.name); }, 'truncate', true], ['Listeners', function (c) { return badge(String((c.listeners || []).length)); }, 'num']],
    mail: [['Subject', function (c) { return c.subject || '(no subject)'; }, 'truncate', true], ['To', function (c) { return h('span', { class: 'muted' }, (c.to || []).join(', ')); }, 'truncate hide-sm'],
      ['', function (c) { return c.error ? badge('failed', 'danger') : c.queued ? badge('queued', 'info') : badge('sent', 'success'); }, 'nowrap']],
    cache: [['Key', function (c) { return h('span', { class: 'mono' }, c.key); }, 'truncate', true], ['Action', function (c) { return resultBadge(c.type); }, 'nowrap']],
    redis: [['Command', function (c) { return h('span', { class: 'mono' }, short(c.command, 200)); }, 'truncate', true], ['Duration', function (c) { return duration(c.time, 50, 250); }, 'num']],
    gate: [['Ability', function (c) { return h('span', { class: 'mono' }, c.ability); }, 'truncate', true], ['Result', function (c) { return resultBadge(c.result); }, 'nowrap']]
  };

  function summaryText(e) {
    var c = e.content || {};
    switch (e.type) {
      case 'request': case 'client_request': return (c.method || '') + ' ' + (c.uri || '') + (c.response_status ? ' · ' + c.response_status : '');
      case 'command': return c.command;
      case 'schedule': return c.name || c.command;
      case 'job': return c.name + (c.status ? ' · ' + c.status : '');
      case 'exception': return c.class + ': ' + short(c.message, 120);
      case 'log': return '[' + (c.level || 'log') + '] ' + short(c.message, 140);
      case 'dump': return short(c.dump, 140);
      case 'query': return short(c.sql, 160);
      case 'model': return (c.action || '') + ' ' + c.model + (c.key !== null && c.key !== undefined ? ':' + c.key : '');
      case 'event': return c.name;
      case 'mail': return c.subject || '(no subject)';
      case 'cache': return (c.type || '') + ' ' + c.key;
      case 'redis': return short(c.command, 140);
      case 'gate': return c.ability + ' · ' + c.result;
      default: return short(c, 140);
    }
  }

  function severity(e) {
    var c = e.content || {};
    if (e.type === 'exception') return 'danger';
    if (e.type === 'log') return levelKind(c.level) === 'danger' ? 'danger' : levelKind(c.level) === 'warning' ? 'warning' : '';
    if (e.type === 'request' || e.type === 'client_request') return c.error ? 'danger' : statusKind(c.response_status) === 'danger' ? 'danger' : statusKind(c.response_status) === 'warning' ? 'warning' : '';
    if (e.type === 'job' || e.type === 'schedule') return c.status === 'failed' ? 'danger' : '';
    if (e.type === 'command') return c.exit_code ? 'danger' : '';
    if (e.type === 'query') return c.error ? 'danger' : c.slow ? 'warning' : '';
    if (e.type === 'gate') return c.result === 'denied' ? 'warning' : '';
    return '';
  }

  function entryTable(entries, type) {
    var cols = COLUMNS[type] || [['Entry', function (c) { return short(c, 120); }, 'truncate', true]];
    var thead = h('thead', null, h('tr', null,
      cols.map(function (c) { return h('th', { class: (c[2] || '').replace('truncate', ''), scope: 'col' }, c[0]); }),
      h('th', { class: 'time', scope: 'col' }, 'Happened')));
    var tbody = h('tbody');
    entries.forEach(function (e) {
      var sev = severity(e);
      var open = function () { window.location.hash = entryLink(e); };
      var row = h('tr', { tabindex: '0', class: sev === 'danger' ? 'row-error' : sev === 'warning' ? 'row-warning' : null,
        'aria-label': summaryText(e), onclick: open, onkeydown: function (ev) { if (ev.key === 'Enter') open(); } },
      cols.map(function (c) {
        var content = c[1](e.content || {});
        return h('td', { class: c[2] || null }, c[3] ? h('div', null, h('div', { class: 'cell-main' }, content), e.tags && e.tags.length ? h('div', { class: 'cell-sub' }, tagChips(e.tags, 3)) : null) : content);
      }),
      h('td', { class: 'time', title: fullTime(e.created_at) }, timeAgo(e.created_at)));
      tbody.appendChild(row);
    });
    return h('div', { class: 'table-wrap' }, h('table', null, thead, tbody));
  }

  function stat(label, value, unit, kind, iconName) {
    return h('div', { class: 'stat' + (kind ? ' ' + kind : '') },
      h('div', { class: 'stat-label' }, iconName ? icon(iconName) : null, label),
      h('div', { class: 'stat-value' }, value, unit ? h('small', null, unit) : null));
  }

  function avg(list) { return list.length ? list.reduce(function (a, b) { return a + b; }, 0) / list.length : 0; }

  function statsFor(type, entries) {
    if (!entries.length) return null;
    var contents = entries.map(function (e) { return e.content || {}; });
    var n = entries.length + (state.complete ? '' : '+');
    var items = [stat('Entries shown', n)];
    var times;
    switch (type) {
      case 'request':
      case 'client_request': {
        times = contents.map(function (c) { return Number(c.duration) || 0; });
        var errors = contents.filter(function (c) { return c.error || Number(c.response_status) >= 500; }).length;
        var clientErrors = contents.filter(function (c) { var s = Number(c.response_status); return s >= 400 && s < 500; }).length;
        items.push(stat('Server errors', errors, null, errors ? 'danger' : 'success'), stat('Client errors', clientErrors, null, clientErrors ? 'warning' : ''),
          stat('Average time', ms(avg(times))), stat('Slowest', ms(Math.max.apply(null, times)), null, Math.max.apply(null, times) >= 1000 ? 'warning' : ''));
        break;
      }
      case 'query': {
        times = contents.map(function (c) { return Number(c.time) || 0; });
        var slow = contents.filter(function (c) { return c.slow; }).length;
        items.push(stat('Slow queries', slow, null, slow ? 'warning' : 'success'), stat('Average time', ms(avg(times))),
          stat('Total time', ms(times.reduce(function (a, b) { return a + b; }, 0))));
        break;
      }
      case 'log': {
        var levels = {};
        contents.forEach(function (c) { levels[c.level || 'log'] = (levels[c.level || 'log'] || 0) + 1; });
        var errs = (levels.error || 0) + (levels.fatal || 0) + (levels.panic || 0);
        items.push(stat('Errors', errs, null, errs ? 'danger' : 'success'), stat('Warnings', levels.warning || 0, null, levels.warning ? 'warning' : ''), stat('Info', levels.info || 0));
        break;
      }
      case 'exception': {
        var total = contents.reduce(function (a, c) { return a + (Number(c.occurrences) || 1); }, 0);
        items = [stat('Distinct exceptions', entries.length), stat('Occurrences', total, null, 'danger')];
        break;
      }
      case 'job': {
        var failed = contents.filter(function (c) { return c.status === 'failed'; }).length;
        items.push(stat('Failed', failed, null, failed ? 'danger' : 'success'), stat('Processed', contents.filter(function (c) { return c.status === 'processed'; }).length));
        break;
      }
      case 'cache': {
        var hits = contents.filter(function (c) { return c.type === 'hit'; }).length, misses = contents.filter(function (c) { return c.type === 'missed'; }).length;
        items.push(stat('Hits', hits, null, 'success'), stat('Misses', misses, null, misses ? 'warning' : ''),
          stat('Hit rate', hits + misses ? Math.round(hits / (hits + misses) * 100) : 0, '%'));
        break;
      }
      case 'gate': {
        var denied = contents.filter(function (c) { return c.result === 'denied'; }).length;
        items.push(stat('Allowed', entries.length - denied, null, 'success'), stat('Denied', denied, null, denied ? 'warning' : ''));
        break;
      }
      default:
        return null;
    }
    return h('div', { class: 'stats' }, items);
  }

  function renderNav() {
    var nav = document.getElementById('nav');
    clear(nav);
    var counts = (state.status && state.status.counts) || {};
    var watchers = (state.status && state.status.watchers) || {};
    var group = null;
    SCREENS.forEach(function (s) {
      if (s.group !== group) { group = s.group; nav.appendChild(h('div', { class: 'nav-label' }, group)); }
      var off = watchers[s.type] === false;
      var active = state.screen === s.slug;
      nav.appendChild(h('a', {
        href: '#/' + s.slug,
        class: (active ? 'active' : '') + (off ? ' off' : ''),
        'aria-current': active ? 'page' : null,
        title: off ? s.title + ' (watcher turned off)' : null
      }, icon(s.slug), h('span', { class: 'label' }, s.title),
      off ? h('span', { class: 'off-tag' }, 'off') : counts[s.type] ? h('span', { class: 'count' }, counts[s.type] > 999 ? '999+' : counts[s.type]) : null));
    });
    nav.appendChild(h('div', { class: 'nav-label' }, 'Tools'));
    nav.appendChild(h('a', { href: '#/monitoring', class: state.screen === 'monitoring' ? 'active' : '', 'aria-current': state.screen === 'monitoring' ? 'page' : null },
      icon('monitoring'), h('span', { class: 'label' }, 'Monitoring')));
  }

  function renderToolbar() {
    var on = state.status ? state.status.recording : true;
    var rec = document.getElementById('recording');
    rec.className = 'tool recording';
    rec.setAttribute('aria-pressed', String(on));
    clear(rec);
    append(rec, [icon(on ? 'pause' : 'play'), h('span', { class: 'tool-label' }, on ? 'Recording' : 'Paused')]);
    rec.setAttribute('aria-label', on ? 'Pause recording' : 'Resume recording');

    var ref = document.getElementById('refresh');
    ref.setAttribute('aria-pressed', String(state.autoRefresh));
    ref.classList.toggle('spinning', state.autoRefresh);
    clear(ref);
    append(ref, [icon('refresh'), h('span', { class: 'tool-label' }, state.autoRefresh ? 'Live' : 'Auto-refresh')]);
    ref.setAttribute('aria-label', state.autoRefresh ? 'Stop auto-refresh' : 'Start auto-refresh');
  }

  function renderStatic() {
    var clearBtn = document.getElementById('clear');
    append(clearBtn, [icon('trash'), h('span', { class: 'tool-label' }, 'Clear')]);
    clearBtn.setAttribute('aria-label', 'Delete all entries');
    document.getElementById('menu').appendChild(icon('menu'));
    document.getElementById('close-menu').appendChild(icon('x'));
    document.getElementById('search-icon').appendChild(icon('search'));
    renderTheme();
  }

  function currentTheme() {
    try { return localStorage.getItem('goralens-theme') || 'system'; } catch (e) { return 'system'; }
  }

  function renderTheme() {
    var t = currentTheme();
    var btn = document.getElementById('theme');
    clear(btn);
    btn.appendChild(icon(t === 'light' ? 'sun' : t === 'dark' ? 'moon' : 'system'));
    btn.title = 'Theme: ' + t;
    btn.setAttribute('aria-label', 'Theme: ' + t + ' (click to change)');
  }

  function setTitle(screen, subtitle) {
    var title = document.getElementById('title');
    clear(title);
    append(title, [icon(screen.slug || 'monitoring'), screen.title]);
    document.getElementById('subtitle').textContent = subtitle || '';
    document.title = screen.title + ' · Goralens';
  }

  function cellValue(v) {
    if (v === null || v === undefined || v === '') return h('span', { class: 'muted' }, '—');
    if (typeof v === 'object') return h('span', { class: 'mono' }, short(JSON.stringify(v), 120));
    return String(v);
  }

  function registerCustom(list) {
    (list || []).forEach(function (w) {
      if (byType[w.type]) return;
      var slug = w.type.replace(/_/g, '-');
      if (bySlug[slug] || slug === 'monitoring') slug = 'custom-' + slug;
      var screen = { type: w.type, slug: slug, title: w.title || w.type, group: w.group || 'Custom', hint: w.hint || 'Entries recorded by the ' + (w.title || w.type) + ' watcher appear here.' };
      SCREENS.push(screen);
      bySlug[slug] = screen;
      byType[w.type] = screen;
      if (w.columns && w.columns.length) {
        COLUMNS[w.type] = w.columns.map(function (col, i) {
          return [col.label, function (c) { return cellValue(c[col.field]); }, i === 0 ? 'truncate' : 'nowrap', i === 0];
        });
      }
    });
  }

  function loadStatus() {
    return request('GET', '/status').then(function (s) {
      state.status = s;
      registerCustom(s.custom);
      renderNav();
      renderToolbar();
    }).catch(function (err) {
      if (err && (err.status === 401 || err.status === 403)) {
        state.denied = true;
        showDenied();
      }
    });
  }

  function skeleton(cols) {
    var rows = [];
    for (var i = 0; i < 6; i++) {
      var cells = [];
      for (var j = 0; j < cols; j++) cells.push(h('td', null, h('span', { style: 'width:' + (j === 1 ? 70 : 40 + (i * 7 % 30)) + '%' })));
      rows.push(h('tr', { class: 'skeleton' }, cells));
    }
    return h('div', { class: 'table-wrap' }, h('table', null, h('tbody', null, rows)));
  }

  function emptyState(screen) {
    var off = state.status && state.status.watchers && state.status.watchers[screen.type] === false;
    if (state.tag) {
      return h('div', { class: 'empty' }, h('div', { class: 'empty-icon' }, icon('tag')), h('h3', null, 'No entries tagged “' + state.tag + '”'),
        h('p', null, 'Nothing on this screen carries that tag. Try another screen or clear the filter.'),
        h('button', { class: 'button', type: 'button', onclick: function () { filterByTag(''); } }, icon('x'), 'Clear filter'));
    }
    if (off) {
      return h('div', { class: 'empty' }, h('div', { class: 'empty-icon' }, icon(screen.slug)), h('h3', null, screen.title + ' watcher is off'),
        h('p', null, 'Turn it on in ', h('code', null, 'config/goralens.go'), ' under ', h('code', null, 'watchers'), ' to record ' + screen.title.toLowerCase() + '.'));
    }
    return h('div', { class: 'empty' }, h('div', { class: 'empty-icon' }, icon(screen.slug)), h('h3', null, 'No ' + screen.title.toLowerCase() + ' yet'),
      h('p', null, screen.hint + ' New entries show up here as they are recorded.'),
      state.autoRefresh ? null : h('button', { class: 'button', type: 'button', onclick: function () { setAutoRefresh(true); } }, icon('refresh'), 'Turn on auto-refresh'));
  }

  function showList(screen) {
    setTitle(screen, state.tag ? 'Filtered by tag “' + state.tag + '”' : screen.hint);
    clear(view);
    var statsBox = h('div');
    var container = card(null, skeleton((COLUMNS[screen.type] || []).length + 1));
    container.id = 'list-card';
    if (state.status && state.status.watchers && state.status.watchers[screen.type] === false) {
      view.appendChild(h('div', { class: 'notice', role: 'note' }, icon('alert'), h('div', null, 'The ' + screen.title + ' watcher is turned off in ', h('code', null, 'config/goralens.go'), '. Older entries are still shown.')));
    }
    view.appendChild(statsBox);
    view.appendChild(container);
    state.entries = [];
    state.complete = false;
    fetchEntries(screen, null).then(function (entries) {
      state.entries = entries;
      state.complete = entries.length < PAGE;
      drawList(container, statsBox, screen);
    }).catch(function (err) {
      clear(container);
      container.appendChild(h('div', { class: 'empty' }, h('div', { class: 'empty-icon' }, icon('alert')), h('h3', null, 'Could not load entries'), h('p', null, err.message)));
    });
  }

  function fetchEntries(screen, before) {
    var params = ['type=' + encodeURIComponent(screen.type), 'take=' + PAGE];
    if (before) params.push('before=' + before);
    if (state.tag) params.push('tag=' + encodeURIComponent(state.tag));
    return request('GET', '/entries?' + params.join('&')).then(function (d) { return d.entries || []; });
  }

  function drawList(container, statsBox, screen) {
    clear(container);
    clear(statsBox);
    if (!state.entries.length) { container.appendChild(emptyState(screen)); return; }
    var s = statsFor(screen.type, state.entries);
    if (s) statsBox.appendChild(s);
    container.appendChild(entryTable(state.entries, screen.type));
    if (!state.complete) {
      var more = h('button', { class: 'button', type: 'button', onclick: function () {
        more.disabled = true;
        more.textContent = 'Loading…';
        var last = state.entries[state.entries.length - 1];
        fetchEntries(screen, last.sequence).then(function (older) {
          state.entries = state.entries.concat(older);
          state.complete = older.length < PAGE;
          drawList(container, statsBox, screen);
        });
      } }, 'Load older entries');
      container.appendChild(h('div', { class: 'more-row' }, more));
    }
  }

  function refreshList() {
    loadStatus();
    var screen = bySlug[state.screen];
    if (!screen || window.location.hash.split('/').length > 2) return;
    fetchEntries(screen, null).then(function (entries) {
      var newest = state.entries.length ? state.entries[0].sequence : 0;
      var fresh = entries.filter(function (e) { return e.sequence > newest; });
      if (!fresh.length) return;
      state.entries = fresh.concat(state.entries);
      var container = document.getElementById('list-card');
      if (container) drawList(container, container.previousSibling, screen);
    }).catch(function () {});
  }

  function hero(e, parts) {
    return h('section', { class: 'card hero' },
      h('div', { class: 'hero-top' }, parts.badges),
      parts.title ? h('h2', { class: 'hero-title' + (parts.titleClass ? ' ' + parts.titleClass : '') }, parts.title) : null,
      parts.sub ? h('p', { class: 'hero-sub' }, parts.sub) : null,
      h('div', { class: 'hero-meta' },
        h('span', { title: fullTime(e.created_at) }, icon('clock'), h('strong', null, timeAgo(e.created_at)), fullTime(e.created_at)),
        (parts.meta || []).map(function (m) { return m && m[1] !== undefined && m[1] !== null && m[1] !== '' ? h('span', null, icon(m[2] || 'hash'), m[0], ' ', h('strong', null, m[1])) : null; })));
  }

  function showEntry(screen, uuid) {
    setTitle(screen, 'Entry details');
    clear(view);
    view.appendChild(card(null, skeleton(3)));
    request('GET', '/entries/' + encodeURIComponent(uuid)).then(function (data) {
      clear(view);
      var e = data.entry;
      view.appendChild(h('a', { class: 'back', href: '#/' + screen.slug }, icon('back'), 'Back to ' + screen.title.toLowerCase()));
      (DETAILS[e.type] || genericDetails)(e, data).forEach(function (node) { if (node) view.appendChild(node); });
      if (e.tags && e.tags.length) view.appendChild(card('Tags', h('div', { class: 'card-body group-chips' }, tagChips(e.tags)), { icon: 'tag' }));
      if (data.batch && data.batch.length) view.appendChild(relatedEntries(e, data.batch));
      view.focus({ preventScroll: true });
    }).catch(function (err) {
      clear(view);
      view.appendChild(card(null, h('div', { class: 'empty' }, h('div', { class: 'empty-icon' }, icon('alert')), h('h3', null, 'Entry not available'),
        h('p', null, err.message), h('a', { class: 'button', href: '#/' + screen.slug }, icon('back'), 'Back to ' + screen.title.toLowerCase()))));
    });
  }

  function relatedEntries(entry, batch) {

    var sorted = batch.slice().sort(function (a, b) { return a.sequence - b.sequence; });
    var start = Math.min.apply(null, sorted.concat([entry]).map(function (e) { return new Date(e.created_at).getTime(); }));
    var counts = {};
    sorted.forEach(function (e) { counts[e.type] = (counts[e.type] || 0) + 1; });
    var chips = h('div', { class: 'group-chips' }, Object.keys(counts).map(function (t) {
      return badge((byType[t] ? byType[t].title : t) + ' ' + counts[t], 'accent');
    }));
    var list = h('ol', { class: 'timeline' }, sorted.map(function (e) {
      var sev = severity(e);
      var offset = new Date(e.created_at).getTime() - start;
      var open = function () { window.location.hash = entryLink(e); };
      var slug = byType[e.type] ? byType[e.type].slug : 'file';
      return h('li', { tabindex: '0', onclick: open, onkeydown: function (ev) { if (ev.key === 'Enter') open(); }, 'aria-label': (byType[e.type] ? byType[e.type].title : e.type) + ': ' + summaryText(e) },
        h('span', { class: 'node' + (sev ? ' ' + sev : '') }, icon(slug)),
        h('span', { class: 'kind' }, byType[e.type] ? byType[e.type].title : e.type),
        h('span', { class: 'what' + (e.type === 'query' || e.type === 'redis' ? ' mono' : '') }, summaryText(e)),
        h('span', { class: 'when', title: fullTime(e.created_at) }, '+' + ms(offset)));
    }));
    return h('section', { class: 'card' },
      h('div', { class: 'card-head' }, h('h2', null, icon('layers'), 'Related entries'), h('span', { class: 'meta' }, batch.length + ' in the same batch')),
      h('div', { class: 'card-body', style: 'padding-bottom:0' }, chips), list);
  }

  function traceList(trace) {
    if (!trace || !trace.length) return null;
    var vendor = /\/(go\/pkg\/mod|goroot|usr\/local\/go|libexec)\//;
    var items = [];
    var hidden = [];
    function frame(f) {
      var app = f.file && !vendor.test(f.file) && (f.function || '').indexOf('runtime.') !== 0;
      var dir = String(f.file || '').split('/');
      var name = dir.pop();
      return h('li', { class: app ? 'app' : null },
        h('div', null, h('div', { class: 'fn' }, f.function || '(unknown)'),
          h('div', { class: 'loc' }, dir.length ? dir.join('/') + '/' : '', h('b', null, name), ':', h('span', { class: 'line' }, f.line))));
    }

    function flush() {
      if (!hidden.length) return;
      if (hidden.length < 3) { hidden.forEach(function (f) { items.push(frame(f)); }); hidden = []; return; }
      var group = hidden.map(frame);
      var toggle = h('li', { class: 'vendor-toggle' }, h('button', { class: 'button small ghost', type: 'button', 'aria-expanded': 'false', onclick: function () {
        group.forEach(function (li) { li.hidden = !li.hidden; });
        var open = !group[0].hidden;
        this.setAttribute('aria-expanded', String(open));
        this.textContent = (open ? 'Hide ' : 'Show ') + group.length + ' framework frames';
      } }, 'Show ' + group.length + ' framework frames'));
      group.forEach(function (li) { li.hidden = true; });
      items.push(toggle);
      items = items.concat(group);
      hidden = [];
    }
    trace.forEach(function (f) {
      var app = f.file && !vendor.test(f.file) && (f.function || '').indexOf('runtime.') !== 0;
      if (app) { flush(); items.push(frame(f)); } else hidden.push(f);
    });
    flush();
    return h('ol', { class: 'trace' }, items);
  }

  function genericDetails(e) { return [hero(e, { badges: badge(e.type, 'accent') }), card('Content', codeBlock(e.content))]; }

  function mono(v) { return v ? h('span', { class: 'mono' }, v) : null; }

  var DETAILS = {
    request: function (e) {
      var c = e.content;
      return [hero(e, { badges: [methodBadge(c.method), statusBadge(c.response_status, true), duration(c.duration)], title: c.uri, titleClass: 'mono',
        meta: [['Duration', ms(c.duration), 'clock'], ['IP', c.ip, 'server'], ['User', c.user ? (c.user.name || c.user.email || '') + ' #' + c.user.id : null, 'user']] }),
      props([['Route', mono(c.route)], ['Handler', mono(c.controller_action)], ['Hostname', c.hostname], ['Memory', c.memory ? c.memory + ' MB' : null]]),
      tabs([['Payload', c.payload], ['Headers', c.headers], ['Session', c.session]], 'Request'),
      tabs([['Response', c.response], ['Response headers', c.response_headers]], 'Response')];
    },
    client_request: function (e) {
      var c = e.content;
      return [hero(e, { badges: [methodBadge(c.method), c.error ? badge('error', 'danger', 'lg') : statusBadge(c.response_status, true), duration(c.duration, 500, 2000)],
        title: c.uri, titleClass: 'mono', sub: c.error }),
      tabs([['Payload', c.payload], ['Headers', c.headers]], 'Request'),
      c.error ? null : tabs([['Response', c.response], ['Response headers', c.response_headers]], 'Response')];
    },
    command: function (e) {
      var c = e.content;
      return [hero(e, { badges: [badge('exit ' + c.exit_code, c.exit_code ? 'danger' : 'success', 'lg'), duration(c.duration, 5000, 30000)], title: c.command, titleClass: 'mono', sub: c.error }),
        card('Arguments', codeBlock(c.arguments || []), { icon: 'commands' })];
    },
    schedule: function (e) {
      var c = e.content;
      return [hero(e, { badges: [resultBadge(c.status), c.duration !== undefined ? duration(c.duration, 5000, 30000) : null], title: c.name || c.command, titleClass: 'mono', sub: c.error,
        meta: [['Expression', c.expression, 'clock'], ['Command', c.command, 'commands']] })];
    },
    job: function (e) {
      var c = e.content;
      return [hero(e, { badges: [resultBadge(c.status), c.duration !== undefined ? duration(c.duration, 1000, 10000) : null], title: c.name, titleClass: 'mono', sub: c.error,
        meta: [['Connection', c.connection, 'server'], ['Queue', c.queue, 'jobs'], ['Delayed until', c.delay_until, 'clock']] }),
      card('Data', codeBlock(c.data || []), { icon: 'jobs' })];
    },
    exception: function (e, data) {
      var c = e.content;
      var occurrences = data.occurrences || 1;
      var nodes = [hero(e, {
        badges: [badge('Exception', 'danger', 'lg'), occurrences > 1 ? h('a', { href: '#/exceptions', onclick: function () { filterByTag(c.class); } }, badge(occurrences + ' occurrences', 'danger')) : null],
        title: c.class, titleClass: 'mono danger', sub: c.message,
        meta: [['At', c.file ? fileLine(c.file, c.line) : null, 'file']] })];
      if (c.file) nodes.push(props([['File', mono(c.file + ':' + c.line)]]));
      if (c.trace && c.trace.length) nodes.push(card('Stack trace', traceList(c.trace), { icon: 'layers', meta: c.trace.length + ' frames' }));
      if (c.context && Object.keys(c.context).length) nodes.push(card('Context', codeBlock(c.context), { icon: 'hash' }));
      return nodes;
    },
    log: function (e) {
      var c = e.content;
      return [hero(e, { badges: [levelBadge(c.level)], title: c.message,
        meta: [['Code', c.code, 'hash'], ['Domain', c.domain, 'tag'], ['Hint', c.hint, 'alert']] }),
      c.context && Object.keys(c.context).length ? card('Context', codeBlock(c.context), { icon: 'hash' }) : null,
      c.request ? card('Request', codeBlock(c.request), { icon: 'requests' }) : null];
    },
    dump: function (e) {
      var c = e.content;
      return [hero(e, { badges: [badge(c.type || 'value', 'violet')], meta: [['At', c.file ? fileLine(c.file, c.line) : null, 'file']] }),
        card('Dump', codeBlock(c.dump), { icon: 'dumps' })];
    },
    query: function (e) {
      var c = e.content;
      return [hero(e, { badges: [c.slow ? badge('slow query', 'warning', 'lg') : badge('query', 'accent', 'lg'), duration(c.time, 100, 500), c.error ? badge('failed', 'danger') : null],
        sub: c.error, meta: [['Connection', c.connection, 'queries'], ['Rows', c.rows !== undefined ? String(c.rows) : null, 'hash'], ['At', c.file ? fileLine(c.file, c.line) : null, 'file']] }),
      card('Query', codeBlock(c.sql, 'sql'), { icon: 'queries' }),
      c.bindings && c.bindings.length ? card('Bindings', codeBlock(c.bindings), { icon: 'hash' }) : null];
    },
    model: function (e) {
      var c = e.content;
      return [hero(e, { badges: [resultBadge(c.action)], title: c.model + (c.key !== null && c.key !== undefined ? ':' + c.key : ''), titleClass: 'mono',
        meta: [['Table', c.table, 'queries'], ['Rows', c.rows !== undefined ? String(c.rows) : null, 'hash']] }),
      c.changes ? card('Changes', codeBlock(c.changes), { icon: 'models' }) : null];
    },
    event: function (e) {
      var c = e.content;
      return [hero(e, { badges: [badge((c.listeners || []).length + ' listeners', 'accent'), c.duration !== undefined ? duration(c.duration) : null], title: c.name, titleClass: 'mono', sub: c.error }),
        card('Payload', codeBlock(c.payload || []), { icon: 'events' }), card('Listeners', codeBlock(c.listeners || []), { icon: 'layers' })];
    },
    mail: function (e) {
      var c = e.content;
      var from = c.from && c.from.Address ? (c.from.Name ? c.from.Name + ' <' + c.from.Address + '>' : c.from.Address) : null;
      var nodes = [hero(e, { badges: [c.error ? badge('failed', 'danger', 'lg') : c.queued ? badge('queued', 'info', 'lg') : badge('sent', 'success', 'lg')], title: c.subject || '(no subject)', sub: c.error,
        meta: [['From', from, 'user'], ['To', (c.to || []).join(', '), 'mail']] }),
      props([['Mailable', mono(c.mailable)], ['Cc', (c.cc || []).join(', ')], ['Bcc', (c.bcc || []).join(', ')]])];
      if (c.html) {

        var frame = h('iframe', { sandbox: '', class: 'mail-frame', title: 'Mail preview' });
        frame.srcdoc = c.html;
        nodes.push(card('Preview', frame, { icon: 'mail' }));
      }
      nodes.push(tabs([['Text', c.text ? codeBlock(c.text, 'text') : null], ['HTML source', c.html ? codeBlock(c.html, 'text') : null],
        ['Attachments', c.attachments], ['Headers', c.headers]], 'Message'));
      return nodes;
    },
    cache: function (e) {
      var c = e.content;
      return [hero(e, { badges: [resultBadge(c.type)], title: c.key, titleClass: 'mono', meta: [['Store', c.store, 'cache'], ['Expires in', c.expiration ? c.expiration + 's' : null, 'clock']] }),
        c.value !== undefined ? card('Value', codeBlock(c.value), { icon: 'cache' }) : null];
    },
    redis: function (e) {
      var c = e.content;
      return [hero(e, { badges: [duration(c.time, 50, 250), c.error ? badge('failed', 'danger') : null], sub: c.error, meta: [['Connection', c.connection, 'redis']] }),
        card('Command', codeBlock(c.command, 'text'), { icon: 'redis' })];
    },
    gate: function (e) {
      var c = e.content;
      return [hero(e, { badges: [resultBadge(c.result)], title: c.ability, titleClass: 'mono', sub: c.message, meta: [['At', c.file ? fileLine(c.file, c.line) : null, 'file']] }),
        card('Arguments', codeBlock(c.arguments || {}), { icon: 'gates' })];
    }
  };

  function showMonitoring() {
    setTitle({ slug: 'monitoring', title: 'Monitoring' }, 'Entries with a monitored tag are always recorded, even when a filter would drop them.');
    clear(view);
    var list = h('ul', { class: 'monitor-list' });
    var input = h('input', { type: 'text', placeholder: 'Tag, e.g. Auth:1, slow or failed', 'aria-label': 'Tag to monitor' });
    var form = h('form', { class: 'monitor-form', onsubmit: function (ev) {
      ev.preventDefault();
      var tag = input.value.trim();
      if (!tag) return;
      request('POST', '/monitored-tags', { tags: [tag] }).then(function (d) { input.value = ''; draw(d.tags); toast('Monitoring “' + tag + '”'); });
    } }, input, h('button', { class: 'button primary', type: 'submit' }, icon('monitoring'), 'Monitor'));
    view.appendChild(card('Monitored tags', [form, list], { icon: 'monitoring' }));

    function draw(tags) {
      clear(list);
      if (!tags.length) {
        list.appendChild(h('li', { style: 'display:block' }, h('div', { class: 'empty', style: 'padding:36px 16px' }, h('div', { class: 'empty-icon' }, icon('tag')),
          h('h3', null, 'No tags are monitored'), h('p', null, 'Add a tag above to always record matching entries.'))));
        return;
      }
      tags.forEach(function (t) {
        list.appendChild(h('li', null, h('button', { class: 'chip', type: 'button', onclick: function () { filterByTag(t, 'requests'); } }, icon('tag'), t),
          h('button', { class: 'button small danger', type: 'button', onclick: function () {
            request('DELETE', '/monitored-tags', { tags: [t] }).then(function (d) { draw(d.tags); });
          } }, icon('x'), 'Stop monitoring')));
      });
    }
    request('GET', '/monitored-tags').then(function (d) { draw(d.tags || []); });
  }

  function setMenu(open) {
    var sidebar = document.getElementById('sidebar');
    sidebar.classList.toggle('open', open);
    document.getElementById('scrim').hidden = !open;
    document.getElementById('menu').setAttribute('aria-expanded', String(open));
    if (open) { var first = sidebar.querySelector('.nav a'); if (first) first.focus(); }
  }

  function showDenied() {
    clear(view);
    document.getElementById('title').textContent = 'Goralens';
    document.getElementById('subtitle').textContent = '';
    view.appendChild(card(null, h('div', { class: 'empty' }, h('div', { class: 'empty-icon' }, icon('alert')),
      h('h3', null, 'You do not have access to Goralens'),
      h('p', null, 'Sign in with an account that is allowed to open the dashboard, then reload this page.'))));
  }

  function route() {
    if (state.denied) return showDenied();
    var parts = window.location.hash.replace(/^#\/?/, '').split('/');
    var slug = parts[0] || 'requests';
    state.screen = slug;
    setMenu(false);
    renderNav();
    window.scrollTo(0, 0);
    if (slug === 'monitoring') return showMonitoring();
    var screen = bySlug[slug] || bySlug.requests;
    state.screen = screen.slug;
    if (parts[1]) showEntry(screen, parts[1]);
    else showList(screen);
  }

  function filterByTag(tag, slug) {
    state.tag = tag;
    document.getElementById('tag').value = tag;
    var target = slug || (bySlug[state.screen] ? state.screen : 'requests');
    if (window.location.hash === '#/' + target) route();
    else window.location.hash = '#/' + target;
  }

  function setAutoRefresh(on) {
    state.autoRefresh = on;
    clearInterval(state.timer);
    if (on) state.timer = setInterval(refreshList, 3000);
    try { localStorage.setItem('goralens-refresh', on ? '1' : '0'); } catch (e) {}
    renderToolbar();
  }

  var tagInput = document.getElementById('tag');
  var tagTimer;
  tagInput.addEventListener('input', function () {
    clearTimeout(tagTimer);
    tagTimer = setTimeout(function () {
      state.tag = tagInput.value.trim();
      var screen = bySlug[state.screen];
      if (screen) {
        if (window.location.hash.split('/').length > 2) window.location.hash = '#/' + screen.slug;
        else showList(screen);
      }
    }, 300);
  });

  document.addEventListener('keydown', function (ev) {
    var typing = /^(INPUT|TEXTAREA|SELECT)$/.test(document.activeElement.tagName);
    if (ev.key === '/' && !typing) { ev.preventDefault(); tagInput.focus(); }
    if (ev.key === 'Escape') {
      if (document.getElementById('sidebar').classList.contains('open')) setMenu(false);
      else if (document.activeElement === tagInput) tagInput.blur();
    }
  });

  document.getElementById('refresh').addEventListener('click', function () { setAutoRefresh(!state.autoRefresh); });

  document.getElementById('recording').addEventListener('click', function () {
    var on = !(state.status && state.status.recording);
    request('POST', '/recording', { recording: on }).then(function () { toast(on ? 'Recording resumed' : 'Recording paused'); return loadStatus(); });
  });

  document.getElementById('clear').addEventListener('click', function () {
    if (!window.confirm('Delete all recorded entries?')) return;
    request('DELETE', '/entries').then(function () { toast('All entries deleted'); loadStatus(); route(); });
  });

  document.getElementById('theme').addEventListener('click', function () {
    var order = ['system', 'light', 'dark'];
    var next = order[(order.indexOf(currentTheme()) + 1) % order.length];
    try { localStorage.setItem('goralens-theme', next); } catch (e) {}
    if (next === 'system') document.documentElement.removeAttribute('data-theme');
    else document.documentElement.setAttribute('data-theme', next);
    renderTheme();
  });

  document.getElementById('menu').addEventListener('click', function () { setMenu(true); });
  document.getElementById('close-menu').addEventListener('click', function () { setMenu(false); });
  document.getElementById('scrim').addEventListener('click', function () { setMenu(false); });

  renderStatic();
  renderToolbar();
  window.addEventListener('hashchange', route);
  loadStatus().then(route);
  var saved = '0';
  try { saved = localStorage.getItem('goralens-refresh') || '0'; } catch (e) {}
  setAutoRefresh(saved === '1');
})();
