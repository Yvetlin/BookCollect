const drop = document.getElementById('drop');
const dropText = document.getElementById('dropText');
const fileInput = document.getElementById('file');
const fileName = document.getElementById('fileName');
const form = document.getElementById('articleForm');
const progress = document.getElementById('progress');
const percent = document.getElementById('percent');
const alertBox = document.getElementById('alert');
const submitBtn = document.getElementById('submitBtn');

function showAlert(text, type) {
    alertBox.style.display = 'block';
    alertBox.textContent = text;
    alertBox.style.borderColor = type === 'error' ? '#ef4444' : '#22c55e';
    alertBox.style.background = type === 'error'
    ? 'color-mix(in oklab, #ef4444, transparent 85%)'
    : 'color-mix(in oklab, #22c55e, transparent 85%)';
}

function clearAlert() {
    alertBox.style.display = 'none';
    alertBox.textContent = '';
}

function onFilePicked(f) {
    if (!f) return;
    fileName.style.display = 'block';
    fileName.textContent = `Файл: ${f.name} (${Math.ceil(f.size/1024)} КБ)`;
}

drop.addEventListener('click', ()=> fileInput.click());
fileInput.addEventListener('change', ()=> onFilePicked(fileInput.files[0]));

drop.addEventListener('dragover', (e)=>{ e.preventDefault(); drop.style.opacity=.9; });
drop.addEventListener('dragleave', ()=> drop.style.opacity=1);
drop.addEventListener('drop', (e)=>{
e.preventDefault(); drop.style.opacity=1;
if (e.dataTransfer.files.length) {
    fileInput.files = e.dataTransfer.files;
    onFilePicked(fileInput.files[0]);
}
});

form.addEventListener('submit', (e)=>{
e.preventDefault();
clearAlert();

const fd = new FormData(form);
const xhr = new XMLHttpRequest();
xhr.open('POST', form.action, true);
xhr.responseType = 'json';

    // блокируем кнопку
submitBtn.disabled = true;
submitBtn.textContent = 'Отправка...';

    // прогресс
progress.style.display = 'block';
percent.style.display = 'block';
percent.textContent = '0%';

xhr.upload.onprogress = (e)=>{
if (e.lengthComputable) {
    const p = Math.round((e.loaded/e.total)*100);
    progress.value = p;
    percent.textContent = p + '%';
} else {
    // если сервер не отдаёт длину — показываем индикатор в неопред. режиме
    progress.removeAttribute('value');
    percent.textContent = 'Загрузка...';
}
};

xhr.onload = ()=>{
submitBtn.disabled = false;
submitBtn.textContent = 'Отправить';
progress.style.display = 'none';
percent.style.display = 'none';

const status = xhr.status;
const body = xhr.response || (function(){ try{ return JSON.parse(xhr.responseText) }catch{ return {} } })();

if (status >= 200 && status < 300) {
    const id = body && body.id ? ` ID заявки: ${body.id}.` : '';
    showAlert('Заявка отправлена.' + id, 'success');
    form.reset();
    fileName.style.display = 'none';
    progress.value = 0;
} else {
    const msg = (body && body.error) ? body.error : 'Ошибка отправки.';
    showAlert(msg, 'error');
}
};

xhr.onerror = ()=>{
    submitBtn.disabled = false;
    submitBtn.textContent = 'Отправить';
    progress.style.display = 'none';
    percent.style.display = 'none';
    showAlert('Сетевая ошибка.', 'error');
};

    xhr.send(fd);
});