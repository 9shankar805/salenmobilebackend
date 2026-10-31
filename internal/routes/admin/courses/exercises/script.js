const DEFAULT_LABS = [
  {
    id: 'e1',
    title: 'Lab 1: VDD_MAIN Voltage Drop Diagnostic',
    description: 'Diagnose short circuits on the main system rail using diode mode voltage drops.',
    questionsCount: 5,
    maxPoints: 50,
    passingPct: 75,
    difficulty: 'Intermediate'
  },
  {
    id: 'e2',
    title: 'Lab 2: BGA Reballing & Solder Joint Inspection',
    description: 'Evaluate stencil alignment, thermal ramp rates, and ball height uniformity.',
    questionsCount: 4,
    maxPoints: 40,
    passingPct: 80,
    difficulty: 'Advanced'
  },
  {
    id: 'e3',
    title: 'Lab 3: Audio Codec Jumper Verification',
    description: 'Verify continuity on 0.02mm enameled jumper trace between Master Clock and C12 pad.',
    questionsCount: 6,
    maxPoints: 60,
    passingPct: 75,
    difficulty: 'Intermediate'
  },
  {
    id: 'e4',
    title: 'Lab 4: Baseband PMU Power Sequences',
    description: 'Probe SMPS inductors L5900_BB through L5904_BB during cellular modem wake cycle.',
    questionsCount: 5,
    maxPoints: 50,
    passingPct: 80,
    difficulty: 'Master'
  }
];

let labs = JSON.parse(localStorage.getItem('smg_labs')) || DEFAULT_LABS;

function renderLabs() {
  const container = document.getElementById('labs-list-container');
  container.innerHTML = labs.map(lab => `
    <div class="lab-card">
      <div class="lab-header">
        <div>
          <span class="badge b-blue" style="margin-bottom:6px">${lab.difficulty}</span>
          <div class="lab-title">${lab.title}</div>
        </div>
      </div>
      <div class="lab-body">
        <p style="font-size:12.5px;color:var(--text3);margin-bottom:14px">${lab.description}</p>
        <div class="lab-meta">
          <span><i class="fas fa-list-ol" style="color:var(--primary)"></i>${lab.questionsCount} Questions</span>
          <span><i class="fas fa-star" style="color:var(--warn)"></i>${lab.maxPoints} Points</span>
          <span><i class="fas fa-bullseye" style="color:var(--success)"></i>${lab.passingPct}% Pass Mark</span>
        </div>
      </div>
      <div class="lab-footer">
        <button class="btn btn-outline btn-sm" onclick="switchView('simulator')"><i class="fas fa-play"></i> Test Lab</button>
        <button class="btn btn-primary btn-sm" onclick="editLab('${lab.id}')"><i class="fas fa-pen"></i> Edit Lab</button>
      </div>
    </div>
  `).join('');

  document.getElementById('kpi-total-labs').innerText = labs.length;
  const totalQ = labs.reduce((sum, l) => sum + l.questionsCount, 0);
  document.getElementById('kpi-total-q').innerText = totalQ;
}

function switchView(view, btn) {
  document.querySelectorAll('.nav-tab').forEach(b => b.classList.remove('active'));
  if (btn) btn.classList.add('active');

  ['catalog', 'builder', 'simulator'].forEach(v => {
    document.getElementById(`view-${v}`).style.display = (v === view ? 'block' : 'none');
  });
}

function showBuilder() {
  switchView('builder');
  document.getElementById('builder-title').value = 'Lab ' + (labs.length + 1) + ': Hardware Diagnostics';
}

function addQuestion() {
  const qContainer = document.getElementById('questions-builder-container');
  const count = qContainer.children.length + 1;
  const div = document.createElement('div');
  div.className = 'question-item';
  div.innerHTML = `
    <div class="question-header">
      <div style="font-weight:700">Question ${count}</div>
      <button class="btn btn-outline btn-sm" onclick="this.closest('.question-item').remove()"><i class="fas fa-trash"></i></button>
    </div>
    <div class="form-group">
      <label>Question Prompt / Test Description</label>
      <input type="text" class="form-input" placeholder="e.g. Measure resistance to ground on PP_VDD_MAIN..."/>
    </div>
    <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px">
      <div class="form-group">
        <label>Correct Answer / Expected Value</label>
        <input type="text" class="form-input" placeholder="e.g. 0.385V"/>
      </div>
      <div class="form-group">
        <label>Points Assigned</label>
        <input type="number" class="form-input" value="10"/>
      </div>
    </div>
  `;
  qContainer.appendChild(div);
}

function saveLab() {
  const title = document.getElementById('builder-title').value.trim();
  const points = parseInt(document.getElementById('builder-points').value, 10);
  if (!title) return alert('Enter lab title');

  labs.push({
    id: 'e' + Date.now(),
    title,
    description: 'Practical hardware diagnostics and test point evaluation.',
    questionsCount: Math.max(1, document.getElementById('questions-builder-container').children.length),
    maxPoints: points || 50,
    passingPct: 75,
    difficulty: 'Intermediate'
  });

  localStorage.setItem('smg_labs', JSON.stringify(labs));
  renderLabs();
  switchView('catalog');
  alert('Diagnostic Lab published successfully!');
}

function editLab(id) {
  const lab = labs.find(x => x.id === id);
  if (!lab) return;
  document.getElementById('builder-title').value = lab.title;
  document.getElementById('builder-points').value = lab.maxPoints;
  switchView('builder');
}

// Initial Boot
renderLabs();
addQuestion();
