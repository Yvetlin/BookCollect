function qs(s, r=document){ return r.querySelector(s); }
async function jsonFetch(url, opts={}) {
    const res = await fetch(url, {credentials:'same-origin', ...opts});
    const ct  = res.headers.get('content-type') || '';
    const body = ct.includes('application/json') ? await res.json() : await res.text();
    if (!res.ok) throw new Error((body && body.error) ? body.error : (typeof body==='string'? body : 'Ошибка запроса'));
    return body;
}

/* ====== СБОРНИКИ ====== */
window.initAdminCollections = function(){
    const T = qs('#tbl tbody');
    const btnNew = qs('#btnNew');
    const dlg = qs('#dlg');
    const frm = qs('#frm');
    const alert = qs('#alert');

    function escapeHtml(s){ return (s||'').replace(/[&<>"']/g, m=>({ '&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;' }[m])); }
    function row(c){
        const year = c.release_year ?? '';
        const num  = c.release_number ?? '';
        return `<tr>
      <td style="padding:8px; border-top:1px solid var(--border)">${c.id}</td>
      <td style="padding:8px; border-top:1px solid var(--border)">${escapeHtml(c.title)}</td>
      <td style="padding:8px; border-top:1px solid var(--border)">${year}${num?(' / № '+num):''}</td>
      <td style="padding:8px; border-top:1px solid var(--border)">
        <button class="btn btn-ghost" data-edit="${c.id}">Редактировать</button>
        <button class="btn btn-ghost" data-del="${c.id}">Удалить</button>
      </td>
    </tr>`;
    }

    async function load(){
        T.innerHTML = '';
        const list = await jsonFetch(window.ADMIN_CFG.listCollections);
        const items = list.collections || list || [];
        items.forEach(c => T.insertAdjacentHTML('beforeend', row(c)));
    }

    btnNew.addEventListener('click', ()=>{
        frm.reset(); qs('[name=id]').value = ''; alert.textContent = ''; dlg.showModal();
    });
    qs('#btnClose').addEventListener('click', ()=> dlg.close());

    T.addEventListener('click', async (e)=>{
        const t = e.target;
        const id = t.dataset.edit || t.dataset.del;
        if (!id) return;

        if (t.dataset.edit){
            const list = await jsonFetch(window.ADMIN_CFG.listCollections);
            const items = list.collections || list || [];
            const item = items.find(x => String(x.id) === String(id));
            if (!item) return;

            frm.reset();
            qs('[name=id]').value = item.id || '';
            qs('[name=title]').value = item.title || '';
            qs('[name=release_year]').value = item.release_year || '';
            qs('[name=release_number]').value = item.release_number || '';
            qs('[name=description]').value = item.description || '';
            qs('[name=publication_link]').value = item.publication_link || '';
            alert.textContent = '';
            dlg.showModal();
        }

        if (t.dataset.del){
            if (!confirm('Удалить сборник?')) return;
            await fetch(window.ADMIN_CFG.deleteCollection(id), { method:'DELETE' });
            await load();
        }
    });

    frm.addEventListener('submit', async (e)=>{
        e.preventDefault();
        alert.textContent = '';
        const id = qs('[name=id]').value.trim();
        const fd = new FormData(frm);

        try {
            if (!id){
                await fetch(window.ADMIN_CFG.createCollection, { method:'POST', body: fd }).then(r=>{ if(!r.ok) throw new Error('Ошибка сохранения'); });
            } else {
                // PUT + multipart нестабилен, поэтому используем POST + _method=PUT
                fd.append("_method", "PUT");
                await fetch(window.ADMIN_CFG.updateCollection(id), { method:'POST', body: fd }).then(r=>{ if(!r.ok) throw new Error('Ошибка сохранения'); });
            }
            dlg.close();
            await load();
        } catch(err){
            alert.textContent = err.message || 'Ошибка';
        }
    });

    load().catch(console.error);
};

/* ====== ЗАЯВКИ ====== */
window.initAdminArticles = function(){
    const T = qs('#tbl tbody');
    function escapeHtml(s){ return (s||'').replace(/[&<>"']/g, m=>({ '&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;' }[m])); }
    function row(a){
        const created = a.created_at ? new Date(a.created_at).toLocaleString() : '';
        return `<tr>
      <td style="padding:8px; border-top:1px solid var(--border)">${a.id}</td>
      <td style="padding:8px; border-top:1px solid var(--border)">${escapeHtml(a.author)}</td>
      <td style="padding:8px; border-top:1px solid var(--border)">${escapeHtml(a.title)}</td>
      <td style="padding:8px; border-top:1px solid var(--border)">${escapeHtml(a.email)}</td>
      <td style="padding:8px; border-top:1px solid var(--border)">${created}</td>
      <td style="padding:8px; border-top:1px solid var(--border)">
        <a class="btn btn-ghost" href="${window.ADMIN_CFG.downloadFile(a.id)}">Скачать</a>
        <button class="btn btn-ghost" data-del="${a.id}">Удалить</button>
      </td>
    </tr>`;
    }
    async function load(){
        T.innerHTML = '';
        const list = await jsonFetch(window.ADMIN_CFG.listArticles);
        const items = list.articles || list || [];
        items.forEach(a => T.insertAdjacentHTML('beforeend', row(a)));
    }
    T.addEventListener('click', async (e)=>{
        const id = e.target.dataset.del;
        if (!id) return;
        if (!confirm('Удалить заявку?')) return;
        await fetch(window.ADMIN_CFG.deleteArticle(id), { method:'DELETE' });
        await load();
    });
    load().catch(console.error);
};
