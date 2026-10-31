const DEFAULT_PEOPLE = [
  { id: 'u1', name: 'Master Salen', email: 'salen@salenmobile.com', avatar: 'MS', role: 'instructor', progress: 100, date: 'Aug 15, 2026' },
  { id: 'u2', name: 'Pravin Shrestha', email: 'pravin.tech@salen.com', avatar: 'PS', role: 'instructor', progress: 100, date: 'Sep 01, 2026' },
  { id: 's1', name: 'Aarav Sharma', email: 'aarav.sharma@gmail.com', avatar: 'AS', role: 'student', progress: 94, date: 'Oct 05, 2026' },
  { id: 's2', name: 'Bikash Gurung', email: 'bikash.g@gmail.com', avatar: 'BG', role: 'student', progress: 88, date: 'Oct 06, 2026' },
  { id: 's3', name: 'Chandra Prakash Thapa', email: 'chandra.thapa@gmail.com', avatar: 'CT', role: 'student', progress: 75, date: 'Oct 07, 2026' },
  { id: 's4', name: 'Dipesh Shrestha', email: 'dipesh.shrestha@hotmail.com', avatar: 'DS', role: 'student', progress: 68, date: 'Oct 08, 2026' },
  { id: 's5', name: 'Elina Tamang', email: 'elina.tamang@gmail.com', avatar: 'ET', role: 'student', progress: 82, date: 'Oct 08, 2026' },
  { id: 's6', name: 'Gopal Adhikari', email: 'gopal.adhikari@yahoo.com', avatar: 'GA', role: 'student', progress: 62, date: 'Oct 09, 2026' },
  { id: 's7', name: 'Hari Bahadur KC', email: 'hari.kc@salen.com', avatar: 'HK', role: 'student', progress: 70, date: 'Oct 10, 2026' },
  { id: 's8', name: 'Ishwar Karki', email: 'ishwar.karki@gmail.com', avatar: 'IK', role: 'student', progress: 58, date: 'Oct 11, 2026' }
];

let people = JSON.parse(localStorage.getItem('smg_course_people')) || DEFAULT_PEOPLE;
let currentRoleFilter = 'all';

function renderPeople() {
  const search = document.getElementById('search-member').value.toLowerCase().trim();
  const filtered = people.filter(p => {
    const matchesSearch = p.name.toLowerCase().includes(search) || p.email.toLowerCase().includes(search);
    if (!matchesSearch) return false;
    if (currentRoleFilter === 'student') return p.role === 'student';
    if (currentRoleFilter === 'instructor') return p.role === 'instructor' || p.role === 'ta';
    return true;
  });

  const body = document.getElementById('people-body');
  if (filtered.length === 0) {
    body.innerHTML = '<tr><td colspan="5" style="text-align:center;padding:40px;color:var(--text4)">No matching members found</td></tr>';
    document.getElementById('table-count').innerText = '0 members';
    return;
  }

  body.innerHTML = filtered.map(p => `
    <tr>
      <td>
        <div class="user-cell">
          <div class="avatar">${p.avatar}</div>
          <div>
            <div class="user-name">${p.name}</div>
            <div class="user-email">
              <span>${p.email}</span>
              <button onclick="navigator.clipboard.writeText('${p.email}');alert('Copied email!')" style="background:transparent;border:none;color:var(--text4);cursor:pointer;font-size:11px" title="Copy Email"><i class="fas fa-copy"></i></button>
            </div>
          </div>
        </div>
      </td>
      <td>
        <select class="role-select" onchange="changeRole('${p.id}', this.value)">
          <option value="student" ${p.role === 'student' ? 'selected' : ''}>Student</option>
          <option value="instructor" ${p.role === 'instructor' ? 'selected' : ''}>Instructor</option>
          <option value="ta" ${p.role === 'ta' ? 'selected' : ''}>Teaching Assistant</option>
        </select>
      </td>
      <td>
        <div style="font-weight:700;font-size:12px;font-family:var(--mono)">${p.progress}%</div>
        <div class="progress-bar-wrap">
          <div class="progress-bar-fill" style="width:${p.progress}%"></div>
        </div>
      </td>
      <td style="color:var(--text3);font-size:12.5px">${p.date}</td>
      <td style="text-align:right">
        <button class="btn btn-outline btn-sm" onclick="removeMember('${p.id}')" title="Remove Member"><i class="fas fa-trash" style="color:var(--danger)"></i></button>
      </td>
    </tr>
  `).join('');

  document.getElementById('table-count').innerText = `Showing ${filtered.length} of ${people.length} members`;
  document.getElementById('kpi-total-people').innerText = people.length;
}

function switchRoleFilter(role, btn) {
  document.querySelectorAll('.role-tab').forEach(b => b.classList.remove('active'));
  btn.classList.add('active');
  currentRoleFilter = role;
  renderPeople();
}

function filterPeople() {
  renderPeople();
}

function changeRole(id, newRole) {
  const p = people.find(x => x.id === id);
  if (p) {
    p.role = newRole;
    localStorage.setItem('smg_course_people', JSON.stringify(people));
    renderPeople();
  }
}

function removeMember(id) {
  if (!confirm('Are you sure you want to remove this member from the course?')) return;
  people = people.filter(x => x.id !== id);
  localStorage.setItem('smg_course_people', JSON.stringify(people));
  renderPeople();
}

function openInviteModal() {
  document.getElementById('new-user-name').value = '';
  document.getElementById('new-user-email').value = '';
  document.getElementById('inviteModal').classList.add('open');
}

function closeInviteModal() {
  document.getElementById('inviteModal').classList.remove('open');
}

function saveNewMember() {
  const name = document.getElementById('new-user-name').value.trim();
  const email = document.getElementById('new-user-email').value.trim();
  const role = document.getElementById('new-user-role').value;
  if (!name || !email) return alert('Please enter name and email');

  const avatar = name.split(' ').map(w => w[0]).join('').slice(0, 2).toUpperCase() || 'ST';
  people.push({
    id: 'u' + Date.now(),
    name,
    email,
    avatar,
    role,
    progress: 0,
    date: new Date().toLocaleDateString('en-US', { month: 'short', day: '2-digit', year: 'numeric' })
  });

  localStorage.setItem('smg_course_people', JSON.stringify(people));
  closeInviteModal();
  renderPeople();
}

// Initial Boot
renderPeople();
