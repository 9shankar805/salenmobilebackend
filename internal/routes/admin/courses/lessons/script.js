const DEFAULT_MODULES = [
  {
    id: 'm1',
    title: 'Module 1: Power Management & Diagnostics',
    lessons: [
      { id: 'l1', title: 'Multimeter Diode Mode Testing Fundamentals', duration: '18m', isPreview: true, videoUrl: 'https://www.youtube.com/embed/dQw4w9WgXcQ', notes: 'Diode mode measures the voltage drop across semiconductor junctions to ground. Always put the RED probe on GND and the BLACK probe on the signal line.' },
      { id: 'l2', title: 'Finding VDD_MAIN & VDD_BOOST Short Circuits', duration: '28m', isPreview: false, videoUrl: 'https://www.youtube.com/embed/dQw4w9WgXcQ', notes: 'Use 1.0V current injection with thermal camera or rosin smoke technique to spot heating capacitors.' },
      { id: 'l3', title: 'DC Power Supply Boot Current Signatures', duration: '32m', isPreview: false, videoUrl: 'https://www.youtube.com/embed/dQw4w9WgXcQ', notes: 'Learn to distinguish between 0.05A stuck current (NAND/CPU), 0.15A loop (I2C bus crash), and high amp short.' }
    ]
  },
  {
    id: 'm2',
    title: 'Module 2: Advanced BGA & Micro-soldering',
    lessons: [
      { id: 'l4', title: 'Safe BGA IC Desoldering Temperature Curves', duration: '22m', isPreview: true, videoUrl: 'https://www.youtube.com/embed/dQw4w9WgXcQ', notes: 'Preheat board to 180°C with lower heat gun, then ramp up to 340°C on top with fine airflow.' },
      { id: 'l5', title: 'PMIC Ball Grid Array Reballing with Stencil', duration: '35m', isPreview: false, videoUrl: 'https://www.youtube.com/embed/dQw4w9WgXcQ', notes: 'Use 183°C medium melting solder paste. Scrape flat with blade and heat evenly.' },
      { id: 'l6', title: 'Under-fill Glue Removal without Tearing Pads', duration: '25m', isPreview: false, videoUrl: 'https://www.youtube.com/embed/dQw4w9WgXcQ', notes: 'Soak edge with resin softener and scrape gently at 220°C.' }
    ]
  },
  {
    id: 'm3',
    title: 'Module 3: Audio, Touch & Display Bus Repair',
    lessons: [
      { id: 'l7', title: 'Audio Codec C12 Pad Micro-Jumpering', duration: '30m', isPreview: false, videoUrl: 'https://www.youtube.com/embed/dQw4w9WgXcQ', notes: 'Run 0.02mm enameled copper wire from master clock test point to pad C12.' },
      { id: 'l8', title: 'Touch IC Bus Troubleshooting on Sandwich Boards', duration: '40m', isPreview: false, videoUrl: 'https://www.youtube.com/embed/dQw4w9WgXcQ', notes: 'Examine interposer solder balls for hairline micro-fractures after drop impacts.' }
    ]
  }
];

let modules = JSON.parse(localStorage.getItem('smg_curriculum_modules')) || DEFAULT_MODULES;
let activeLessonId = 'l1';

function saveCurriculum() {
  localStorage.setItem('smg_curriculum_modules', JSON.stringify(modules));
  updateKPIs();
}

function updateKPIs() {
  let totalLessons = 0;
  let previews = 0;
  modules.forEach(m => {
    totalLessons += m.lessons.length;
    m.lessons.forEach(l => {
      if (l.isPreview) previews++;
    });
  });

  document.getElementById('kpi-modules').innerText = modules.length;
  document.getElementById('kpi-lessons').innerText = totalLessons;
  document.getElementById('kpi-preview').innerText = previews;
  document.getElementById('tree-summary').innerText = `${modules.length} Modules &bull; ${totalLessons} Lessons`;
}

function renderTree() {
  const container = document.getElementById('module-accordion-container');
  container.innerHTML = modules.map((m, mIdx) => `
    <div class="module-item">
      <div class="module-header" onclick="toggleModuleAccordion('${m.id}')">
        <div class="module-title">
          <i class="fas fa-folder-open" style="color:var(--primary);font-size:12px"></i>
          <span>${m.title}</span>
        </div>
        <span class="badge badge-time">${m.lessons.length}</span>
      </div>
      <div class="lesson-list" id="lesson-list-${m.id}">
        ${m.lessons.map(l => `
          <div class="lesson-item ${l.id === activeLessonId ? 'active' : ''}" onclick="selectLesson('${l.id}')">
            <div class="lesson-title">
              <i class="fas fa-circle-play" style="font-size:11px"></i>
              <span>${l.title}</span>
            </div>
            <div style="display:flex;align-items:center;gap:6px">
              ${l.isPreview ? '<span class="badge badge-preview">Free</span>' : ''}
              <span class="badge badge-time">${l.duration}</span>
            </div>
          </div>
        `).join('')}
      </div>
    </div>
  `).join('');
}

function findLesson(id) {
  for (const m of modules) {
    for (const l of m.lessons) {
      if (l.id === id) return { lesson: l, module: m };
    }
  }
  return null;
}

function selectLesson(id) {
  activeLessonId = id;
  const res = findLesson(id);
  if (!res) return;

  const { lesson, module } = res;
  document.getElementById('active-module-title').innerText = module.title;
  document.getElementById('active-lesson-title').innerText = lesson.title;
  document.getElementById('lesson-title-input').value = lesson.title;
  document.getElementById('lesson-duration-input').value = lesson.duration;
  document.getElementById('lesson-video-url').value = lesson.videoUrl || '';
  document.getElementById('lesson-notes').value = lesson.notes || '';
  document.getElementById('lesson-preview-input').checked = !!lesson.isPreview;

  updateVideoPreview();
  renderTree();
}

function updateVideoPreview() {
  const url = document.getElementById('lesson-video-url').value.trim();
  const iframe = document.getElementById('video-iframe');
  if (url.includes('youtube.com/watch?v=')) {
    const vId = url.split('v=')[1]?.split('&')[0];
    iframe.src = `https://www.youtube.com/embed/${vId}`;
  } else if (url.includes('youtu.be/')) {
    const vId = url.split('youtu.be/')[1]?.split('?')[0];
    iframe.src = `https://www.youtube.com/embed/${vId}`;
  } else if (url.startsWith('http')) {
    iframe.src = url;
  }
}

function switchStudioTab(tab, btn) {
  document.querySelectorAll('.studio-tabs .tab-btn').forEach(b => b.classList.remove('active'));
  btn.classList.add('active');

  ['video', 'notes', 'attachments', 'settings'].forEach(t => {
    document.getElementById(`tab-pane-${t}`).style.display = (t === tab ? 'block' : 'none');
  });
}

function navLesson(direction) {
  const allLessons = [];
  modules.forEach(m => m.lessons.forEach(l => allLessons.push(l.id)));
  const curIdx = allLessons.indexOf(activeLessonId);
  const nextIdx = curIdx + direction;
  if (nextIdx >= 0 && nextIdx < allLessons.length) {
    selectLesson(allLessons[nextIdx]);
  }
}

function saveActiveLesson() {
  const res = findLesson(activeLessonId);
  if (!res) return;

  res.lesson.title = document.getElementById('lesson-title-input').value.trim();
  res.lesson.duration = document.getElementById('lesson-duration-input').value.trim();
  res.lesson.videoUrl = document.getElementById('lesson-video-url').value.trim();
  res.lesson.notes = document.getElementById('lesson-notes').value.trim();
  res.lesson.isPreview = document.getElementById('lesson-preview-input').checked;

  saveCurriculum();
  selectLesson(activeLessonId);
  alert('Lesson saved successfully!');
}

// Modals
function openAddModuleModal() {
  document.getElementById('new-module-title').value = '';
  document.getElementById('new-module-desc').value = '';
  document.getElementById('addModuleModal').classList.add('open');
}

function closeAddModuleModal() {
  document.getElementById('addModuleModal').classList.remove('open');
}

function saveNewModule() {
  const title = document.getElementById('new-module-title').value.trim();
  if (!title) return alert('Enter module title');

  const id = 'm' + (modules.length + 1);
  modules.push({ id, title, lessons: [] });
  saveCurriculum();
  closeAddModuleModal();
  renderTree();
}

function openAddLessonModal() {
  const select = document.getElementById('new-lesson-module-select');
  select.innerHTML = modules.map(m => `<option value="${m.id}">${m.title}</option>`).join('');
  document.getElementById('new-lesson-title').value = '';
  document.getElementById('addLessonModal').classList.add('open');
}

function closeAddLessonModal() {
  document.getElementById('addLessonModal').classList.remove('open');
}

function saveNewLesson() {
  const mId = document.getElementById('new-lesson-module-select').value;
  const title = document.getElementById('new-lesson-title').value.trim();
  const duration = document.getElementById('new-lesson-duration').value.trim() || '20m';
  const url = document.getElementById('new-lesson-url').value.trim();
  const isPreview = document.getElementById('new-lesson-preview').checked;
  if (!title) return alert('Enter lesson title');

  const mod = modules.find(m => m.id === mId);
  if (!mod) return;

  const id = 'l' + Date.now();
  const newL = { id, title, duration, videoUrl: url, notes: '', isPreview };
  mod.lessons.push(newL);

  saveCurriculum();
  closeAddLessonModal();
  selectLesson(id);
}

// Initial Boot
updateKPIs();
renderTree();
selectLesson(activeLessonId);
