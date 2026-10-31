// Mock / Initial Data
const DEFAULT_LECTURES = [
  { id: 'L1', code: 'L01', title: 'Soldering Safety & ESD Basics', date: 'Oct 10' },
  { id: 'L2', code: 'L02', title: 'Multimeter & DC Power Supply', date: 'Oct 12' },
  { id: 'L3', code: 'L03', title: 'Schematics & Boardview Analysis', date: 'Oct 15' },
  { id: 'L4', code: 'L04', title: 'Short Circuit Finding: VDD_MAIN', date: 'Oct 18' },
  { id: 'L5', code: 'L05', title: 'Thermal Camera Diagnosis', date: 'Oct 21' },
  { id: 'L6', code: 'L06', title: 'BGA IC Removal & Cleaning', date: 'Oct 24' },
  { id: 'L7', code: 'L07', title: 'Reballing PMIC with Stencil', date: 'Oct 27' },
  { id: 'L8', code: 'L08', title: 'iPhone 13 No Power Practical', date: 'Oct 30' }
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

// State
let lectures = JSON.parse(localStorage.getItem('smg_lectures')) || DEFAULT_LECTURES;
let students = JSON.parse(localStorage.getItem('smg_students')) || DEFAULT_STUDENTS;
let attendanceRecords = JSON.parse(localStorage.getItem('smg_attendance_records')) || {};

// Initialize records if empty
if (Object.keys(attendanceRecords).length === 0) {
  students.forEach((s, sIdx) => {
    attendanceRecords[s.id] = {};
    lectures.forEach((l, lIdx) => {
      // 90% attendance simulation
      attendanceRecords[s.id][l.id] = !((sIdx + lIdx) % 11 === 0);
    });
  });
  saveRecords();
}

function saveRecords() {
  localStorage.setItem('smg_attendance_records', JSON.stringify(attendanceRecords));
  localStorage.setItem('smg_lectures', JSON.stringify(lectures));
  updateKPIs();
}

function updateKPIs() {
  let totalOpportunities = 0;
  let totalPresent = 0;
  let perfectCount = 0;

  students.forEach(s => {
    let sPresent = 0;
    lectures.forEach(l => {
      totalOpportunities++;
      if (attendanceRecords[s.id] && attendanceRecords[s.id][l.id]) {
        totalPresent++;
        sPresent++;
      }
    });
    if (lectures.length > 0 && sPresent === lectures.length) {
      perfectCount++;
    }
  });

  const overallRate = totalOpportunities > 0 ? ((totalPresent / totalOpportunities) * 100).toFixed(1) : 0;
  document.getElementById('kpi-rate').innerText = overallRate + '%';
  document.getElementById('kpi-students').innerText = students.length;
  document.getElementById('kpi-lectures').innerText = lectures.length;
  document.getElementById('kpi-perfect').innerText = perfectCount;
}

function getStudentRate(studentId) {
  if (lectures.length === 0) return 0;
  let present = 0;
  lectures.forEach(l => {
    if (attendanceRecords[studentId] && attendanceRecords[studentId][l.id]) present++;
  });
  return Math.round((present / lectures.length) * 100);
}

function toggleAttendance(studentId, lectureId) {
  if (!attendanceRecords[studentId]) attendanceRecords[studentId] = {};
  attendanceRecords[studentId][lectureId] = !attendanceRecords[studentId][lectureId];
  saveRecords();
  renderMatrix();
}

function markAll(status) {
  const isPresent = status === 'present';
  const filter = document.getElementById('lecture-filter').value;
  const targetLectures = filter === 'last' ? [lectures[lectures.length - 1]] : lectures;

  students.forEach(s => {
    if (!attendanceRecords[s.id]) attendanceRecords[s.id] = {};
    targetLectures.forEach(l => {
      attendanceRecords[s.id][l.id] = isPresent;
    });
  });

  saveRecords();
  renderMatrix();
}

function filterStudents() {
  renderMatrix();
}

function renderMatrix() {
  const search = document.getElementById('search-student').value.toLowerCase().trim();
  const filter = document.getElementById('lecture-filter').value;
  const filteredStudents = students.filter(s => s.name.toLowerCase().includes(search) || s.email.toLowerCase().includes(search));

  const displayLectures = filter === 'last' && lectures.length > 0 ? [lectures[lectures.length - 1]] : lectures;

  // Render Header
  const headRow = document.getElementById('matrix-head-row');
  headRow.innerHTML = `
    <th class="col-student">Student (${filteredStudents.length})</th>
    <th style="text-align:center;width:90px">Overall %</th>
    ${displayLectures.map(l => `
      <th style="text-align:center;min-width:85px" title="${l.title} (${l.date})">
        <div>${l.code}</div>
        <div style="font-size:10px;font-weight:400;color:var(--text4)">${l.date}</div>
      </th>
    `).join('')}
  `;

  // Render Body
  const body = document.getElementById('matrix-body');
  if (filteredStudents.length === 0) {
    body.innerHTML = `<tr><td colspan="${displayLectures.length + 2}" style="text-align:center;padding:40px;color:var(--text4)">No students matching criteria</td></tr>`;
    document.getElementById('table-info').innerText = '0 students';
    return;
  }

  body.innerHTML = filteredStudents.map(s => {
    const rate = getStudentRate(s.id);
    const rateClass = rate >= 85 ? 'rate-high' : rate >= 70 ? 'rate-med' : 'rate-low';

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
        <td style="text-align:center">
          <span class="rate-badge ${rateClass}">${rate}%</span>
        </td>
        ${displayLectures.map(l => {
          const isPresent = attendanceRecords[s.id] && attendanceRecords[s.id][l.id];
          return `
            <td class="att-cell">
              <button class="att-btn ${isPresent ? 'present' : 'absent'}" onclick="toggleAttendance('${s.id}', '${l.id}')" title="${s.name} - ${l.code}: ${isPresent ? 'Present' : 'Absent'}">
                <i class="fas ${isPresent ? 'fa-check' : 'fa-times'}"></i>
              </button>
            </td>
          `;
        }).join('')}
      </tr>
    `;
  }).join('');

  document.getElementById('table-info').innerText = `Showing ${filteredStudents.length} of ${students.length} students`;
}

// Modal handling
function openAddLectureModal() {
  document.getElementById('new-lecture-title').value = `L0${lectures.length + 1}: Practical Diagnostic Session`;
  document.getElementById('new-lecture-date').value = new Date().toISOString().split('T')[0];
  document.getElementById('addLectureModal').classList.add('open');
}

function closeAddLectureModal() {
  document.getElementById('addLectureModal').classList.remove('open');
}

function saveNewLecture() {
  const title = document.getElementById('new-lecture-title').value.trim();
  const dateStr = document.getElementById('new-lecture-date').value;
  if (!title) return alert('Please enter lecture title');

  const id = 'L' + (lectures.length + 1);
  const code = 'L' + String(lectures.length + 1).padStart(2, '0');
  const d = new Date(dateStr);
  const formattedDate = d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });

  lectures.push({ id, code, title, date: formattedDate });
  
  // Set all students present by default for newly added lecture
  students.forEach(s => {
    if (!attendanceRecords[s.id]) attendanceRecords[s.id] = {};
    attendanceRecords[s.id][id] = true;
  });

  saveRecords();
  closeAddLectureModal();
  renderMatrix();
}

function exportCSV() {
  let csv = 'Student ID,Student Name,Email,Attendance Rate %,' + lectures.map(l => `${l.code} (${l.date})`).join(',') + '\n';
  students.forEach(s => {
    const rate = getStudentRate(s.id);
    const row = [
      s.id,
      `"${s.name}"`,
      s.email,
      rate + '%',
      ...lectures.map(l => (attendanceRecords[s.id] && attendanceRecords[s.id][l.id] ? 'Present' : 'Absent'))
    ];
    csv += row.join(',') + '\n';
  });

  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `attendance_smg_${new Date().toISOString().split('T')[0]}.csv`;
  a.click();
  URL.revokeObjectURL(url);
}

// Initial Boot
updateKPIs();
renderMatrix();
