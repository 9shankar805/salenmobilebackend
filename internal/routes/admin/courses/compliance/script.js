const DEFAULT_COMPLIANCE = [
  { id: '1', name: 'Aarav Sharma', status: 'Compliant', expiry: 'In 324 days', last: 'Oct 15, 2026', cycle: 'Annual' },
  { id: '2', name: 'Bikash Gurung', status: 'Compliant', expiry: 'In 290 days', last: 'Sep 20, 2026', cycle: 'Annual' },
  { id: '3', name: 'Chandra Prakash Thapa', status: 'Expiring Soon', expiry: 'In 18 days', last: 'Nov 18, 2025', cycle: 'Annual' },
  { id: '4', name: 'Dipesh Shrestha', status: 'Expiring Soon', expiry: 'In 24 days', last: 'Nov 24, 2025', cycle: 'Annual' },
  { id: '5', name: 'Elina Tamang', status: 'Waived', expiry: 'Exempt until 2027', last: 'External Lab', cycle: 'External' }
];

let complianceList = JSON.parse(localStorage.getItem('smg_compliance_records')) || DEFAULT_COMPLIANCE;

function renderCompliance() {
  document.getElementById('compliance-body').innerHTML = complianceList.map(item => {
    let bClass = 'b-compliant';
    let icon = 'fa-circle-check';
    if (item.status === 'Expiring Soon') { bClass = 'b-warning'; icon = 'fa-triangle-exclamation'; }
    if (item.status === 'Overdue') { bClass = 'b-overdue'; icon = 'fa-circle-xmark'; }
    if (item.status === 'Waived') { bClass = 'b-waived'; icon = 'fa-shield-halved'; }

    return `
      <tr>
        <td style="font-weight:600">${item.name}</td>
        <td><span class="badge ${bClass}"><i class="fas ${icon}"></i> ${item.status}</span></td>
        <td style="font-family:var(--mono);color:var(--text3)">${item.expiry}</td>
        <td style="color:var(--text4)">${item.last}</td>
        <td>${item.cycle}</td>
        <td style="text-align:right">
          <button class="btn btn-outline btn-sm" onclick="extendDue('${item.id}')"><i class="fas fa-calendar-plus"></i> Extend</button>
        </td>
      </tr>
    `;
  }).join('');
}

function openWaiveModal() {
  document.getElementById('waive-date').value = '2027-10-31';
  document.getElementById('waiveModal').classList.add('open');
}

function closeWaiveModal() {
  document.getElementById('waiveModal').classList.remove('open');
}

function applyWaiver() {
  const name = document.getElementById('waive-student-select').value;
  complianceList.push({
    id: String(Date.now()),
    name,
    status: 'Waived',
    expiry: 'Exempt until 2027',
    last: 'Granted Waiver',
    cycle: 'Exempt'
  });
  localStorage.setItem('smg_compliance_records', JSON.stringify(complianceList));
  closeWaiveModal();
  renderCompliance();
}

function extendDue(id) {
  const item = complianceList.find(x => x.id === id);
  if (item) {
    item.status = 'Compliant';
    item.expiry = 'Extended +90 Days';
    localStorage.setItem('smg_compliance_records', JSON.stringify(complianceList));
    renderCompliance();
    alert(`Extended compliance deadline for ${item.name} by 90 days!`);
  }
}

// Initial Boot
renderCompliance();
