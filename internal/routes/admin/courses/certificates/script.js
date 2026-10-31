const DEFAULT_CERTS = [
  { id: 'c1', name: 'Aarav Sharma', serial: 'SMG-CERT-8841', date: 'Oct 28, 2026', status: 'Valid' },
  { id: 'c2', name: 'Bikash Gurung', serial: 'SMG-CERT-8842', date: 'Oct 28, 2026', status: 'Valid' },
  { id: 'c3', name: 'Chandra Prakash Thapa', serial: 'SMG-CERT-8843', date: 'Oct 27, 2026', status: 'Valid' },
  { id: 'c4', name: 'Dipesh Shrestha', serial: 'SMG-CERT-8844', date: 'Oct 26, 2026', status: 'Valid' },
  { id: 'c5', name: 'Elina Tamang', serial: 'SMG-CERT-8845', date: 'Oct 26, 2026', status: 'Valid' }
];

let certs = JSON.parse(localStorage.getItem('smg_issued_certs')) || DEFAULT_CERTS;

function renderLedger() {
  document.getElementById('cert-ledger-body').innerHTML = certs.map(c => `
    <tr>
      <td style="font-weight:600">${c.name}</td>
      <td style="font-family:var(--mono);color:var(--gold)">${c.serial}</td>
      <td style="color:var(--text3)">${c.date}</td>
      <td><span style="font-weight:700;color:var(--success)"><i class="fas fa-circle-check"></i> ${c.status}</span></td>
      <td style="text-align:right">
        <button class="btn btn-outline btn-sm" onclick="alert('Downloading PDF for ${c.serial}...')"><i class="fas fa-download"></i> PDF</button>
      </td>
    </tr>
  `).join('');
}

function switchTab(tab, btn) {
  document.querySelectorAll('.nav-tab').forEach(b => b.classList.remove('active'));
  btn.classList.add('active');

  ['designer', 'ledger', 'rules'].forEach(t => {
    document.getElementById(`tab-${t}`).style.display = (t === tab ? 'block' : 'none');
  });
}

function updateCertPreview() {
  document.getElementById('preview-title').innerText = document.getElementById('cert-title-input').value;
  document.getElementById('preview-sign').innerText = document.getElementById('cert-sign-input').value;
}

function updateTheme() {
  const val = document.getElementById('cert-theme-select').value;
  const frame = document.getElementById('live-cert-frame');
  if (val === 'cyan') {
    frame.style.borderColor = 'var(--cyan)';
  } else if (val === 'white') {
    frame.style.borderColor = '#ffffff';
  } else {
    frame.style.borderColor = 'var(--gold)';
  }
}

function issueCertPrompt() {
  const name = prompt('Enter Student Name to issue certificate:');
  if (!name) return;
  const serial = 'SMG-CERT-' + Math.floor(1000 + Math.random() * 9000);
  certs.unshift({
    id: 'c' + Date.now(),
    name,
    serial,
    date: new Date().toLocaleDateString('en-US', { month: 'short', day: '2-digit', year: 'numeric' }),
    status: 'Valid'
  });
  localStorage.setItem('smg_issued_certs', JSON.stringify(certs));
  renderLedger();
  alert(`Certificate ${serial} issued to ${name}!`);
}

// Initial Boot
renderLedger();
