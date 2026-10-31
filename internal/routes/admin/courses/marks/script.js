const DEFAULT_ASSESSMENTS = [
  { id: 'a1', title: 'Lab 1: VDD_MAIN Voltage Drop', maxPoints: 50 },
  { id: 'a2', title: 'Quiz 1: BGA & SMD Components', maxPoints: 20 },
  { id: 'a3', title: 'Lab 2: iPhone PMIC Reballing', maxPoints: 50 },
  { id: 'a4', title: 'Lab 3: Audio IC Jumpering', maxPoints: 50 },
  { id: 'a5', title: 'Midterm Practical Board Exam', maxPoints: 100 },
  { id: 'a6', title: 'Final Motherboard Troubleshooting', maxPoints: 100 }
];

const DEFAULT_STUDENTS = [
  { id: 's1', name: 'Aarav Sharma', email: 'aarav.sharma@gmail.com', avatar: 'AS' },
  { id: 's2', name: 'Bikash Gurung', email: 'bikash.g@gmail.com', avatar: 'BG' },
  { id: 's3', name: 'Chandra Prakash Thapa', email: 'chandra.thapa@gmail.com', avatar: 'CT' },
  { id: 's4', name: 'Dipesh Shrestha', email: 'dipesh.shrestha@hotmail.com', avatar: 'DS' },
  { id: 's5', name: 'Elina Tamang', email: 'elina.tamang@gmail.com', avatar: 'ET' },
  { id: 's6', name: 'Gopal Adhikari', email: 'gopal.adhikari@yahoo.com', avatar: 'GA' },
  { id: 's7', name: 'Hari Bahadur KC', email: 'hari.kc@salen.com', avatar: 'HK' },
  { id: 's8', name: 'Ishwar Karki', email: 'ishwar.karki@gmail.com', avatar: 'IK' },
  { id: 's9', name: 'Jeevan Rai', email: 'jeevan.rai@nepalmail.com', avatar: 'JR' },
  { id: 's10', name: 'Kiran Maharjan', email: 'kiran.m@gmail.com', avatar: 'KM' },
  { id: 's11', name: 'Laxman Chaudhary', email: 'laxman.c@gmail.com', avatar: 'LC' },
  { id: 's12', name: 'Manoj Pokhrel', email: 'manoj.pokhrel@gmail.com', avatar: 'MP' },
  { id: 's13', name: 'Nabina Shrestha', email: 'nabina.s@gmail.com', avatar: 'NS' },
  { id: 's14', name: 'Pooja Khatri', email: 'pooja.khatri@gmail.com', avatar: 'PK' },
  { id: 's15', name: 'Rabin Silwal', email: 'rabin.silwal@gmail.com', avatar: 'RS' },
  { id: 's16', name: 'Sanjay Subedi', email: 'sanjay.subedi@gmail.com', avatar: 'SS' },
  { id: 's17', name: 'Tara Nath Gautam', email: 'taranath.g@gmail.com', avatar: 'TG' },
  { id: 's18', name: 'Umesh Pandey', email: 'umesh.pandey@gmail.com', avatar: 'UP' }
];

let assessments = JSON.parse(localStorage.getItem('smg_assessments')) || DEFAULT_ASSESSMENTS;
let students = JSON.parse(localStorage.getItem('smg_students')) || DEFAULT_STUDENTS;
let studentMarks = JSON.parse(localStorage.getItem('smg_student_marks')) || {};

// Mock initial marks
if (Object.keys(studentMarks).length === 0) {
  students.forEach((s, sIdx) => {
    studentMarks[s.id] = {};
    assessments.forEach(a => {
      // Generate scores between 65% and 98%
      const basePct = 0.70 + ((sIdx * 17 + a.maxPoints) % 28) / 100;
      studentMarks[s.id][a.id] = Math.round(a.maxPoints * Math.min(0.98, Math.max(0.60, basePct)));
    });
  });
  saveMarks();
}

function saveMarks() {
  localStorage.setItem('smg_student_marks', JSON.stringify(studentMarks));
  localStorage.setItem('smg_assessments', JSON.stringify(assessments));
  updateKPIs();
}

function calculateAverage(studentId) {
  let totalScored = 0;
  let totalPossible = 0;
  assessments.forEach(a => {
    const score = studentMarks[studentId]?.[a.id];
    if (score !== undefined) {
      totalScored += score;
      totalPossible += a.maxPoints;
    }
  });
  return totalPossible > 0 ? (totalScored / totalPossible) * 100 : 0;
}

function getLetterGrade(pct) {
  if (pct >= 90) return { letter: 'A+', class: 'grade-90' };
  if (pct >= 80) return { letter: 'A', class: 'grade-80' };
  if (pct >= 70) return { letter: 'B', class: 'grade-70' };
  return { letter: 'F', class: 'grade-low' };
}

function updateKPIs() {
  let sumAvg = 0;
  let passCount = 0;
  let distCount = 0;
  let topStudent = null;
  let topScore = -1;

  students.forEach(s => {
    const avg = calculateAverage(s.id);
    sumAvg += avg;
    if (avg >= 70) passCount++;
    if (avg >= 90) distCount++;
    if (avg > topScore) {
      topScore = avg;
      topStudent = s;
    }
  });

  const overallAvg = students.length > 0 ? (sumAvg / students.length).toFixed(1) : 0;
  document.getElementById('kpi-avg').innerText = overallAvg + '%';
  document.getElementById('kpi-pass-rate').innerText = ((passCount / students.length) * 100).toFixed(1) + '%';
  document.getElementById('kpi-distinction').innerText = distCount;
  if (topStudent) {
    document.getElementById('kpi-top').innerText = topStudent.name;
  }
}

function renderTable() {
  const search = document.getElementById('search-student').value.toLowerCase().trim();
  const filter = document.getElementById('grade-filter').value;

  const filteredStudents = students.filter(s => {
    const matchesSearch = s.name.toLowerCase().includes(search) || s.email.toLowerCase().includes(search);
    if (!matchesSearch) return false;
    const avg = calculateAverage(s.id);
    if (filter === 'distinction') return avg >= 90;
    if (filter === 'first') return avg >= 80 && avg < 90;
    if (filter === 'pass') return avg >= 70 && avg < 80;
    if (filter === 'risk') return avg < 70;
    return true;
  });

  // Render Header
  const headRow = document.getElementById('marks-head-row');
  headRow.innerHTML = `
    <th class="col-student">Student Name (${filteredStudents.length})</th>
    <th style="text-align:center;width:110px">Overall Average</th>
    ${assessments.map(a => `
      <th style="text-align:center;min-width:130px" title="${a.title}">
        <div style="white-space:normal;max-width:160px;margin:0 auto;line-height:1.2">${a.title}</div>
        <div style="font-size:10px;font-weight:400;color:var(--text4);margin-top:2px">Max: ${a.maxPoints} pts</div>
      </th>
    `).join('')}
  `;

  // Render Body
  const body = document.getElementById('marks-body');
  if (filteredStudents.length === 0) {
    body.innerHTML = `<tr><td colspan="${assessments.length + 2}" style="text-align:center;padding:40px;color:var(--text4)">No matching student records found</td></tr>`;
    document.getElementById('table-info').innerText = '0 students';
    return;
  }

  body.innerHTML = filteredStudents.map(s => {
    const avg = calculateAverage(s.id);
    const letter = getLetterGrade(avg);

    return `
      <tr>
        <td class="col-student">
          <div class="student-cell">
            <div class="avatar">${s.avatar}</div>
            <div class="student-meta">
              <div class="student-name">${s.name}</div>
              <div class="student-email">${s.email}</div>
            </div>
          </div>
        </td>
        <td class="avg-cell">
          <span style="font-weight:800;font-size:14px">${avg.toFixed(1)}%</span>
          <span class="letter-pill ${letter.class}">${letter.letter}</span>
        </td>
        ${assessments.map(a => {
          const score = studentMarks[s.id]?.[a.id] ?? 0;
          const pct = Math.round((score / a.maxPoints) * 100);
          const gradeCls = pct >= 90 ? 'grade-90' : pct >= 80 ? 'grade-80' : pct >= 70 ? 'grade-70' : 'grade-low';

          return `
            <td class="score-cell">
              <div class="score-badge ${gradeCls}" onclick="openEditModal('${s.id}', '${a.id}')" title="Click to edit: ${s.name} - ${a.title}">
                <span class="score-pts">${score}</span>
                <span class="score-pct">${pct}%</span>
              </div>
            </td>
          `;
        }).join('')}
      </tr>
    `;
  }).join('');

  document.getElementById('table-info').innerText = `Showing ${filteredStudents.length} of ${students.length} students`;
}

function filterGrades() {
  renderTable();
}

// Edit Modal
let activeEditStudentId = null;
let activeEditAssessmentId = null;

function openEditModal(studentId, assessmentId) {
  activeEditStudentId = studentId;
  activeEditAssessmentId = assessmentId;
  const s = students.find(x => x.id === studentId);
  const a = assessments.find(x => x.id === assessmentId);
  if (!s || !a) return;

  const currentScore = studentMarks[studentId]?.[assessmentId] ?? 0;
  document.getElementById('modal-student-desc').innerHTML = `Updating grade for <strong>${s.name}</strong> on <em>${a.title}</em>`;
  document.getElementById('modal-score-input').value = currentScore;
  document.getElementById('modal-score-input').max = a.maxPoints;
  document.getElementById('modal-max-points').innerText = `/ ${a.maxPoints} pts`;
  document.getElementById('editScoreModal').classList.add('open');
}

function closeEditModal() {
  document.getElementById('editScoreModal').classList.remove('open');
}

function saveScore() {
  const val = parseInt(document.getElementById('modal-score-input').value, 10);
  const a = assessments.find(x => x.id === activeEditAssessmentId);
  if (isNaN(val) || val < 0 || (a && val > a.maxPoints)) {
    return alert(`Score must be between 0 and ${a.maxPoints}`);
  }

  if (!studentMarks[activeEditStudentId]) studentMarks[activeEditStudentId] = {};
  studentMarks[activeEditStudentId][activeEditAssessmentId] = val;
  saveMarks();
  closeEditModal();
  renderTable();
}

// Add Assessment Column
function openAddAssessmentModal() {
  document.getElementById('new-assessment-title').value = '';
  document.getElementById('new-assessment-max').value = 50;
  document.getElementById('addAssessmentModal').classList.add('open');
}

function closeAddAssessmentModal() {
  document.getElementById('addAssessmentModal').classList.remove('open');
}

function saveNewAssessment() {
  const title = document.getElementById('new-assessment-title').value.trim();
  const max = parseInt(document.getElementById('new-assessment-max').value, 10);
  if (!title) return alert('Please enter assessment title');

  const id = 'a' + (assessments.length + 1);
  assessments.push({ id, title, maxPoints: max });

  // Default initial scores
  students.forEach(s => {
    if (!studentMarks[s.id]) studentMarks[s.id] = {};
    studentMarks[s.id][id] = Math.round(max * 0.85);
  });

  saveMarks();
  closeAddAssessmentModal();
  renderTable();
}

function exportCSV() {
  let csv = 'Student ID,Name,Email,Overall Average %,' + assessments.map(a => `"${a.title} (${a.maxPoints}pts)"`).join(',') + '\n';
  students.forEach(s => {
    const avg = calculateAverage(s.id).toFixed(1);
    const row = [
      s.id,
      `"${s.name}"`,
      s.email,
      avg + '%',
      ...assessments.map(a => studentMarks[s.id]?.[a.id] ?? 0)
    ];
    csv += row.join(',') + '\n';
  });

  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `gradebook_smg_${new Date().toISOString().split('T')[0]}.csv`;
  a.click();
  URL.revokeObjectURL(url);
}

// Initial Boot
updateKPIs();
renderTable();
