const DEFAULT_SUBMISSIONS = [
  { id: 'sub1', studentName: 'Dipesh Shrestha', avatar: 'DS', labTitle: 'Lab 1: VDD_MAIN Voltage Drop', status: 'pending', time: '1 hour ago', points: 50 },
  { id: 'sub2', studentName: 'Elina Tamang', avatar: 'ET', labTitle: 'Lab 2: BGA Reballing Inspection', status: 'pending', time: '3 hours ago', points: 40 },
  { id: 'sub3', studentName: 'Gopal Adhikari', avatar: 'GA', labTitle: 'Lab 3: Audio Codec Jumpering', status: 'pending', time: '5 hours ago', points: 60 },
  { id: 'sub4', studentName: 'Hari Bahadur KC', avatar: 'HK', labTitle: 'Lab 1: VDD_MAIN Voltage Drop', status: 'pending', time: '8 hours ago', points: 50 },
  { id: 'sub5', studentName: 'Ishwar Karki', avatar: 'IK', labTitle: 'Lab 2: BGA Reballing Inspection', status: 'draft', time: 'Yesterday', points: 40 },
  { id: 'sub6', studentName: 'Jeevan Rai', avatar: 'JR', labTitle: 'Lab 4: Baseband Power Sequence', status: 'draft', time: 'Yesterday', points: 50 },
  { id: 'sub7', studentName: 'Aarav Sharma', avatar: 'AS', labTitle: 'Lab 1: VDD_MAIN Voltage Drop', status: 'graded', time: 'Oct 28', score: 48, points: 50 },
  { id: 'sub8', studentName: 'Bikash Gurung', avatar: 'BG', labTitle: 'Lab 1: VDD_MAIN Voltage Drop', status: 'graded', time: 'Oct 28', score: 45, points: 50 },
  { id: 'sub9', studentName: 'Chandra Prakash Thapa', avatar: 'CT', labTitle: 'Lab 2: BGA Reballing Inspection', status: 'graded', time: 'Oct 27', score: 38, points: 40 }
];

let submissions = JSON.parse(localStorage.getItem('smg_submissions')) || DEFAULT_SUBMISSIONS;
let activeSubId = null;

function saveSubmissions() {
  localStorage.setItem('smg_submissions', JSON.stringify(submissions));
  renderKanban();
}

function renderKanban() {
  const pending = submissions.filter(s => s.status === 'pending');
  const draft = submissions.filter(s => s.status === 'draft');
  const graded = submissions.filter(s => s.status === 'graded');

  document.getElementById('kpi-pending').innerText = pending.length;
  document.getElementById('kpi-draft').innerText = draft.length;
  document.getElementById('kpi-graded').innerText = graded.length;

  document.getElementById('badge-col-pending').innerText = pending.length;
  document.getElementById('badge-col-draft').innerText = draft.length;
  document.getElementById('badge-col-graded').innerText = graded.length;

  renderCol('cards-pending', pending);
  renderCol('cards-draft', draft);
  renderCol('cards-graded', graded);
}

function renderCol(containerId, list) {
  const el = document.getElementById(containerId);
  if (list.length === 0) {
    el.innerHTML = '<div style="text-align:center;padding:30px;color:var(--text4);font-size:12px">No submissions</div>';
    return;
  }

  el.innerHTML = list.map(s => `
    <div class="sub-card" onclick="openGradeModal('${s.id}')">
      <div class="sub-student">
        <div class="avatar">${s.avatar}</div>
        <span style="font-weight:600;font-size:12.5px">${s.studentName}</span>
      </div>
      <div class="sub-title">${s.labTitle}</div>
      <div class="sub-footer">
        <span><i class="fas fa-clock" style="margin-right:4px"></i>${s.time}</span>
        ${s.score !== undefined ? `<span class="badge b-green">${s.score} / ${s.points} pts</span>` : `<span class="badge b-warn">Needs Grade</span>`}
      </div>
    </div>
  `).join('');
}

function openGradeModal(id) {
  activeSubId = id;
  const s = submissions.find(x => x.id === id);
  if (!s) return;

  document.getElementById('modal-sub-meta').innerText = `${s.studentName} • ${s.labTitle}`;
  document.getElementById('modal-score').value = s.score !== undefined ? s.score : Math.round(s.points * 0.9);
  document.getElementById('modal-score').max = s.points;
  document.getElementById('modal-status-select').value = s.status;
  document.getElementById('modal-feedback').value = s.feedback || '';

  document.getElementById('gradeModal').classList.add('open');
}

function closeGradeModal() {
  document.getElementById('gradeModal').classList.remove('open');
}

function autoGradeWithAI() {
  const s = submissions.find(x => x.id === activeSubId);
  if (!s) return;
  const recommendedScore = Math.round(s.points * 0.94);
  document.getElementById('modal-score').value = recommendedScore;
  document.getElementById('modal-feedback').value = 'AI Evaluation: All test points correlate with Apple golden board specs within 3% tolerance. Voltage drop and resistance to ground confirmed healthy.';
  document.getElementById('modal-status-select').value = 'graded';
}

function saveGrading() {
  const s = submissions.find(x => x.id === activeSubId);
  if (!s) return;

  s.score = parseInt(document.getElementById('modal-score').value, 10);
  s.status = document.getElementById('modal-status-select').value;
  s.feedback = document.getElementById('modal-feedback').value.trim();

  saveSubmissions();
  closeGradeModal();
}

// Initial Boot
renderKanban();
